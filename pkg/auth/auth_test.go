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

	_, readErr := readCachedSession()
	if readErr == nil {
		t.Fatal("cache should be cleared after failed verify")
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

	sessionPath, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	pmDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(pmDir, 0700); err != nil {
		t.Fatal(err)
	}
	stale := session{
		Token:    "stale-token",
		Uid:      "old@example.com",
		Client:   "old-client",
		CachedAt: time.Now(),
	}
	b, _ := json.Marshal(stale)
	if err := os.WriteFile(sessionPath, b, 0600); err != nil {
		t.Fatal(err)
	}

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

func TestCacheEmployeeIDStoresOnValidSession(t *testing.T) {
	setupAuthTestHome(t)
	valid := session{
		Token:    "cached-token",
		Uid:      "user@example.com",
		Client:   "client-id",
		CachedAt: time.Now(),
	}
	if err := writeCachedSession(&valid); err != nil {
		t.Fatal(err)
	}
	if err := CacheEmployeeID("2751858"); err != nil {
		t.Fatal(err)
	}
	if got := GetCachedEmployeeID(); got != "2751858" {
		t.Fatalf("employee ID: got %q", got)
	}
}

func TestVerifyCredentialsClearsEmployeeIDForNewSession(t *testing.T) {
	setupAuthTestHome(t)
	old := session{
		Token:      "cached-token",
		Uid:        "old@example.com",
		Client:     "client-id",
		EmployeeID: "old-employee",
		CachedAt:   time.Now(),
	}
	if err := writeCachedSession(&old); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "new-token",
			"client_id": "client-2",
			"data":      map[string]any{"login": "new@example.com"},
		})
	}))
	defer srv.Close()
	oldURL := signInURL
	signInURL = srv.URL
	t.Cleanup(func() { signInURL = oldURL })

	if err := VerifyCredentials("new@example.com", "secret"); err != nil {
		t.Fatal(err)
	}
	if got := GetCachedEmployeeID(); got != "" {
		t.Fatalf("new session should not retain employee ID, got %q", got)
	}
}

func setupAuthTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
}
