package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"pm-cli/pkg/auth"
)

type myTimeBreakResponse struct {
	Employee struct {
		ID stringID `json:"id"`
	} `json:"employee"`
}

func (c *Client) fetchMyTimeBreak(ctx context.Context) (json.RawMessage, string, error) {
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodGet, c.baseURL()+"/employees/my_time_break", nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", statusError(resp)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	var env struct {
		Employee json.RawMessage `json:"employee"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, "", err
	}
	if len(env.Employee) == 0 {
		return nil, "", fmt.Errorf("employee missing from PontoMais response")
	}
	var parsed myTimeBreakResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, "", err
	}
	employeeID := string(parsed.Employee.ID)
	if employeeID == "" {
		return nil, "", fmt.Errorf("employee ID missing from PontoMais response")
	}
	return env.Employee, employeeID, nil
}

func (c *Client) discoverEmployeeID(ctx context.Context) (string, error) {
	_, employeeID, err := c.fetchMyTimeBreak(ctx)
	if err != nil {
		return "", err
	}
	if err := auth.CacheEmployeeID(employeeID); err != nil {
		return "", err
	}
	return employeeID, nil
}
