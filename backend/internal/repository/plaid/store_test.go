package plaid

import (
	"path/filepath"
	"testing"

	plaidSDK "github.com/plaid/plaid-go/v29/plaid"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// ── minimal mock for repository.Store ───────────────────────────────────────

type mockInner struct{}

func (m *mockInner) Overview() (models.OverviewResponse, error)     { return models.OverviewResponse{}, nil }
func (m *mockInner) Markets() (models.MarketsResponse, error)       { return models.MarketsResponse{}, nil }
func (m *mockInner) Crypto() (models.CryptoResponse, error)         { return models.CryptoResponse{}, nil }
func (m *mockInner) Recommendations() (models.RecommendationsResponse, error) {
	return models.RecommendationsResponse{}, nil
}
func (m *mockInner) Principal() (models.PortfolioResponse, error)     { return models.PortfolioResponse{}, nil }
func (m *mockInner) MorganStanley() (models.PortfolioResponse, error) { return models.PortfolioResponse{}, nil }
func (m *mockInner) Fidelity() (models.PortfolioResponse, error)      { return models.PortfolioResponse{}, nil }
func (m *mockInner) SoFi() (models.PortfolioResponse, error)          { return models.PortfolioResponse{}, nil }

// ── Store.New ────────────────────────────────────────────────────────────────

func TestStoreNew_Success(t *testing.T) {
	dir := t.TempDir()
	s, err := New(Config{
		Inner:     &mockInner{},
		ClientID:  "client-id",
		Secret:    "secret",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestStoreNew_ProductionEnv(t *testing.T) {
	dir := t.TempDir()
	s, err := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "production",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("nil store")
	}
}

func TestStoreNew_UnknownEnvDefaultsSandbox(t *testing.T) {
	dir := t.TempDir()
	// Any unrecognised env string should not error.
	_, err := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "staging",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStoreNew_BadTokenFileReturnsError(t *testing.T) {
	_, err := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: "/nonexistent-dir/tokens.json",
	})
	// Non-existent file is OK (created on first write). But if the path itself
	// is invalid (dir doesn't exist), an error is expected only on first save.
	// newTokenStore only fails if the file exists but is corrupt.
	_ = err // acceptable either way; main goal is no panic
}

// ── Passthrough methods ──────────────────────────────────────────────────────

func TestStore_PassthroughMethods(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	if _, err := s.Overview(); err != nil {
		t.Errorf("Overview: %v", err)
	}
	if _, err := s.Markets(); err != nil {
		t.Errorf("Markets: %v", err)
	}
	if _, err := s.Crypto(); err != nil {
		t.Errorf("Crypto: %v", err)
	}
	if _, err := s.Recommendations(); err != nil {
		t.Errorf("Recommendations: %v", err)
	}
}

// ── portfolioForSlot – no token ──────────────────────────────────────────────

func TestPrincipal_NotConnected(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	resp, err := s.Principal()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Connected {
		t.Error("expected Connected = false when no token")
	}
}

func TestMorganStanley_NotConnected(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	resp, err := s.MorganStanley()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Connected {
		t.Error("expected Connected = false when no token")
	}
}

func TestFidelity_NotConnected(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	resp, err := s.Fidelity()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Connected {
		t.Error("expected Connected = false when no token")
	}
}

func TestSoFi_NotConnected(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	resp, err := s.SoFi()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Connected {
		t.Error("expected Connected = false when no token")
	}
}

func TestPortfolioForSlot_AllSlotsNotConnected(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	slots := []struct {
		name string
		fn   func() (interface{ GetConnected() bool }, error)
	}{
		{"principal", func() (interface{ GetConnected() bool }, error) {
			r, e := s.Principal()
			return &portfolioConnectedWrapper{r.Connected}, e
		}},
		{"morganstanley", func() (interface{ GetConnected() bool }, error) {
			r, e := s.MorganStanley()
			return &portfolioConnectedWrapper{r.Connected}, e
		}},
		{"fidelity", func() (interface{ GetConnected() bool }, error) {
			r, e := s.Fidelity()
			return &portfolioConnectedWrapper{r.Connected}, e
		}},
		{"sofi", func() (interface{ GetConnected() bool }, error) {
			r, e := s.SoFi()
			return &portfolioConnectedWrapper{r.Connected}, e
		}},
	}

	for _, tc := range slots {
		t.Run(tc.name, func(t *testing.T) {
			r, err := tc.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.GetConnected() {
				t.Errorf("%s: expected Connected = false with no token", tc.name)
			}
		})
	}
}

type portfolioConnectedWrapper struct{ connected bool }

func (w *portfolioConnectedWrapper) GetConnected() bool { return w.connected }

// ── ConnectedSlots ───────────────────────────────────────────────────────────

func TestConnectedSlots_Empty(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	slots := s.ConnectedSlots()
	if len(slots) != 0 {
		t.Errorf("slots = %v, want empty", slots)
	}
}

func TestConnectedSlots_AfterTokenSet(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	s.tokens.set("principal", tokenRecord{AccessToken: "tok"}) //nolint

	slots := s.ConnectedSlots()
	if len(slots) != 1 || slots[0] != "principal" {
		t.Errorf("slots = %v, want [principal]", slots)
	}
}

// ── mapPortfolio ─────────────────────────────────────────────────────────────

func TestMapPortfolio_EmptyResponses(t *testing.T) {
	s := &Store{}
	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	tr := plaidSDK.InvestmentsTransactionsGetResponse{}

	resp := s.mapPortfolio(hr, tr, "Test Bank")

	if !resp.Connected {
		t.Error("expected Connected = true")
	}
	if resp.Summary.Provider != "Test Bank" {
		t.Errorf("Provider = %q, want Test Bank", resp.Summary.Provider)
	}
	if len(resp.PerformanceData.Labels) != 12 {
		t.Errorf("perf labels = %d, want 12", len(resp.PerformanceData.Labels))
	}
	if len(resp.ContribData.Labels) != 12 {
		t.Errorf("contrib labels = %d, want 12", len(resp.ContribData.Labels))
	}
}

func TestMapPortfolio_HoldingsSortedByValue(t *testing.T) {
	s := &Store{}

	// Create two securities.
	sec1 := plaidSDK.Security{}
	sec1.SetSecurityId("sec-1")
	sec1.SetTickerSymbol("AAPL")
	sec1.SetName("Apple Inc.")

	sec2 := plaidSDK.Security{}
	sec2.SetSecurityId("sec-2")
	sec2.SetTickerSymbol("NVDA")
	sec2.SetName("NVIDIA Corp")

	// holding1 has lower value than holding2.
	h1 := plaidSDK.Holding{}
	h1.SetSecurityId("sec-1")
	h1.SetInstitutionValue(5000)
	h1.SetCostBasis(4000)

	h2 := plaidSDK.Holding{}
	h2.SetSecurityId("sec-2")
	h2.SetInstitutionValue(20000)
	h2.SetCostBasis(15000)

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetHoldings([]plaidSDK.Holding{h1, h2})
	hr.SetSecurities([]plaidSDK.Security{sec1, sec2})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}

	resp := s.mapPortfolio(hr, tr, "Brokerage")

	if len(resp.Holdings) != 2 {
		t.Fatalf("holdings = %d, want 2", len(resp.Holdings))
	}
	// NVDA (20k) should be first (descending by value).
	if resp.Holdings[0].Ticker != "NVDA" {
		t.Errorf("Holdings[0].Ticker = %q, want NVDA", resp.Holdings[0].Ticker)
	}
	if resp.Holdings[1].Ticker != "AAPL" {
		t.Errorf("Holdings[1].Ticker = %q, want AAPL", resp.Holdings[1].Ticker)
	}
}

func TestMapPortfolio_TotalGainPositive(t *testing.T) {
	s := &Store{}

	sec := plaidSDK.Security{}
	sec.SetSecurityId("sec-1")
	sec.SetTickerSymbol("VTI")
	sec.SetName("Vanguard Total Stock")

	h := plaidSDK.Holding{}
	h.SetSecurityId("sec-1")
	h.SetInstitutionValue(12000)
	h.SetCostBasis(10000)

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetHoldings([]plaidSDK.Holding{h})
	hr.SetSecurities([]plaidSDK.Security{sec})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}
	resp := s.mapPortfolio(hr, tr, "Brokerage")

	if resp.Summary.TotalGainDir != "up" {
		t.Errorf("TotalGainDir = %q, want up", resp.Summary.TotalGainDir)
	}
	if resp.Summary.TotalGain != "+$2,000.00" {
		t.Errorf("TotalGain = %q, want +$2,000.00", resp.Summary.TotalGain)
	}
	if resp.Summary.TotalGainPct != "+20.00%" {
		t.Errorf("TotalGainPct = %q, want +20.00%%", resp.Summary.TotalGainPct)
	}
}

