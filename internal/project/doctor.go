package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Doctor(ctx context.Context, root string, out io.Writer) error {
	root, err := ResolveRoot(root)
	if err != nil {
		return err
	}
	if out == nil {
		out = io.Discard
	}

	facts := DetectHostFacts(ctx, root)
	fmt.Fprintln(out, "PM project doctor")
	fmt.Fprintf(out, "Root: %s\n", root)
	fmt.Fprintf(out, "Host: %s/%s", facts.GOOS, facts.GoEnvGOARCH)
	if facts.Machine != "" && facts.Machine != facts.GoEnvGOARCH {
		fmt.Fprintf(out, " on %s", facts.Machine)
	}
	fmt.Fprintln(out)

	var failed []string
	checkTool := func(required bool, label string, name string, args ...string) {
		detail, ok := toolVersion(ctx, root, name, args...)
		marker := "✓"
		if !ok {
			marker = "!"
			detail = "not found"
			if required {
				failed = append(failed, label)
			}
		}
		fmt.Fprintf(out, "%s %s: %s\n", marker, label, detail)
	}

	checkTool(true, "Go", "go", "version")
	checkTool(true, "Node.js", "node", "--version")
	checkTool(true, "npm", "npm", "--version")
	checkTool(false, "Wails CLI for dev", "wails", "version")

	if facts.GOOS == "linux" {
		checkPkgConfig := func(required bool, label string, packages ...string) {
			ok := false
			for _, pkg := range packages {
				if pkgConfigExists(ctx, root, pkg) {
					fmt.Fprintf(out, "✓ %s: %s\n", label, pkg)
					ok = true
					break
				}
			}
			if !ok {
				fmt.Fprintf(out, "! %s: missing %s\n", label, strings.Join(packages, " or "))
				if required {
					failed = append(failed, label)
				}
			}
		}
		checkTool(true, "pkg-config", "pkg-config", "--version")
		checkPkgConfig(true, "GTK development headers", "gtk+-3.0")
		checkPkgConfig(true, "WebKitGTK development headers", "webkit2gtk-4.1", "webkit2gtk-4.0")
	}

	if len(failed) > 0 {
		return fmt.Errorf("doctor found missing required tools: %s", strings.Join(failed, ", "))
	}
	return nil
}

func toolVersion(ctx context.Context, dir string, name string, args ...string) (string, bool) {
	if _, err := exec.LookPath(name); err != nil {
		return "", false
	}
	out := commandOutput(ctx, dir, name, args...)
	if strings.TrimSpace(out) == "" {
		return "available", true
	}
	return firstLine(out), true
}

func pkgConfigExists(ctx context.Context, dir string, pkg string) bool {
	if _, err := exec.LookPath("pkg-config"); err != nil {
		return false
	}
	cmd := exec.CommandContext(ctx, "pkg-config", "--exists", pkg)
	cmd.Dir = dir
	return cmd.Run() == nil
}

func EnsureTool(name string, installHint string) error {
	if _, err := exec.LookPath(name); err == nil {
		return nil
	}
	if installHint == "" {
		return fmt.Errorf("%s was not found on PATH", name)
	}
	return errors.New(installHint)
}
