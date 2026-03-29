// Package plaid provides a [repository.Store] that sources personal
// portfolio data from the Plaid API (https://plaid.com).
//
// Market data methods (Overview, Markets, Crypto, Recommendations) are
// delegated to the wrapped inner store. Only Principal and MorganStanley
// are sourced from Plaid; they require the user to connect their accounts
// via the Plaid Link flow before returning real data.
package plaid

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	plaidSDK "github.com/plaid/plaid-go/v29/plaid"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

const (
	SlotPrincipal     = "principal"
	SlotMorganStanley = "morganstanley"
)

// Config holds construction parameters for the Plaid store.
type Config struct {
	// Inner is used for all non-portfolio methods (Overview, Markets, etc.).
	Inner repository.Store

	ClientID string
	Secret   string
	// Env is "sandbox" or "production". Defaults to "sandbox".
	Env string
	// TokenFile is the path to the JSON file that persists access tokens.
	TokenFile string
}

// Store implements repository.Store using Plaid for portfolio data.
// All methods are safe for concurrent use.
type Store struct {
	inner  repository.Store
	client *plaidSDK.APIClient
	tokens *tokenStore
}

// New creates a Plaid-backed Store.
func New(cfg Config) (*Store, error) {
	plaidCfg := plaidSDK.NewConfiguration()
	plaidCfg.AddDefaultHeader("PLAID-CLIENT-ID", cfg.ClientID)
	plaidCfg.AddDefaultHeader("PLAID-SECRET", cfg.Secret)

	if strings.ToLower(cfg.Env) == "production" {
		plaidCfg.UseEnvironment(plaidSDK.Production)
	} else {
		plaidCfg.UseEnvironment(plaidSDK.Sandbox)
	}

	ts, err := newTokenStore(cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("plaid: token store: %w", err)
	}

	return &Store{
		inner:  cfg.Inner,
		client: plaidSDK.NewAPIClient(plaidCfg),
		tokens: ts,
	}, nil
}

// ── repository.Store passthrough methods ────────────────────────────────────

func (s *Store) Overview() (models.OverviewResponse, error) { return s.inner.Overview() }
func (s *Store) Markets() (models.MarketsResponse, error)   { return s.inner.Markets() }
func (s *Store) Crypto() (models.CryptoResponse, error)     { return s.inner.Crypto() }
func (s *Store) Recommendations() (models.RecommendationsResponse, error) {
	return s.inner.Recommendations()
}

// ── Portfolio methods ────────────────────────────────────────────────────────

func (s *Store) Principal() (models.PortfolioResponse, error) {
	return s.portfolioForSlot(SlotPrincipal)
}

func (s *Store) MorganStanley() (models.PortfolioResponse, error) {
	return s.portfolioForSlot(SlotMorganStanley)
}

