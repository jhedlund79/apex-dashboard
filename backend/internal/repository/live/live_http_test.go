package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ── mock HTTP transport ───────────────────────────────────────────────────────

// mockTransport dispatches responses based on URL substring matches.
type mockTransport struct {
	routes []mockRoute
}

type mockRoute struct {
	contains   string
	statusCode int
	body       string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u := req.URL.String()
	for _, r := range m.routes {
		if strings.Contains(u, r.contains) {
			return &http.Response{
				StatusCode: r.statusCode,
				Body:       io.NopCloser(strings.NewReader(r.body)),
				Header:     make(http.Header),
			}, nil
		}
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Header:     make(http.Header),
	}, nil
}

// newTestStore builds a Store with a mock HTTP client and a pre-closed
// warmupDone channel so tests don't block on the background goroutine.
func newTestStore(routes []mockRoute) *Store {
	done := make(chan struct{})
	close(done)
	return &Store{
		polygonKey: "test-key",
		client:     &http.Client{Transport: &mockTransport{routes: routes}},
		gbCache:    make(map[string]map[string]float64),
		warmupDone: done,
	}
}

// groupedBarsJSON builds a Polygon grouped bars JSON response for the given
// ticker→close map.
func groupedBarsJSON(bars map[string]float64) string {
	results := make([]pgGroupedBar, 0, len(bars))
	for sym, c := range bars {
		results = append(results, pgGroupedBar{Sym: sym, C: c})
	}
	b, _ := json.Marshal(pgGroupedResp{Results: results})
	return string(b)
}

// aggsJSON builds a Polygon aggregates response JSON.
func aggsJSON(bars []pgBar) string {
	b, _ := json.Marshal(pgAggsResp{Results: bars})
	return string(b)
}

// baseGroupedBarsRoutes returns the minimal set of routes needed for
// fetchSnapshots (two consecutive trading-day grouped bars calls).
func baseGroupedBarsRoutes() []mockRoute {
	barData := groupedBarsJSON(map[string]float64{
		"SPY": 500, "DIA": 400, "QQQ": 450, "IWM": 200, "GLD": 185,
		"META": 580, "AMZN": 185, "NVDA": 850, "MSFT": 415, "MU": 95,
		"LMT": 460, "NOC": 440, "AVAV": 190,
		"EWG": 35, "EWU": 30, "EWQ": 28, "EWI": 35, "VGK": 65,
		"EWJ": 68, "ASHR": 26, "MCHI": 22, "EWH": 18, "EWY": 62,
		"ITA": 140, "XLE": 88, "SOXX": 215, "XLV": 145, "XLU": 67,
		"XLF": 43, "XLI": 120, "XLY": 195, "IGV": 380, "JETS": 18,
		"KWEB": 27,
		"NVDA_": 700, // 1-month-ago value (same route, different call)
	})
	return []mockRoute{
		{contains: "/v2/aggs/grouped/locale/us/market/stocks/", statusCode: 200, body: barData},
	}
}

// ── fetchGroupedBars ──────────────────────────────────────────────────────────

func TestFetchGroupedBars_Success(t *testing.T) {
	body := groupedBarsJSON(map[string]float64{"AAPL": 170, "MSFT": 415})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: body},
	})
	bars, err := s.fetchGroupedBars("2024-03-12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bars["AAPL"] != 170 {
		t.Errorf("AAPL = %v, want 170", bars["AAPL"])
	}
}

func TestFetchGroupedBars_CacheHit(t *testing.T) {
	s := newTestStore(nil) // no routes — cache must serve the request
	s.gbCache["2024-03-12"] = map[string]float64{"AAPL": 170}
	bars, err := s.fetchGroupedBars("2024-03-12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bars["AAPL"] != 170 {
		t.Errorf("AAPL = %v, want 170", bars["AAPL"])
	}
}

func TestFetchGroupedBars_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 500, body: `{}`},
	})
	_, err := s.fetchGroupedBars("2024-03-12")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestFetchGroupedBars_DecodeError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: `not-json`},
	})
	_, err := s.fetchGroupedBars("2024-03-12")
	if err == nil {
		t.Error("expected error for bad JSON")
	}
}

// ── fetchAggregates ───────────────────────────────────────────────────────────

