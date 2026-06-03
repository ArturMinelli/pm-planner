package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type DesktopMenuOptions struct {
	Root      string
	SkipBuild bool
	Host      HostFacts
}

type DesktopMenuPaths struct {
	Binary         string
	ApplicationDir string
	DesktopFile    string
	IconDir        string
	IconFile       string
	Template       string
	IconSource     string
}

func InstallDesktopMenu(ctx context.Context, options DesktopMenuOptions, runner Runner) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("desktop-menu install is only supported on Linux")
	}

	root, err := ResolveRoot(options.Root)
	if err != nil {
		return err
	}

	host := options.Host
	if host.GOOS == "" {
		host = DetectHostFacts(ctx, root)
	}

	if !options.SkipBuild {
		if err := BuildDesktop(ctx, DesktopBuildOptions{
			Root: root,
			Host: host,
		}, runner); err != nil {
			return err
		}
	}

	paths, err := LinuxDesktopMenuPaths(root, "", nil)
	if err != nil {
		return err
	}

	sourceBinary := DefaultDesktopOutput(root, host.GOOS)
	if err := os.MkdirAll(filepath.Dir(paths.Binary), 0755); err != nil {
		return err
	}
	if err := copyFile(sourceBinary, paths.Binary, 0755); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	if err := os.MkdirAll(paths.IconDir, 0755); err != nil {
		return err
	}
	if err := copyFile(paths.IconSource, paths.IconFile, 0644); err != nil {
		return fmt.Errorf("install icon: %w", err)
	}
	_ = os.Chtimes(filepath.Dir(paths.IconDir), time.Now(), time.Now())

	template, err := os.ReadFile(paths.Template)
	if err != nil {
		return err
	}
	rendered := RenderDesktopEntry(string(template), paths.Binary)
	if err := os.MkdirAll(paths.ApplicationDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(paths.DesktopFile, []byte(rendered), 0644); err != nil {
		return err
	}

	runOptional(ctx, paths.ApplicationDir, "update-desktop-database", paths.ApplicationDir)
	runOptional(ctx, root, "gtk-update-icon-cache", "-f", "-t", filepath.Dir(paths.IconDir))
	return nil
}

func LinuxDesktopMenuPaths(root string, home string, env map[string]string) (DesktopMenuPaths, error) {
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return DesktopMenuPaths{}, err
		}
	}
	getenv := func(key string) string {
		if env != nil {
			return env[key]
		}
		return os.Getenv(key)
	}

	dataHome := strings.TrimSpace(getenv("XDG_DATA_HOME"))
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}
	binHome := strings.TrimSpace(getenv("XDG_BIN_HOME"))
	if binHome == "" {
		binHome = filepath.Join(home, ".local", "bin")
	}

	iconDir := filepath.Join(dataHome, "icons", "hicolor", "scalable", "apps")
	appDir := filepath.Join(dataHome, "applications")
	return DesktopMenuPaths{
		Binary:         filepath.Join(binHome, "pm-desktop"),
		ApplicationDir: appDir,
		DesktopFile:    filepath.Join(appDir, "pm-desktop.desktop"),
		IconDir:        iconDir,
		IconFile:       filepath.Join(iconDir, "pm-desktop.svg"),
		Template:       filepath.Join(root, "packaging", "pm-desktop.desktop"),
		IconSource:     filepath.Join(root, "packaging", "icons", "pm-desktop.svg"),
	}, nil
}

func RenderDesktopEntry(template string, executable string) string {
	lines := strings.Split(template, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Exec=") {
			lines[i] = "Exec=" + executable
			continue
		}
		if strings.HasPrefix(line, "TryExec=") {
			lines[i] = "TryExec=" + executable
		}
	}
	return strings.Join(lines, "\n")
}

func copyFile(source string, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(destination, mode)
}

func runOptional(ctx context.Context, dir string, name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		return
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	_ = cmd.Run()
}
