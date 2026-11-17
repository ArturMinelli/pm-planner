package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// Location is the position submitted with a punch.
type Location struct {
	Latitude, Longitude, Accuracy float64
	Address                       string
}

// Punch is a time card entry as returned by PontoMais.
type Punch struct {
	ID        int64   `json:"id"`
	Date      string  `json:"date"`
	Time      string  `json:"time"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Disabled  bool    `json:"disabled"`
}

// deviceUUID is generated once per process and sent as the _device uuid and uuid header.
var deviceUUID = sync.OnceValues(newUUIDv4)

func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// setWebHeaders mirrors the headers the PontoMais web app sends (seen in the HAR capture).
func setWebHeaders(req *http.Request, uuid string) {
	req.Header.Set("api-version", "2")
	req.Header.Set("uuid", uuid)
	req.Header.Set("Origin", "https://app2.pontomais.com.br")
	req.Header.Set("Referer", "https://app2.pontomais.com.br/")
}

func (c *Client) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if uuid, uerr := deviceUUID(); uerr == nil {
		setWebHeaders(req, uuid)
	}
	resp, err := c.DoAuthenticated(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return statusError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// FetchEmployee returns the raw session employee object (embedded verbatim in punches)
// and its id.
func (c *Client) FetchEmployee(ctx context.Context) (json.RawMessage, string, error) {
	var env struct {
		Session struct {
			Employee json.RawMessage `json:"employee"`
		} `json:"session"`
	}
	if err := c.getJSON(ctx, c.baseURL()+"/session", &env); err != nil {
		return nil, "", err
	}
	var emp struct {
		ID json.Number `json:"id"`
	}
	if len(env.Session.Employee) > 0 {
		if err := json.Unmarshal(env.Session.Employee, &emp); err != nil {
			return nil, "", err
		}
	}
	return env.Session.Employee, emp.ID.String(), nil
}

// ReverseGeocode returns the service-formatted address for a coordinate.
func (c *Client) ReverseGeocode(ctx context.Context, lat, lng float64, employeeID string) (string, error) {
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lng", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("employee_id", employeeID)
	var env struct {
		Address struct {
			FormattedAddress string `json:"formatted_address"`
		} `json:"address"`
	}
	if err := c.getJSON(ctx, c.baseURL()+"/geo_locations?"+q.Encode(), &env); err != nil {
		return "", err
	}
	if env.Address.FormattedAddress == "" {
		return "", errors.New("geo_locations: empty formatted_address")
	}
	return env.Address.FormattedAddress, nil
}

// RegisterPunch submits a punch at loc, replaying the web client's request shape.
func (c *Client) RegisterPunch(ctx context.Context, loc Location) (*Punch, error) {
	if err := guardRealPunch(c); err != nil {
		return nil, err
	}
	employee, _, err := c.FetchEmployee(ctx)
	if err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}
	uuid, err := deviceUUID()
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"image":    nil,
		"employee": employee,
		"time_card": map[string]any{
			"latitude":           loc.Latitude,
			"longitude":          loc.Longitude,
			"address":            loc.Address,
			"original_latitude":  loc.Latitude,
			"original_longitude": loc.Longitude,
			"original_address":   loc.Address,
			"location_edited":    false,
			"accuracy":           loc.Accuracy,
			"accuracy_method":    nil,
			"image":              nil,
			"info":               nil,
		},
		"_path":       "/registrar-ponto",
		"_appVersion": "0.10.32",
		"_device":     map[string]any{"manufacturer": nil, "model": nil, "uuid": uuid, "version": nil},
	}
	req, err := c.NewAuthenticatedRequest(ctx, http.MethodPost, c.baseURL()+"/time_cards/register", body)
	if err != nil {
		return nil, err
	}
	setWebHeaders(req, uuid)
	resp, err := c.DoAuthenticated(req)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("register: %w", statusError(resp))
	}
	var env struct {
		TimeCard *Punch `json:"time_card"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if env.TimeCard == nil {
		return &Punch{}, nil
	}
	return env.TimeCard, nil
}

// FetchTodayPunches returns the punches of the current day from the cached last-state endpoint.
func (c *Client) FetchTodayPunches(ctx context.Context) ([]Punch, error) {
	var env struct {
		TimeCards []Punch `json:"time_cards"`
	}
	if err := c.getJSON(ctx, c.baseURL()+"/time_cards/current/last_cached", &env); err != nil {
		return nil, err
	}
	return env.TimeCards, nil
}
