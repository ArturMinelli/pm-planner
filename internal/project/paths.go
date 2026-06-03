package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const moduleName = "pm-cli"

func ResolveRoot(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		root, err := filepath.Abs(filepath.Clean(override))
		if err != nil {
			return "", err
		}
		if !isProjectRoot(root) {
			return "", fmt.Errorf("%s is not the pm-planner project root", root)
		}
		return root, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindRoot(wd)
}

func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(filepath.Clean(start))
	if err != nil {
		return "", err
	}

	for {
		if isProjectRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find pm-planner project root")
		}
		dir = parent
	}
}

func isProjectRoot(dir string) bool {
	modulePath := filepath.Join(dir, "go.mod")
	moduleBytes, err := os.ReadFile(modulePath)
	if err != nil {
		return false
	}
	if !declaresModule(moduleBytes, moduleName) {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "desktop", "wails.json")); err != nil {
		return false
	}
	return true
}

func declaresModule(contents []byte, want string) bool {
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.TrimSpace(line) == "module "+want {
			return true
		}
	}
	return false
}

func ResolveOutputPath(root string, output string) string {
	if filepath.IsAbs(output) {
		return filepath.Clean(output)
	}
	return filepath.Join(root, filepath.Clean(output))
}

func BinaryName(base string, goos string) string {
	if goos == "" {
		goos = runtime.GOOS
	}
	if goos == "windows" {
		return base + ".exe"
	}
	return base
}

func DefaultCLIOutput(root string, goos string) string {
	return filepath.Join(root, "bin", BinaryName("pm", goos))
}

func DefaultDesktopOutput(root string, goos string) string {
	return filepath.Join(root, "bin", BinaryName("pm-desktop", goos))
}

func CleanTargets(root string) []string {
	return []string{
		filepath.Join(root, "bin"),
		filepath.Join(root, "desktop", "frontend", "dist"),
		filepath.Join(root, "desktop", "frontend", "node_modules", ".tmp"),
		filepath.Join(root, "main"),
		filepath.Join(root, "pm"),
		filepath.Join(root, "pm.exe"),
		filepath.Join(root, "pm-desktop"),
		filepath.Join(root, "pm-desktop.exe"),
	}
}
