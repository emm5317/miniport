package handler

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Credentials maps usernames to SHA-256 hex password hashes.
// File format: one "username:sha256hex" per line. Generate hashes with:
//
//	echo -n 'password' | sha256sum | cut -d' ' -f1
type Credentials map[string]string

// LoadHtpasswd reads an auth file (username:sha256hex per line).
func LoadHtpasswd(path string) (Credentials, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open auth file: %w", err)
	}
	defer f.Close()

	creds := make(Credentials)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		user, hash, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		creds[user] = hash
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read auth file: %w", err)
	}
	if len(creds) == 0 {
		return nil, fmt.Errorf("auth file contains no entries")
	}
	return creds, nil
}

// checkPassword verifies a plaintext password against a stored SHA-256 hex hash.
func (c Credentials) checkPassword(user, password string) bool {
	hash, ok := c[user]
	if !ok {
		// Constant-time operation even for unknown users.
		subtle.ConstantTimeCompare([]byte(sha256Hex(password)), []byte(strings.Repeat("0", 64)))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sha256Hex(password)), []byte(hash)) == 1
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// BasicAuth wraps a handler with HTTP Basic Authentication.
// The healthz endpoint is excluded so monitoring probes work without credentials.
func BasicAuth(creds Credentials, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || !creds.checkPassword(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="MiniPort"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
