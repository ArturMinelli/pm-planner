package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"pm-cli/pkg/auth"
)

const BaseURL = "https://api.pontomais.com.br/api"

// fakeEnv, when set, points the client at a localhost stand-in (cmd/pm-fake-pontomais)
// and disables sign-in entirely.
const fakeEnv = "PM_FAKE_PONTOMAIS_URL"

// refusedURL is a dead address used whenever fake mode must not reach a server.
const refusedURL = "http://127.0.0.1:1/invalid"

var errFakeRefused = errors.New("fake mode: refusing non-localhost URL")

// FakeMode reports whether PM_FAKE_PONTOMAIS_URL is set. A set but empty or non-localhost
// value fails closed rather than falling back to production.
func FakeMode() bool {
	_, ok := os.LookupEnv(fakeEnv)
	return ok
}

type Client struct {
	HTTP        *http.Client
	BaseURL     string
	AuthHeaders func() (map[string]string, error)
}

func New() *Client {
	if FakeMode() {
		return newFake()
	}
	return &Client{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		BaseURL:     BaseURL,
		AuthHeaders: auth.GetAuthHeaders,
	}
}

func newFake() *Client {
	c := &Client{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		BaseURL:     refusedURL,
		AuthHeaders: func() (map[string]string, error) { return nil, errFakeRefused },
	}
	raw := os.Getenv(fakeEnv)
	u, err := url.Parse(raw)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		return c
	}
	c.BaseURL = raw
	c.AuthHeaders = func() (map[string]string, error) {
		return map[string]string{
			"access-token": "fake",
			"client":       "fake",
			"uid":          "test@example.invalid",
			"token":        "fake",
		}, nil
	}
	return c
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	if FakeMode() {
		return refusedURL
	}
	return BaseURL
}

// DoAuthenticated executes req and retries once with a fresh session on 401/403.
// Callers should use NewAuthenticatedRequest to build req.
//
// PontoMais rotates the auth token (and sometimes uid/client) on every authenticated
// request and rejects the previous one outside a short batch-request grace window, so
// every response that carries rotation headers is persisted back to session.json —
// otherwise the cached session would stop working after this very call.
func (c *Client) DoAuthenticated(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if FakeMode() {
		// Fake mode has no session cache and must never re-sign-in against production.
		return resp, nil
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		auth.ApplyRotatedHeaders(resp.Header)
		return resp, nil
	}
	// Stale/invalid session — discard cached token, re-login, and retry once.
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if _, refreshErr := auth.RefreshAuth(); refreshErr != nil {
		// Can't refresh; surface the original auth error.
		return nil, &authError{status: resp.StatusCode, body: string(body)}
	}
	// Rebuild the request with fresh auth headers (original req is already spent).
	retryReq, err := c.NewAuthenticatedRequest(req.Context(), req.Method, req.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	// Carry over the caller's body and non-auth headers (e.g. api-version, uuid).
	if req.GetBody != nil {
		if retryReq.Body, err = req.GetBody(); err != nil {
			return nil, err
		}
		retryReq.ContentLength = req.ContentLength
	}
	for k, v := range req.Header {
		if retryReq.Header.Get(k) == "" {
			retryReq.Header[k] = v
		}
	}
	retryResp, err := c.HTTP.Do(retryReq)
	if err == nil {
		auth.ApplyRotatedHeaders(retryResp.Header)
	}
	return retryResp, err
}

type authError struct {
	status int
	body   string
}

func (e *authError) Error() string {
	return "HTTP " + http.StatusText(e.status) + ": " + e.body
}

func (c *Client) NewAuthenticatedRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	getHeaders := c.AuthHeaders
	if getHeaders == nil {
		if FakeMode() {
			return nil, errFakeRefused
		}
		getHeaders = auth.GetAuthHeaders
	}
	headers, err := getHeaders()
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// VerifyAccess performs a lightweight authenticated request to confirm the current
// session is actually accepted by the API — not just that sign-in succeeded. A
// successful sign-in alone does not guarantee the resulting session headers are valid
// for subsequent calls (see DoAuthenticated).
func (c *Client) VerifyAccess(ctx context.Context) error {
	_, err := c.FetchWorkDay(ctx, time.Now().Format("2006-01-02"))
	return err
}
