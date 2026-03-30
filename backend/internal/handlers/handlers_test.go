package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"apex-dashboard/backend/internal/models"
	"apex-dashboard/backend/internal/repository"
)

// ── Mock Store ───────────────────────────────────────────────────────────────

type mockStore struct {
	overviewResp      models.OverviewResponse
	overviewErr       error
	marketsResp       models.MarketsResponse
	marketsErr        error
	cryptoResp        models.CryptoResponse
	cryptoErr         error
	recsResp          models.RecommendationsResponse
	recsErr           error
	principalResp     models.PortfolioResponse
	principalErr      error
	morganStanleyResp models.PortfolioResponse
	morganStanleyErr  error
	fidelityResp      models.PortfolioResponse
	fidelityErr       error
	sofiResp          models.PortfolioResponse
	sofiErr           error
}

func (m *mockStore) Overview() (models.OverviewResponse, error) {
	return m.overviewResp, m.overviewErr
}
func (m *mockStore) Markets() (models.MarketsResponse, error) {
	return m.marketsResp, m.marketsErr
}
func (m *mockStore) Crypto() (models.CryptoResponse, error) {
	return m.cryptoResp, m.cryptoErr
}
func (m *mockStore) Recommendations() (models.RecommendationsResponse, error) {
	return m.recsResp, m.recsErr
}
func (m *mockStore) Principal() (models.PortfolioResponse, error) {
	return m.principalResp, m.principalErr
}
func (m *mockStore) MorganStanley() (models.PortfolioResponse, error) {
	return m.morganStanleyResp, m.morganStanleyErr
}
func (m *mockStore) Fidelity() (models.PortfolioResponse, error) {
	return m.fidelityResp, m.fidelityErr
}
func (m *mockStore) SoFi() (models.PortfolioResponse, error) {
	return m.sofiResp, m.sofiErr
}

// ── Mock PlaidManager ────────────────────────────────────────────────────────

type mockPlaid struct {
	linkToken    string
	linkTokenErr error
	exchangeErr  error
	slots        []string
}

func (m *mockPlaid) CreateLinkToken(_ context.Context, _ string) (string, error) {
	return m.linkToken, m.linkTokenErr
}
func (m *mockPlaid) ExchangeToken(_ context.Context, _, _ string) error {
	return m.exchangeErr
}
func (m *mockPlaid) ConnectedSlots() []string {
	return m.slots
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func newHandler(t *testing.T, store *mockStore, plaid PlaidManager) *Handler {
	t.Helper()
	h, err := New(Config{Store: store, Plaid: plaid})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return h
}

func doRequest(h http.HandlerFunc, method, path string, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func assertJSON(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d (body: %s)", w.Code, want, w.Body.String())
	}
}

func decodeData(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v (body: %s)", err, w.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			t.Fatalf("decode data: %v", err)
		}
	}
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) (code, message string) {
	t.Helper()
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	return env.Error.Code, env.Error.Message
}

// ── Handler.New ──────────────────────────────────────────────────────────────

func TestNew_NilStore(t *testing.T) {
	_, err := New(Config{Store: nil})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("err = %v, want ErrInvalidConfig", err)
	}
}

func TestNew_NilLoggerDefaults(t *testing.T) {
	h, err := New(Config{Store: &mockStore{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.logger == nil {
		t.Error("logger should not be nil")
	}
}

func TestNew_Success(t *testing.T) {
	h, err := New(Config{Store: &mockStore{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Error("handler should not be nil")
	}
}

// ── Market endpoint handlers ─────────────────────────────────────────────────

type marketHandlerCase struct {
	name        string
	storeErr    error
	wantStatus  int
	wantErrCode string
}

var commonCases = []marketHandlerCase{
	{
		name:       "success",
		wantStatus: http.StatusOK,
	},
	{
		name:        "unavailable",
		storeErr:    repository.ErrUnavailable,
		wantStatus:  http.StatusServiceUnavailable,
		wantErrCode: "SERVICE_UNAVAILABLE",
	},
	{
		name:        "internal error",
		storeErr:    errors.New("boom"),
		wantStatus:  http.StatusInternalServerError,
		wantErrCode: "INTERNAL_ERROR",
	},
}

func TestHandleOverview(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				overviewResp: models.OverviewResponse{
					Header: models.HeaderInfo{FearGreed: "FEAR: 20"},
				},
				overviewErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleOverview, http.MethodGet, "/api/v1/overview", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.OverviewResponse
				decodeData(t, w, &data)
				if data.Header.FearGreed != "FEAR: 20" {
					t.Errorf("FearGreed = %q, want FEAR: 20", data.Header.FearGreed)
				}
			} else {
				code, _ := decodeError(t, w)
				if code != tc.wantErrCode {
					t.Errorf("error code = %q, want %q", code, tc.wantErrCode)
				}
			}
		})
	}
}

func TestHandleMarkets(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				marketsResp: models.MarketsResponse{
					USSummary: models.USSummary{MarchReturn: "+3.2%"},
				},
				marketsErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleMarkets, http.MethodGet, "/api/v1/markets", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.MarketsResponse
				decodeData(t, w, &data)
				if data.USSummary.MarchReturn != "+3.2%" {
					t.Errorf("MarchReturn = %q", data.USSummary.MarchReturn)
				}
			} else {
				code, _ := decodeError(t, w)
				if code != tc.wantErrCode {
					t.Errorf("error code = %q, want %q", code, tc.wantErrCode)
				}
			}
		})
	}
}

