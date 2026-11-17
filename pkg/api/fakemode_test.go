package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pm-cli/pkg/auth"
)

const fakeEnvName = "PM_FAKE_PONTOMAIS_URL"

func unsetFakeEnv(t *testing.T) {
	t.Helper()
	t.Setenv(fakeEnvName, "x") // registers restore on cleanup
	os.Unsetenv(fakeEnvName)
}

func TestFakeModeOffByDefault(t *testing.T) {
	unsetFakeEnv(t)
	if FakeMode() {
		t.Fatal("FakeMode() = true with env unset")
	}
	c := New()
	if c.BaseURL != BaseURL {
		t.Fatalf("BaseURL = %q, want production %q", c.BaseURL, BaseURL)
	}
}

func TestFakeModeLocalhostUsesStaticHeaders(t *testing.T) {
	for _, u := range []string{"http://127.0.0.1:3850/api", "http://localhost:3850/api"} {
		t.Run(u, func(t *testing.T) {
			t.Setenv(fakeEnvName, u)
			if !FakeMode() {
				t.Fatal("FakeMode() = false with env set")
			}
			c := New()
			if c.BaseURL != u {
				t.Fatalf("BaseURL = %q, want %q", c.BaseURL, u)
			}
			h, err := c.AuthHeaders()
			if err != nil {
				t.Fatal(err)
			}
			if h["access-token"] != "fake" || h["client"] != "fake" || h["uid"] != "test@example.invalid" || h["token"] != "fake" {
				t.Fatalf("headers = %#v", h)
			}
		})
	}
}

func TestFakeModeRefusesNonLocalhost(t *testing.T) {
	for _, u := range []string{
		"https://api.pontomais.com.br/api",
		"http://localhost.pontomais.com.br/api",
		"http://10.0.0.5:3850/api",
		"127.0.0.1:3850/api",
		"",
	} {
		t.Run(u, func(t *testing.T) {
			t.Setenv(fakeEnvName, u)
			c := New()
			if c.BaseURL != refusedURL {
				t.Fatalf("BaseURL = %q, want refused %q", c.BaseURL, refusedURL)
			}
			if _, err := c.AuthHeaders(); err == nil || !strings.Contains(err.Error(), "refusing non-localhost URL") {
				t.Fatalf("AuthHeaders err = %v, want refusal", err)
			}
			if _, err := c.NewAuthenticatedRequest(context.Background(), http.MethodGet, c.baseURL()+"/x", nil); err == nil {
				t.Fatal("NewAuthenticatedRequest succeeded; want refusal")
			}
		})
	}
}

func TestFakeModeNeverSignsIn(t *testing.T) {
	// Any sign-in hit fails the test; a 401 must surface without re-login.
	signIn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("sign-in reached in fake mode: %s", r.URL)
	}))
	defer signIn.Close()
	oldSignIn := auth.SignInURL
	auth.SignInURL = signIn.URL
	t.Cleanup(func() { auth.SignInURL = oldSignIn })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	t.Setenv(fakeEnvName, srv.URL)

	c := New()
	req, err := c.NewAuthenticatedRequest(context.Background(), http.MethodGet, srv.URL+"/session", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.DoAuthenticated(req)
	if err != nil {
		t.Fatalf("DoAuthenticated err = %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 passed through", resp.StatusCode)
	}
}
