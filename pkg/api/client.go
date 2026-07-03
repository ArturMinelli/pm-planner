package api

import (
	"context"
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
