//go:build darwin

package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const macOSAppBundleName = "PM Planner.app"

func macOSApplicationsDir(systemInstall bool) (string, error) {
	if systemInstall {
		return "/Applications", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Applications"), nil
}

func InstallDesktopDarwin(ctx context.Context, options DesktopInstallOptions, runner Runner) error {
	root, err := ResolveRoot(options.Root)
	if err != nil {
		return err
	}

	host := options.Host
	if host.GOOS == "" {
		host = DetectHostFacts(ctx, root)
	}

	if !options.SkipBuild {
		if err := BuildDesktopBundle(ctx, DesktopBuildOptions{
			Root:         root,
			SkipFrontend: false,
			Host:         host,
		}, runner); err != nil {
			return err
		}
	}

	sourceBundle := MacOSAppBundleArtifact(root)
	if _, err := os.Stat(sourceBundle); err != nil {
		return fmt.Errorf("app bundle not found at %s; run without --skip-build", sourceBundle)
	}

	applicationsDir, err := macOSApplicationsDir(options.System)
	if err != nil {
		return err
	}
	installPath := filepath.Join(applicationsDir, macOSAppBundleName)

	if err := os.RemoveAll(installPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing install: %w", err)
	}
	if err := copyDir(sourceBundle, installPath); err != nil {
		return fmt.Errorf("install app bundle: %w", err)
	}

	runOptional(ctx, root, "xattr", "-dr", "com.apple.quarantine", installPath)

	fmt.Fprintf(runner.Stdout, "PM Planner instalado em %s — abra pelo Finder ou Launchpad.\n", installPath)
	return nil
}
