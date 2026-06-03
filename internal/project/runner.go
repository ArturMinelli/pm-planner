package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type CommandSpec struct {
	Dir  string
	Name string
	Args []string
	Env  []string
}

func (c CommandSpec) Display() string {
	parts := make([]string, 0, 1+len(c.Args))
	parts = append(parts, c.Name)
	parts = append(parts, c.Args...)
	return strings.Join(parts, " ")
}

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Env    []string
}

func (r Runner) RunAll(ctx context.Context, commands []CommandSpec) error {
	for _, command := range commands {
		if err := r.Run(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) Run(ctx context.Context, command CommandSpec) error {
	stdout := r.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := r.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if command.Dir != "" {
		fmt.Fprintf(stdout, "→ (%s) %s\n", command.Dir, command.Display())
	} else {
		fmt.Fprintf(stdout, "→ %s\n", command.Display())
	}

	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = mergeEnv(os.Environ(), append(r.Env, command.Env...))
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func mergeEnv(base []string, overrides []string) []string {
	merged := make([]string, 0, len(base)+len(overrides))
	positions := make(map[string]int, len(base)+len(overrides))

	add := func(entry string) {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			return
		}
		if existing, ok := positions[key]; ok {
			merged[existing] = entry
			return
		}
		positions[key] = len(merged)
		merged = append(merged, entry)
	}

	for _, entry := range base {
		add(entry)
	}
	for _, entry := range overrides {
		add(entry)
	}
	return merged
}