func TestFetchAggregates_Success(t *testing.T) {
	now := time.Now()
	bars := []pgBar{
		{T: now.AddDate(0, -1, 0).UnixMilli(), C: 490},
		{T: now.UnixMilli(), C: 500},
	}
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/ticker/", statusCode: 200, body: aggsJSON(bars)},
	})
	got, err := s.fetchAggregates("SPY", now.AddDate(0, -1, 0), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestFetchAggregates_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/ticker/", statusCode: 429, body: `{}`},
	})
	_, err := s.fetchAggregates("SPY", time.Now().AddDate(0, -1, 0), time.Now())
	if err == nil {
		t.Error("expected error for HTTP 429")
	}
}

func TestFetchAggregates_DecodeError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/ticker/", statusCode: 200, body: `bad`},
	})
	_, err := s.fetchAggregates("SPY", time.Now().AddDate(0, -1, 0), time.Now())
	if err == nil {
		t.Error("expected decode error")
	}
}

// ── fetchCGMarkets ────────────────────────────────────────────────────────────

func TestFetchCGMarkets_Success(t *testing.T) {
	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", CurrentPrice: 82000, PriceChangePerc24h: -2.1, MarketCap: 1.6e12, TotalVolume: 30e9},
	}
	body, _ := json.Marshal(coins)
	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 200, body: string(body)},
	})
	got, err := s.fetchCGMarkets([]string{"bitcoin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "bitcoin" {
		t.Errorf("got %+v", got)
	}
}

func TestFetchCGMarkets_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 503, body: `{}`},
	})
	_, err := s.fetchCGMarkets([]string{"bitcoin"})
	if err == nil {
		t.Error("expected error for HTTP 503")
	}
}

func TestFetchCGMarkets_DecodeError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 200, body: `not-an-array`},
	})
	_, err := s.fetchCGMarkets([]string{"bitcoin"})
	if err == nil {
		t.Error("expected decode error")
	}
}

func TestFetchCGMarkets_WithAPIKey(t *testing.T) {
	coins := []cgCoin{{ID: "bitcoin", Name: "Bitcoin"}}
	body, _ := json.Marshal(coins)
	s := newTestStore([]mockRoute{{contains: "/coins/markets", statusCode: 200, body: string(body)}})
	s.coinGeckoKey = "cg-key"
	got, err := s.fetchCGMarkets([]string{"bitcoin"})
	if err != nil || len(got) == 0 {
		t.Fatalf("unexpected result: %v %v", got, err)
	}
}

// ── fetchBTCHistory ───────────────────────────────────────────────────────────

func TestFetchBTCHistory_Success(t *testing.T) {
	chart := cgMarketChart{
		Prices: [][]float64{
			{float64(time.Now().AddDate(-1, 0, 0).UnixMilli()), 45000},
			{float64(time.Now().UnixMilli()), 82000},
		},
	}
	body, _ := json.Marshal(chart)
	s := newTestStore([]mockRoute{
		{contains: "/coins/bitcoin/market_chart", statusCode: 200, body: string(body)},
	})
	got, err := s.fetchBTCHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestFetchBTCHistory_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/coins/bitcoin/market_chart", statusCode: 429, body: `{}`},
	})
	_, err := s.fetchBTCHistory()
	if err == nil {
		t.Error("expected error for HTTP 429")
	}
}

func TestFetchBTCHistory_DecodeError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/coins/bitcoin/market_chart", statusCode: 200, body: `bad`},
	})
	_, err := s.fetchBTCHistory()
	if err == nil {
		t.Error("expected decode error")
	}
}

// ── fetchFREDLatest ───────────────────────────────────────────────────────────

const fredCSV = "DATE,DCOILBRENTEU\n2024-01-01,75.5\n2024-02-01,78.2\n2024-03-01,80.1\n"
const fredCSVMissing = "DATE,UMCSENT\n2024-01-01,.\n2024-02-01,.\n"

func TestFetchFREDLatest_Success(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "fred.stlouisfed.org", statusCode: 200, body: fredCSV},
	})
	latest, prev, err := s.fetchFREDLatest("DCOILBRENTEU")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest != 80.1 {
		t.Errorf("latest = %v, want 80.1", latest)
	}
	if prev != 78.2 {
		t.Errorf("prev = %v, want 78.2", prev)
	}
}

func TestFetchFREDLatest_SingleValue(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "fred.stlouisfed.org", statusCode: 200, body: "DATE,V\n2024-03-01,5.0\n"},
	})
	latest, prev, err := s.fetchFREDLatest("DGS10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest != 5.0 || prev != 0 {
		t.Errorf("latest=%v prev=%v", latest, prev)
	}
}

