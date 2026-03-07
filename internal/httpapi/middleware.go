package httpapi

import (
	"net/http"
	"strings"
)

func (h *Handler) withAuth(next http.Handler) http.Handler {
	// If API_KEY is not set, keep the API open (convenient for localhost dev).
	// Operators should set API_KEY when exposing this service.
	if h.apiKey == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Support either X-API-Key: <key> or Authorization: Bearer <key>
		key := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if key == "" {
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				key = strings.TrimSpace(auth[7:])
			}
		}

		if key != h.apiKey {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
