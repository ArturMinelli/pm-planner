package update

import (
	"os"
	"path/filepath"
	"testing"

	"pm-cli/pkg/message"
)

func TestResultRoundTrip(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	written := &Result{
		OK: true,
		Message: message.New(message.KeyUpdateResultSuccessCommit, map[string]string{
			"commit": "a1b2c3d",
		}),
		LogPath:    "/tmp/update.log",
		Commit:     "a1b2c3d",
		FinishedAt: "2026-08-05T10:00:00-03:00",
	}
	if err := WriteResult(written); err != nil {
		t.Fatalf("WriteResult: %v", err)
	}

	read, err := ConsumeResult()
	if err != nil {
		t.Fatalf("ConsumeResult: %v", err)
	}
	if read == nil {
		t.Fatal("expected a stored result")
	}
	if read.OK != written.OK || read.Message.Key != written.Message.Key || read.LogPath != written.LogPath {
		t.Fatalf("read %+v, want %+v", *read, *written)
	}
}

func TestConsumeResultClearsTheFile(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	if err := WriteResult(&Result{OK: false, Message: message.New(message.KeyUpdateResultFailed, nil)}); err != nil {
		t.Fatalf("WriteResult: %v", err)
	}
	if _, err := ConsumeResult(); err != nil {
		t.Fatalf("first ConsumeResult: %v", err)
	}

	second, err := ConsumeResult()
	if err != nil {
		t.Fatalf("second ConsumeResult: %v", err)
	}
	if second != nil {
		t.Fatalf("expected the result to be consumed once, got %+v", second)
	}
}

func TestConsumeResultWithoutPendingUpdate(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	result, err := ConsumeResult()
	if err != nil {
		t.Fatalf("ConsumeResult: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUpdateScriptLivesInTheInstallRoot(t *testing.T) {
	root := t.TempDir()
	script := UpdateScript(root)

	if filepath.Dir(script) != filepath.Join(root, "scripts") {
		t.Fatalf("UpdateScript = %q, want it under %q", script, filepath.Join(root, "scripts"))
	}
}

func TestStagedScriptSurvivesTheSourceBeingRewritten(t *testing.T) {
	root := t.TempDir()
	source := UpdateScript(root)
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatalf("create scripts dir: %v", err)
	}
	if err := os.WriteFile(source, []byte("original\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	staged, cleanup, err := stageUpdateScript(root)
	if err != nil {
		t.Fatalf("stageUpdateScript: %v", err)
	}
	defer cleanup()

	if err := os.WriteFile(source, []byte("pulled from origin\n"), 0o755); err != nil {
		t.Fatalf("rewrite script: %v", err)
	}

	contents, err := os.ReadFile(staged)
	if err != nil {
		t.Fatalf("read staged script: %v", err)
	}
	if string(contents) != "original\n" {
		t.Fatalf("staged script = %q, want the copy taken before the update", contents)
	}

	cleanup()
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatalf("expected the staged copy to be removed, stat error: %v", err)
	}
}

func TestResolveInstallRootFindsTheProjectFromCwd(t *testing.T) {
	root := t.TempDir()
	writeProjectMarkers(t, root)
	chdir(t, root)

	resolved, err := ResolveInstallRoot()
	if err != nil {
		t.Fatalf("ResolveInstallRoot: %v", err)
	}
	if resolved != root {
		// macOS reports /private-prefixed temp dirs, so compare resolved paths.
		expected, _ := filepath.EvalSymlinks(root)
		actual, _ := filepath.EvalSymlinks(resolved)
		if expected != actual {
			t.Fatalf("ResolveInstallRoot = %q, want %q", resolved, root)
		}
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
}

func writeProjectMarkers(t *testing.T, root string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module pm-cli\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "desktop"), 0o755); err != nil {
		t.Fatalf("create desktop dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "desktop", "wails.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write wails.json: %v", err)
	}
}