func TestFetchFREDLatest_MissingValues(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "fred.stlouisfed.org", statusCode: 200, body: fredCSVMissing},
	})
	_, _, err := s.fetchFREDLatest("UMCSENT")
	if err == nil {
		t.Error("expected error for all-missing FRED data")
	}
}

func TestFetchFREDLatest_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "fred.stlouisfed.org", statusCode: 503, body: ``},
	})
	_, _, err := s.fetchFREDLatest("DGS10")
	if err == nil {
		t.Error("expected error for HTTP 503")
	}
}

// ── fetchVIX ─────────────────────────────────────────────────────────────────

const vixCSV = "DATE,OPEN,HIGH,LOW,CLOSE\n2024-03-11,20.5,21.0,19.8,20.2\n2024-03-12,19.5,20.1,18.9,19.4\n"

func TestFetchVIX_Success(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "cdn.cboe.com", statusCode: 200, body: vixCSV},
	})
	val, pct, dir := s.fetchVIX()
	if val == "—" {
		t.Error("expected numeric VIX value")
	}
	_ = pct
	if dir != "down" && dir != "up" && dir != "neutral" {
		t.Errorf("unexpected dir: %q", dir)
	}
}

func TestFetchVIX_SingleRow(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "cdn.cboe.com", statusCode: 200, body: "DATE,OPEN,HIGH,LOW,CLOSE\n2024-03-12,19.5,20.1,18.9,19.4\n"},
	})
	val, pct, _ := s.fetchVIX()
	if val == "—" {
		t.Error("expected value for single row")
	}
	if pct != 0 {
		t.Errorf("expected pct=0 for single row, got %v", pct)
	}
}

func TestFetchVIX_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "cdn.cboe.com", statusCode: 500, body: ``},
	})
	val, _, _ := s.fetchVIX()
	if val != "—" {
		t.Errorf("expected '—' for HTTP error, got %q", val)
	}
}

func TestFetchVIX_Empty(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "cdn.cboe.com", statusCode: 200, body: "DATE,OPEN,HIGH,LOW,CLOSE\n"},
	})
	val, _, _ := s.fetchVIX()
	if val != "—" {
		t.Errorf("expected '—' for empty CSV, got %q", val)
	}
}

// ── fetchUMCSent ──────────────────────────────────────────────────────────────

func TestFetchUMCSent_Success(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "fred.stlouisfed.org/graph/fredgraph.csv?id=UMCSENT", statusCode: 200, body: "DATE,UMCSENT\n2024-01-01,78.5\n2024-02-01,80.2\n"},
	})
	val := s.fetchUMCSent()
	if val == "—" {
		t.Error("expected numeric value")
	}
}

func TestFetchUMCSent_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "UMCSENT", statusCode: 500, body: ``},
	})
	val := s.fetchUMCSent()
	if val != "—" {
		t.Errorf("expected '—' for HTTP error, got %q", val)
	}
}

func TestFetchUMCSent_Empty(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "UMCSENT", statusCode: 200, body: "DATE,UMCSENT\n"},
	})
	val := s.fetchUMCSent()
	if val != "—" {
		t.Errorf("expected '—' for empty CSV, got %q", val)
	}
}

// ── fetchFearGreed ────────────────────────────────────────────────────────────

func TestFetchFearGreed_Success(t *testing.T) {
	body := `{"data":[{"value":"12","value_classification":"Extreme Fear"}]}`
	s := newTestStore([]mockRoute{
		{contains: "alternative.me", statusCode: 200, body: body},
	})
	val, cls := s.fetchFearGreed()
	if val != "12" || cls != "Extreme Fear" {
		t.Errorf("val=%q cls=%q", val, cls)
	}
}

func TestFetchFearGreed_HTTPError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "alternative.me", statusCode: 500, body: ``},
	})
	val, cls := s.fetchFearGreed()
	if val != "—" || cls != "Unknown" {
		t.Errorf("val=%q cls=%q", val, cls)
	}
}

func TestFetchFearGreed_EmptyData(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "alternative.me", statusCode: 200, body: `{"data":[]}`},
	})
	val, cls := s.fetchFearGreed()
	if val != "—" || cls != "Unknown" {
		t.Errorf("val=%q cls=%q", val, cls)
	}
}

