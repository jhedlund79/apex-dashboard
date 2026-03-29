// Package live provides a [repository.Store] backed by live market data.
//
// Equity data is sourced from Polygon.io (https://polygon.io).
// Crypto data is sourced from CoinGecko's public API (no key required).
// The Fear & Greed index is sourced from alternative.me (no key required).
//
// Required environment variable:
//
//	POLYGON_API_KEY — Polygon.io API key
//
// Data that requires specialized APIs not covered by Polygon's free tier
// (VIX, Brent crude futures, 10-year Treasury yield) falls back to static
// seed values and is clearly labelled in the response.
package live

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// Store implements [repository.Store] using live market data.
// All methods are safe for concurrent use.
type Store struct {
	polygonKey   string
	coinGeckoKey string // optional; set via CoinGecko API key for higher rate limits
	client       *http.Client

	// grouped bars cache: the same date's data is reused across all endpoint
	// methods so we only call the grouped bars API once per trading day.
	gbMu    sync.Mutex
	gbCache map[string]map[string]float64 // date → ticker → close

	// warmupDone is closed once the background pre-fetch of historic grouped
	// bars (1-month and 3-month ago) completes. Markets() blocks on it once.
	warmupDone chan struct{}
}

const polygonBaseURL = "https://api.massive.com"
const cgBaseURL = "https://api.coingecko.com/api/v3"

// New returns a *Store ready for use.
// polygonKey is required; coinGeckoKey is optional (empty string uses the public API).
func New(polygonKey, coinGeckoKey string) (*Store, error) {
	if polygonKey == "" {
		return nil, fmt.Errorf("live: POLYGON_API_KEY is required")
	}
	s := &Store{
		polygonKey:   polygonKey,
		coinGeckoKey: coinGeckoKey,
		client:       &http.Client{Timeout: 10 * time.Second},
		gbCache:      make(map[string]map[string]float64),
		warmupDone:   make(chan struct{}),
	}
	go s.warmupHistoricBars()
	return s, nil
}

// warmupHistoricBars pre-fetches grouped daily bars for 1-month and 3-month
// ago into the gbCache. It staggers requests by 13 seconds to stay within
// Polygon's free-tier rate limit (5 req/min). The server starts immediately;
// Markets() blocks on warmupDone until this completes.
func (s *Store) warmupHistoricBars() {
	defer close(s.warmupDone)
	now := time.Now()
	// Overview() fetches today + yesterday on first call, so wait before
	// making additional requests to avoid colliding with those.
	time.Sleep(13 * time.Second)
	s.fetchGroupedBarsForDate(now.AddDate(0, -1, 0))
	time.Sleep(13 * time.Second)
	s.fetchGroupedBarsForDate(now.AddDate(0, -3, 0))
}

// ── Polygon API types ────────────────────────────────────────────────────────

type pgTicker struct {
	Ticker           string  `json:"ticker"`
	TodaysChangePerc float64 `json:"todaysChangePerc"`
	TodaysChange     float64 `json:"todaysChange"`
	Day              pgOHLC  `json:"day"`
	PrevDay          pgOHLC  `json:"prevDay"`
}

type pgOHLC struct {
	C float64 `json:"c"`
}

type pgAggsResp struct {
	Results []pgBar `json:"results"`
}

type pgBar struct {
	T int64   `json:"t"` // unix milliseconds
	C float64 `json:"c"` // close price
}

type pgGroupedResp struct {
	Results []pgGroupedBar `json:"results"`
}

type pgGroupedBar struct {
	Sym string  `json:"T"` // ticker symbol (uppercase key)
	Ts  int64   `json:"t"` // timestamp unix ms — must be declared so the decoder
	//                         doesn't case-fold "t" onto the string Sym field
	C float64 `json:"c"` // close price
}

// ── CoinGecko API types ──────────────────────────────────────────────────────

type cgCoin struct {
	ID                 string  `json:"id"`
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	CurrentPrice       float64 `json:"current_price"`
	MarketCap          float64 `json:"market_cap"`
	TotalVolume        float64 `json:"total_volume"`
	PriceChangePerc24h float64 `json:"price_change_percentage_24h"`
}

type cgMarketChart struct {
	Prices [][]float64 `json:"prices"` // [[unix_ms, price], ...]
}

// ── Alternative.me Fear & Greed types ───────────────────────────────────────

type fngResp struct {
	Data []struct {
		Value               string `json:"value"`
		ValueClassification string `json:"value_classification"`
	} `json:"data"`
}

// ── Fetch helpers ────────────────────────────────────────────────────────────

// lastTradingDays returns the two most recent trading day dates (Mon–Fri)
// as YYYY-MM-DD strings, starting from (but not including) now.
func lastTradingDays(now time.Time) (day1, day2 string) {
	count := 0
	d := now.AddDate(0, 0, -1)
	var days []string
	for count < 2 {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			days = append(days, d.Format("2006-01-02"))
			count++
		}
		d = d.AddDate(0, 0, -1)
	}
	return days[0], days[1]
}

