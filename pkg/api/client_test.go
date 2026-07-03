package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"pm-cli/pkg/auth"
)

func TestDoAuthenticatedPersistsRotatedHeadersOnSuccess(t *testing.T) {
	setupBalanceSession(t, "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("access-token", "rotated-token")
		w.Header().Set("client", "rotated-client")
		w.Header().Set("uid", "user@example.com")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &Client{HTTP: http.DefaultClient, BaseURL: srv.URL}
	req, err := client.NewAuthenticatedRequest(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.DoAuthenticated(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	headers, err := auth.GetAuthHeaders()
	if err != nil {
		t.Fatal(err)
	}
	if headers["Access-Token"] != "rotated-token" {
		t.Fatalf("expected rotated access-token to be persisted, got %#v", headers)
	}
	if headers["Client"] != "rotated-client" {
		t.Fatalf("expected rotated client to be persisted, got %#v", headers)
	}
}

func TestDoAuthenticatedPersistsRotatedHeadersAfterRetry(t *testing.T) {
	setupBalanceSession(t, "")

	oldSignInURL := auth.SignInURL
	oldLogin := viper.GetString("login")
	oldPassword := viper.GetString("password")
	t.Cleanup(func() {
		auth.SignInURL = oldSignInURL
		viper.Set("login", oldLogin)
		viper.Set("password", oldPassword)
	})
	viper.Set("login", "user@example.com")
	viper.Set("password", "secret")

	signInSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("access-token", "relogin-token")
		w.Header().Set("client", "relogin-client")
		w.Header().Set("uid", "user@example.com")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "relogin-token",
			"client_id": "relogin-client",
			"data":      map[string]any{"login": "user@example.com"},
		})
	}))
	defer signInSrv.Close()
	auth.SignInURL = signInSrv.URL

	var calls int
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Faça login para continuar.","redirect_to_login":true}`))
			return
		}
		// Simulate the server rotating the token again on the retried request.
		w.Header().Set("access-token", "rotated-after-retry")
		w.Header().Set("client", "relogin-client")
		w.Header().Set("uid", "user@example.com")
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	client := &Client{HTTP: http.DefaultClient, BaseURL: apiSrv.URL}
	req, err := client.NewAuthenticatedRequest(context.Background(), http.MethodGet, apiSrv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.DoAuthenticated(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if calls != 2 {
		t.Fatalf("expected exactly one retry (2 calls), got %d", calls)
	}

	headers, err := auth.GetAuthHeaders()
	if err != nil {
		t.Fatal(err)
	}
	if headers["Access-Token"] != "rotated-after-retry" {
		t.Fatalf("expected the retry response's rotated token to be persisted, got %#v", headers)
	}
}

func TestVerifyAccessSucceedsWhenWorkDayFetchSucceeds(t *testing.T) {
	setupBalanceSession(t, "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"work_day":{"date":"2026-07-03"}}`))
	}))
	defer srv.Close()

	client := &Client{HTTP: http.DefaultClient, BaseURL: srv.URL}
	if err := client.VerifyAccess(context.Background()); err != nil {
		t.Fatalf("expected VerifyAccess to succeed, got %v", err)
	}
}

func TestVerifyAccessFailsOnPersistent403(t *testing.T) {
	setupBalanceSession(t, "")

	oldLogin := viper.GetString("login")
	oldPassword := viper.GetString("password")
	t.Cleanup(func() {
		viper.Set("login", oldLogin)
		viper.Set("password", oldPassword)
	})
	// Clear credentials so the 401/403 auto-refresh has nothing to sign in with,
	// forcing VerifyAccess to surface the original API error.
	viper.Set("login", "")
	viper.Set("password", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"Faça login para continuar.","redirect_to_login":true}`))
	}))
	defer srv.Close()

	client := &Client{HTTP: http.DefaultClient, BaseURL: srv.URL}
	if err := client.VerifyAccess(context.Background()); err == nil {
		t.Fatal("expected VerifyAccess to fail on a persistent 403")
	}
}
