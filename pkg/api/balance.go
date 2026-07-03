package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"pm-cli/pkg/auth"
)

// HTTPStatusError exposes response status codes that callers can recover from.
type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// EmployeeBalance is the current signed time-bank balance returned by PontoMais.
type EmployeeBalance struct {
	EmployeeID      string `json:"employeeId"`
	TimeBalanceSecs int64  `json:"timeBalanceSecs"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

type stringID string

func (v *stringID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*v = stringID(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*v = stringID(n.String())
	return nil
}

type signedSeconds int64

func (v *signedSeconds) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = 0
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		return v.setWholeSeconds(n.String())
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return v.setWholeSeconds(s)
}

func (v *signedSeconds) setWholeSeconds(value string) error {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		*v = signedSeconds(seconds)
		return nil
	}
	rational, ok := new(big.Rat).SetString(value)
	if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
		return fmt.Errorf("invalid whole-second balance %q", value)
	}
	seconds := rational.Num().Int64()
	*v = signedSeconds(seconds)
	return nil
}

func (c *Client) discoverEmployeeID(ctx context.Context) (string, error) {
	url := c.baseURL() + "/employees/my_time_break"
	resp, err := c.DoAuthenticated(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", statusError(resp)
	}
	var env struct {
		Employee struct {
			ID stringID `json:"id"`
		} `json:"employee"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return "", err
	}
	employeeID := string(env.Employee.ID)
	if employeeID == "" {
		return "", fmt.Errorf("employee ID missing from PontoMais response")
	}
	if err := auth.CacheEmployeeID(employeeID); err != nil {
		return "", err
	}
	return employeeID, nil
}

func (c *Client) fetchEmployeeBalance(ctx context.Context, employeeID string) (*EmployeeBalance, error) {
	url := fmt.Sprintf("%s/employees/statuses/%s", c.baseURL(), employeeID)
	resp, err := c.DoAuthenticated(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, statusError(resp)
	}
	var env struct {
		Statuses struct {
			TimeBalance signedSeconds `json:"time_balance"`
			LastSettle  *struct {
				UpdatedAt string `json:"updated_at"`
			} `json:"last_settle_time_balance"`
		} `json:"statuses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	balance := &EmployeeBalance{
		EmployeeID:      employeeID,
		TimeBalanceSecs: int64(env.Statuses.TimeBalance),
	}
	if env.Statuses.LastSettle != nil {
		balance.UpdatedAt = env.Statuses.LastSettle.UpdatedAt
	}
	return balance, nil
}

// FetchEmployeeBalance discovers and caches the current employee ID, then loads their time bank.
// A stale cached employee ID is cleared and rediscovered once after HTTP 403 or 404.
func (c *Client) FetchEmployeeBalance(ctx context.Context) (*EmployeeBalance, error) {
	employeeID := auth.GetCachedEmployeeID()
	usedCachedID := employeeID != ""
	if employeeID == "" {
		var err error
		employeeID, err = c.discoverEmployeeID(ctx)
		if err != nil {
			return nil, err
		}
	}
	balance, err := c.fetchEmployeeBalance(ctx, employeeID)
	if err == nil || !usedCachedID || !isStaleEmployeeIDError(err) {
		return balance, err
	}
	_ = auth.ClearCachedEmployeeID()
	employeeID, err = c.discoverEmployeeID(ctx)
	if err != nil {
		return nil, err
	}
	return c.fetchEmployeeBalance(ctx, employeeID)
}

func isStaleEmployeeIDError(err error) bool {
	statusErr, ok := err.(*HTTPStatusError)
	return ok && (statusErr.StatusCode == http.StatusForbidden || statusErr.StatusCode == http.StatusNotFound)
}

func statusError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(body)}
}
