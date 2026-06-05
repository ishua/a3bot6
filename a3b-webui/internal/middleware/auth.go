package middleware

import (
	"net/http"
)

// SessionValidator validates session cookies.
type SessionValidator interface {
	ValidateCookie(r *http.Request) (string, error)
}

// AuthMiddleware checks session cookie and redirects to /login if invalid.
func AuthMiddleware(next http.Handler, sv SessionValidator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login/logout paths
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		_, err := sv.ValidateCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
