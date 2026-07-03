package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestVerifyCredentials_rejectsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	oldURL := SignInURL
	SignInURL = srv.URL
	t.Cleanup(func() { SignInURL = oldURL })

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

	oldURL := SignInURL
	SignInURL = srv.URL
	t.Cleanup(func() { SignInURL = oldURL })

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
	oldURL := SignInURL
	SignInURL = srv.URL
	t.Cleanup(func() { SignInURL = oldURL })

	if err := VerifyCredentials("new@example.com", "secret"); err != nil {
		t.Fatal(err)
	}
	if got := GetCachedEmployeeID(); got != "" {
		t.Fatalf("new session should not retain employee ID, got %q", got)
	}
}

func TestVerifyCredentialsPrefersHeaderAuthOverBody(t *testing.T) {
	setupAuthTestHome(t)

	expiry := time.Now().Add(time.Hour).Unix()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("access-token", "header-token")
		w.Header().Set("client", "header-client")
		w.Header().Set("uid", "header-uid@example.com")
		w.Header().Set("expiry", strconv.FormatInt(expiry, 10))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "body-token",
			"client_id": "body-client",
			"data":      map[string]any{"login": "body-uid@example.com"},
		})
	}))
	defer srv.Close()

	oldURL := SignInURL
	SignInURL = srv.URL
	t.Cleanup(func() { SignInURL = oldURL })

	if err := VerifyCredentials("user@example.com", "secret"); err != nil {
		t.Fatal(err)
	}

	got, err := readCachedSession()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "header-token" || got.AccessToken != "header-token" {
		t.Fatalf("expected header access-token to win over body token, got %#v", got)
	}
	if got.Client != "header-client" {
		t.Fatalf("expected header client to win over body client_id, got %#v", got)
	}
	if got.Uid != "header-uid@example.com" {
		t.Fatalf("expected header uid to win over body data.login, got %#v", got)
	}
	if got.Expiry != expiry {
		t.Fatalf("expected expiry header to be captured, got %d", got.Expiry)
	}
}

func TestVerifyCredentialsFallsBackToBodyWhenHeadersMissing(t *testing.T) {
	setupAuthTestHome(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "body-token",
			"client_id": "body-client",
			"data":      map[string]any{"login": "body-uid@example.com"},
		})
	}))
	defer srv.Close()

	oldURL := SignInURL
	SignInURL = srv.URL
	t.Cleanup(func() { SignInURL = oldURL })

	if err := VerifyCredentials("user@example.com", "secret"); err != nil {
		t.Fatal(err)
	}

	got, err := readCachedSession()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "body-token" || got.Client != "body-client" || got.Uid != "body-uid@example.com" {
		t.Fatalf("expected body fields to be used as fallback, got %#v", got)
	}
}

func TestApplyRotatedHeadersUpdatesSession(t *testing.T) {
	setupAuthTestHome(t)
	original := session{
		Token:    "old-token",
		Uid:      "user@example.com",
		Client:   "old-client",
		CachedAt: time.Now().Add(-time.Hour),
	}
	if err := writeCachedSession(&original); err != nil {
		t.Fatal(err)
	}

	expiry := time.Now().Add(2 * time.Hour).Unix()
	header := http.Header{}
	header.Set("access-token", "new-token")
	header.Set("client", "new-client")
	header.Set("uid", "user@example.com")
	header.Set("expiry", strconv.FormatInt(expiry, 10))

	ApplyRotatedHeaders(header)

	got, err := readCachedSession()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "new-token" || got.AccessToken != "new-token" {
		t.Fatalf("expected token to be rotated, got %#v", got)
	}
	if got.Client != "new-client" {
		t.Fatalf("expected client to be rotated, got %#v", got)
	}
	if got.Expiry != expiry {
		t.Fatalf("expected expiry to be persisted, got %d", got.Expiry)
	}
}

func TestApplyRotatedHeadersNoopWithoutAccessToken(t *testing.T) {
	setupAuthTestHome(t)
	original := session{
		Token:    "old-token",
		Uid:      "user@example.com",
		Client:   "old-client",
		CachedAt: time.Now(),
	}
	if err := writeCachedSession(&original); err != nil {
		t.Fatal(err)
	}

	ApplyRotatedHeaders(http.Header{})

	got, err := readCachedSession()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "old-token" || got.Client != "old-client" {
		t.Fatalf("session should be left untouched, got %#v", got)
	}
}

func TestApplyRotatedHeadersNoopWithoutCachedSession(t *testing.T) {
	setupAuthTestHome(t)

	header := http.Header{}
	header.Set("access-token", "new-token")
	ApplyRotatedHeaders(header)

	if _, err := readCachedSession(); err == nil {
		t.Fatal("expected no session file to be created out of thin air")
	}
}

func setupAuthTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
}
