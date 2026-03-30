package handlers

import "net/http"

// HandleFidelity serves GET /api/v1/portfolio/fidelity.
func (h *Handler) HandleFidelity(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Fidelity()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