func TestFetchFearGreed_DecodeError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "alternative.me", statusCode: 200, body: `bad`},
	})
	val, cls := s.fetchFearGreed()
	if val != "—" || cls != "Unknown" {
		t.Errorf("val=%q cls=%q", val, cls)
	}
}

// ── fetchSnapshots ────────────────────────────────────────────────────────────

func TestFetchSnapshots_Success(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500, "QQQ": 450})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
	})
	snaps, err := s.fetchSnapshots([]string{"SPY", "QQQ"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps["SPY"].Day.C != 500 {
		t.Errorf("SPY close = %v, want 500", snaps["SPY"].Day.C)
	}
}

func TestFetchSnapshots_MissingTicker(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
	})
	snaps, err := s.fetchSnapshots([]string{"SPY", "MISSING"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := snaps["MISSING"]; ok {
		t.Error("did not expect MISSING ticker in result")
	}
}

// ── fetchGroupedBarsForDate ───────────────────────────────────────────────────

func TestFetchGroupedBarsForDate_Success(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 490})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
	})
	// Use a known weekday
	target := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC) // Tuesday
	bars, err := s.fetchGroupedBarsForDate(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bars["SPY"] != 490 {
		t.Errorf("SPY = %v, want 490", bars["SPY"])
	}
}

func TestFetchGroupedBarsForDate_EmptyBarsSkipsToNextDay(t *testing.T) {
	callCount := 0
	tr := &funcTransport{fn: func(req *http.Request) (*http.Response, error) {
		callCount++
		var body string
		if callCount == 1 {
			body = groupedBarsJSON(map[string]float64{}) // empty → skip
		} else {
			body = groupedBarsJSON(map[string]float64{"SPY": 480})
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}}
	done := make(chan struct{})
	close(done)
	s := &Store{
		polygonKey: "test",
		client:     &http.Client{Transport: tr},
		gbCache:    make(map[string]map[string]float64),
		warmupDone: done,
	}
	target := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	bars, err := s.fetchGroupedBarsForDate(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bars["SPY"] != 480 {
		t.Errorf("SPY = %v, want 480", bars["SPY"])
	}
}

func TestFetchGroupedBarsForDate_NoDataError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: groupedBarsJSON(map[string]float64{})},
	})
	target := time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC)
	_, err := s.fetchGroupedBarsForDate(target)
	if err == nil {
		t.Error("expected error when all dates return empty bars")
	}
}

// ── FetchQuote ────────────────────────────────────────────────────────────────

func TestFetchQuote_Success(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"NVDA": 850, "NVDA_prev": 820})
	// Need two calls (today + yesterday) returning data for NVDA
	barDataWithNVDA := groupedBarsJSON(map[string]float64{"NVDA": 850})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barDataWithNVDA},
	})
	_ = barData
	q, err := s.FetchQuote("NVDA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Ticker != "NVDA" {
		t.Errorf("ticker = %q", q.Ticker)
	}
	if q.Price != 850 {
		t.Errorf("price = %v, want 850", q.Price)
	}
}

func TestFetchQuote_NotFound(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
	})
	_, err := s.FetchQuote("MISSING")
	if err == nil {
		t.Error("expected error for missing ticker")
	}
}

func TestFetchQuote_Directions(t *testing.T) {
	cases := []struct {
		today, yesterday float64
		wantDir          string
	}{
		{520, 500, "up"},
		{480, 500, "down"},
		{500, 500, "neutral"},
	}
	for _, tc := range cases {
		callN := 0
		tr := &funcTransport{fn: func(req *http.Request) (*http.Response, error) {
			callN++
			price := tc.today
			if callN > 1 {
				price = tc.yesterday
			}
			body := groupedBarsJSON(map[string]float64{"TEST": price})
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}}
		done := make(chan struct{})
		close(done)
		s := &Store{
			polygonKey: "test",
			client:     &http.Client{Transport: tr},
			gbCache:    make(map[string]map[string]float64),
			warmupDone: done,
		}
		q, err := s.FetchQuote("TEST")
		if err != nil {
			t.Fatalf("FetchQuote error: %v", err)
		}
		if q.Dir != tc.wantDir {
			t.Errorf("today=%v yesterday=%v dir=%q want %q", tc.today, tc.yesterday, q.Dir, tc.wantDir)
		}
	}
}

// ── funcTransport helper ──────────────────────────────────────────────────────

type funcTransport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f *funcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.fn(req)
}

// ── Overview ──────────────────────────────────────────────────────────────────

