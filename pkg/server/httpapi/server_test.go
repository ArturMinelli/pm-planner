package httpapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"
	"testing"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	"pm-cli/pkg/punchlog"
	pmserver "pm-cli/pkg/server"
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

func TestTestModeEndpointReflectsFakeEnv(t *testing.T) {
	const env = "PM_FAKE_PONTOMAIS_URL"
	get := func() bool {
		rec := httptest.NewRecorder()
		httpapi.New("3847").Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/testmode", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var body struct {
			TestMode bool `json:"testMode"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return body.TestMode
	}

	t.Setenv(env, "x")
	os.Unsetenv(env)
	if get() {
		t.Fatal("testMode = true with env unset")
	}
	t.Setenv(env, "http://127.0.0.1:3850/api")
	if !get() {
		t.Fatal("testMode = false with env set")
	}
}

func TestAuthTestSkipsSignInInFakeMode(t *testing.T) {
	t.Setenv("PM_FAKE_PONTOMAIS_URL", "http://127.0.0.1:1/unused")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/test", strings.NewReader(`{"login":"a@example.invalid","password":"p"}`))
	httpapi.New("3847").Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("status = %d body %s; want ok without sign-in", rec.Code, rec.Body.String())
	}
}

type stubRemote struct{ calls int }

func (s *stubRemote) FetchTodayPunches(context.Context) ([]api.Punch, error) {
	s.calls++
	return []api.Punch{{ID: 3, Date: "2026-11-17", Time: "09:00:00"}}, nil
}

func punchServer(t *testing.T) (http.Handler, *stubRemote) {
	t.Helper()
	remote := &stubRemote{}
	s := punchlog.NewSyncer(punchlog.NewStore(filepath.Join(t.TempDir(), "punches.json")), remote)
	s.Now = func() time.Time { return time.Date(2026, 11, 17, 10, 0, 0, 0, time.Local) }
	pmserver.SetPunchLog(s)
	t.Cleanup(func() { pmserver.SetPunchLog(nil) })
	return httpapi.New("3847").Handler(), remote
}

func doReq(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestPunchesEndpoints(t *testing.T) {
	h, remote := punchServer(t)

	rec := doReq(h, http.MethodGet, "/api/punches")
	var st punchlog.State
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &st) != nil || st.Date != "2026-11-17" || len(st.Punches) != 0 || remote.calls != 0 {
		t.Fatalf("get: %d %s calls=%d", rec.Code, rec.Body.String(), remote.calls)
	}
	if !strings.Contains(rec.Body.String(), `"punches":[]`) || !strings.Contains(rec.Body.String(), `"pendingCount":0`) {
		t.Fatalf("contract: %s", rec.Body.String())
	}

	rec = doReq(h, http.MethodPost, "/api/punches/sync")
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &st) != nil || len(st.Punches) != 1 || st.Punches[0].RemoteID != 3 || st.LastSyncAt == nil {
		t.Fatalf("sync: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"remoteId":3`) || !strings.Contains(rec.Body.String(), `"source":"remote"`) {
		t.Fatalf("camelCase: %s", rec.Body.String())
	}
	// Debounced second call: no new remote call.
	doReq(h, http.MethodPost, "/api/punches/sync")
	if remote.calls != 1 {
		t.Fatalf("remote calls = %d, want 1", remote.calls)
	}

	if rec = doReq(h, http.MethodGet, "/api/punches?date=2026-11-17"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"remoteId":3`) {
		t.Fatalf("by date: %s", rec.Body.String())
	}
	if rec = doReq(h, http.MethodGet, "/api/punches?date=nope"); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad date status = %d", rec.Code)
	}
}
