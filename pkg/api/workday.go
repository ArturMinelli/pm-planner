package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type WorkDay struct {
	Date      string  `json:"date"`
	ShiftTime float64 `json:"shift_time"`
	ShiftDay  struct {
		Periods []struct {
			EnterTime string `json:"enter_time"`
			LeaveTime string `json:"leave_time"`
		} `json:"periods"`
	} `json:"shift_day"`
	TimeCards []struct {
		Time string `json:"time"`
	} `json:"time_cards"`
}

type workdayEnvelope struct {
	WorkDay WorkDay `json:"work_day"`
}

// RateLimitError carries retry timing when the API returns HTTP 429.
type RateLimitError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("HTTP %d: rate limited; retry after %s", e.StatusCode, e.RetryAfter)
	}
	return fmt.Sprintf("HTTP %d: rate limited", e.StatusCode)
}

func (c *Client) FetchWorkDay(ctx context.Context, date string) (*WorkDay, error) {
	url := fmt.Sprintf("%s/time_card_control/current/work_days/%s", c.baseURL(), date)
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoAuthenticated(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, &RateLimitError{
				StatusCode: resp.StatusCode,
				Body:       string(b),
				RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
			}
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	var env workdayEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	return &env.WorkDay, nil
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		d := time.Until(when)
		if d > 0 {
			return d
		}
	}
	return 0
}

func ParseHHMMOnDate(hhmm string, date time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
}
