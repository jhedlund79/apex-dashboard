package handlers

import "net/http"

// HandleSoFi serves GET /api/v1/portfolio/sofi.
func (h *Handler) HandleSoFi(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.SoFi()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
