package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	LocaleEN   = "en"
	LocalePTBR = "pt-BR"
)

// SupportedLocales lists locales the desktop app can use.
var SupportedLocales = []string{LocaleEN, LocalePTBR}

// DetectOSLocale maps the host locale to a supported locale, defaulting to pt-BR.
func DetectOSLocale() string {
	locale := os.Getenv("LANG")
	if locale == "" {
		locale = os.Getenv("LC_ALL")
	}
	if locale == "" {
		return LocalePTBR
	}
	locale = strings.ToLower(locale)
	if strings.HasPrefix(locale, "en") {
		return LocaleEN
	}
	if strings.HasPrefix(locale, "pt") {
		return LocalePTBR
	}
	return LocalePTBR
}

// NormalizeLocale returns a supported locale or an error.
func NormalizeLocale(locale string) (string, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return "", nil
	}
	for _, supported := range SupportedLocales {
		if locale == supported {
			return supported, nil
		}
	}
	return "", fmt.Errorf("unsupported locale %q (use %q or %q)", locale, LocaleEN, LocalePTBR)
}

// ResolveLocale returns the configured locale or detects it from the OS.
func ResolveLocale(file *File) string {
	if file != nil {
		if locale, err := NormalizeLocale(file.Locale); err == nil && locale != "" {
			return locale
		}
	}
	return DetectOSLocale()
}

// ValidateLocale rejects unknown locale values on save.
func ValidateLocale(locale string) error {
	if strings.TrimSpace(locale) == "" {
		return nil
	}
	_, err := NormalizeLocale(locale)
	return err
}
