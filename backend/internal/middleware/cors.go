// Package middleware provides reusable HTTP middleware for the APEX server.
package middleware

import "net/http"

// CORS wraps next with permissive CORS headers suitable for development.
// All origins are allowed. OPTIONS preflight requests are answered with
// 204 and no body.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