// fetchGroupedBars fetches close prices for all US stocks on a given date
// using the grouped daily bars endpoint — a single API call regardless of
// how many tickers are needed. Results are cached in-process so multiple
// endpoint methods can share the same day's data without extra API calls.
func (s *Store) fetchGroupedBars(date string) (map[string]float64, error) {
	s.gbMu.Lock()
	if cached, ok := s.gbCache[date]; ok {
		s.gbMu.Unlock()
		return cached, nil
	}
	s.gbMu.Unlock()

	params := url.Values{}
	params.Set("adjusted", "true")
	params.Set("apiKey", s.polygonKey)
	u := fmt.Sprintf("%s/v2/aggs/grouped/locale/us/market/stocks/%s?%s", polygonBaseURL, date, params.Encode())
	resp, err := s.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("polygon grouped bars: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("polygon grouped bars: HTTP %d", resp.StatusCode)
	}
	var body pgGroupedResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("polygon grouped bars decode: %w", err)
	}
	out := make(map[string]float64, len(body.Results))
	for _, r := range body.Results {
		out[r.Sym] = r.C
	}

	s.gbMu.Lock()
	s.gbCache[date] = out
	s.gbMu.Unlock()

	return out, nil
}

// fetchSnapshots returns price and daily change data for the given tickers.
// Uses two grouped daily bars calls (one per trading day) so the total API
// call count is 2 regardless of how many tickers are requested.
func (s *Store) fetchSnapshots(tickers []string) (map[string]pgTicker, error) {
	day1, day2 := lastTradingDays(time.Now())

	today, err := s.fetchGroupedBars(day1)
	if err != nil {
		return nil, err
	}
	prev, err := s.fetchGroupedBars(day2)
	if err != nil {
		return nil, err
	}

	want := make(map[string]bool, len(tickers))
	for _, t := range tickers {
		want[t] = true
	}

	out := make(map[string]pgTicker, len(tickers))
	for ticker := range want {
		close, ok := today[ticker]
		if !ok || close == 0 {
			continue
		}
		prevClose := prev[ticker]
		var change, changePct float64
		if prevClose != 0 {
			change = close - prevClose
			changePct = (change / prevClose) * 100
		}
		out[ticker] = pgTicker{
			Ticker:           ticker,
			TodaysChange:     change,
			TodaysChangePerc: changePct,
			Day:              pgOHLC{C: close},
			PrevDay:          pgOHLC{C: prevClose},
		}
	}
	return out, nil
}

// fetchAggregates retrieves Polygon daily OHLC bars for ticker between from and to.
func (s *Store) fetchAggregates(ticker string, from, to time.Time) ([]pgBar, error) {
	path := fmt.Sprintf("/v2/aggs/ticker/%s/range/1/day/%s/%s",
		ticker,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)
	params := url.Values{}
	params.Set("adjusted", "true")
	params.Set("sort", "asc")
	params.Set("limit", "365")
	params.Set("apiKey", s.polygonKey)
	u := polygonBaseURL + path + "?" + params.Encode()
	resp, err := s.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("polygon aggs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("polygon aggs: HTTP %d", resp.StatusCode)
	}
	var body pgAggsResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("polygon aggs decode: %w", err)
	}
	return body.Results, nil
}

// cgRequest builds a CoinGecko API request, attaching the API key header when configured.
func (s *Store) cgRequest(rawURL string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if s.coinGeckoKey != "" {
		req.Header.Set("x-cg-demo-api-key", s.coinGeckoKey)
	}
	return req, nil
}

// fetchCGMarkets retrieves CoinGecko market data for the given coin IDs.
func (s *Store) fetchCGMarkets(ids []string) ([]cgCoin, error) {
	params := url.Values{}
	params.Set("vs_currency", "usd")
	params.Set("ids", strings.Join(ids, ","))
	params.Set("order", "market_cap_desc")
	params.Set("sparkline", "false")
	req, err := s.cgRequest(cgBaseURL + "/coins/markets?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("coingecko markets: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko markets: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko markets: HTTP %d", resp.StatusCode)
	}
	var coins []cgCoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("coingecko markets decode: %w", err)
	}
	return coins, nil
}

// fetchBTCHistory retrieves one year of daily BTC/USD prices from CoinGecko.
func (s *Store) fetchBTCHistory() ([][]float64, error) {
	req, err := s.cgRequest(cgBaseURL + "/coins/bitcoin/market_chart?vs_currency=usd&days=365")
	if err != nil {
		return nil, fmt.Errorf("coingecko btc history: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko btc history: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko btc history: HTTP %d", resp.StatusCode)
	}
	var chart cgMarketChart
	if err := json.NewDecoder(resp.Body).Decode(&chart); err != nil {
		return nil, fmt.Errorf("coingecko btc history decode: %w", err)
	}
	return chart.Prices, nil
}

// fetchSPYBars fetches daily SPY aggregates from the given start date to today.
func (s *Store) fetchSPYBars(from time.Time) ([]pgBar, error) {
	return s.fetchAggregates("SPY", from, time.Now())
}