func TestOverview_Success(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{
		"SPY": 500, "DIA": 400, "QQQ": 450, "IWM": 200, "GLD": 185,
		"EWG": 35, "EWU": 30, "EWQ": 28, "EWI": 35, "VGK": 65,
		"EWJ": 68, "ASHR": 26, "MCHI": 22,
	})
	aggData := aggsJSON([]pgBar{
		{T: time.Now().AddDate(0, -1, 0).UnixMilli(), C: 480},
		{T: time.Now().UnixMilli(), C: 500},
	})
	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", CurrentPrice: 82000, PriceChangePerc24h: -2.1},
		{ID: "ethereum", Name: "Ethereum", CurrentPrice: 3000, PriceChangePerc24h: 1.5},
	}
	coinsBody, _ := json.Marshal(coins)
	fngBody := `{"data":[{"value":"12","value_classification":"Extreme Fear"}]}`
	vixBody := vixCSV
	fredBody := fredCSV

	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/v2/aggs/ticker/", statusCode: 200, body: aggData},
		{contains: "/coins/markets", statusCode: 200, body: string(coinsBody)},
		{contains: "/coins/bitcoin/market_chart", statusCode: 200, body: `{"prices":[]}`},
		{contains: "alternative.me", statusCode: 200, body: fngBody},
		{contains: "cdn.cboe.com", statusCode: 200, body: vixBody},
		{contains: "fred.stlouisfed.org", statusCode: 200, body: fredBody},
	})

	resp, err := s.Overview()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Header.MarketStatus == "" {
		t.Error("MarketStatus should not be empty")
	}
	if len(resp.Indices) == 0 {
		t.Error("expected at least one index")
	}
}

func TestOverview_FetchSnapshotsError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 500, body: `{}`},
	})
	_, err := s.Overview()
	if err == nil {
		t.Error("expected error when snapshots fail")
	}
}

func TestOverview_FetchAggregatesError(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500, "DIA": 400, "QQQ": 450, "IWM": 200, "GLD": 185})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/v2/aggs/ticker/", statusCode: 500, body: `{}`},
	})
	_, err := s.Overview()
	if err == nil {
		t.Error("expected error when aggregates fail")
	}
}

func TestOverview_CoinGeckoError(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500, "DIA": 400, "QQQ": 450, "IWM": 200, "GLD": 185})
	aggData := aggsJSON([]pgBar{{T: time.Now().UnixMilli(), C: 500}})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/v2/aggs/ticker/", statusCode: 200, body: aggData},
		{contains: "/coins/markets", statusCode: 503, body: `{}`},
	})
	_, err := s.Overview()
	if err == nil {
		t.Error("expected error when CoinGecko fails")
	}
}

// ── staticDrivers / staticMarketContext ───────────────────────────────────────

func TestStaticDrivers(t *testing.T) {
	drivers := staticDrivers()
	if len(drivers) == 0 {
		t.Error("expected at least one driver")
	}
	for _, d := range drivers {
		if d.Title == "" || d.Text == "" {
			t.Errorf("driver missing title or text: %+v", d)
		}
	}
}

func TestStaticMarketContext(t *testing.T) {
	ctx := staticMarketContext()
	if ctx == "" {
		t.Error("expected non-empty market context")
	}
}

// ── computeUSSummary ──────────────────────────────────────────────────────────

func TestComputeUSSummary_Empty(t *testing.T) {
	s := computeUSSummary(nil, time.Now())
	if s.MarchReturn != "—" {
		t.Errorf("MarchReturn = %q, want '—'", s.MarchReturn)
	}
}

func TestComputeUSSummary_WithData(t *testing.T) {
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	febDate := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	yoyDate := now.AddDate(-1, 0, 0)
	bars := []pgBar{
		{T: yoyDate.UnixMilli(), C: 400},
		{T: febDate.UnixMilli(), C: 450},
		{T: now.UnixMilli(), C: 500},
	}
	s := computeUSSummary(bars, now)
	if s.MarchReturn == "—" {
		t.Error("expected March return to be calculated")
	}
	if s.YoYReturn == "—" {
		t.Error("expected YoY return to be calculated")
	}
}

func TestComputeUSSummary_NoFebBar(t *testing.T) {
	// All bars are in March — no February close found
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	bars := []pgBar{
		{T: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 495},
		{T: now.UnixMilli(), C: 500},
	}
	s := computeUSSummary(bars, now)
	if s.MarchReturn != "—" {
		t.Errorf("MarchReturn = %q, want '—' when no Feb bar", s.MarchReturn)
	}
}

