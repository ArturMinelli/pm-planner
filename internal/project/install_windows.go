//go:build windows

package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const windowsAppName = "PM Planner"

func windowsProgramDir() (string, error) {
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return filepath.Join(localAppData, "Programs", windowsAppName), nil
}

func windowsStartMenuShortcutPath() (string, error) {
	appData := strings.TrimSpace(os.Getenv("APPDATA"))
	if appData == "" {
		return "", fmt.Errorf("APPDATA is not set")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", windowsAppName+".lnk"), nil
}

func windowsDesktopShortcutPath() (string, error) {
	userProfile := strings.TrimSpace(os.Getenv("USERPROFILE"))
	if userProfile == "" {
		return "", fmt.Errorf("USERPROFILE is not set")
	}
	return filepath.Join(userProfile, "Desktop", windowsAppName+".lnk"), nil
}

func InstallDesktopWindows(ctx context.Context, options DesktopInstallOptions, runner Runner) error {
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

	sourceBinary := DefaultDesktopOutput(root, host.GOOS)
	if _, err := os.Stat(sourceBinary); err != nil {
		return fmt.Errorf("desktop binary not found at %s; run without --skip-build", sourceBinary)
	}

	programDir, err := windowsProgramDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(programDir, 0755); err != nil {
		return err
	}

	installedBinary := filepath.Join(programDir, BinaryName("pm-desktop", "windows"))
	if err := copyFile(sourceBinary, installedBinary, 0755); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	startMenuShortcut, err := windowsStartMenuShortcutPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(startMenuShortcut), 0755); err != nil {
		return err
	}
	if err := createWindowsShortcut(installedBinary, startMenuShortcut, windowsAppName); err != nil {
		return fmt.Errorf("create Start Menu shortcut: %w", err)
	}

	if options.DesktopShortcut {
		desktopShortcut, err := windowsDesktopShortcutPath()
		if err != nil {
			return err
		}
		if err := createWindowsShortcut(installedBinary, desktopShortcut, windowsAppName); err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}
	}

	fmt.Fprintln(runner.Stdout, "PM Planner instalado — procure no Menu Iniciar.")
	if options.DesktopShortcut {
		fmt.Fprintln(runner.Stdout, "Atalho na Área de Trabalho criado.")
	}
	return nil
}

func createWindowsShortcut(targetExecutable string, shortcutPath string, description string) error {
	workingDirectory := filepath.Dir(targetExecutable)
	script := fmt.Sprintf(
		`$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%s'); $Shortcut.TargetPath = '%s'; $Shortcut.WorkingDirectory = '%s'; $Shortcut.Description = '%s'; $Shortcut.Save()`,
		escapePowerShellSingleQuoted(shortcutPath),
		escapePowerShellSingleQuoted(targetExecutable),
		escapePowerShellSingleQuoted(workingDirectory),
		escapePowerShellSingleQuoted(description),
	)
	return runPowerShell(script)
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func runPowerShell(script string) error {
	command := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