// fetchFREDLatest fetches the most recent two values from a FRED data series.
// FRED series values of "." (missing) are skipped.
// Returns an error if no numeric values are found.
func (s *Store) fetchFREDLatest(seriesID string) (latest, prev float64, err error) {
	resp, err := s.client.Get("https://fred.stlouisfed.org/graph/fredgraph.csv?id=" + seriesID)
	if err != nil {
		return 0, 0, fmt.Errorf("FRED %s: %w", seriesID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("FRED %s: HTTP %d", seriesID, resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	var vals []float64
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "DATE") {
			continue
		}
		_, valStr, ok := strings.Cut(line, ",")
		if !ok {
			continue
		}
		valStr = strings.TrimSpace(valStr)
		if valStr == "." {
			continue // FRED uses "." for missing observations
		}
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		vals = append(vals, v)
	}
	if len(vals) == 0 {
		return 0, 0, fmt.Errorf("FRED %s: no data", seriesID)
	}
	latest = vals[len(vals)-1]
	if len(vals) >= 2 {
		prev = vals[len(vals)-2]
	}
	return latest, prev, nil
}

// fetchVIX retrieves the latest VIX close and daily change from the CBOE
// free historical CSV. Returns placeholder values on any failure.
func (s *Store) fetchVIX() (value string, changePct float64, dir string) {
	resp, err := s.client.Get("https://cdn.cboe.com/api/global/us_indices/daily_prices/VIX_History.csv")
	if err != nil {
		return "—", 0, "neutral"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "—", 0, "neutral"
	}
	// CSV: DATE,OPEN,HIGH,LOW,CLOSE — last two rows are latest and prior session.
	scanner := bufio.NewScanner(resp.Body)
	var closes []float64
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "DATE") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 5 {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		if err != nil {
			continue
		}
		closes = append(closes, v)
	}
	if len(closes) == 0 {
		return "—", 0, "neutral"
	}
	latest := closes[len(closes)-1]
	var pct float64
	if len(closes) >= 2 {
		p := closes[len(closes)-2]
		if p > 0 {
			pct = (latest - p) / p * 100
		}
	}
	return fmt.Sprintf("%.2f", latest), pct, dirStr(pct)
}

// fetchGroupedBarsForDate returns grouped daily bars for the nearest trading
// day on or before target, skipping weekends. Tries up to 10 days back.
// Results are served from gbCache when available.
func (s *Store) fetchGroupedBarsForDate(target time.Time) (map[string]float64, error) {
	for i := 0; i < 10; i++ {
		d := target.AddDate(0, 0, -i)
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		date := d.Format("2006-01-02")
		bars, err := s.fetchGroupedBars(date)
		if err != nil {
			return nil, err
		}
		if len(bars) > 0 {
			return bars, nil
		}
	}
	return nil, fmt.Errorf("no trading data found near %s", target.Format("2006-01-02"))
}

// scoreFromPct maps a percentage return to a 0–100 score centred at 50.
// scale controls sensitivity: 2.5 means ±20% maps to 0 or 100.
func scoreFromPct(pct, scale float64) float64 {
	return math.Round(math.Max(0, math.Min(100, 50+pct*scale)))
}

// fetchUMCSent fetches the latest University of Michigan Consumer Sentiment
// reading from the FRED public CSV endpoint. Returns "—" on any failure so
// callers can treat it as non-fatal.
func (s *Store) fetchUMCSent() string {
	resp, err := s.client.Get("https://fred.stlouisfed.org/graph/fredgraph.csv?id=UMCSENT")
	if err != nil {
		return "—"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "—"
	}
	// CSV format: DATE,UMCSENT\n2024-01-01,78.8\n...
	// We want the last data row.
	scanner := bufio.NewScanner(resp.Body)
	var lastVal string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "DATE") {
			continue
		}
		if _, val, ok := strings.Cut(line, ","); ok {
			lastVal = strings.TrimSpace(val)
		}
	}
	if lastVal == "" {
		return "—"
	}
	return lastVal
}

// fetchFearGreed retrieves the current Fear & Greed index from alternative.me.
// Returns placeholder strings on any failure so callers can treat it as non-fatal.
func (s *Store) fetchFearGreed() (value, classification string) {
	resp, err := s.client.Get("https://api.alternative.me/fng/?limit=1")
	if err != nil {
		return "—", "Unknown"
	}
	defer resp.Body.Close()
	var body fngResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || len(body.Data) == 0 {
		return "—", "Unknown"
	}
	return body.Data[0].Value, body.Data[0].ValueClassification
}

// ── Formatting helpers ───────────────────────────────────────────────────────

func dirStr(pct float64) string {
	if pct > 0 {
		return "up"
	}
	if pct < 0 {
		return "down"
	}
	return "neutral"
}

func fmtPct(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}

// fmtDollar formats a float as a dollar amount with comma thousands separators.
func fmtDollar(v float64) string {
	var whole int64
	var decimals string
	if v >= 1000 {
		whole = int64(math.Round(v))
		decimals = ""
	} else if v >= 1 {
		whole = int64(v)
		frac := math.Round((v-float64(whole))*100) / 100
		decimals = fmt.Sprintf(".%02d", int(frac*100))
	} else {
		return fmt.Sprintf("$%.4f", v)
	}
	s := strconv.FormatInt(whole, 10)
	var buf []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, c)
	}
	return "$" + string(buf) + decimals
}

// marketStatus returns a human-readable US market open/close string.
func marketStatus() string {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	if wd := now.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return "Markets Closed · " + now.Format("Mon Jan 2")
	}
	open := time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0, 0, loc)
	close := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, loc)
	if now.Before(open) || now.After(close) {
		return "Markets Closed · " + now.Format("Mon Jan 2")
	}
	return "Markets Open · " + now.Format("15:04") + " ET"
}

