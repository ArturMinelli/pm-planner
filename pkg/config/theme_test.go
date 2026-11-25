package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSanitizeThemes(t *testing.T) {
	tests := []struct {
		name string
		in   *File
		want *File
	}{
		{
			name: "drops invalid hex and unknown keys",
			in:   &File{Theme: map[string]string{"accent": "#112233", "bg": "red", "nope": "#000000", "text": "#12345"}},
			want: &File{Theme: map[string]string{"accent": "#112233"}},
		},
		{
			name: "empty theme becomes nil",
			in:   &File{Theme: map[string]string{"bg": "bad"}},
			want: &File{},
		},
		{
			name: "bad preset id dropped",
			in: &File{Themes: map[string]ThemePreset{
				"Bad Id": {Colors: map[string]string{"bg": "#000000"}},
				"ok-1":   {Colors: map[string]string{"bg": "#000000"}},
			}},
			want: &File{Themes: map[string]ThemePreset{
				"ok-1": {Colors: map[string]string{"bg": "#000000"}},
			}},
		},
		{
			name: "preset id lowercased and trimmed, label trimmed",
			in: &File{Themes: map[string]ThemePreset{
				"  Night ": {Label: "  Noite  ", Colors: map[string]string{"accent": "#7c5cff"}},
			}},
			want: &File{Themes: map[string]ThemePreset{
				"night": {Label: "Noite", Colors: map[string]string{"accent": "#7c5cff"}},
			}},
		},
		{
			name: "preset with no valid colors dropped",
			in: &File{Themes: map[string]ThemePreset{
				"empty": {Colors: map[string]string{"bg": "#zzzzzz", "foo": "#000000"}},
			}},
			want: &File{},
		},
		{
			name: "default and name trimmed",
			in:   &File{DefaultTheme: " midnight ", ThemeName: "\tauto "},
			want: &File{DefaultTheme: "midnight", ThemeName: "auto"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SanitizeThemes(tc.in)
			if !reflect.DeepEqual(tc.in, tc.want) {
				t.Fatalf("got %+v, want %+v", tc.in, tc.want)
			}
		})
	}
}

func TestThemes_yamlRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	want := &File{
		DefaultTheme: "midnight",
		ThemeName:    "midnight",
		Theme:        map[string]string{"accent": "#ff8800"},
		Themes: map[string]ThemePreset{
			"midnight": {Label: "Meia-noite", Colors: map[string]string{"bg": "#000000", "accent": "#7c5cff"}},
		},
	}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultTheme != want.DefaultTheme || got.ThemeName != want.ThemeName {
		t.Fatalf("names: got %q/%q", got.DefaultTheme, got.ThemeName)
	}
	if !reflect.DeepEqual(got.Theme, want.Theme) {
		t.Fatalf("theme: got %v", got.Theme)
	}
	if !reflect.DeepEqual(got.Themes, want.Themes) {
		t.Fatalf("themes: got %+v", got.Themes)
	}
}

func TestRead_handWrittenThemes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `default_theme: midnight
theme_name: midnight
themes:
  midnight:
    label: Meia-noite
    colors: {bg: "#000000", panel: "#0a0a0a", text: "#f5f5f5", muted: "#a0a0a0", border: "#262626", accent: "#7c5cff", ok: "#4ade80", warning: "#facc15", danger: "#f87171"}
  Bad Id:
    colors: {bg: "#000000"}
  broken:
    colors: {bg: "nope", unknown: "#ffffff"}
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.DefaultTheme != "midnight" || f.ThemeName != "midnight" {
		t.Fatalf("names: got %q/%q", f.DefaultTheme, f.ThemeName)
	}
	if len(f.Themes) != 1 {
		t.Fatalf("themes: got %+v", f.Themes)
	}
	m := f.Themes["midnight"]
	if m.Label != "Meia-noite" || len(m.Colors) != 9 || m.Colors["accent"] != "#7c5cff" {
		t.Fatalf("midnight: got %+v", m)
	}
}

func TestSanitizeThemesHiddenIDs(t *testing.T) {
	f := &File{HiddenThemes: []string{" Ocean ", "light", "light", "Bad Id", ""}}
	SanitizeThemes(f)
	if got, want := strings.Join(f.HiddenThemes, ","), "light,ocean"; got != want {
		t.Fatalf("hidden themes: got %q, want %q", got, want)
	}

	empty := &File{HiddenThemes: []string{"!!"}}
	SanitizeThemes(empty)
	if empty.HiddenThemes != nil {
		t.Fatalf("expected nil, got %v", empty.HiddenThemes)
	}
}
