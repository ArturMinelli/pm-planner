package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"pm-cli/pkg/config"
)

// signInURL is the login endpoint; tests may override it.
var signInURL = "https://api.pontomais.com.br/api/auth/sign_in"

type session struct {
	AccessToken string    `json:"access_token"`
	Token       string    `json:"token"`
	Uid         string    `json:"uid"`
	Client      string    `json:"client"`
	EmployeeID  string    `json:"employee_id,omitempty"`
	CachedAt    time.Time `json:"cached_at"`
	Expiry      int64     `json:"expiry"` // optional epoch seconds if provided by API
}

func cachePath() (string, error) {
	dir, err := config.DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
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

// GetCachedEmployeeID returns the employee ID associated with the current valid session.
func GetCachedEmployeeID() string {
	s, err := readCachedSession()
	if err != nil || !isSessionValid(s) {
		return ""
	}
	return s.EmployeeID
}

// CacheEmployeeID stores the employee ID alongside the current authenticated session.
func CacheEmployeeID(employeeID string) error {
	s, err := readCachedSession()
	if err != nil {
		return err
	}
	if !isSessionValid(s) {
		return errors.New("cannot cache employee ID without a valid session")
	}
	s.EmployeeID = employeeID
	return writeCachedSession(s)
}

// ClearCachedEmployeeID removes a stale employee ID without discarding authentication.
func ClearCachedEmployeeID() error {
	s, err := readCachedSession()
	if err != nil {
		return err
	}
	s.EmployeeID = ""
	return writeCachedSession(s)
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

func sessionToHeaders(s *session) map[string]string {
	access := s.Token
	if s.AccessToken != "" {
		access = s.AccessToken
	}
	return map[string]string{
		"Access-Token": access,
		"Token":        s.Token,
		"Uid":          s.Uid,
		"Client":       s.Client,
	}
}

func signIn(loginID, password string) (*session, error) {
	body := map[string]string{
		"login":    loginID,
		"password": password,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, signInURL, bytes.NewReader(b))
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
		AccessToken: lr.Token,
		Token:       lr.Token,
		Uid:         lr.Data.Login,
		Client:      lr.ClientID,
		CachedAt:    time.Now(),
	}
	if s.Token == "" || s.Uid == "" || s.Client == "" {
		return nil, errors.New("login succeeded but required fields missing in response")
	}
	return s, nil
}

// VerifyCredentials performs a fresh sign-in with the given credentials, ignoring any cached session.
// On success it updates session.json; on failure the existing cache is left unchanged.
func VerifyCredentials(loginID, password string) error {
	s, err := signIn(loginID, password)
	if err != nil {
		return err
	}
	return writeCachedSession(s)
}

func configLogin() string {
	if login := strings.TrimSpace(viper.GetString("login")); login != "" {
		return login
	}
	return strings.TrimSpace(viper.GetString("email"))
}

// GetAuthHeaders ensures we have valid headers, logging in if needed.
func GetAuthHeaders() (map[string]string, error) {
	if s, err := readCachedSession(); err == nil && isSessionValid(s) {
		return sessionToHeaders(s), nil
	}

	loginID := configLogin()
	password := viper.GetString("password")
	if loginID == "" || password == "" {
		path, err := config.DefaultPath()
		if err != nil {
			return nil, errors.New("missing credentials: set 'login' and 'password' in the PM config file")
		}
		return nil, fmt.Errorf("missing credentials: set 'login' and 'password' in %s", path)
	}

	s, err := signIn(loginID, password)
	if err != nil {
		return nil, err
	}
	_ = writeCachedSession(s)
	return sessionToHeaders(s), nil
}
