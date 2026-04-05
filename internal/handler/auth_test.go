package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	creds := Credentials{"admin": sha256Hex("secret")}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	h := BasicAuth(creds, inner)

	t.Run("no credentials returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 401 {
			t.Fatalf("got %d, want 401", rr.Code)
		}
		if rr.Header().Get("WWW-Authenticate") == "" {
			t.Fatal("missing WWW-Authenticate header")
		}
	})

	t.Run("wrong password returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("admin", "wrong")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 401 {
			t.Fatalf("got %d, want 401", rr.Code)
		}
	})

	t.Run("correct credentials returns 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("admin", "secret")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("got %d, want 200", rr.Code)
		}
	})

	t.Run("unknown user returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("nobody", "secret")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 401 {
			t.Fatalf("got %d, want 401", rr.Code)
		}
	})

	t.Run("healthz skips auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("got %d, want 200", rr.Code)
		}
	})
}

func TestLoadHtpasswd(t *testing.T) {
	hash := sha256Hex("pass")
	content := "admin:" + hash + "\n# comment\n\n"

	f, err := os.CreateTemp(t.TempDir(), "htpasswd")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()

	creds, err := LoadHtpasswd(f.Name())
	if err != nil {
		t.Fatalf("LoadHtpasswd: %v", err)
	}
	if !creds.checkPassword("admin", "pass") {
		t.Fatal("expected valid password check")
	}
	if creds.checkPassword("admin", "wrong") {
		t.Fatal("expected invalid password check")
	}
}

func TestLoadHtpasswd_empty(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "htpasswd")
	f.WriteString("# only comments\n")
	f.Close()

	_, err := LoadHtpasswd(f.Name())
	if err == nil {
		t.Fatal("expected error for empty auth file")
	}
}
