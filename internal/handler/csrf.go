package handler

import (
	"net/http"
	"net/url"
	"strings"
)

// CheckOrigin rejects state-changing requests (POST, PUT, DELETE, PATCH)
// whose Origin or Referer header doesn't match the request's Host.
// Requests with no Origin/Referer are allowed (non-browser clients like curl).
func CheckOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		// No Origin or Referer — allow (non-browser client).
		if origin == "" && referer == "" {
			next.ServeHTTP(w, r)
			return
		}

		host := r.Host
		if host == "" {
			host = r.URL.Host
		}

		if origin != "" {
			if !hostMatches(origin, host) {
				http.Error(w, "Forbidden: origin mismatch", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Fall back to Referer.
		if !hostMatches(referer, host) {
			http.Error(w, "Forbidden: referer mismatch", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// hostMatches parses a URL string and checks if its host matches the expected host.
func hostMatches(rawURL, expectedHost string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, expectedHost)
}