func (s *Store) portfolioForSlot(slot string) (models.PortfolioResponse, error) {
	rec, ok := s.tokens.get(slot)
	if !ok {
		return models.PortfolioResponse{Connected: false}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return s.fetchPortfolio(ctx, rec.AccessToken, rec.InstitutionName)
}

func (s *Store) fetchPortfolio(ctx context.Context, accessToken, institutionName string) (models.PortfolioResponse, error) {
	type holdingsResult struct {
		resp plaidSDK.InvestmentsHoldingsGetResponse
		err  error
	}
	type txResult struct {
		resp plaidSDK.InvestmentsTransactionsGetResponse
		err  error
	}

	hCh := make(chan holdingsResult, 1)
	tCh := make(chan txResult, 1)

	go func() {
		req := plaidSDK.NewInvestmentsHoldingsGetRequest(accessToken)
		resp, _, err := s.client.PlaidApi.InvestmentsHoldingsGet(ctx).InvestmentsHoldingsGetRequest(*req).Execute()
		hCh <- holdingsResult{resp, err}
	}()

	go func() {
		endDate := time.Now().Format("2006-01-02")
		startDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		req := plaidSDK.NewInvestmentsTransactionsGetRequest(accessToken, startDate, endDate)
		opts := plaidSDK.InvestmentsTransactionsGetRequestOptions{}
		opts.SetCount(500)
		req.SetOptions(opts)
		resp, _, err := s.client.PlaidApi.InvestmentsTransactionsGet(ctx).InvestmentsTransactionsGetRequest(*req).Execute()
		tCh <- txResult{resp, err}
	}()

	hr := <-hCh
	tr := <-tCh

	if hr.err != nil {
		return models.PortfolioResponse{}, fmt.Errorf("plaid holdings: %w", hr.err)
	}
	if tr.err != nil {
		return models.PortfolioResponse{}, fmt.Errorf("plaid transactions: %w", tr.err)
	}

	return s.mapPortfolio(hr.resp, tr.resp, institutionName), nil
}

func (s *Store) mapPortfolio(
	hr plaidSDK.InvestmentsHoldingsGetResponse,
	tr plaidSDK.InvestmentsTransactionsGetResponse,
	institutionName string,
) models.PortfolioResponse {
	// Build security lookup map.
	secByID := map[string]plaidSDK.Security{}
	for _, sec := range hr.GetSecurities() {
		secByID[sec.GetSecurityId()] = sec
	}

	// Sum total current value and cost basis.
	totalValue := 0.0
	totalCost := 0.0
	for _, h := range hr.GetHoldings() {
		totalValue += h.GetInstitutionValue()
		totalCost += h.GetCostBasis()
	}

	// Build holdings list, sorted by value descending.
	type holdingItem struct {
		models.PortfolioHolding
		val float64
	}
	items := make([]holdingItem, 0, len(hr.GetHoldings()))
	for _, h := range hr.GetHoldings() {
		sec := secByID[h.GetSecurityId()]
		gain := h.GetInstitutionValue() - h.GetCostBasis()
		gainDir := "up"
		if gain < 0 {
			gainDir = "down"
		}
		alloc := 0.0
		if totalValue > 0 {
			alloc = h.GetInstitutionValue() / totalValue * 100
		}
		ticker := sec.GetTickerSymbol()
		if ticker == "" {
			ticker = truncate(sec.GetName(), 10)
		}
		items = append(items, holdingItem{
			val: h.GetInstitutionValue(),
			PortfolioHolding: models.PortfolioHolding{
				Ticker:     ticker,
				Name:       sec.GetName(),
				Value:      fmtDollar(h.GetInstitutionValue()),
				Allocation: fmt.Sprintf("%.1f%%", alloc),
				Gain:       fmtSignedDollar(gain),
				GainDir:    gainDir,
			},
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].val > items[j].val })
	holdings := make([]models.PortfolioHolding, len(items))
	for i, it := range items {
		holdings[i] = it.PortfolioHolding
	}

	// Detect account type from institution name.
	accountType := "Investment"
	lowerInst := strings.ToLower(institutionName)
	if strings.Contains(lowerInst, "401") || strings.Contains(lowerInst, "principal") {
		accountType = "401(k)"
	} else if strings.Contains(lowerInst, "ira") {
		accountType = "IRA"
	}

	// Total gain metrics.
	totalGain := totalValue - totalCost
	totalGainDir := "up"
	if totalGain < 0 {
		totalGainDir = "down"
	}
	totalGainPct := 0.0
	if totalCost > 0 {
		totalGainPct = totalGain / totalCost * 100
	}

	// Change metrics and contribution history from transactions.
	weekly, monthly, ytd := computeChangeMetrics(tr.GetInvestmentTransactions(), totalValue)
	contributions := buildContributions(tr.GetInvestmentTransactions())
	contribData, perfData := buildChartData(tr.GetInvestmentTransactions(), totalValue)

	// Use the first account name if available.
	accountName := institutionName
	for _, acct := range hr.GetAccounts() {
		if acct.GetName() != "" {
			accountName = acct.GetName()
			break
		}
	}

	return models.PortfolioResponse{
		Connected: true,
		Summary: models.PortfolioSummary{
			AccountName:   accountName,
			AccountType:   accountType,
			Provider:      institutionName,
			CurrentValue:  fmtDollar(totalValue),
			TotalCost:     fmtDollar(totalCost),
			TotalGain:     fmtSignedDollar(totalGain),
			TotalGainPct:  fmtSignedPct(totalGainPct),
			TotalGainDir:  totalGainDir,
			WeeklyChange:  weekly,
			MonthlyChange: monthly,
			YTDChange:     ytd,
		},
		PerformanceData: perfData,
		ContribData:     contribData,
		Holdings:        holdings,
		Contributions:   contributions,
	}
}

// computeChangeMetrics estimates weekly, monthly, and YTD change by summing
// net cash flows from investment transactions over each window.
// Note: these are net flow approximations — market price movements are
// not included. Use for directional guidance only.
func computeChangeMetrics(txns []plaidSDK.InvestmentTransaction, currentValue float64) (weekly, monthly, ytd models.ChangeMetric) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	var weekNet, monthNet, ytdNet float64
	for _, tx := range txns {
		d, err := time.Parse("2006-01-02", tx.GetDate())
		if err != nil {
			continue
		}
		// Plaid investment amounts: positive = cash leaving account, negative = cash entering.
		// Invert so that contributions appear as positive changes.
		amt := -tx.GetAmount()
		if d.After(weekAgo) {
			weekNet += amt
		}
		if d.After(monthAgo) {
			monthNet += amt
		}
		if !d.Before(yearStart) {
			ytdNet += amt
		}
	}

	weekPct := pct(weekNet, currentValue-weekNet)
	monthPct := pct(monthNet, currentValue-monthNet)
	ytdPct := pct(ytdNet, currentValue-ytdNet)

	return toChangeMetric(weekNet, weekPct),
		toChangeMetric(monthNet, monthPct),
		toChangeMetric(ytdNet, ytdPct)
}