// sampleBars reduces a bar slice to n evenly-spaced points for chart display.
func sampleBars(bars []pgBar, n int) models.ChartData {
	if len(bars) == 0 {
		return models.ChartData{}
	}
	step := len(bars) / n
	if step < 1 {
		step = 1
	}
	var labels []string
	var data []float64
	for i := 0; i < len(bars) && len(labels) < n; i += step {
		t := time.UnixMilli(bars[i].T)
		labels = append(labels, t.Format("Jan 2"))
		data = append(data, math.Round(bars[i].C*100)/100)
	}
	return models.ChartData{Labels: labels, Data: data}
}

// samplePriceHistory reduces a CoinGecko [[ms, price]] slice to n points.
func samplePriceHistory(prices [][]float64, n int) models.ChartData {
	if len(prices) == 0 {
		return models.ChartData{}
	}
	step := len(prices) / n
	if step < 1 {
		step = 1
	}
	var labels []string
	var data []float64
	for i := 0; i < len(prices) && len(labels) < n; i += step {
		t := time.UnixMilli(int64(prices[i][0]))
		labels = append(labels, t.Format("Jan 06"))
		data = append(data, math.Round(prices[i][1]))
	}
	return models.ChartData{Labels: labels, Data: data}
}

// ── Store methods ────────────────────────────────────────────────────────────

