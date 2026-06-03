package project

import (
	"os"
	"path/filepath"
	"testing"
)

func testProjectRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module pm-cli\n\ngo 1.23.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "desktop", "frontend"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "desktop", "wails.json"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCLIBuildCommandsDefaultWindowsOutput(t *testing.T) {
	root := testProjectRoot(t)

	commands, err := CLIBuildCommands(CLIBuildOptions{
		Root:   root,
		GOOS:   "windows",
		GOARCH: "amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 1 {
		t.Fatalf("commands: got %d", len(commands))
	}

	command := commands[0]
	if command.Dir != root {
		t.Fatalf("dir: got %q", command.Dir)
	}
	wantOutput := filepath.Join(root, "bin", "pm.exe")
	if got := command.Args[2]; got != wantOutput {
		t.Fatalf("output: got %q want %q", got, wantOutput)
	}
	if got := command.Args[len(command.Args)-1]; got != "./cmd/pm" {
		t.Fatalf("build target: got %q", got)
	}
	assertEnv(t, command.Env, "GOOS=windows")
	assertEnv(t, command.Env, "GOARCH=amd64")
}

func TestCLIBuildCommandsResolvesRelativeOutput(t *testing.T) {
	root := testProjectRoot(t)

	commands, err := CLIBuildCommands(CLIBuildOptions{
		Root:   root,
		Output: "out/pm-custom",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(root, "out", "pm-custom")
	if got := commands[0].Args[2]; got != want {
		t.Fatalf("output: got %q want %q", got, want)
	}
}

func TestDesktopBuildCommandsLinuxArchOverride(t *testing.T) {
	root := testProjectRoot(t)

	commands, err := DesktopBuildCommands(DesktopBuildOptions{
		Root: root,
		Host: HostFacts{
			GOOS:          "linux",
			RuntimeGOARCH: "386",
			GoEnvGOARCH:   "386",
			Machine:       "x86_64",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 3 {
		t.Fatalf("commands: got %d", len(commands))
	}

	build := commands[2]
	assertEnv(t, build.Env, "CGO_ENABLED=1")
	assertEnv(t, build.Env, "GOOS=linux")
	assertEnv(t, build.Env, "GOARCH=amd64")
	if got := build.Args[2]; got != "production,webkit2_41" {
		t.Fatalf("tags: got %q", got)
	}
}

func TestDesktopBuildCommandsCanForceHostArch(t *testing.T) {
	root := testProjectRoot(t)

	commands, err := DesktopBuildCommands(DesktopBuildOptions{
		Root:        root,
		ForceGoHost: true,
		Host: HostFacts{
			GOOS:          "linux",
			RuntimeGOARCH: "386",
			GoEnvGOARCH:   "386",
			Machine:       "x86_64",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEnv(t, commands[2].Env, "GOARCH=386")
}

func TestDesktopBuildCommandsSkipFrontend(t *testing.T) {
	root := testProjectRoot(t)

	commands, err := DesktopBuildCommands(DesktopBuildOptions{
		Root:         root,
		SkipFrontend: true,
		Host: HostFacts{
			GOOS:          "darwin",
			RuntimeGOARCH: "arm64",
			GoEnvGOARCH:   "arm64",
			Machine:       "arm64",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 1 {
		t.Fatalf("commands: got %d", len(commands))
	}
	command := commands[0]
	if command.Name != "wails" {
		t.Fatalf("command: got %q", command.Name)
	}
	assertArg(t, command.Args, "-s")
	assertArg(t, command.Args, "-skipbindings")
	assertArgPair(t, command.Args, "-platform", "darwin/arm64")
	assertArgPair(t, command.Args, "-o", "pm-desktop")
}

func TestWailsDesktopArtifactUsesPlatformBinaryName(t *testing.T) {
	root := testProjectRoot(t)

	got := WailsDesktopArtifact(root, "windows")
	want := filepath.Join(root, "desktop", "build", "bin", "pm-desktop.exe")
	if got != want {
		t.Fatalf("artifact: got %q want %q", got, want)
	}
}

func assertEnv(t *testing.T, env []string, want string) {
	t.Helper()
	for _, entry := range env {
		if entry == want {
			return
		}
	}
	t.Fatalf("env missing %q in %#v", want, env)
}

func assertArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, arg := range args {
		if arg == want {
			return
		}
	}
	t.Fatalf("arg missing %q in %#v", want, args)
}

func assertArgPair(t *testing.T, args []string, key string, wantValue string) {
	t.Helper()
	for index := 0; index < len(args)-1; index++ {
		if args[index] == key && args[index+1] == wantValue {
			return
		}
	}
	t.Fatalf("arg pair missing %q %q in %#v", key, wantValue, args)
}