func pct(change, base float64) float64 {
	if base <= 0 {
		return 0
	}
	return change / base * 100
}

func toChangeMetric(amount, p float64) models.ChangeMetric {
	dir := "up"
	if amount < 0 {
		dir = "down"
	} else if amount == 0 {
		dir = "flat"
	}
	return models.ChangeMetric{
		Amount: fmtSignedDollar(amount),
		Pct:    fmtSignedPct(p),
		Dir:    dir,
	}
}

// buildContributions extracts contribution/deposit/rollover transactions.
func buildContributions(txns []plaidSDK.InvestmentTransaction) []models.Contribution {
	var out []models.Contribution
	for _, tx := range txns {
		subtype := tx.GetSubtype()
		if subtype != plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION &&
			subtype != plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_DEPOSIT &&
			subtype != plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_TRANSFER {
			continue
		}
		ctype := "deposit"
		if subtype == plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_CONTRIBUTION {
			ctype = "employee"
		} else if subtype == plaidSDK.INVESTMENTTRANSACTIONSUBTYPE_TRANSFER {
			ctype = "rollover"
		}
		out = append(out, models.Contribution{
			Date:   formatDate(tx.GetDate()),
			Amount: fmtDollar(math.Abs(tx.GetAmount())),
			Type:   ctype,
			Note:   tx.GetName(),
		})
	}
	return out
}

