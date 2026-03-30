package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"apex-dashboard/backend/internal/models"
)

// HandleSimulation serves GET /api/v1/simulation.
// Returns a portfolio summary enriched with live prices for each open position.
func (h *Handler) HandleSimulation(w http.ResponseWriter, r *http.Request) {
	if h.sim == nil {
		h.respond(w, r, http.StatusOK, models.SimulationSummaryResponse{
			CashBalance:     100_000,
			StartingCash:    100_000,
			Positions:       []models.SimulationPositionResponse{},
			Transactions:    []models.SimulationTx{},
			PerformanceData: models.ChartData{Labels: []string{}, Data: []float64{}},
		})
		return
	}

	account, err := h.sim.Summary()
	if err != nil {
		h.storeError(w, r, err)
		return
	}

	// Enrich positions with live prices.
	positions := make([]models.SimulationPositionResponse, 0, len(account.Positions))
	portfolioValue := 0.0
	for _, p := range account.Positions {
		pr := models.SimulationPositionResponse{
			Ticker:  p.Ticker,
			Shares:  p.Shares,
			AvgCost: p.AvgCost,
		}
		if h.quoter != nil {
			if q, qErr := h.quoter.FetchQuote(p.Ticker); qErr == nil {
				pr.CurrentPrice = q.Price
			}
		}
		if pr.CurrentPrice == 0 {
			pr.CurrentPrice = p.AvgCost // fallback to cost basis
		}
		pr.CurrentValue = pr.Shares * pr.CurrentPrice
		cost := pr.Shares * pr.AvgCost
		pr.Gain = pr.CurrentValue - cost
		if cost > 0 {
			pr.GainPct = (pr.Gain / cost) * 100
		}
		pr.GainDir = simGainDir(pr.Gain)
		portfolioValue += pr.CurrentValue
		positions = append(positions, pr)
	}

	totalValue := account.CashBalance + portfolioValue
	totalGain := totalValue - account.StartingCash
	var totalGainPct float64
	if account.StartingCash > 0 {
		totalGainPct = (totalGain / account.StartingCash) * 100
	}

	// Take a snapshot of today's total value (best-effort).
	if h.sim != nil {
		_ = h.sim.AddSnapshot(totalValue)
	}

	// Build performance chart from snapshots (capped at 90 days).
	snaps := account.Snapshots
	if len(snaps) > 90 {
		snaps = snaps[len(snaps)-90:]
	}
	labels := make([]string, len(snaps))
	values := make([]float64, len(snaps))
	for i, s := range snaps {
		labels[i] = s.Date
		values[i] = s.Value
	}

	resp := models.SimulationSummaryResponse{
		CashBalance:    account.CashBalance,
		StartingCash:   account.StartingCash,
		PortfolioValue: portfolioValue,
		TotalValue:     totalValue,
		TotalGain:      totalGain,
		TotalGainPct:   totalGainPct,
		Positions:      positions,
		Transactions:   account.Transactions,
		PerformanceData: models.ChartData{
			Labels: labels,
			Data:   values,
		},
	}
	h.respond(w, r, http.StatusOK, resp)
}

// HandleSimulationQuote serves GET /api/v1/simulation/quote?ticker=NVDA.
func (h *Handler) HandleSimulationQuote(w http.ResponseWriter, r *http.Request) {
	ticker := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ticker")))
	if ticker == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"BAD_REQUEST","message":"ticker is required"}}`))
		return
	}
	if h.quoter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"SERVICE_UNAVAILABLE","message":"live data not available"}}`))
		return
	}
	q, err := h.quoter.FetchQuote(ticker)
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, q)
}

// HandleSimulationTrade serves POST /api/v1/simulation/trade.
// Body: {"ticker":"NVDA","action":"buy","shares":10}
func (h *Handler) HandleSimulationTrade(w http.ResponseWriter, r *http.Request) {
	if h.sim == nil || h.quoter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"SERVICE_UNAVAILABLE","message":"simulation not available"}}`))
		return
	}

	var req models.TradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"BAD_REQUEST","message":"invalid request body"}}`))
		return
	}
	req.Ticker = strings.ToUpper(strings.TrimSpace(req.Ticker))
	if req.Ticker == "" || req.Shares <= 0 || (req.Action != "buy" && req.Action != "sell") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"BAD_REQUEST","message":"ticker, shares (>0), and action (buy|sell) are required"}}`))
		return
	}

	q, err := h.quoter.FetchQuote(req.Ticker)
	if err != nil {
		h.storeError(w, r, err)
		return
	}

	if err := h.sim.Trade(req.Action, req.Ticker, req.Shares, q.Price); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "TRADE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	account, _ := h.sim.Summary()
	h.respond(w, r, http.StatusOK, models.TradeResponse{
		Success:     true,
		Message:     req.Action + " executed",
		Price:       q.Price,
		Total:       req.Shares * q.Price,
		CashBalance: account.CashBalance,
	})
}

// HandleSimulationReset serves POST /api/v1/simulation/reset.
// Wipes all positions and transactions, restoring the starting cash.
func (h *Handler) HandleSimulationReset(w http.ResponseWriter, r *http.Request) {
	if h.sim == nil {
		h.respond(w, r, http.StatusOK, map[string]bool{"success": true})
		return
	}
	if err := h.sim.Reset(); err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, map[string]bool{"success": true})
}

func simGainDir(gain float64) string {
	if gain > 0 {
		return "up"
	}
	if gain < 0 {
		return "down"
	}
	return "flat"
}