// Overview fetches live index prices (Polygon), S&P 500 chart (Polygon SPY),
// crypto 24h changes (CoinGecko), and the Fear & Greed index (alternative.me).
// Gold is approximated from the GLD ETF (1 share ≈ 0.0931 troy oz).
// VIX, Brent crude, and 10-year yield require specialized futures/index APIs
// and fall back to static seed values labelled accordingly.
func (s *Store) Overview() (models.OverviewResponse, error) {
	// Batch snapshot: index ETFs + GLD for gold approximation
	snaps, err := s.fetchSnapshots([]string{"SPY", "DIA", "QQQ", "IWM", "GLD"})
	if err != nil {
		return models.OverviewResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// SPY 6-month daily aggregates for S&P 500 chart
	now := time.Now()
	bars, err := s.fetchAggregates("SPY", now.AddDate(0, -6, 0), now)
	if err != nil {
		return models.OverviewResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// CoinGecko: top 6 coins for the crypto bar chart
	cgIDs := []string{"bitcoin", "ethereum", "solana", "ripple", "binancecoin", "dogecoin"}
	coins, err := s.fetchCGMarkets(cgIDs)
	if err != nil {
		return models.OverviewResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	fngVal, fngClass := s.fetchFearGreed() // non-fatal

	// VIX from CBOE free CSV — non-fatal
	vixVal, vixPct, vixDir := s.fetchVIX()

	// Brent crude spot (DCOILBRENTEU) and 10Y Treasury yield (DGS10) from FRED
	brentLatest, brentPrev, brentErr := s.fetchFREDLatest("DCOILBRENTEU")
	yieldLatest, yieldPrev, yieldErr := s.fetchFREDLatest("DGS10")

	// Intl ETF proxies for chart — reuses grouped bars already cached by fetchSnapshots
	intlChartDefs := []struct{ etf, label string }{
		{"EWG", "DAX"}, {"EWU", "FTSE"}, {"EWQ", "CAC 40"}, {"EWI", "MIB"},
		{"VGK", "STOXX"}, {"EWJ", "NIKKEI"}, {"ASHR", "CSI 300"}, {"MCHI", "SHCOMP"},
	}
	var intlETFTickers []string
	for _, d := range intlChartDefs {
		intlETFTickers = append(intlETFTickers, d.etf)
	}
	intlSnaps, _ := s.fetchSnapshots(intlETFTickers) // cache hit — no extra API calls

	// Build index cards from ETF snapshots
	indexDefs := []struct {
		etf   string
		label string
		note  string
	}{
		{"SPY", "S&P 500", "via SPDR S&P 500 ETF (SPY)"},
		{"DIA", "Dow Jones", "via SPDR Dow Jones ETF (DIA)"},
		{"QQQ", "Nasdaq 100", "via Invesco QQQ ETF"},
		{"IWM", "Russell 2000", "via iShares Russell 2000 ETF"},
	}
	indices := make([]models.Index, 0, len(indexDefs))
	for _, def := range indexDefs {
		t, ok := snaps[def.etf]
		if !ok {
			continue
		}
		price := t.Day.C
		if price == 0 {
			price = t.PrevDay.C
		}
		indices = append(indices, models.Index{
			Ticker: def.label,
			Value:  fmtDollar(price),
			Change: fmt.Sprintf("%+.2f", t.TodaysChange),
			Pct:    fmtPct(t.TodaysChangePerc),
			Dir:    dirStr(t.TodaysChangePerc),
			Note:   def.note,
		})
	}

	// Gold QuickStat: GLD share ≈ 0.0931 troy oz → multiply for spot approx
	const gldOzFactor = 1.0 / 0.0931
	goldStat := models.QuickStat{
		Label: "GOLD (approx)", Value: "—", Change: "—", ChangeDir: "neutral",
	}
	if gld, ok := snaps["GLD"]; ok {
		price := gld.Day.C
		if price == 0 {
			price = gld.PrevDay.C
		}
		goldStat = models.QuickStat{
			Label:     "GOLD (approx)",
			Value:     fmtDollar(price * gldOzFactor),
			Change:    fmtPct(gld.TodaysChangePerc),
			ChangeDir: dirStr(gld.TodaysChangePerc),
		}
	}

	// Crypto bar chart: 24h % change per coin
	cgColorMap := map[string]string{
		"bitcoin": "#f7931a", "ethereum": "#627eea", "solana": "#2775ca",
		"ripple": "#c3a634", "binancecoin": "#f3ba2f", "dogecoin": "#c2a633",
	}
	cgSymMap := map[string]string{
		"bitcoin": "BTC", "ethereum": "ETH", "solana": "SOL",
		"ripple": "XRP", "binancecoin": "BNB", "dogecoin": "DOGE",
	}
	var cgLabels []string
	var cgData []float64
	var cgColors []string
	for _, c := range coins {
		sym := cgSymMap[c.ID]
		if sym == "" {
			sym = strings.ToUpper(c.Symbol)
		}
		cgLabels = append(cgLabels, sym)
		cgData = append(cgData, math.Round(c.PriceChangePerc24h*100)/100)
		col := cgColorMap[c.ID]
		if col == "" {
			col = "#8888a0"
		}
		cgColors = append(cgColors, col)
	}

	return models.OverviewResponse{
		Header: models.HeaderInfo{
			FearGreed:    fmt.Sprintf("FEAR & GREED: %s — %s", fngVal, strings.ToUpper(fngClass)),
			MarketStatus: marketStatus(),
		},
		Indices: indices,
		QuickStats: func() []models.QuickStat {
			stats := []models.QuickStat{goldStat}

			stats = append(stats, models.QuickStat{
				Label:     "VIX INDEX",
				Value:     vixVal,
				Change:    fmtPct(vixPct),
				ChangeDir: vixDir,
			})

			brentStat := models.QuickStat{Label: "BRENT CRUDE", Value: "—", Change: "—", ChangeDir: "neutral"}
			if brentErr == nil {
				bpct := 0.0
				if brentPrev > 0 {
					bpct = (brentLatest - brentPrev) / brentPrev * 100
				}
				brentStat = models.QuickStat{
					Label:     "BRENT CRUDE",
					Value:     fmt.Sprintf("$%.2f", brentLatest),
					Change:    fmtPct(bpct),
					ChangeDir: dirStr(bpct),
				}
			}
			stats = append(stats, brentStat)

			yieldStat := models.QuickStat{Label: "10Y YIELD", Value: "—", Change: "—", ChangeDir: "neutral"}
			if yieldErr == nil {
				ypct := 0.0
				if yieldPrev > 0 {
					ypct = (yieldLatest - yieldPrev) / yieldPrev * 100
				}
				yieldStat = models.QuickStat{
					Label:     "10Y YIELD",
					Value:     fmt.Sprintf("%.2f%%", yieldLatest),
					Change:    fmtPct(ypct),
					ChangeDir: dirStr(ypct),
				}
			}
			stats = append(stats, yieldStat)

			return stats
		}(),
		MarketDrivers:  staticDrivers(),
		SP500ChartData: sampleBars(bars, 17),
		CryptoChartData: models.CryptoChartData{
			Labels: cgLabels, Data: cgData, Colors: cgColors,
		},
		IntlChartData: func() models.ChartData {
			var labels []string
			var data []float64
			for _, d := range intlChartDefs {
				labels = append(labels, d.label)
				pct := 0.0
				if snap, ok := intlSnaps[d.etf]; ok {
					pct = math.Round(snap.TodaysChangePerc*100) / 100
				}
				data = append(data, pct)
			}
			return models.ChartData{Labels: labels, Data: data}
		}(),
	}, nil
}

// Markets fetches live data for US movers, international indices (ETF proxies),
// sector performance (ETF proxies), and US summary stats.
func (s *Store) Markets() (models.MarketsResponse, error) {
	now := time.Now()

	// ETF definitions
	moverDefs := []struct{ ticker, name, driver string }{
		{"META", "Meta Platforms", "AI infrastructure & advertising revenue"},
		{"AMZN", "Amazon", "AWS cloud & e-commerce"},
		{"NVDA", "Nvidia", "AI accelerator demand"},
		{"MSFT", "Microsoft", "Cloud + Copilot AI integration"},
		{"MU", "Micron", "Memory chip supply/demand cycle"},
		{"LMT", "Lockheed Martin", "Defense spending uplift"},
		{"NOC", "Northrop Grumman", "Defense sector rally"},
		{"AVAV", "AeroVironment", "Drone & autonomous systems demand"},
	}
	intlDefs := []struct{ etf, index, region, driver string }{
		{"EWG", "DAX", "Germany", "Fiscal stimulus + energy & industrial exposure"},
		{"EWU", "FTSE 100", "UK", "Energy, financials & mining; BoE policy"},
		{"EWQ", "CAC 40", "France", "Luxury, energy & industrials"},
		{"EWI", "FTSE MIB", "Italy", "Banks & energy; ECB rate sensitive"},
		{"VGK", "STOXX 600", "Pan-Europe", "Broad European; miners & tech leading"},
		{"EWJ", "NIKKEI 225", "Japan", "Exports & manufacturing; yen-sensitive"},
		{"ASHR", "CSI 300", "China", "A-shares; domestic consumption & tech"},
		{"MCHI", "Shanghai/China", "China", "Broad China; trade tensions + policy"},
		{"EWH", "HANG SENG", "Hong Kong", "Financial hub; China policy sensitive"},
		{"EWY", "KOSPI", "S. Korea", "Chips, autos & shipbuilding"},
	}
	sectorDefs := []struct{ etf, label string }{
		{"ITA", "Defense"}, {"XLE", "Energy"}, {"SOXX", "AI/Semis"},
		{"XLV", "Healthcare"}, {"XLU", "Utilities"}, {"XLF", "Financials"},
		{"XLI", "Industrials"}, {"XLY", "Consumer Disc."},
		{"IGV", "SaaS"}, {"IWM", "US Small Cap"}, {"JETS", "Airlines"}, {"KWEB", "China Tech"},
	}

	// Collect all tickers for a single batched snapshot call
	var allTickers []string
	for _, d := range moverDefs {
		allTickers = append(allTickers, d.ticker)
	}
	for _, d := range intlDefs {
		allTickers = append(allTickers, d.etf)
	}
	for _, d := range sectorDefs {
		allTickers = append(allTickers, d.etf)
	}

	snaps, err := s.fetchSnapshots(allTickers)
	if err != nil {
		return models.MarketsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// SPY aggregates for US summary stats
	spyBars, err := s.fetchSPYBars(now.AddDate(-1, -1, 0))
	if err != nil {
		return models.MarketsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}
	usSummary := computeUSSummary(spyBars, now)
	usSummary.ConsumerSentiment = s.fetchUMCSent()
	usSummary.SentimentNote = "Univ. of Michigan survey (UMCSENT via FRED)"

	// Wait for the background warmup to finish pre-fetching 1-month and 3-month
	// grouped bars. After first call this is instant (channel is already closed).
	<-s.warmupDone

	// Retrieve from gbCache — these are populated by warmupHistoricBars.
	oneMonthBars, err := s.fetchGroupedBarsForDate(now.AddDate(0, -1, 0))
	if err != nil {
		return models.MarketsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}
	threeMonthBars, err := s.fetchGroupedBarsForDate(now.AddDate(0, -3, 0))
	if err != nil {
		return models.MarketsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}
	todayBars, err := s.fetchGroupedBarsForDate(now)
	if err != nil {
		return models.MarketsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// Build US movers
	usMovers := make([]models.USMover, 0, len(moverDefs))
	for _, def := range moverDefs {
		t, ok := snaps[def.ticker]
		if !ok {
			continue
		}
		usMovers = append(usMovers, models.USMover{
			Ticker: def.ticker,
			Name:   def.name,
			Move:   fmtPct(t.TodaysChangePerc),
			Dir:    dirStr(t.TodaysChangePerc),
			Driver: def.driver,
		})
	}

	// Build international indices from ETF daily % changes
	intlIndices := make([]models.IntlIndex, 0, len(intlDefs))
	for _, def := range intlDefs {
		t, ok := snaps[def.etf]
		change, dir := "—", "neutral"
		if ok {
			change = fmtPct(t.TodaysChangePerc)
			dir = dirStr(t.TodaysChangePerc)
		}
		intlIndices = append(intlIndices, models.IntlIndex{
			Ticker: def.index,
			Region: def.region,
			Change: change,
			Dir:    dir,
			Driver: def.driver,
		})
	}

	// Build sector data: momentum from 1-month ETF return, outlook from 3-month.
	// Scores mapped to 0–100 centred at 50 (0% change).
	var sectorLabels []string
	var momentum, outlook []float64
	for _, def := range sectorDefs {
		sectorLabels = append(sectorLabels, def.label)
		todayClose := todayBars[def.etf]
		oneMonthClose := oneMonthBars[def.etf]
		threeMonthClose := threeMonthBars[def.etf]

		var momPct, outPct float64
		if oneMonthClose > 0 && todayClose > 0 {
			momPct = (todayClose - oneMonthClose) / oneMonthClose * 100
		}
		if threeMonthClose > 0 && todayClose > 0 {
			outPct = (todayClose - threeMonthClose) / threeMonthClose * 100
		}
		momentum = append(momentum, scoreFromPct(momPct, 2.5))
		outlook = append(outlook, scoreFromPct(outPct, 2.0))
	}

	// Rank sectors by combined score (momentum weight 0.4, outlook weight 0.6).
	type sectorScore struct {
		label    string
		combined float64
	}
	scores := make([]sectorScore, len(sectorLabels))
	for i, label := range sectorLabels {
		scores[i] = sectorScore{label: label, combined: momentum[i]*0.4 + outlook[i]*0.6}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].combined > scores[j].combined })

	overweight := make([]string, 0, 5)
	for _, s := range scores[:5] {
		overweight = append(overweight, s.label)
	}
	underweight := make([]string, 0, 5)
	for _, s := range scores[len(scores)-5:] {
		underweight = append(underweight, s.label)
	}

	return models.MarketsResponse{
		USSummary:   usSummary,
		USMovers:    usMovers,
		IntlIndices: intlIndices,
		SectorData: models.SectorData{
			Labels:      sectorLabels,
			Momentum:    momentum,
			Outlook:     outlook,
			Overweight:  overweight,
			Underweight: underweight,
		},
	}, nil
}

// computeUSSummary derives S&P 500 March MTD and YoY returns from SPY bars.
// bars must be sorted ascending by date (as Polygon returns them).
func computeUSSummary(bars []pgBar, now time.Time) models.USSummary {
	if len(bars) == 0 {
		return models.USSummary{MarchReturn: "—", YoYReturn: "—"}
	}

	latest := bars[len(bars)-1].C

	// March MTD: find the last bar of February (month before current if current is March,
	// otherwise last February in the dataset).
	marchStart := time.Date(now.Year(), time.March, 1, 0, 0, 0, 0, time.UTC)
	var febClose float64
	for i := len(bars) - 1; i >= 0; i-- {
		t := time.UnixMilli(bars[i].T)
		if t.Before(marchStart) {
			febClose = bars[i].C
			break
		}
	}

	marchReturn := "—"
	marchNote := ""
	if febClose > 0 {
		pct := (latest - febClose) / febClose * 100
		marchReturn = fmtPct(pct)
		if pct < 0 {
			marchNote = fmt.Sprintf("SPY MTD as of %s", now.Format("Jan 2"))
		} else {
			marchNote = fmt.Sprintf("SPY MTD as of %s", now.Format("Jan 2"))
		}
	}

	// YoY: find the bar closest to one year ago.
	oneYearAgo := now.AddDate(-1, 0, 0)
	var yoyClose float64
	var bestDiff time.Duration = 1<<62 - 1
	for _, b := range bars {
		t := time.UnixMilli(b.T)
		diff := t.Sub(oneYearAgo)
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			bestDiff = diff
			yoyClose = b.C
		}
	}

	yoYReturn := "—"
	yoYNote := ""
	if yoyClose > 0 {
		pct := (latest - yoyClose) / yoyClose * 100
		yoYReturn = fmtPct(pct)
		yoYNote = fmt.Sprintf("SPY vs ~%s", oneYearAgo.Format("Jan 2 2006"))
	}

	return models.USSummary{
		MarchReturn: marchReturn,
		MarchNote:   marchNote,
		YoYReturn:   yoYReturn,
		YoYNote:     yoYNote,
	}
}

// Crypto fetches live prices, 24h changes, and market cap data for the top
// coins from CoinGecko, plus one year of BTC price history.
func (s *Store) Crypto() (models.CryptoResponse, error) {
	mainIDs := []string{"bitcoin", "ethereum", "solana", "ripple"}
	smallIDs := []string{"binancecoin", "dogecoin"}
	allCoins, err := s.fetchCGMarkets(append(mainIDs, smallIDs...))
	if err != nil {
		return models.CryptoResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	prices, err := s.fetchBTCHistory()
	if err != nil {
		return models.CryptoResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	coinMap := make(map[string]cgCoin, len(allCoins))
	for _, c := range allCoins {
		coinMap[c.ID] = c
	}

	assetMeta := map[string]struct {
		symbol string
		bg     string
		color  string
	}{
		"bitcoin":  {"₿", "linear-gradient(135deg,#f7931a,#ff6b00)", "#fff"},
		"ethereum": {"Ξ", "linear-gradient(135deg,#627eea,#3b5998)", "#fff"},
		"solana":   {"S", "linear-gradient(135deg,#2775ca,#1a5bb0)", "#fff"},
		"ripple":   {"X", "linear-gradient(135deg,#c3a634,#ffd700)", "#000"},
	}

	cryptoAssets := make([]models.CryptoAsset, 0, len(mainIDs))
	for _, id := range mainIDs {
		c, ok := coinMap[id]
		if !ok {
			continue
		}
		meta := assetMeta[id]
		cryptoAssets = append(cryptoAssets, models.CryptoAsset{
			Symbol: meta.symbol,
			Name:   c.Name,
			Ticker: strings.ToUpper(c.Symbol),
			Price:  fmtDollar(c.CurrentPrice),
			Change: fmtPct(c.PriceChangePerc24h),
			Dir:    dirStr(c.PriceChangePerc24h),
			Bg:     meta.bg,
			Color:  meta.color,
			Meta:   fmt.Sprintf("MCap: $%.0fB · Vol: $%.1fB", c.MarketCap/1e9, c.TotalVolume/1e9),
		})
	}

	cryptoSmall := make([]models.CryptoSmall, 0, len(smallIDs)+1)
	for _, id := range smallIDs {
		c, ok := coinMap[id]
		if !ok {
			continue
		}
		cryptoSmall = append(cryptoSmall, models.CryptoSmall{
			Ticker: c.Name,
			Price:  fmtDollar(c.CurrentPrice),
			Change: fmtPct(c.PriceChangePerc24h),
			Dir:    dirStr(c.PriceChangePerc24h),
		})
	}
	// Total market cap requires a separate CoinGecko /global call; omit for now.
	cryptoSmall = append(cryptoSmall, models.CryptoSmall{
		Ticker: "Total Market",
		Price:  "—",
		Change: "See CoinGecko /global",
		Dir:    "neutral",
	})

	return models.CryptoResponse{
		CryptoAssets:   cryptoAssets,
		CryptoSmall:    cryptoSmall,
		BtcHistoryData: samplePriceHistory(prices, 12),
		MarketContext:  staticMarketContext(),
	}, nil
}

// Recommendations dynamically selects the top 3 picks per category by
// 1-month price momentum from the screener universe, then hydrates each
// card with live price, 24h change, and 1-month return from cached data.
// No additional API calls are made beyond what Markets() already caches.
func (s *Store) Recommendations() (models.RecommendationsResponse, error) {
	// Wait for warmup so grouped bars are available.
	<-s.warmupDone

	now := time.Now()
	todayBars, err := s.fetchGroupedBarsForDate(now)
	if err != nil {
		return models.RecommendationsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}
	oneMonthBars, err := s.fetchGroupedBarsForDate(now.AddDate(0, -1, 0))
	if err != nil {
		return models.RecommendationsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// Today's grouped bars give us 24h change via yesterday's close.
	yesterdayBars, err := s.fetchGroupedBarsForDate(now.AddDate(0, 0, -1))
	if err != nil {
		return models.RecommendationsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	// Fetch crypto universe from CoinGecko.
	cgIDs := make([]string, 0, len(cryptoUniverse))
	for _, c := range cryptoUniverse {
		cgIDs = append(cgIDs, c.cgID)
	}
	coins, err := s.fetchCGMarkets(cgIDs)
	if err != nil {
		return models.RecommendationsResponse{}, fmt.Errorf("%w: %w", repository.ErrUnavailable, err)
	}

	toRec := func(r ranked, i int) models.Recommendation {
		badge, recType := badgeForRank(i, r.badge, r.recType)
		price := r.price
		changePct := 0.0
		if prev := yesterdayBars[r.ticker]; prev > 0 && price > 0 {
			changePct = (price - prev) / prev * 100
		}
		return models.Recommendation{
			Ticker:  r.ticker,
			Name:    r.name,
			Badge:   badge,
			Type:    recType,
			Thesis:  r.thesis,
			Stats: []models.RecStat{
				{Label: "Price", Value: fmtDollar(price), Dir: ""},
				{Label: "24h Chg", Value: fmtPct(changePct), Dir: dirStr(changePct)},
				{Label: "1M Mom", Value: fmtMom(r.momPct), Dir: dirStr(r.momPct)},
			},
		}
	}

	toCryptoRec := func(r ranked, i int) models.Recommendation {
		badge, recType := badgeForRank(i, r.badge, r.recType)
		return models.Recommendation{
			Ticker:  r.ticker,
			Name:    r.name,
			Badge:   badge,
			Type:    recType,
			Thesis:  r.thesis,
			Stats: []models.RecStat{
				{Label: "Price", Value: fmtDollar(r.price), Dir: ""},
				{Label: "24h Chg", Value: fmtPct(r.changePct), Dir: dirStr(r.changePct)},
				{Label: "1M Mom", Value: fmtMom(r.momPct), Dir: dirStr(r.momPct)},
			},
		}
	}

	aiRanked := s.rankByMomentum(aiSemisUniverse, todayBars, oneMonthBars, 3)
	disRanked := s.rankByMomentum(disruptorsUniverse, todayBars, oneMonthBars, 3)
	defRanked := s.rankByMomentum(defenseUniverse, todayBars, oneMonthBars, 3)
	cryRanked := rankCrypto(coins, cryptoUniverse, 3)

	toRecs := func(rs []ranked, fn func(ranked, int) models.Recommendation) []models.Recommendation {
		out := make([]models.Recommendation, 0, len(rs))
		for i, r := range rs {
			out = append(out, fn(r, i))
		}
		return out
	}

	return models.RecommendationsResponse{
		AISemis:    toRecs(aiRanked, toRec),
		Disruptors: toRecs(disRanked, toRec),
		Defense:    toRecs(defRanked, toRec),
		Crypto:     toRecs(cryRanked, toCryptoRec),
	}, nil
}

// ── Static fallback data ─────────────────────────────────────────────────────

func staticDrivers() []models.MarketDriver {
	return []models.MarketDriver{
		{
			Icon:  "🛢",
			Title: "Energy Markets",
			Color: "danger",
			Text:  "Oil prices and energy supply disruptions remain key macro drivers. Monitor Brent crude for stagflation signals.",
		},
		{
			Icon:  "📉",
			Title: "Tech Valuations",
			Color: "warn",
			Text:  "AI-driven valuation compression in high-multiple tech stocks. Nasdaq under pressure from rate sensitivity.",
		},
		{
			Icon:  "🌍",
			Title: "Global Rotation",
			Color: "blue",
			Text:  "Capital rotating from US mega-cap tech into cheaper European and Asian markets. Germany fiscal stimulus, Japan reforms.",
		},
	}
}


func staticMarketContext() string {
	return "Crypto markets remain sensitive to macro risk-off sentiment. Institutional adoption continues via ETF flows. Monitor on-chain activity and exchange reserves as leading indicators of trend reversals."
}

