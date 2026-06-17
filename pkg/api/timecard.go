package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const registerAppVersion = "0.10.32"

// CachedTimeCard is a recent punch returned by last_cached.
type CachedTimeCard struct {
	ID        int64    `json:"id"`
	Date      string   `json:"date"`
	Time      string   `json:"time"`
	Address   *string  `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// ClockInStatus summarizes the most recent punch for display.
type ClockInStatus struct {
	Date      string `json:"date"`
	Time      string `json:"time"`
	Address   string `json:"address,omitempty"`
	HasRecent bool   `json:"hasRecent"`
}

// TimeCardLocation is the geolocation payload for clock-in.
type TimeCardLocation struct {
	Latitude  float64
	Longitude float64
	Address   string
	Accuracy  float64
}

// RegisterTimeCardResult is returned after a successful punch.
type RegisterTimeCardResult struct {
	Time    string `json:"time"`
	Date    string `json:"date"`
	Address string `json:"address"`
}

// FetchLastCachedTimeCards loads recent punches from PontoMais.
func (c *Client) FetchLastCachedTimeCards(ctx context.Context) ([]CachedTimeCard, error) {
	query := url.Values{"_t": {strconv.FormatInt(time.Now().UnixMilli(), 10)}}
	endpoint := fmt.Sprintf("%s/time_cards/current/last_cached?%s", c.baseURL(), query.Encode())
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, statusError(resp)
	}
	var env struct {
		TimeCards []CachedTimeCard `json:"time_cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	return env.TimeCards, nil
}

// FetchClockInStatus returns the latest cached punch for UI display.
func (c *Client) FetchClockInStatus(ctx context.Context) (*ClockInStatus, error) {
	cards, err := c.FetchLastCachedTimeCards(ctx)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return &ClockInStatus{}, nil
	}
	latest := cards[0]
	status := &ClockInStatus{
		Date:      latest.Date,
		Time:      latest.Time,
		HasRecent: true,
	}
	if latest.Address != nil {
		status.Address = *latest.Address
	}
	return status, nil
}

// FetchEmployeeForRegister loads the employee profile required by register.
func (c *Client) FetchEmployeeForRegister(ctx context.Context) (json.RawMessage, error) {
	employee, _, err := c.fetchMyTimeBreak(ctx)
	return employee, err
}

type registerRequest struct {
	Image    any             `json:"image"`
	Employee json.RawMessage `json:"employee"`
	TimeCard registerTimeCard `json:"time_card"`
	Path     string          `json:"_path"`
	AppVer   string          `json:"_appVersion"`
	Device   registerDevice  `json:"_device"`
}

type registerTimeCard struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Address          string  `json:"address"`
	OriginalLatitude float64 `json:"original_latitude"`
	OriginalLongitude float64 `json:"original_longitude"`
	OriginalAddress  string  `json:"original_address"`
	LocationEdited   bool    `json:"location_edited"`
	Accuracy         float64 `json:"accuracy"`
	AccuracyMethod   any     `json:"accuracy_method"`
	Image            any     `json:"image"`
	Info             any     `json:"info"`
}

type registerDevice struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Version      string `json:"version"`
}

type registerResponse struct {
	UntreatedTimeCard *struct {
		Time    string `json:"time"`
		Date    string `json:"date"`
		Address string `json:"address"`
	} `json:"untreated_time_card"`
}

// RegisterTimeCard records a punch with the given location.
func (c *Client) RegisterTimeCard(ctx context.Context, location TimeCardLocation) (*RegisterTimeCardResult, error) {
	employee, err := c.FetchEmployeeForRegister(ctx)
	if err != nil {
		return nil, err
	}

	payload := registerRequest{
		Image:    nil,
		Employee: employee,
		TimeCard: registerTimeCard{
			Latitude:          location.Latitude,
			Longitude:         location.Longitude,
			Address:           location.Address,
			OriginalLatitude:  location.Latitude,
			OriginalLongitude: location.Longitude,
			OriginalAddress:   location.Address,
			LocationEdited:    false,
			Accuracy:          location.Accuracy,
			AccuracyMethod:    nil,
			Image:             nil,
			Info:              nil,
		},
		Path:   "/registrar-ponto",
		AppVer: registerAppVersion,
		Device: registerDevice{
			Manufacturer: "null",
			Model:        "null",
			Version:      "null",
		},
	}

	endpoint := c.baseURL() + "/time_cards/register"
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var parsed registerResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.UntreatedTimeCard == nil {
		return nil, fmt.Errorf("register succeeded but untreated_time_card missing")
	}
	return &RegisterTimeCardResult{
		Time:    parsed.UntreatedTimeCard.Time,
		Date:    parsed.UntreatedTimeCard.Date,
		Address: parsed.UntreatedTimeCard.Address,
	}, nil
}
