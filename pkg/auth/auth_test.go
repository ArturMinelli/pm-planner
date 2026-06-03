package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifyCredentials_rejectsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	oldURL := signInURL
	signInURL = srv.URL
	t.Cleanup(func() { signInURL = oldURL })

	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", dir)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfigHome)
	})

	sessionPath, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	pmDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(pmDir, 0700); err != nil {
		t.Fatal(err)
	}
	valid := session{
		Token:    "cached-token",
		Uid:      "user@example.com",
		Client:   "client-id",
		CachedAt: time.Now(),
	}
	b, _ := json.Marshal(valid)
	if err := os.WriteFile(sessionPath, b, 0600); err != nil {
		t.Fatal(err)
	}

	err = VerifyCredentials("user@example.com", "wrong")
	if err == nil {
		t.Fatal("expected login error for wrong password")
	}

	got, readErr := readCachedSession()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got.Token != "cached-token" {
		t.Fatalf("cache should be unchanged, got token %q", got.Token)
	}
}

func TestVerifyCredentials_acceptsAndCaches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "new-token",
			"client_id": "client-1",
			"data":      map[string]any{"login": "user@example.com"},
		})
	}))
	defer srv.Close()

	oldURL := signInURL
	signInURL = srv.URL
	t.Cleanup(func() { signInURL = oldURL })

	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", dir)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfigHome)
	})

	if err := VerifyCredentials("user@example.com", "secret"); err != nil {
		t.Fatal(err)
	}

	got, err := readCachedSession()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "new-token" {
		t.Fatalf("token: got %q", got.Token)
	}
}