// buildChartData produces monthly contribution totals and an approximate
// portfolio balance history by working backwards from the current balance.
func buildChartData(txns []plaidSDK.InvestmentTransaction, currentValue float64) (contribData, perfData models.ChartData) {
	type monthKey struct{ year, month int }

	// Net cash flow per month (positive = net inflow).
	netByMonth := map[monthKey]float64{}
	for _, tx := range txns {
		d, err := time.Parse("2006-01-02", tx.GetDate())
		if err != nil {
			continue
		}
		k := monthKey{d.Year(), int(d.Month())}
		netByMonth[k] -= tx.GetAmount() // invert Plaid sign convention
	}

	now := time.Now()
	labels := make([]string, 12)
	contribAmts := make([]float64, 12)
	perfAmts := make([]float64, 12)

	// Walk backwards from current balance to approximate monthly snapshots.
	runningBalance := currentValue
	for i := 11; i >= 0; i-- {
		t := now.AddDate(0, -(11-i), 0)
		k := monthKey{t.Year(), int(t.Month())}
		labels[i] = t.Format("Jan '06")

		net := netByMonth[k]
		contribAmts[i] = math.Max(0, net) // only positive flows (contributions/deposits)

		if i < 11 {
			runningBalance -= net
		}
		perfAmts[i] = math.Max(0, runningBalance)
	}
	perfAmts[11] = currentValue

	return models.ChartData{Labels: labels, Data: contribAmts},
		models.ChartData{Labels: labels, Data: perfAmts}
}

// ── Plaid Link management ────────────────────────────────────────────────────

// CreateLinkToken creates a Plaid Link token for the given account slot.
func (s *Store) CreateLinkToken(ctx context.Context, slot string) (string, error) {
	user := plaidSDK.NewLinkTokenCreateRequestUser("apex-user")
	req := plaidSDK.NewLinkTokenCreateRequest(
		"APEX Dashboard",
		"en",
		[]plaidSDK.CountryCode{plaidSDK.COUNTRYCODE_US},
		*user,
	)
	req.SetProducts([]plaidSDK.Products{plaidSDK.PRODUCTS_INVESTMENTS})

	resp, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", fmt.Errorf("plaid link token: %w", err)
	}
	return resp.GetLinkToken(), nil
}

// ExchangeToken exchanges a Plaid public token for a persistent access token
// stored under the given slot name.
func (s *Store) ExchangeToken(ctx context.Context, slot, publicToken string) error {
	req := plaidSDK.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return fmt.Errorf("plaid exchange: %w", err)
	}

	// Look up the institution name for the item.
	instName := slot
	itemResp, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(
		*plaidSDK.NewItemGetRequest(resp.GetAccessToken()),
	).Execute()
	if err == nil && itemResp.Item.InstitutionId.IsSet() {
		if instID := itemResp.Item.InstitutionId.Get(); instID != nil {
			instResp, _, err2 := s.client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(
				*plaidSDK.NewInstitutionsGetByIdRequest(*instID, []plaidSDK.CountryCode{plaidSDK.COUNTRYCODE_US}),
			).Execute()
			if err2 == nil {
				instName = instResp.Institution.GetName()
			}
		}
	}

	return s.tokens.set(slot, tokenRecord{
		AccessToken:     resp.GetAccessToken(),
		ItemID:          resp.GetItemId(),
		InstitutionName: instName,
		ConnectedAt:     time.Now(),
	})
}

// ConnectedSlots returns the slot names that have a stored access token.
func (s *Store) ConnectedSlots() []string {
	return s.tokens.slots()
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func fmtDollar(v float64) string {
	return "$" + formatCommas(v)
}

func fmtSignedDollar(v float64) string {
	if v >= 0 {
		return "+$" + formatCommas(v)
	}
	return "-$" + formatCommas(math.Abs(v))
}

func fmtSignedPct(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}

func formatCommas(v float64) string {
	s := fmt.Sprintf("%.2f", math.Abs(v))
	parts := strings.SplitN(s, ".", 2)
	n := parts[0]
	var out []byte
	for i := range n {
		if i > 0 && (len(n)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, n[i])
	}
	if len(parts) > 1 {
		return string(out) + "." + parts[1]
	}
	return string(out)
}

func formatDate(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	return t.Format("Jan 2, 2006")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
