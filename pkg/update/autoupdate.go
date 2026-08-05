package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	linuxAutoUpdateEntry  = "pm-planner-autoupdate.desktop"
	darwinAutoUpdateLabel = "com.pm-planner.autoupdate"
	windowsAutoUpdateTask = "PM Planner Auto Update"
)

// RemoveScheduledAutoUpdate unregisters the daily auto-update the old setup
// scripts installed. Updates are manual now, so any surviving OS entry from a
// previous install is stale. Safe to call repeatedly and when nothing exists.
func RemoveScheduledAutoUpdate() error {
	switch runtime.GOOS {
	case "darwin":
		plist, err := darwinAutoUpdatePlist()
		if err != nil {
			return err
		}
		if _, err := os.Stat(plist); err != nil {
			return nil
		}
		_ = exec.Command("launchctl", "unload", plist).Run()
		return removeIfExists(plist)
	case "windows":
		// schtasks fails when the task is absent, which is the common case.
		_ = exec.Command("schtasks", "/Delete", "/TN", windowsAutoUpdateTask, "/F").Run()
		return nil
	default:
		entry, err := linuxAutoUpdateDesktopFile()
		if err != nil {
			return err
		}
		return removeIfExists(entry)
	}
}

func linuxAutoUpdateDesktopFile() (string, error) {
	configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "autostart", linuxAutoUpdateEntry), nil
}

func darwinAutoUpdatePlist() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", darwinAutoUpdateLabel+".plist"), nil
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
