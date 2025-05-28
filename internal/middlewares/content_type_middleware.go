package middlewares

import (
	"net/http"
)

// ContentTypeMiddleware sets the Content-Type header to application/json for all responses
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
