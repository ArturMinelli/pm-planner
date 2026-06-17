package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// UserFacingError returns a concise message suitable for UI display.
func UserFacingError(err error) string {
	if err == nil {
		return ""
	}
	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		if msg := parseAPIErrorBody(statusErr.Body); msg != "" {
			return msg
		}
		if statusErr.StatusCode == 401 || statusErr.StatusCode == 403 {
			return "sessão expirada ou sem permissão. Verifique login em Configurações."
		}
		return fmt.Sprintf("PontoMais retornou erro %d.", statusErr.StatusCode)
	}
	return err.Error()
}

func parseAPIErrorBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	var env struct {
		Errors  []string `json:"errors"`
		Error   string   `json:"error"`
		Message string   `json:"message"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		return ""
	}
	for _, item := range env.Errors {
		if msg := strings.TrimSpace(item); msg != "" {
			return msg
		}
	}
	if msg := strings.TrimSpace(env.Message); msg != "" {
		return msg
	}
	if msg := strings.TrimSpace(env.Error); msg != "" {
		return msg
	}
	return ""
}