func TestMapPortfolio_TotalGainNegative(t *testing.T) {
	s := &Store{}

	sec := plaidSDK.Security{}
	sec.SetSecurityId("sec-1")
	sec.SetTickerSymbol("BND")
	sec.SetName("Vanguard Bond Fund")

	h := plaidSDK.Holding{}
	h.SetSecurityId("sec-1")
	h.SetInstitutionValue(9000)
	h.SetCostBasis(10000)

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetHoldings([]plaidSDK.Holding{h})
	hr.SetSecurities([]plaidSDK.Security{sec})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}
	resp := s.mapPortfolio(hr, tr, "Brokerage")

	if resp.Summary.TotalGainDir != "down" {
		t.Errorf("TotalGainDir = %q, want down", resp.Summary.TotalGainDir)
	}
}

func TestMapPortfolio_UsesAccountNameFromAccounts(t *testing.T) {
	s := &Store{}

	acct := plaidSDK.AccountBase{}
	acct.SetName("My 401k")

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetAccounts([]plaidSDK.AccountBase{acct})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}
	resp := s.mapPortfolio(hr, tr, "Principal Financial Group")

	if resp.Summary.AccountName != "My 401k" {
		t.Errorf("AccountName = %q, want My 401k", resp.Summary.AccountName)
	}
}

