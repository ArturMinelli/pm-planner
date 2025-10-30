package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type session struct {
	AccessToken string    `json:"access_token"`
	Token       string    `json:"token"`
	Uid         string    `json:"uid"`
	Client      string    `json:"client"`
	CachedAt    time.Time `json:"cached_at"`
	Expiry      int64     `json:"expiry"` // optional epoch seconds if provided by API
}

func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pm", "session.json"), nil
}

func readCachedSession() (*session, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func writeCachedSession(s *session) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func isSessionValid(s *session) bool {
	if s == nil || s.Token == "" || s.Uid == "" || s.Client == "" {
		return false
	}
	// Prefer API-provided expiry if present
	if s.Expiry > 0 {
		return time.Now().Unix() < s.Expiry
	}
	// Fallback TTL (hours) from config, default 8h
	ttlHours := viper.GetInt("cache_ttl_hours")
	if ttlHours <= 0 {
		ttlHours = 8
	}
	return time.Since(s.CachedAt) < time.Duration(ttlHours)*time.Hour
}

// GetAuthHeaders ensures we have valid headers, logging in if needed.
func GetAuthHeaders() (map[string]string, error) {
	if s, err := readCachedSession(); err == nil && isSessionValid(s) {
		return map[string]string{
			"Access-Token": func() string {
				if s.AccessToken != "" {
					return s.AccessToken
				}
				return s.Token
			}(),
			"Token":  s.Token,
			"Uid":    s.Uid,
			"Client": s.Client,
		}, nil
	}

	email := viper.GetString("email")
	password := viper.GetString("password")
	if email == "" || password == "" {
		return nil, errors.New("missing credentials: set 'email' and 'password' in $HOME/.pm/config.yaml")
	}

	body := map[string]string{
		"login":    email,
		"password": password,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, "https://api.pontomais.com.br/api/auth/sign_in", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Parse JSON response body
	type loginResponse struct {
		Success  string `json:"success"`
		Token    string `json:"token"`
		ClientID string `json:"client_id"`
		Data     struct {
			Login        string `json:"login"`
			SignInCount  int    `json:"sign_in_count"`
			LastSignInIP string `json:"last_sign_in_ip"`
			LastSignInAt int64  `json:"last_sign_in_at"`
		} `json:"data"`
	}
	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	s := &session{
		AccessToken: lr.Token, // Some endpoints expect both Access-Token and Token with same value
		Token:       lr.Token,
		Uid:         lr.Data.Login,
		Client:      lr.ClientID,
		CachedAt:    time.Now(),
		// Expiry not provided by this response; fallback TTL logic applies
	}
	if s.Token == "" || s.Uid == "" || s.Client == "" {
		return nil, errors.New("login succeeded but required fields missing in response")
	}
	_ = writeCachedSession(s)

	return map[string]string{
		"Access-Token": s.AccessToken,
		"Token":        s.Token,
		"Uid":          s.Uid,
		"Client":       s.Client,
	}, nil
}
