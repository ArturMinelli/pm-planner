package config

import (
	"regexp"
	"sort"
	"strings"
)

// ThemePreset is a user-defined color preset stored under `themes` in config.yaml.
type ThemePreset struct {
	Label  string            `json:"label,omitempty" yaml:"label,omitempty"`
	Colors map[string]string `json:"colors" yaml:"colors"`
}

var (
	themeColorKeys = map[string]bool{"bg": true, "panel": true, "text": true, "muted": true, "border": true, "accent": true, "ok": true, "warning": true, "danger": true}
	themeHexRe     = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	themeIDRe      = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

// SanitizeThemes drops invalid colors, presets and ids in place. It never fails:
// bad entries are silently removed. Missing preset keys are left for the frontend to fill.
func SanitizeThemes(f *File) {
	if f == nil {
		return
	}
	f.DefaultTheme = strings.TrimSpace(f.DefaultTheme)
	f.ThemeName = strings.TrimSpace(f.ThemeName)
	f.Theme = sanitizeColors(f.Theme)
	f.HiddenThemes = sanitizeThemeIDs(f.HiddenThemes)

	if f.Themes == nil {
		return
	}
	themes := map[string]ThemePreset{}
	// sorted so that ids differing only by case resolve deterministically
	keys := make([]string, 0, len(f.Themes))
	for k := range f.Themes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, rawID := range keys {
		p := f.Themes[rawID]
		id := strings.ToLower(strings.TrimSpace(rawID))
		colors := sanitizeColors(p.Colors)
		if !themeIDRe.MatchString(id) || len(colors) == 0 {
			continue
		}
		themes[id] = ThemePreset{Label: strings.TrimSpace(p.Label), Colors: colors}
	}
	f.Themes = nil
	if len(themes) > 0 {
		f.Themes = themes
	}
}

// sanitizeColors keeps only known keys with #rrggbb values; returns nil when nothing survives.
func sanitizeColors(in map[string]string) map[string]string {
	var out map[string]string
	for k, v := range in {
		if !themeColorKeys[k] || !themeHexRe.MatchString(v) {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		out[k] = v
	}
	return out
}

// sanitizeThemeIDs lowercases, validates, dedupes and sorts theme ids; returns nil when none survive.
func sanitizeThemeIDs(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range in {
		id := strings.ToLower(strings.TrimSpace(raw))
		if !themeIDRe.MatchString(id) || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
