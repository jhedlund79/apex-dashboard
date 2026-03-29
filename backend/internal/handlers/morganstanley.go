package handlers

import "net/http"

// HandleMorganStanley serves GET /api/v1/portfolio/morgan-stanley.
func (h *Handler) HandleMorganStanley(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.MorganStanley()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
