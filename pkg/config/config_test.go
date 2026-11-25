package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead_legacyEmailKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("email: user@example.com\npassword: secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Login != "user@example.com" {
		t.Fatalf("login: got %q", f.Login)
	}
	if f.Password != "secret" {
		t.Fatalf("password: got %q", f.Password)
	}
}

func TestRead_loginKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("login: 52998224725\npassword: secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Login != "52998224725" {
		t.Fatalf("login: got %q", f.Login)
	}
}

func TestRead_prefersLoginOverLegacyEmail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "login: cpf-user\nemail: legacy@example.com\npassword: secret\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Login != "cpf-user" {
		t.Fatalf("login: got %q", f.Login)
	}
}

func TestSave_writesLoginOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := Save(path, &File{Login: "user@example.com", Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, "login:") {
		t.Fatalf("expected login key in %q", content)
	}
	if strings.Contains(content, "email:") {
		t.Fatalf("did not expect legacy email key in %q", content)
	}
}

func TestSave_roundTripsTheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	want := map[string]string{"accent": "#112233"}
	if err := Save(path, &File{Login: "user", Password: "secret", Theme: want}); err != nil {
		t.Fatal(err)
	}

	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Theme) != len(want) || f.Theme["accent"] != want["accent"] {
		t.Fatalf("theme: got %v, want %v", f.Theme, want)
	}
}
