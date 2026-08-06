package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"pm-cli/pkg/config"
	"pm-cli/pkg/server"
)

const defaultPort = "3847"

// Server is the local HTTP API for dev-browser access to shared Go domain logic.
type Server struct {
	addr string
}

// New constructs a server bound to 127.0.0.1 on the given port (default 3847).
func New(port string) *Server {
	if port == "" {
		port = defaultPort
	}
	return &Server{addr: "127.0.0.1:" + port}
}

// ListenAndServe starts the HTTP server until the context is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	handler := s.handler()

	server := &http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// Handler returns the HTTP handler for tests and server startup.
func (s *Server) Handler() http.Handler {
	return s.handler()
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/auth/test", s.handleAuthTest)
	mux.HandleFunc("POST /api/planner/recalculate", s.handleRecalculate)
	mux.HandleFunc("GET /api/planner/{date}", s.handlePlannerDate)
	mux.HandleFunc("/api/reminders/status", s.handleReminderStatus)
	mux.HandleFunc("/api/reminders/settings", s.handleReminderSettings)
	mux.HandleFunc("/api/reminders/plan", s.handleReminderPlan)
	mux.HandleFunc("/api/updates/check", s.handleUpdateCheck)
	return withCORS(mux)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		file, err := server.GetConfig()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, file)
	case http.MethodPut:
		file, err := decodeJSON[*config.File](r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if file == nil {
			writeError(w, http.StatusBadRequest, errors.New("config required"))
			return
		}
		if err := server.SaveConfig(file); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *Server) handleAuthTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	body, err := decodeJSON[authTestRequest](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result := server.TestAuth(r.Context(), body.Login, body.Password)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handlePlannerDate(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if date == "" {
		writeError(w, http.StatusBadRequest, errors.New("date required"))
		return
	}
	payload, err := server.LoadPlanner(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleRecalculate(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[server.RecalculateRequest](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	summary, err := server.RecalculateDay(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleReminderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	status, err := server.BrowserReminderStatus()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleReminderSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	settings, err := decodeJSON[config.Reminders](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := server.SaveReminderSettingsToConfig(settings); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReminderPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	date := r.URL.Query().Get("date")
	plan, err := server.BuildReminderPlan(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	writeJSON(w, http.StatusOK, server.DesktopOnlyUpdateStatus())
}

type authTestRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func decodeJSON[T any](body io.Reader) (T, error) {
	var value T
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:")
}