func TestHandleCrypto(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				cryptoResp: models.CryptoResponse{MarketContext: "bearish"},
				cryptoErr:  tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleCrypto, http.MethodGet, "/api/v1/crypto", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.CryptoResponse
				decodeData(t, w, &data)
				if data.MarketContext != "bearish" {
					t.Errorf("MarketContext = %q", data.MarketContext)
				}
			} else {
				code, _ := decodeError(t, w)
				if code != tc.wantErrCode {
					t.Errorf("error code = %q, want %q", code, tc.wantErrCode)
				}
			}
		})
	}
}

func TestHandleRecommendations(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				recsResp: models.RecommendationsResponse{
					AISemis: []models.Recommendation{{Ticker: "NVDA", Name: "NVIDIA"}},
				},
				recsErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleRecommendations, http.MethodGet, "/api/v1/recommendations", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.RecommendationsResponse
				decodeData(t, w, &data)
				if len(data.AISemis) != 1 || data.AISemis[0].Ticker != "NVDA" {
					t.Errorf("unexpected AISemis: %v", data.AISemis)
				}
			}
		})
	}
}

func TestHandlePrincipal(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				principalResp: models.PortfolioResponse{
					Connected: true,
					Summary:   models.PortfolioSummary{CurrentValue: "$150k"},
				},
				principalErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandlePrincipal, http.MethodGet, "/api/v1/portfolio/principal", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.PortfolioResponse
				decodeData(t, w, &data)
				if !data.Connected {
					t.Error("expected Connected = true")
				}
				if data.Summary.CurrentValue != "$150k" {
					t.Errorf("CurrentValue = %q", data.Summary.CurrentValue)
				}
			}
		})
	}
}

func TestHandleMorganStanley(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				morganStanleyResp: models.PortfolioResponse{
					Connected: true,
					Summary:   models.PortfolioSummary{CurrentValue: "$62k"},
				},
				morganStanleyErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleMorganStanley, http.MethodGet, "/api/v1/portfolio/morgan-stanley", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.PortfolioResponse
				decodeData(t, w, &data)
				if data.Summary.CurrentValue != "$62k" {
					t.Errorf("CurrentValue = %q", data.Summary.CurrentValue)
				}
			}
		})
	}
}

func TestHandleFidelity(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				fidelityResp: models.PortfolioResponse{
					Connected: true,
					Summary:   models.PortfolioSummary{CurrentValue: "$43k"},
				},
				fidelityErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleFidelity, http.MethodGet, "/api/v1/portfolio/fidelity", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.PortfolioResponse
				decodeData(t, w, &data)
				if data.Summary.CurrentValue != "$43k" {
					t.Errorf("CurrentValue = %q", data.Summary.CurrentValue)
				}
			}
		})
	}
}

func TestHandleSoFi(t *testing.T) {
	for _, tc := range commonCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStore{
				sofiResp: models.PortfolioResponse{
					Connected: true,
					Summary:   models.PortfolioSummary{CurrentValue: "$8k"},
				},
				sofiErr: tc.storeErr,
			}
			h := newHandler(t, store, nil)
			w := doRequest(h.HandleSoFi, http.MethodGet, "/api/v1/portfolio/sofi", "")

			assertStatus(t, w, tc.wantStatus)
			assertJSON(t, w)

			if tc.storeErr == nil {
				var data models.PortfolioResponse
				decodeData(t, w, &data)
				if data.Summary.CurrentValue != "$8k" {
					t.Errorf("CurrentValue = %q", data.Summary.CurrentValue)
				}
			}
		})
	}
}

