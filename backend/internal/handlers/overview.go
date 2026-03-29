package handlers

import "net/http"

// HandleOverview serves GET /api/v1/overview.
func (h *Handler) HandleOverview(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Overview()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
