package project

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxDesktopMenuPathsDefaultXDG(t *testing.T) {
	paths, err := LinuxDesktopMenuPaths("/repo", "/home/alex", nil)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := paths.Binary, filepath.Join("/home/alex", ".local", "bin", "pm-desktop"); got != want {
		t.Fatalf("binary: got %q want %q", got, want)
	}
	if got, want := paths.DesktopFile, filepath.Join("/home/alex", ".local", "share", "applications", "pm-desktop.desktop"); got != want {
		t.Fatalf("desktop file: got %q want %q", got, want)
	}
	if got, want := paths.IconFile, filepath.Join("/home/alex", ".local", "share", "icons", "hicolor", "scalable", "apps", "pm-desktop.svg"); got != want {
		t.Fatalf("icon file: got %q want %q", got, want)
	}
}

func TestLinuxDesktopMenuPathsHonorsXDGOverrides(t *testing.T) {
	paths, err := LinuxDesktopMenuPaths("/repo", "/home/alex", map[string]string{
		"XDG_DATA_HOME": "/xdg/data",
		"XDG_BIN_HOME":  "/xdg/bin",
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := paths.Binary, filepath.Join("/xdg", "bin", "pm-desktop"); got != want {
		t.Fatalf("binary: got %q want %q", got, want)
	}
	if got, want := paths.DesktopFile, filepath.Join("/xdg", "data", "applications", "pm-desktop.desktop"); got != want {
		t.Fatalf("desktop file: got %q want %q", got, want)
	}
}

func TestRenderDesktopEntryRewritesExecutableFields(t *testing.T) {
	rendered := RenderDesktopEntry("Name=PM\nExec=pm-desktop\nTryExec=pm-desktop\nIcon=pm-desktop\n", "/home/alex/.local/bin/pm-desktop")

	if !strings.Contains(rendered, "Exec=/home/alex/.local/bin/pm-desktop") {
		t.Fatalf("missing rewritten Exec line:\n%s", rendered)
	}
	if !strings.Contains(rendered, "TryExec=/home/alex/.local/bin/pm-desktop") {
		t.Fatalf("missing rewritten TryExec line:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Icon=pm-desktop") {
		t.Fatalf("icon should remain unchanged:\n%s", rendered)
	}
}
