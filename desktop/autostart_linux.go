//go:build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func linuxAutostartDir() (string, error) {
	configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "autostart"), nil
}

func autostartPath() (string, error) {
	dir, err := linuxAutostartDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pm-planner-reminders.desktop"), nil
}

func autostartEnable(executable string) error {
	path, err := autostartPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(renderAutostartDesktop(executable)), 0644)
}

func autostartDisable() error {
	path, err := autostartPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func autostartIsEnabled() (bool, error) {
	path, err := autostartPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func renderAutostartDesktop(executable string) string {
	return fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=PM Planner Reminders
Comment=Lembretes de jornada do PM Planner
Exec=%s --daemon
Terminal=false
X-GNOME-Autostart-enabled=true
`, quoteDesktopExec(executable))
}

func quoteDesktopExec(value string) string {
	if !strings.ContainsAny(value, " \t\n\"'\\") {
		return value
	}
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}
