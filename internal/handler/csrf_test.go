package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	h := CheckOrigin(inner)

	tests := []struct {
		name     string
		method   string
		origin   string
		referer  string
		host     string
		wantCode int
	}{
		{"GET always passes", "GET", "", "", "localhost:8092", 200},
		{"POST no headers passes (curl)", "POST", "", "", "localhost:8092", 200},
		{"POST matching origin passes", "POST", "http://localhost:8092", "", "localhost:8092", 200},
		{"POST mismatched origin blocked", "POST", "http://evil.com", "", "localhost:8092", 403},
		{"POST matching referer passes", "POST", "", "http://localhost:8092/page", "localhost:8092", 200},
		{"POST mismatched referer blocked", "POST", "", "http://evil.com/page", "localhost:8092", 403},
		{"DELETE mismatched origin blocked", "DELETE", "http://evil.com", "", "localhost:8092", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/containers/abc/start", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.wantCode {
				t.Errorf("got %d, want %d", rr.Code, tt.wantCode)
			}
		})
	}
}
