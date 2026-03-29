package handlers

import "net/http"

// HandleRecommendations serves GET /api/v1/recommendations.
func (h *Handler) HandleRecommendations(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Recommendations()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
