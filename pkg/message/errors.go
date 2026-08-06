package message

import (
	"strings"
)

// FromError maps known errors to i18n keys; unknown errors use errors.generic.
func FromError(err error) Message {
	if err == nil {
		return New(KeyErrorsGeneric, nil)
	}

	text := strings.ToLower(err.Error())

	if strings.Contains(text, "missing credentials") {
		return New(KeyErrorsAuthMissingCredentials, nil)
	}
	if strings.Contains(text, "login failed") {
		return New(KeyErrorsAuthLoginFailed, nil)
	}
	if strings.Contains(text, "faça login") || strings.Contains(text, "login para continuar") {
		return New(KeyErrorsAuthSessionInvalid, nil)
	}
	if strings.Contains(text, "employee id missing") {
		return New(KeyErrorsBalanceUnavailable, nil)
	}
	if strings.Contains(text, "http 401") || strings.Contains(text, "http 403") {
		return New(KeyErrorsAuthSessionInvalid, nil)
	}
	if strings.Contains(text, "http 429") || strings.Contains(text, "rate limited") {
		return New(KeyErrorsBalanceUnavailable, nil)
	}
	if strings.Contains(text, "http ") {
		return New(KeyErrorsBalanceUnavailable, nil)
	}

	return New(KeyErrorsGeneric, nil)
}

// BlockerEnglish returns the English CLI text for an update blocker key.
func BlockerEnglish(message Message) string {
	switch message.Key {
	case KeyUpdateBlockersInstallNotFound:
		return "PM Planner installation not found on disk. Run setup again."
	case KeyUpdateBlockersPMNotFound:
		return "pm CLI not found. Run setup again to reinstall it."
	case KeyUpdateBlockersGoNotFound:
		return "Go not found on PATH — required to build PM Planner."
	case KeyUpdateBlockersNodeNotFound:
		return "Node.js not found on PATH — required to build the frontend."
	case KeyUpdateBlockersDirtyWorkingTree:
		if message.Params != nil && message.Params["root"] != "" {
			return "Uncommitted local changes in " + message.Params["root"] + ". Commit or stash before updating."
		}
		return "Uncommitted local changes. Commit or stash before updating."
	case KeyUpdateBlockersFetchFailed:
		if message.Params != nil && message.Params["detail"] != "" {
			return "Could not fetch from remote repository: " + message.Params["detail"]
		}
		return "Could not fetch from remote repository."
	case KeyUpdateBlockersCompareFailed:
		if message.Params != nil && message.Params["branch"] != "" && message.Params["detail"] != "" {
			return "Could not compare with " + message.Params["branch"] + ": " + message.Params["detail"]
		}
		return "Could not compare with remote branch."
	case KeyUpdateBlockersCheckFailed:
		if message.Params != nil && message.Params["detail"] != "" {
			return "Could not check for updates: " + message.Params["detail"]
		}
		return "Could not check for updates."
	default:
		return message.Key
	}
}

// ResultEnglish returns the English CLI text for an update result key.
func ResultEnglish(message Message) string {
	switch message.Key {
	case KeyUpdateResultSuccess:
		return "PM Planner updated successfully."
	case KeyUpdateResultSuccessCommit:
		if message.Params != nil && message.Params["commit"] != "" {
			return "PM Planner updated to " + message.Params["commit"] + "."
		}
		return "PM Planner updated successfully."
	case KeyUpdateResultFailed:
		if message.Params != nil && message.Params["detail"] != "" && message.Params["logPath"] != "" {
			return "Update failed: " + message.Params["detail"] + ". See " + message.Params["logPath"] + " for details."
		}
		return "Update failed."
	default:
		return message.Key
	}
}
