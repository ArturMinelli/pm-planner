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
	"slices"
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

// resolveToolPath finds a build tool on PATH or in known install locations.
// GUI launches inherit a minimal PATH that often excludes Go and Node.
func resolveToolPath(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	binary := project.BinaryName(name, runtime.GOOS)
	for _, dir := range toolDirs(name) {
		candidate := filepath.Join(dir, binary)
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}
	return "", exec.ErrNotFound
}

func toolDirs(name string) []string {
	switch name {
	case "go":
		return goToolchainDirs()
	case "node":
		return nodeBinDirs()
	default:
		return nil
	}
}

// goToolchainDirs lists directories that may contain the go compiler itself
// (distinct from GOPATH/bin, which holds go-installed tools like pm).
func goToolchainDirs() []string {
	dirs := make([]string, 0, 8)
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "go", "bin"))
	}
	dirs = append(dirs, "/usr/local/go/bin")
	dirs = append(dirs, commonBinDirs()...)
	return dirs
}

// nodeBinDirs lists directories that may contain the node binary, including
// version managers whose install paths are outside a typical GUI PATH.
func nodeBinDirs() []string {
	dirs := make([]string, 0, 16)
	dirs = append(dirs, versionManagerNodeBins(
		envOrHome("NVM_DIR", ".nvm"),
		filepath.Join("versions", "node"),
		"bin",
	)...)
	dirs = append(dirs, versionManagerNodeBins(
		envOrHome("FNM_DIR", filepath.Join(".local", "share", "fnm")),
		"node-versions",
		filepath.Join("installation", "bin"),
	)...)
	dirs = append(dirs, commonBinDirs()...)
	return dirs
}

func envOrHome(envKey, homeRelative string) string {
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, homeRelative)
}

// versionManagerNodeBins returns bin dirs under a version-manager root,
// newest version first so resolveToolPath prefers a current install.
func versionManagerNodeBins(root, versionsRel, binRel string) []string {
	if root == "" {
		return nil
	}
	versionsDir := filepath.Join(root, versionsRel)
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	slices.Reverse(names)

	dirs := make([]string, 0, len(names))
	for _, name := range names {
		dirs = append(dirs, filepath.Join(versionsDir, name, binRel))
	}
	return dirs
}

func commonBinDirs() []string {
	dirs := []string{"/usr/local/bin", "/usr/bin", "/snap/bin"}
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, "/opt/homebrew/bin")
	default:
		dirs = append(dirs, "/home/linuxbrew/.linuxbrew/bin")
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	return dirs
}

// augmentedEnv returns the process environment with known Go/Node bin
// directories prepended to PATH so update subprocesses inherit a usable PATH.
func augmentedEnv() []string {
	extra := append(goToolchainDirs(), nodeBinDirs()...)
	extra = append(extra, goBinDirs()...)
	path := prependPath(os.Getenv("PATH"), extra...)
	return replaceEnv(os.Environ(), "PATH", path)
}

func prependPath(current string, dirs ...string) string {
	seen := make(map[string]bool)
	parts := make([]string, 0, len(dirs)+1)
	for _, dir := range dirs {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		parts = append(parts, dir)
	}
	if current != "" {
		for _, dir := range filepath.SplitList(current) {
			if dir == "" || seen[dir] {
				continue
			}
			seen[dir] = true
			parts = append(parts, dir)
		}
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

func replaceEnv(environ []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(environ)+1)
	replaced := false
	for _, entry := range environ {
		if strings.HasPrefix(entry, prefix) {
			out = append(out, prefix+value)
			replaced = true
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
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
