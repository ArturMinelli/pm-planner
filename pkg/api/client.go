package api

import (
	"context"
	"net/http"
	"time"

	"pm-cli/pkg/auth"
)

const BaseURL = "https://api.pontomais.com.br/api"

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) NewAuthenticatedRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	headers, err := auth.GetAuthHeaders()
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}
