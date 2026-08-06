package config

import (
	"fmt"
	"strings"
)

const LocalePTBR = "pt-BR"

// ResolveLocale always returns pt-BR.
func ResolveLocale(_ *File) string {
	return LocalePTBR
}

// ValidateLocale rejects any locale other than empty or pt-BR.
func ValidateLocale(locale string) error {
	locale = strings.TrimSpace(locale)
	if locale == "" || locale == LocalePTBR {
		return nil
	}
	return fmt.Errorf("unsupported locale %q (only %q is supported)", locale, LocalePTBR)
}