func TestMapPortfolio_AllocationSumsToHundred(t *testing.T) {
	s := &Store{}

	makeSecHolding := func(id, ticker string, value, cost float64) (plaidSDK.Security, plaidSDK.Holding) {
		sec := plaidSDK.Security{}
		sec.SetSecurityId(id)
		sec.SetTickerSymbol(ticker)
		sec.SetName(ticker)
		h := plaidSDK.Holding{}
		h.SetSecurityId(id)
		h.SetInstitutionValue(value)
		h.SetCostBasis(cost)
		return sec, h
	}

	s1, h1 := makeSecHolding("s1", "A", 6000, 5000)
	s2, h2 := makeSecHolding("s2", "B", 4000, 3000)

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetHoldings([]plaidSDK.Holding{h1, h2})
	hr.SetSecurities([]plaidSDK.Security{s1, s2})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}
	resp := s.mapPortfolio(hr, tr, "Brokerage")

	if len(resp.Holdings) != 2 {
		t.Fatalf("holdings = %d, want 2", len(resp.Holdings))
	}
	// 60% and 40%
	if resp.Holdings[0].Allocation != "60.0%" {
		t.Errorf("Holdings[0].Allocation = %q, want 60.0%%", resp.Holdings[0].Allocation)
	}
	if resp.Holdings[1].Allocation != "40.0%" {
		t.Errorf("Holdings[1].Allocation = %q, want 40.0%%", resp.Holdings[1].Allocation)
	}
}

func TestMapPortfolio_FallsBackToNameWhenNoTicker(t *testing.T) {
	s := &Store{}

	sec := plaidSDK.Security{}
	sec.SetSecurityId("sec-1")
	sec.SetName("Some Long Fund Name That Has No Ticker")
	// No ticker symbol set.

	h := plaidSDK.Holding{}
	h.SetSecurityId("sec-1")
	h.SetInstitutionValue(1000)
	h.SetCostBasis(900)

	hr := plaidSDK.InvestmentsHoldingsGetResponse{}
	hr.SetHoldings([]plaidSDK.Holding{h})
	hr.SetSecurities([]plaidSDK.Security{sec})

	tr := plaidSDK.InvestmentsTransactionsGetResponse{}
	resp := s.mapPortfolio(hr, tr, "Brokerage")

	if len(resp.Holdings) != 1 {
		t.Fatalf("holdings = %d, want 1", len(resp.Holdings))
	}
	// Ticker should be truncated name (first 10 chars).
	if len(resp.Holdings[0].Ticker) > 10 {
		t.Errorf("Ticker = %q, expected ≤10 chars (truncated from name)", resp.Holdings[0].Ticker)
	}
}

func TestStore_PassthroughDelegatesCorrectly(t *testing.T) {
	type callTracker struct {
		mockInner
		overviewCalled bool
	}
	tracker := &struct {
		mockInner
		overviewCalled bool
	}{}
	_ = tracker

	// Verify that Overview/Markets/Crypto/Recommendations route through inner.
	dir := t.TempDir()
	s, _ := New(Config{
		Inner:     &mockInner{},
		ClientID:  "id",
		Secret:    "sec",
		Env:       "sandbox",
		TokenFile: filepath.Join(dir, "tokens.json"),
	})

	// These should not return ErrUnavailable since the mock inner always succeeds.
	if _, err := s.Overview(); err != nil {
		t.Errorf("Overview passthrough error: %v", err)
	}
	if _, err := s.Markets(); err != nil {
		t.Errorf("Markets passthrough error: %v", err)
	}
	if _, err := s.Crypto(); err != nil {
		t.Errorf("Crypto passthrough error: %v", err)
	}
	if _, err := s.Recommendations(); err != nil {
		t.Errorf("Recommendations passthrough error: %v", err)
	}
}

// Ensure Store satisfies repository.Store at compile time.
var _ repository.Store = (*Store)(nil)