// ── Plaid handlers ───────────────────────────────────────────────────────────

func TestHandlePlaidStatus_PlaidNotConfigured(t *testing.T) {
	h := newHandler(t, &mockStore{}, nil) // no plaid
	w := doRequest(h.HandlePlaidStatus, http.MethodGet, "/api/v1/plaid/status", "")

	assertStatus(t, w, http.StatusOK)
	var data struct {
		Enabled       bool `json:"enabled"`
		Principal     bool `json:"principal"`
		Morganstanley bool `json:"morganstanley"`
		Fidelity      bool `json:"fidelity"`
		Sofi          bool `json:"sofi"`
	}
	decodeData(t, w, &data)
	if data.Enabled {
		t.Error("enabled should be false when Plaid is not configured")
	}
	if data.Principal || data.Morganstanley || data.Fidelity || data.Sofi {
		t.Error("slots should be false when Plaid is not configured")
	}
}

func TestHandlePlaidStatus_WithConnectedSlots(t *testing.T) {
	plaid := &mockPlaid{slots: []string{"principal", "fidelity"}}
	h := newHandler(t, &mockStore{}, plaid)
	w := doRequest(h.HandlePlaidStatus, http.MethodGet, "/api/v1/plaid/status", "")

	assertStatus(t, w, http.StatusOK)
	var data struct {
		Enabled       bool `json:"enabled"`
		Principal     bool `json:"principal"`
		Morganstanley bool `json:"morganstanley"`
		Fidelity      bool `json:"fidelity"`
		Sofi          bool `json:"sofi"`
	}
	decodeData(t, w, &data)
	if !data.Enabled {
		t.Error("enabled should be true when Plaid is configured")
	}
	if !data.Principal {
		t.Error("principal should be true")
	}
	if data.Morganstanley {
		t.Error("morganstanley should be false")
	}
	if !data.Fidelity {
		t.Error("fidelity should be true")
	}
	if data.Sofi {
		t.Error("sofi should be false")
	}
}

func TestHandlePlaidLinkToken_NoPlaid(t *testing.T) {
	h := newHandler(t, &mockStore{}, nil)
	w := doRequest(h.HandlePlaidLinkToken, http.MethodPost, "/api/v1/plaid/link-token",
		`{"slot":"principal"}`)
	assertStatus(t, w, http.StatusServiceUnavailable)
}