// ── Markets ───────────────────────────────────────────────────────────────────

func TestMarkets_Success(t *testing.T) {
	allTickers := map[string]float64{
		"META": 580, "AMZN": 185, "NVDA": 850, "MSFT": 415, "MU": 95,
		"LMT": 460, "NOC": 440, "AVAV": 190,
		"EWG": 35, "EWU": 30, "EWQ": 28, "EWI": 35, "VGK": 65,
		"EWJ": 68, "ASHR": 26, "MCHI": 22, "EWH": 18, "EWY": 62,
		"ITA": 140, "XLE": 88, "SOXX": 215, "XLV": 145, "XLU": 67,
		"XLF": 43, "XLI": 120, "XLY": 195, "IGV": 380, "JETS": 18, "KWEB": 27,
		"SPY": 500,
	}
	barData := groupedBarsJSON(allTickers)
	aggData := aggsJSON([]pgBar{
		{T: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 400},
		{T: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC).UnixMilli(), C: 460},
		{T: time.Now().UnixMilli(), C: 500},
	})

	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/v2/aggs/ticker/", statusCode: 200, body: aggData},
		{contains: "UMCSENT", statusCode: 200, body: "DATE,UMCSENT\n2024-01-01,78.5\n"},
	})

	resp, err := s.Markets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.IntlIndices) == 0 {
		t.Error("expected international indices")
	}
	if len(resp.SectorData.Labels) == 0 {
		t.Error("expected sector labels")
	}
}

func TestMarkets_SnapshotsError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 500, body: `{}`},
	})
	_, err := s.Markets()
	if err == nil {
		t.Error("expected error when snapshots fail")
	}
}

func TestMarkets_SPYAggregatesError(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"SPY": 500, "META": 580})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/v2/aggs/ticker/", statusCode: 500, body: `{}`},
	})
	_, err := s.Markets()
	if err == nil {
		t.Error("expected error when SPY aggregates fail")
	}
}

// ── Crypto ────────────────────────────────────────────────────────────────────

func TestCrypto_Success(t *testing.T) {
	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", CurrentPrice: 82000, PriceChangePerc24h: -2.1, MarketCap: 1.6e12, TotalVolume: 30e9},
		{ID: "ethereum", Name: "Ethereum", Symbol: "eth", CurrentPrice: 3000, PriceChangePerc24h: 1.5, MarketCap: 0.36e12, TotalVolume: 15e9},
		{ID: "solana", Name: "Solana", Symbol: "sol", CurrentPrice: 140, PriceChangePerc24h: -3.2, MarketCap: 0.06e12, TotalVolume: 5e9},
		{ID: "ripple", Name: "XRP", Symbol: "xrp", CurrentPrice: 0.5, PriceChangePerc24h: 0.8, MarketCap: 0.027e12, TotalVolume: 2e9},
		{ID: "binancecoin", Name: "BNB", Symbol: "bnb", CurrentPrice: 590, PriceChangePerc24h: -1.0, MarketCap: 0.085e12, TotalVolume: 1.8e9},
		{ID: "dogecoin", Name: "Dogecoin", Symbol: "doge", CurrentPrice: 0.16, PriceChangePerc24h: -4.5, MarketCap: 0.024e12, TotalVolume: 1.5e9},
	}
	coinsBody, _ := json.Marshal(coins)
	btcHistory := cgMarketChart{Prices: [][]float64{
		{float64(time.Now().AddDate(-1, 0, 0).UnixMilli()), 45000},
		{float64(time.Now().UnixMilli()), 82000},
	}}
	btcBody, _ := json.Marshal(btcHistory)

	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 200, body: string(coinsBody)},
		{contains: "/coins/bitcoin/market_chart", statusCode: 200, body: string(btcBody)},
	})

	resp, err := s.Crypto()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.CryptoAssets) == 0 {
		t.Error("expected crypto assets")
	}
	if resp.BtcHistoryData.Labels == nil {
		t.Error("expected BTC history labels")
	}
	if resp.MarketContext == "" {
		t.Error("expected market context")
	}
}

func TestCrypto_CoinGeckoMarketsError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 503, body: `{}`},
	})
	_, err := s.Crypto()
	if err == nil {
		t.Error("expected error when CoinGecko fails")
	}
}

