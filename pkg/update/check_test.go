package update

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeCommand struct {
	output string
	err    error
}

// fakeEnvironment builds an environment whose git calls are keyed by the first
// two arguments (for example "rev-parse --short").
func fakeEnvironment(root string, isGit bool, commands map[string]fakeCommand, missingTools ...string) environment {
	missing := make(map[string]bool, len(missingTools))
	for _, tool := range missingTools {
		missing[tool] = true
	}

	return environment{
		root:  root,
		isGit: isGit,
		run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			key := name
			if len(args) > 0 {
				key += " " + args[0]
			}
			command, ok := commands[key]
			if !ok {
				return "", errors.New("unexpected command: " + key + " " + strings.Join(args, " "))
			}
			return command.output, command.err
		},
		lookPath: func(name string) (string, error) {
			if missing[name] {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + name, nil
		},
	}
}

func gitCommands() map[string]fakeCommand {
	return map[string]fakeCommand{
		"git rev-parse": {output: "a1b2c3d\n"},
		"git log":       {output: "2026-08-05T10:00:00-03:00\n"},
		"git status":    {output: ""},
		"git fetch":     {output: ""},
		"git rev-list":  {output: "3\n"},
	}
}

func TestCheckReportsAvailableUpdate(t *testing.T) {
	status := check(context.Background(), fakeEnvironment("/root", true, gitCommands()))

	if !status.UpdateAvailable {
		t.Fatalf("expected update to be available, blockers: %v", status.Blockers)
	}
	if status.Behind != 3 {
		t.Fatalf("Behind = %d, want 3", status.Behind)
	}
	if status.CommitSHA != "a1b2c3d" {
		t.Fatalf("CommitSHA = %q, want a1b2c3d", status.CommitSHA)
	}
	if status.CommitDate != "2026-08-05T10:00:00-03:00" {
		t.Fatalf("CommitDate = %q", status.CommitDate)
	}
	if status.Dirty {
		t.Fatal("expected clean working tree")
	}
}

func TestCheckReportsUpToDate(t *testing.T) {
	commands := gitCommands()
	commands["git rev-list"] = fakeCommand{output: "0\n"}

	status := check(context.Background(), fakeEnvironment("/root", true, commands))

	if status.UpdateAvailable {
		t.Fatal("expected no update to be available")
	}
	if len(status.Blockers) != 0 {
		t.Fatalf("unexpected blockers: %v", status.Blockers)
	}
}

func TestCheckBlocksOnDirtyWorkingTree(t *testing.T) {
	commands := gitCommands()
	commands["git status"] = fakeCommand{output: " M scripts/update.sh\n"}

	status := check(context.Background(), fakeEnvironment("/root", true, commands))

	if !status.Dirty {
		t.Fatal("expected dirty working tree")
	}
	if status.UpdateAvailable {
		t.Fatal("dirty tree must not offer an update")
	}
	if len(status.Blockers) != 1 || !strings.Contains(status.Blockers[0], "alterações locais") {
		t.Fatalf("blockers = %v", status.Blockers)
	}
}

func TestCheckBlocksOnMissingBuildTools(t *testing.T) {
	status := check(context.Background(), fakeEnvironment("/root", true, gitCommands(), "go", "node"))

	if len(status.Blockers) != 2 {
		t.Fatalf("blockers = %v, want one per missing tool", status.Blockers)
	}
	if status.UpdateAvailable {
		t.Fatal("missing build tools must not offer an update")
	}
}

func TestCheckBlocksWhenRemoteIsUnreachable(t *testing.T) {
	commands := gitCommands()
	commands["git fetch"] = fakeCommand{err: errors.New("could not resolve host: github.com\nfatal: unable to access")}

	status := check(context.Background(), fakeEnvironment("/root", true, commands))

	if status.UpdateAvailable {
		t.Fatal("unreachable remote must not offer an update")
	}
	if len(status.Blockers) != 1 || !strings.Contains(status.Blockers[0], "could not resolve host") {
		t.Fatalf("blockers = %v", status.Blockers)
	}
	if strings.Contains(status.Blockers[0], "\n") {
		t.Fatalf("blocker should be a single line: %q", status.Blockers[0])
	}
}

func TestCheckOffersUpdateForNonGitInstall(t *testing.T) {
	status := check(context.Background(), fakeEnvironment("/root", false, nil))

	if !status.UpdateAvailable {
		t.Fatalf("tarball installs should always offer an update, blockers: %v", status.Blockers)
	}
	if status.CommitSHA != "" || status.Behind != 0 {
		t.Fatalf("no commit information expected for tarball installs: %+v", status)
	}
}

func TestParseCount(t *testing.T) {
	cases := map[string]int{
		"7\n":   7,
		"  0  ": 0,
		"":      0,
		"abc":   0,
		"-2":    0,
	}
	for input, want := range cases {
		if got := parseCount(input); got != want {
			t.Errorf("parseCount(%q) = %d, want %d", input, got, want)
		}
	}
}
