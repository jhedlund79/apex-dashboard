package handlers

import "net/http"

// HandlePrincipal serves GET /api/v1/portfolio/principal.
func (h *Handler) HandlePrincipal(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Principal()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
