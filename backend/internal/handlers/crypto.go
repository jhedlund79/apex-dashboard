package handlers

import "net/http"

// HandleCrypto serves GET /api/v1/crypto.
func (h *Handler) HandleCrypto(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.Crypto()
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	h.respond(w, r, http.StatusOK, data)
}