func TestHandlePlaidLinkToken_MissingSlot(t *testing.T) {
	h := newHandler(t, &mockStore{}, &mockPlaid{linkToken: "link-token-123"})
	w := doRequest(h.HandlePlaidLinkToken, http.MethodPost, "/api/v1/plaid/link-token", `{}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandlePlaidLinkToken_InvalidBody(t *testing.T) {
	h := newHandler(t, &mockStore{}, &mockPlaid{linkToken: "link-token-123"})
	w := doRequest(h.HandlePlaidLinkToken, http.MethodPost, "/api/v1/plaid/link-token", `not json`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandlePlaidLinkToken_Success(t *testing.T) {
	plaid := &mockPlaid{linkToken: "link-sandbox-abc123"}
	h := newHandler(t, &mockStore{}, plaid)
	w := doRequest(h.HandlePlaidLinkToken, http.MethodPost, "/api/v1/plaid/link-token",
		`{"slot":"principal"}`)

	assertStatus(t, w, http.StatusOK)
	var data struct {
		LinkToken string `json:"linkToken"`
	}
	decodeData(t, w, &data)
	if data.LinkToken != "link-sandbox-abc123" {
		t.Errorf("linkToken = %q, want link-sandbox-abc123", data.LinkToken)
	}
}

func TestHandlePlaidLinkToken_PlaidError(t *testing.T) {
	plaid := &mockPlaid{linkTokenErr: errors.New("plaid down")}
	h := newHandler(t, &mockStore{}, plaid)
	w := doRequest(h.HandlePlaidLinkToken, http.MethodPost, "/api/v1/plaid/link-token",
		`{"slot":"principal"}`)
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestHandlePlaidExchange_NoPlaid(t *testing.T) {
	h := newHandler(t, &mockStore{}, nil)
	w := doRequest(h.HandlePlaidExchange, http.MethodPost, "/api/v1/plaid/exchange",
		`{"slot":"principal","publicToken":"pub-token"}`)
	assertStatus(t, w, http.StatusServiceUnavailable)
}

func TestHandlePlaidExchange_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing publicToken", `{"slot":"principal"}`},
		{"missing slot", `{"publicToken":"tok"}`},
		{"invalid json", `not json`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, &mockStore{}, &mockPlaid{})
			w := doRequest(h.HandlePlaidExchange, http.MethodPost, "/api/v1/plaid/exchange", tc.body)
			assertStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestHandlePlaidExchange_Success(t *testing.T) {
	h := newHandler(t, &mockStore{}, &mockPlaid{})
	w := doRequest(h.HandlePlaidExchange, http.MethodPost, "/api/v1/plaid/exchange",
		`{"slot":"principal","publicToken":"public-sandbox-xyz"}`)

	assertStatus(t, w, http.StatusOK)
	var data struct {
		Success bool `json:"success"`
	}
	decodeData(t, w, &data)
	if !data.Success {
		t.Error("expected success = true")
	}
}

func TestHandlePlaidExchange_PlaidError(t *testing.T) {
	plaid := &mockPlaid{exchangeErr: errors.New("exchange failed")}
	h := newHandler(t, &mockStore{}, plaid)
	w := doRequest(h.HandlePlaidExchange, http.MethodPost, "/api/v1/plaid/exchange",
		`{"slot":"principal","publicToken":"public-sandbox-xyz"}`)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── Mock SimulationManager ────────────────────────────────────────────────────

type mockSimulation struct {
	account     models.SimulationAccount
	summaryErr  error
	tradeErr    error
	snapshotErr error
	resetErr    error
}

func (m *mockSimulation) Summary() (models.SimulationAccount, error) {
	return m.account, m.summaryErr
}
func (m *mockSimulation) Trade(action, ticker string, shares, price float64) error {
	return m.tradeErr
}
func (m *mockSimulation) AddSnapshot(totalValue float64) error {
	return m.snapshotErr
}
func (m *mockSimulation) Reset() error {
	return m.resetErr
}

// ── Mock Quoter ───────────────────────────────────────────────────────────────

type mockQuoter struct {
	quote    models.QuoteResponse
	quoteErr error
}

func (m *mockQuoter) FetchQuote(_ string) (models.QuoteResponse, error) {
	return m.quote, m.quoteErr
}

// ── Simulation handlers ───────────────────────────────────────────────────────

func newSimHandler(t *testing.T, sim SimulationManager, quoter Quoter) *Handler {
	t.Helper()
	h, err := New(Config{Store: &mockStore{}, Simulation: sim, Quoter: quoter})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return h
}

func TestHandleSimulation_NoSimStore(t *testing.T) {
	h := newSimHandler(t, nil, nil)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")
	assertStatus(t, w, http.StatusOK)
	var data models.SimulationSummaryResponse
	decodeData(t, w, &data)
	if data.CashBalance != 100_000 {
		t.Errorf("CashBalance = %v, want 100000", data.CashBalance)
	}
}

func TestHandleSimulation_Success(t *testing.T) {
	sim := &mockSimulation{
		account: models.SimulationAccount{
			CashBalance:  90_000,
			StartingCash: 100_000,
			Positions: []models.SimulationPosition{
				{Ticker: "NVDA", Shares: 10, AvgCost: 100},
			},
			Transactions: []models.SimulationTx{},
			Snapshots:    []models.PortfolioSnapshot{},
		},
	}
	quoter := &mockQuoter{quote: models.QuoteResponse{Ticker: "NVDA", Price: 120}}
	h := newSimHandler(t, sim, quoter)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")

	assertStatus(t, w, http.StatusOK)
	var data models.SimulationSummaryResponse
	decodeData(t, w, &data)
	if data.CashBalance != 90_000 {
		t.Errorf("CashBalance = %v, want 90000", data.CashBalance)
	}
	if len(data.Positions) != 1 {
		t.Fatalf("positions len = %d, want 1", len(data.Positions))
	}
	if data.Positions[0].CurrentPrice != 120 {
		t.Errorf("CurrentPrice = %v, want 120", data.Positions[0].CurrentPrice)
	}
	if data.Positions[0].GainDir != "up" {
		t.Errorf("GainDir = %q, want up", data.Positions[0].GainDir)
	}
}

func TestHandleSimulation_SummaryError(t *testing.T) {
	sim := &mockSimulation{summaryErr: errors.New("disk error")}
	h := newSimHandler(t, sim, nil)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestHandleSimulationQuote_MissingTicker(t *testing.T) {
	h := newSimHandler(t, nil, &mockQuoter{})
	w := doRequest(h.HandleSimulationQuote, http.MethodGet, "/api/v1/simulation/quote", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandleSimulationQuote_NoQuoter(t *testing.T) {
	h := newSimHandler(t, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation/quote?ticker=NVDA", nil)
	w := httptest.NewRecorder()
	h.HandleSimulationQuote(w, req)
	assertStatus(t, w, http.StatusServiceUnavailable)
}

func TestHandleSimulationQuote_Success(t *testing.T) {
	quoter := &mockQuoter{quote: models.QuoteResponse{Ticker: "NVDA", Price: 135.50, Dir: "up"}}
	h := newSimHandler(t, nil, quoter)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation/quote?ticker=NVDA", nil)
	w := httptest.NewRecorder()
	h.HandleSimulationQuote(w, req)

	assertStatus(t, w, http.StatusOK)
	var data models.QuoteResponse
	decodeData(t, w, &data)
	if data.Price != 135.50 {
		t.Errorf("Price = %v, want 135.50", data.Price)
	}
}

func TestHandleSimulationTrade_NoSimStore(t *testing.T) {
	h := newSimHandler(t, nil, &mockQuoter{})
	w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade",
		`{"ticker":"NVDA","action":"buy","shares":5}`)
	assertStatus(t, w, http.StatusServiceUnavailable)
}

func TestHandleSimulationTrade_InvalidBody(t *testing.T) {
	h := newSimHandler(t, &mockSimulation{}, &mockQuoter{})
	w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade", `not json`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandleSimulationTrade_BadRequest(t *testing.T) {
	tests := []struct{ name, body string }{
		{"missing ticker", `{"action":"buy","shares":5}`},
		{"zero shares", `{"ticker":"NVDA","action":"buy","shares":0}`},
		{"invalid action", `{"ticker":"NVDA","action":"hold","shares":5}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newSimHandler(t, &mockSimulation{}, &mockQuoter{quote: models.QuoteResponse{Price: 100}})
			w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade", tc.body)
			assertStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestHandleSimulationTrade_Success(t *testing.T) {
	sim := &mockSimulation{
		account: models.SimulationAccount{CashBalance: 95_000, StartingCash: 100_000},
	}
	quoter := &mockQuoter{quote: models.QuoteResponse{Price: 100}}
	h := newSimHandler(t, sim, quoter)
	w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade",
		`{"ticker":"NVDA","action":"buy","shares":5}`)

	assertStatus(t, w, http.StatusOK)
	var data models.TradeResponse
	decodeData(t, w, &data)
	if !data.Success {
		t.Error("expected success = true")
	}
	if data.Price != 100 {
		t.Errorf("Price = %v, want 100", data.Price)
	}
}

func TestHandleSimulationTrade_TradeFailed(t *testing.T) {
	sim := &mockSimulation{tradeErr: errors.New("insufficient funds")}
	quoter := &mockQuoter{quote: models.QuoteResponse{Price: 100}}
	h := newSimHandler(t, sim, quoter)
	w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade",
		`{"ticker":"NVDA","action":"buy","shares":5000}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandleSimulationReset_NoSimStore(t *testing.T) {
	h := newSimHandler(t, nil, nil)
	w := doRequest(h.HandleSimulationReset, http.MethodPost, "/api/v1/simulation/reset", "")
	assertStatus(t, w, http.StatusOK)
}

func TestHandleSimulationReset_Success(t *testing.T) {
	h := newSimHandler(t, &mockSimulation{}, nil)
	w := doRequest(h.HandleSimulationReset, http.MethodPost, "/api/v1/simulation/reset", "")
	assertStatus(t, w, http.StatusOK)
	var data struct {
		Success bool `json:"success"`
	}
	decodeData(t, w, &data)
	if !data.Success {
		t.Error("expected success = true")
	}
}

func TestHandleSimulationReset_Error(t *testing.T) {
	sim := &mockSimulation{resetErr: errors.New("disk full")}
	h := newSimHandler(t, sim, nil)
	w := doRequest(h.HandleSimulationReset, http.MethodPost, "/api/v1/simulation/reset", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── simGainDir ────────────────────────────────────────────────────────────────

func TestSimGainDir(t *testing.T) {
	cases := []struct {
		gain float64
		want string
	}{
		{100, "up"},
		{0.001, "up"},
		{-100, "down"},
		{-0.001, "down"},
		{0, "flat"},
	}
	for _, tc := range cases {
		got := simGainDir(tc.gain)
		if got != tc.want {
			t.Errorf("simGainDir(%v) = %q, want %q", tc.gain, got, tc.want)
		}
	}
}

// ── HandleSimulation – flat gain dir ─────────────────────────────────────────

func TestHandleSimulation_FlatGain(t *testing.T) {
	sim := &mockSimulation{
		account: models.SimulationAccount{
			CashBalance:  100_000,
			StartingCash: 100_000,
			Positions: []models.SimulationPosition{
				// avgCost == currentPrice → gain = 0 → "flat"
				{Ticker: "VTI", Shares: 1, AvgCost: 200},
			},
			Transactions: []models.SimulationTx{},
			Snapshots:    []models.PortfolioSnapshot{},
		},
	}
	quoter := &mockQuoter{quote: models.QuoteResponse{Ticker: "VTI", Price: 200}}
	h := newSimHandler(t, sim, quoter)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")

	assertStatus(t, w, http.StatusOK)
	var data models.SimulationSummaryResponse
	decodeData(t, w, &data)
	if len(data.Positions) != 1 || data.Positions[0].GainDir != "flat" {
		t.Errorf("GainDir = %q, want flat", data.Positions[0].GainDir)
	}
}

// ── HandleSimulation – quoter error falls back to avgCost ────────────────────

func TestHandleSimulation_QuoterErrorFallback(t *testing.T) {
	sim := &mockSimulation{
		account: models.SimulationAccount{
			CashBalance:  90_000,
			StartingCash: 100_000,
			Positions: []models.SimulationPosition{
				{Ticker: "NVDA", Shares: 10, AvgCost: 1000},
			},
			Transactions: []models.SimulationTx{},
			Snapshots:    []models.PortfolioSnapshot{},
		},
	}
	quoter := &mockQuoter{quoteErr: errors.New("api down")}
	h := newSimHandler(t, sim, quoter)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")

	assertStatus(t, w, http.StatusOK)
	var data models.SimulationSummaryResponse
	decodeData(t, w, &data)
	// CurrentPrice should fall back to AvgCost
	if data.Positions[0].CurrentPrice != 1000 {
		t.Errorf("CurrentPrice = %v, want 1000 (fallback to avgCost)", data.Positions[0].CurrentPrice)
	}
}

// ── HandleSimulationQuote – quoter error ──────────────────────────────────────

func TestHandleSimulationQuote_QuoterError(t *testing.T) {
	quoter := &mockQuoter{quoteErr: errors.New("ticker not found")}
	h := newSimHandler(t, nil, quoter)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation/quote?ticker=FAKE", nil)
	w := httptest.NewRecorder()
	h.HandleSimulationQuote(w, req)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── HandleSimulationTrade – quoter error ──────────────────────────────────────

func TestHandleSimulationTrade_QuoterError(t *testing.T) {
	quoter := &mockQuoter{quoteErr: errors.New("market closed")}
	h := newSimHandler(t, &mockSimulation{}, quoter)
	w := doRequest(h.HandleSimulationTrade, http.MethodPost, "/api/v1/simulation/trade",
		`{"ticker":"NVDA","action":"buy","shares":5}`)
	assertStatus(t, w, http.StatusInternalServerError)
}

// ── HandleSimulation – with snapshots ─────────────────────────────────────────

func TestHandleSimulation_WithSnapshots(t *testing.T) {
	sim := &mockSimulation{
		account: models.SimulationAccount{
			CashBalance:  100_000,
			StartingCash: 100_000,
			Positions:    []models.SimulationPosition{},
			Transactions: []models.SimulationTx{},
			Snapshots: []models.PortfolioSnapshot{
				{Date: "2026-01-01", Value: 100_000},
				{Date: "2026-01-15", Value: 105_000},
				{Date: "2026-02-01", Value: 110_000},
			},
		},
	}
	h := newSimHandler(t, sim, nil)
	w := doRequest(h.HandleSimulation, http.MethodGet, "/api/v1/simulation", "")

	assertStatus(t, w, http.StatusOK)
	var data models.SimulationSummaryResponse
	decodeData(t, w, &data)
	if len(data.PerformanceData.Labels) != 3 {
		t.Errorf("perf labels = %d, want 3", len(data.PerformanceData.Labels))
	}
	if data.PerformanceData.Labels[0] != "2026-01-01" {
		t.Errorf("first label = %q, want 2026-01-01", data.PerformanceData.Labels[0])
	}
}
