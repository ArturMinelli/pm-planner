package httpapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"pm-cli/pkg/config"
	"pm-cli/pkg/server/httpapi"
)

func TestUpdateCheckReturnsDesktopOnlyBlocker(t *testing.T) {
	server := httpapi.New("3847")
	request := httptest.NewRequest(http.MethodGet, "/api/updates/check", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	body, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var status struct {
		UpdateAvailable bool `json:"updateAvailable"`
		Blockers        []struct {
			Key string `json:"key"`
		} `json:"blockers"`
	}
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if status.UpdateAvailable {
		t.Fatal("expected updateAvailable false")
	}
	if len(status.Blockers) != 1 || status.Blockers[0].Key != "common.desktopOnlyUpdates" {
		t.Fatalf("blockers = %v, want desktopOnlyUpdates blocker", status.Blockers)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	if err := config.Init(""); err != nil {
		t.Fatalf("config init: %v", err)
	}

	server := httpapi.New("3847")

	getRequest := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	getRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRecorder.Code, http.StatusOK)
	}

	putBody := `{"login":"dev@example.com","password":"secret","locale":"pt-BR"}`
	putRequest := httptest.NewRequest(http.MethodPut, "/api/config", io.NopCloser(strings.NewReader(putBody)))
	putRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(putRecorder, putRequest)
	if putRecorder.Code != http.StatusNoContent {
		t.Fatalf("put status = %d, want %d; body %s", putRecorder.Code, http.StatusNoContent, putRecorder.Body.String())
	}

	getRecorder2 := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder2, getRequest)
	if getRecorder2.Code != http.StatusOK {
		t.Fatalf("get2 status = %d, want %d", getRecorder2.Code, http.StatusOK)
	}

	var file config.File
	if err := json.NewDecoder(getRecorder2.Body).Decode(&file); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if file.Login != "dev@example.com" {
		t.Fatalf("login = %q, want dev@example.com", file.Login)
	}
	if file.Locale != config.LocalePTBR {
		t.Fatalf("locale = %q, want %q", file.Locale, config.LocalePTBR)
	}
}

func TestConfigRejectsUnsupportedLocale(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	if err := config.Init(""); err != nil {
		t.Fatalf("config init: %v", err)
	}

	server := httpapi.New("3847")
	putBody := `{"login":"dev@example.com","password":"secret","locale":"en"}`
	putRequest := httptest.NewRequest(http.MethodPut, "/api/config", io.NopCloser(strings.NewReader(putBody)))
	putRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(putRecorder, putRequest)
	if putRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("put status = %d, want %d; body %s", putRecorder.Code, http.StatusInternalServerError, putRecorder.Body.String())
	}
}
