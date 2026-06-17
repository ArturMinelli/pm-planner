package auth

import "errors"

// BuildRegisterDevice returns the _device payload expected by time_cards/register.
func BuildRegisterDevice() (map[string]any, error) {
	s, err := readCachedSession()
	if err != nil {
		return nil, err
	}
	if !isSessionValid(s) || !hasCompleteAuth(s) {
		return nil, errors.New("valid session required to build register device payload")
	}
	uuid := ensureSessionUUID(s)
	deviceUUID := map[string]any{
		"token":     s.Token,
		"client_id": s.Client,
		"uuid":      uuid,
		"data": map[string]any{
			"login":            s.Uid,
			"sign_in_count":    s.SignInCount,
			"last_sign_in_ip":  s.LastSignInIP,
			"last_sign_in_at":  s.LastSignInAt,
		},
	}
	if s.SignInSuccess != "" {
		deviceUUID["success"] = s.SignInSuccess
	}
	return map[string]any{
		"manufacturer": "null",
		"model":        "null",
		"version":      "null",
		"uuid":         deviceUUID,
	}, nil
}
