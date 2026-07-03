package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"pm-cli/pkg/auth"
)

const BaseURL = "https://api.pontomais.com.br/api"

type Client struct {
	HTTP        *http.Client
	BaseURL     string
	AuthHeaders func() (map[string]string, error)
}

func New() *Client {
	return &Client{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		BaseURL:     BaseURL,
		AuthHeaders: auth.GetAuthHeaders,
	}
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return BaseURL
}

// DoAuthenticated executes req and retries once with a fresh session on 401/403.
// Callers should use NewAuthenticatedRequest to build req.
func (c *Client) DoAuthenticated(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
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
	return c.HTTP.Do(retryReq)
}

type authError struct {
	status int
	body   string
}

func (e *authError) Error() string {
	return "HTTP " + http.StatusText(e.status) + ": " + e.body
}

func (c *Client) NewAuthenticatedRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	getHeaders := c.AuthHeaders
	if getHeaders == nil {
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
