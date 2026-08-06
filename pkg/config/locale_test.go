package config_test

import (
	"testing"

	"pm-cli/pkg/config"
)

func TestValidateLocaleRejectsEnglish(t *testing.T) {
	err := config.ValidateLocale("en")
	if err == nil {
		t.Fatal("expected error for locale en")
	}
}

func TestValidateLocaleAcceptsPortuguese(t *testing.T) {
	if err := config.ValidateLocale(config.LocalePTBR); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocaleAcceptsEmpty(t *testing.T) {
	if err := config.ValidateLocale(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveLocaleAlwaysReturnsPortuguese(t *testing.T) {
	if got := config.ResolveLocale(&config.File{Locale: "en"}); got != config.LocalePTBR {
		t.Fatalf("ResolveLocale = %q, want %q", got, config.LocalePTBR)
	}
	if got := config.ResolveLocale(nil); got != config.LocalePTBR {
		t.Fatalf("ResolveLocale(nil) = %q, want %q", got, config.LocalePTBR)
	}
}

func TestSaveForcesPortugueseLocale(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	file := &config.File{
		Login:    "user@example.com",
		Password: "secret",
		Locale:   "pt-BR",
	}
	if err := config.Save(path, file); err != nil {
		t.Fatalf("save: %v", err)
	}
	read, err := config.Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if read.Locale != config.LocalePTBR {
		t.Fatalf("locale = %q, want %q", read.Locale, config.LocalePTBR)
	}
}

func TestSaveRejectsEnglishLocale(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	file := &config.File{
		Login:    "user@example.com",
		Password: "secret",
		Locale:   "en",
	}
	if err := config.Save(path, file); err == nil {
		t.Fatal("expected error saving locale en")
	}
}
