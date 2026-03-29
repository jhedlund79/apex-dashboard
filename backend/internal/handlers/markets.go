package handlers

import "net/http"

// HandleMarkets serves GET /api/v1/markets.
func (h *Handler) HandleMarkets(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Markets()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
