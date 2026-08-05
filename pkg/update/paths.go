// Package update resolves, inspects and applies updates of the local
// pm-planner installation. It backs both the `pm update` CLI command and the
// update card in the desktop Settings page.
package update

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"pm-cli/internal/project"
)

const (
	cacheDirName   = "pm"
	resultFileName = "update-result.json"
	logFileName    = "update.log"
)

// ErrRootNotFound is returned when no pm-planner source tree can be located.
var ErrRootNotFound = errors.New("instalação do PM Planner não encontrada")

// ResolveInstallRoot locates the source tree that backs this installation.
// The desktop app is launched with an arbitrary working directory, so the
// walk-up search used by the CLI is only the first of several candidates.
func ResolveInstallRoot() (string, error) {
	if wd, err := os.Getwd(); err == nil {
		if root, err := project.FindRoot(wd); err == nil {
			return root, nil
		}
	}

	if executable, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		if root, err := project.FindRoot(filepath.Dir(executable)); err == nil {
			return root, nil
		}
	}

	for _, candidate := range installDirCandidates() {
		if root, err := project.FindRoot(candidate); err == nil && root == candidate {
			return root, nil
		}
	}

	return "", ErrRootNotFound
}

// installDirCandidates lists the directories setup.sh/setup.ps1 install into,
// most recent convention first, ending with the legacy ~/pm-planner location.
func installDirCandidates() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := make([]string, 0, 3)
	switch runtime.GOOS {
	case "darwin":
		candidates = append(candidates, filepath.Join(home, "Library", "Application Support", "pm-planner"))
	case "windows":
		candidates = append(candidates, filepath.Join(home, "pm-planner"))
	default:
		dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		candidates = append(candidates, filepath.Join(dataHome, "pm-planner"))
	}
	legacy := filepath.Join(home, "pm-planner")
	if len(candidates) == 0 || candidates[0] != legacy {
		candidates = append(candidates, legacy)
	}
	return candidates
}

// UpdateScript returns the platform update script inside root.
func UpdateScript(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "scripts", "update.ps1")
	}
	return filepath.Join(root, "scripts", "update.sh")
}

// updateScriptCommand returns the interpreter invocation for an update script.
func updateScriptCommand(script string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{
			"-NoProfile", "-NonInteractive",
			"-ExecutionPolicy", "Bypass",
			"-File", script,
		}
	}
	return "bash", []string{script}
}

// CacheDir is where the update log and the hand-off result file are kept.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, cacheDirName), nil
}

// LogPath is the file the update script output is written to.
func LogPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
}

// ResultPath is the file an applied update leaves behind for the relaunched app.
func ResultPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, resultFileName), nil
}

// ResolvePMBinary finds the pm CLI. A GUI launch on Linux inherits a minimal
// PATH that usually excludes the Go bin directories, so PATH alone is not enough.
func ResolvePMBinary() (string, error) {
	name := project.BinaryName("pm", runtime.GOOS)
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	for _, dir := range goBinDirs() {
		candidate := filepath.Join(dir, name)
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	if root, err := ResolveInstallRoot(); err == nil {
		candidate := project.DefaultCLIOutput(root, runtime.GOOS)
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	return "", errors.New("binário pm não encontrado no PATH nem nos diretórios padrão do Go")
}

func goBinDirs() []string {
	dirs := make([]string, 0, 3)
	if gopath := strings.TrimSpace(os.Getenv("GOPATH")); gopath != "" {
		for _, entry := range filepath.SplitList(gopath) {
			dirs = append(dirs, filepath.Join(entry, "bin"))
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, "go", "bin"),
			filepath.Join(home, ".local", "go", "bin"),
		)
	}
	return dirs
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}

// DesktopLaunchCommand returns how to start the installed desktop app.
func DesktopLaunchCommand() (string, []string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}

	switch runtime.GOOS {
	case "darwin":
		bundle := filepath.Join(home, "Applications", "PM Planner.app")
		if _, err := os.Stat(bundle); err == nil {
			return "open", []string{"-a", bundle}, nil
		}
		binary := filepath.Join(bundle, "Contents", "MacOS", "pm-desktop")
		if isExecutableFile(binary) {
			return binary, nil, nil
		}
	case "windows":
		localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
		if localAppData != "" {
			binary := filepath.Join(localAppData, "Programs", "PM Planner", "pm-desktop.exe")
			if isExecutableFile(binary) {
				return binary, nil, nil
			}
		}
	default:
		binHome := strings.TrimSpace(os.Getenv("XDG_BIN_HOME"))
		if binHome == "" {
			binHome = filepath.Join(home, ".local", "bin")
		}
		binary := filepath.Join(binHome, "pm-desktop")
		if isExecutableFile(binary) {
			return binary, nil, nil
		}
	}

	return "", nil, errors.New("app desktop instalado não encontrado")
}
