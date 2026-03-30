package handlers

import (
	"encoding/json"
	"net/http"
)

// HandlePlaidStatus serves GET /api/v1/plaid/status.
// Returns which portfolio slots have a connected Plaid account.
func (h *Handler) HandlePlaidStatus(w http.ResponseWriter, r *http.Request) {
	if h.plaid == nil {
		h.respond(w, r, http.StatusOK, map[string]any{
			"enabled":       false,
			"principal":     false,
			"morganstanley": false,
			"fidelity":      false,
			"sofi":          false,
		})
		return
	}

	connected := map[string]bool{}
	for _, slot := range h.plaid.ConnectedSlots() {
		connected[slot] = true
	}
	h.respond(w, r, http.StatusOK, map[string]any{
		"enabled":       true,
		"principal":     connected["principal"],
		"morganstanley": connected["morganstanley"],
		"fidelity":      connected["fidelity"],
		"sofi":          connected["sofi"],
	})
}

// HandlePlaidLinkToken serves POST /api/v1/plaid/link-token.
// Body: {"slot": "principal" | "morganstanley"}
// Returns a Plaid Link token for the frontend to open Plaid Link.
func (h *Handler) HandlePlaidLinkToken(w http.ResponseWriter, r *http.Request) {
	if h.plaid == nil {
		http.Error(w, `{"error":{"code":"NOT_CONFIGURED","message":"Plaid not configured"}}`, http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Slot string `json:"slot"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Slot == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"slot required"}}`, http.StatusBadRequest)
		return
	}

	token, err := h.plaid.CreateLinkToken(r.Context(), body.Slot)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "plaid link token", "err", err)
		http.Error(w, `{"error":{"code":"PLAID_ERROR","message":"failed to create link token"}}`, http.StatusInternalServerError)
		return
	}

	h.respond(w, r, http.StatusOK, map[string]string{"linkToken": token})
}

// HandlePlaidExchange serves POST /api/v1/plaid/exchange.
// Body: {"slot": "principal" | "morganstanley", "publicToken": "<token>"}
// Exchanges the Plaid public token for a persistent access token.
func (h *Handler) HandlePlaidExchange(w http.ResponseWriter, r *http.Request) {
	if h.plaid == nil {
		http.Error(w, `{"error":{"code":"NOT_CONFIGURED","message":"Plaid not configured"}}`, http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Slot        string `json:"slot"`
		PublicToken string `json:"publicToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Slot == "" || body.PublicToken == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"slot and publicToken required"}}`, http.StatusBadRequest)
		return
	}

	if err := h.plaid.ExchangeToken(r.Context(), body.Slot, body.PublicToken); err != nil {
		h.logger.ErrorContext(r.Context(), "plaid exchange", "err", err)
		http.Error(w, `{"error":{"code":"PLAID_ERROR","message":"failed to exchange token"}}`, http.StatusInternalServerError)
		return
	}

	h.respond(w, r, http.StatusOK, map[string]bool{"success": true})
}
