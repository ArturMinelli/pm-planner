package update

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const remoteBranch = "origin/main"

// Status describes what an update would do, and what stands in its way.
type Status struct {
	Root string `json:"root"`
	// IsGit is false for tarball installs, where no commit comparison is possible.
	IsGit           bool     `json:"isGit"`
	CommitSHA       string   `json:"commitSha"`
	CommitDate      string   `json:"commitDate"`
	Behind          int      `json:"behind"`
	Dirty           bool     `json:"dirty"`
	Blockers        []string `json:"blockers"`
	UpdateAvailable bool     `json:"updateAvailable"`
}

// environment isolates Check from the host so its decisions can be tested.
type environment struct {
	root     string
	isGit    bool
	run      func(ctx context.Context, dir string, name string, args ...string) (string, error)
	lookPath func(name string) (string, error)
}

// Check inspects the installation and reports whether an update can be applied.
// User-actionable problems are returned as blockers rather than errors, so the
// caller can always render a status.
func Check(ctx context.Context) (*Status, error) {
	root, err := ResolveInstallRoot()
	if err != nil {
		return &Status{
			Blockers: []string{"Instalação do PM Planner não encontrada no disco. Execute o setup novamente."},
		}, nil
	}

	return check(ctx, environment{
		root:     root,
		isGit:    isGitCheckout(root),
		run:      runCapture,
		lookPath: resolveToolPath,
	}), nil
}

func check(ctx context.Context, env environment) *Status {
	status := &Status{Root: env.root, IsGit: env.isGit}

	for _, tool := range []struct{ name, reason string }{
		{"go", "Go não encontrado no PATH — necessário para compilar o PM Planner."},
		{"node", "Node.js não encontrado no PATH — necessário para compilar o frontend."},
	} {
		if _, err := env.lookPath(tool.name); err != nil {
			status.Blockers = append(status.Blockers, tool.reason)
		}
	}

	if !env.isGit {
		status.UpdateAvailable = len(status.Blockers) == 0
		return status
	}

	if sha, err := env.run(ctx, env.root, "git", "rev-parse", "--short", "HEAD"); err == nil {
		status.CommitSHA = strings.TrimSpace(sha)
	}
	if date, err := env.run(ctx, env.root, "git", "log", "-1", "--format=%cI"); err == nil {
		status.CommitDate = strings.TrimSpace(date)
	}

	if dirty, err := env.run(ctx, env.root, "git", "status", "--porcelain"); err == nil {
		status.Dirty = strings.TrimSpace(dirty) != ""
		if status.Dirty {
			status.Blockers = append(status.Blockers, fmt.Sprintf(
				"Há alterações locais não commitadas em %s. Faça commit ou stash antes de atualizar.", env.root))
		}
	}

	if _, err := env.run(ctx, env.root, "git", "fetch", "origin", "--prune"); err != nil {
		status.Blockers = append(status.Blockers, "Não foi possível consultar o repositório remoto: "+firstLine(err.Error()))
		return status
	}

	behind, err := env.run(ctx, env.root, "git", "rev-list", "--count", "HEAD.."+remoteBranch)
	if err != nil {
		status.Blockers = append(status.Blockers, "Não foi possível comparar com "+remoteBranch+": "+firstLine(err.Error()))
		return status
	}
	status.Behind = parseCount(behind)
	status.UpdateAvailable = len(status.Blockers) == 0 && status.Behind > 0
	return status
}

func parseCount(output string) int {
	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil || count < 0 {
		return 0
	}
	return count
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(value), "\n")
	return line
}

func isGitCheckout(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func runCapture(ctx context.Context, dir string, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%s", message)
	}
	return stdout.String(), nil
}