func TestCrypto_BTCHistoryError(t *testing.T) {
	coins := []cgCoin{{ID: "bitcoin", Name: "Bitcoin"}}
	coinsBody, _ := json.Marshal(coins)
	s := newTestStore([]mockRoute{
		{contains: "/coins/markets", statusCode: 200, body: string(coinsBody)},
		{contains: "/coins/bitcoin/market_chart", statusCode: 503, body: `{}`},
	})
	_, err := s.Crypto()
	if err == nil {
		t.Error("expected error when BTC history fails")
	}
}

// ── Recommendations ───────────────────────────────────────────────────────────

func TestRecommendations_Success(t *testing.T) {
	// Need grouped bars for today, 1-month, 3-month, and yesterday
	allBars := make(map[string]float64)
	// Add all candidates from aiSemis, disruptors, defense universes
	for _, c := range aiSemisUniverse {
		allBars[c.ticker] = 100 + float64(len(c.ticker))
	}
	for _, c := range disruptorsUniverse {
		allBars[c.ticker] = 50 + float64(len(c.ticker))
	}
	for _, c := range defenseUniverse {
		allBars[c.ticker] = 200 + float64(len(c.ticker))
	}
	barData := groupedBarsJSON(allBars)

	// Lower prices for 1-month bars (simulate positive momentum)
	lowerBars := make(map[string]float64)
	for k, v := range allBars {
		lowerBars[k] = v * 0.9
	}
	lowerData := groupedBarsJSON(lowerBars)

	coins := []cgCoin{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", CurrentPrice: 82000, PriceChangePerc24h: -2.1},
		{ID: "ethereum", Name: "Ethereum", Symbol: "eth", CurrentPrice: 3000, PriceChangePerc24h: 1.5},
		{ID: "solana", Name: "Solana", Symbol: "sol", CurrentPrice: 140, PriceChangePerc24h: -3.2},
		{ID: "ripple", Name: "XRP", Symbol: "xrp", CurrentPrice: 0.5, PriceChangePerc24h: 0.8},
		{ID: "binancecoin", Name: "BNB", Symbol: "bnb", CurrentPrice: 590, PriceChangePerc24h: -1.0},
	}
	coinsBody, _ := json.Marshal(coins)

	callN := 0
	tr := &funcTransport{fn: func(req *http.Request) (*http.Response, error) {
		u := req.URL.String()
		if strings.Contains(u, "/coins/markets") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(coinsBody))), Header: make(http.Header)}, nil
		}
		// Grouped bars: first 2 calls return current prices, subsequent return lower (1-month)
		callN++
		body := barData
		if callN > 2 {
			body = lowerData
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	}}
	done := make(chan struct{})
	close(done)
	s := &Store{
		polygonKey: "test",
		client:     &http.Client{Transport: tr},
		gbCache:    make(map[string]map[string]float64),
		warmupDone: done,
	}

	resp, err := s.Recommendations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.AISemis) == 0 {
		t.Error("expected AI semis recommendations")
	}
	if len(resp.Crypto) == 0 {
		t.Error("expected crypto recommendations")
	}
}

func TestRecommendations_GroupedBarsError(t *testing.T) {
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 500, body: `{}`},
	})
	_, err := s.Recommendations()
	if err == nil {
		t.Error("expected error when grouped bars fail")
	}
}

func TestRecommendations_CoinGeckoError(t *testing.T) {
	barData := groupedBarsJSON(map[string]float64{"NVDA": 850})
	s := newTestStore([]mockRoute{
		{contains: "/v2/aggs/grouped/", statusCode: 200, body: barData},
		{contains: "/coins/markets", statusCode: 503, body: `{}`},
	})
	_, err := s.Recommendations()
	if err == nil {
		t.Error("expected error when CoinGecko fails")
	}
}

// ── New ───────────────────────────────────────────────────────────────────────

func TestNew_EmptyPolygonKey(t *testing.T) {
	_, err := New("", "")
	if err == nil {
		t.Error("expected error for empty polygon key")
	}
}

func TestNew_ValidKey(t *testing.T) {
	// New starts a goroutine (warmupHistoricBars) — we don't wait for it.
	// Just verify the store is created without error.
	s, err := New("test-polygon-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Error("expected non-nil store")
	}
	// Drain to avoid goroutine leak in test (warmup will fail since URLs are fake, returns quickly)
	fmt.Sprintf("%p", s) // just use s
}
