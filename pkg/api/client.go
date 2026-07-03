package api

import (
	"bytes"
	"context"
	"encoding/json"
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

func (c *Client) NewAuthenticatedRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// DoAuthenticated executes an authenticated request and retries once after refreshing auth on 401/403.
func (c *Client) DoAuthenticated(ctx context.Context, method, url string, body any) (*http.Response, error) {
	return c.doAuthenticated(ctx, method, url, body, nil)
}

// DoAuthenticatedWithHeaders is like DoAuthenticated but sets additional request headers.
func (c *Client) DoAuthenticatedWithHeaders(ctx context.Context, method, url string, body any, extra map[string]string) (*http.Response, error) {
	return c.doAuthenticated(ctx, method, url, body, extra)
}

func (c *Client) doAuthenticated(ctx context.Context, method, url string, body any, extra map[string]string) (*http.Response, error) {
	resp, err := c.doOnce(ctx, method, url, body, extra)
	if err != nil {
		return nil, err
	}
	if !isAuthStatus(resp.StatusCode) {
		return resp, nil
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if _, err := auth.RefreshAuth(); err != nil {
		return nil, &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(bodyBytes)}
	}
	resp, err = c.doOnce(ctx, method, url, body, extra)
	if err != nil {
		return nil, err
	}
	if isAuthStatus(resp.StatusCode) {
		bodyBytes, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(bodyBytes)}
	}
	return resp, nil
}

func (c *Client) doOnce(ctx context.Context, method, url string, body any, extra map[string]string) (*http.Response, error) {
	req, err := c.NewAuthenticatedRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	return c.HTTP.Do(req)
}

func isAuthStatus(code int) bool {
	return code == http.StatusUnauthorized || code == http.StatusForbidden
}
