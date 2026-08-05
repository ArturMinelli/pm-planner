package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result is the outcome an applied update leaves behind for the relaunched app.
type Result struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	LogPath    string `json:"logPath"`
	Commit     string `json:"commit"`
	FinishedAt string `json:"finishedAt"`
}

// ApplyOptions tunes how the update script is run.
type ApplyOptions struct {
	// Output mirrors the script output for interactive callers. May be nil.
	Output io.Writer
	// Relaunch starts the desktop app again once the update finishes.
	Relaunch bool
}

// Apply runs the platform update script, records the outcome for the next app
// start, and optionally relaunches the desktop app. The returned error covers
// only failures to run the script at all; a failed update is reported through
// the Result so the user still sees it after the relaunch.
func Apply(ctx context.Context, options ApplyOptions) (*Result, error) {
	root, err := ResolveInstallRoot()
	if err != nil {
		return nil, err
	}

	logPath, err := LogPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()

	var output io.Writer = logFile
	if options.Output != nil {
		output = io.MultiWriter(logFile, options.Output)
	}

	fmt.Fprintf(output, "=== PM Planner update: %s ===\n", time.Now().Format(time.RFC3339))

	// The script pulls new code over itself, and the interpreter reads it
	// incrementally, so run a copy taken before the source tree changes. Both
	// scripts locate the project from the working directory, not from their own
	// path, so running from a temporary location is safe.
	script, cleanup, err := stageUpdateScript(root)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	name, args := updateScriptCommand(script)
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = root
	command.Env = augmentedEnv()
	command.Stdout = output
	command.Stderr = output
	runErr := command.Run()

	result := &Result{
		OK:         runErr == nil,
		LogPath:    logPath,
		Commit:     currentCommit(ctx, root),
		FinishedAt: time.Now().Format(time.RFC3339),
	}
	if runErr == nil {
		result.Message = "PM Planner atualizado com sucesso."
		if result.Commit != "" {
			result.Message = "PM Planner atualizado para " + result.Commit + "."
		}
	} else {
		result.Message = "A atualização falhou: " + firstLine(runErr.Error()) + ". Detalhes em " + logPath
	}

	if err := WriteResult(result); err != nil {
		fmt.Fprintf(output, "aviso: não foi possível gravar o resultado da atualização: %v\n", err)
	}

	if options.Relaunch {
		if err := LaunchDesktop(); err != nil {
			fmt.Fprintf(output, "aviso: não foi possível reabrir o app desktop: %v\n", err)
		}
	}

	return result, nil
}

// LaunchDesktop starts the installed desktop app detached from this process.
func LaunchDesktop() error {
	name, args, err := DesktopLaunchCommand()
	if err != nil {
		return err
	}
	return SpawnDetached("", name, args...)
}

// WriteResult stores the outcome for the next app start to pick up.
func WriteResult(result *Result) error {
	path, err := ResultPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o644)
}

// ConsumeResult reads and deletes the pending update outcome. It returns nil
// when no update has run since the last read.
func ConsumeResult() (*Result, error) {
	path, err := ResultPath()
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	_ = os.Remove(path)

	result := &Result{}
	if err := json.Unmarshal(contents, result); err != nil {
		return nil, err
	}
	return result, nil
}

// stageUpdateScript copies the update script somewhere the update itself will
// not overwrite. The returned cleanup removes the copy.
func stageUpdateScript(root string) (string, func(), error) {
	source := UpdateScript(root)
	contents, err := os.ReadFile(source)
	if err != nil {
		return "", nil, err
	}

	directory, err := os.MkdirTemp("", "pm-update-")
	if err != nil {
		return "", nil, err
	}
	staged := filepath.Join(directory, filepath.Base(source))
	if err := os.WriteFile(staged, contents, 0o755); err != nil {
		_ = os.RemoveAll(directory)
		return "", nil, err
	}
	return staged, func() { _ = os.RemoveAll(directory) }, nil
}

func currentCommit(ctx context.Context, root string) string {
	if !isGitCheckout(root) {
		return ""
	}
	sha, err := runCapture(ctx, root, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(sha)
}
