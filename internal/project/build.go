package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type HostFacts struct {
	GOOS          string
	RuntimeGOARCH string
	GoEnvGOARCH   string
	Machine       string
}

type CLIBuildOptions struct {
	Root   string
	Output string
	GOOS   string
	GOARCH string
}

type DesktopBuildOptions struct {
	Root         string
	Output       string
	SkipFrontend bool
	ForceGoHost  bool
	Host         HostFacts
}

func DetectHostFacts(ctx context.Context, root string) HostFacts {
	return HostFacts{
		GOOS:          runtime.GOOS,
		RuntimeGOARCH: runtime.GOARCH,
		GoEnvGOARCH:   firstLine(commandOutput(ctx, root, "go", "env", "GOARCH")),
		Machine:       detectMachine(ctx, root),
	}
}

func CLIBuildCommands(options CLIBuildOptions) ([]CommandSpec, error) {
	root, err := ResolveRoot(options.Root)
	if err != nil {
		return nil, err
	}

	targetGOOS := options.GOOS
	if targetGOOS == "" {
		targetGOOS = runtime.GOOS
	}
	output := options.Output
	if output == "" {
		output = DefaultCLIOutput(root, targetGOOS)
	} else {
		output = ResolveOutputPath(root, output)
	}

	env := make([]string, 0, 2)
	if options.GOOS != "" {
		env = append(env, "GOOS="+options.GOOS)
	}
	if options.GOARCH != "" {
		env = append(env, "GOARCH="+options.GOARCH)
	}

	return []CommandSpec{
		{
			Dir:  root,
			Name: "go",
			Args: []string{"build", "-o", output, "./cmd/pm"},
			Env:  env,
		},
	}, nil
}

func DesktopBuildCommands(options DesktopBuildOptions) ([]CommandSpec, error) {
	root, err := ResolveRoot(options.Root)
	if err != nil {
		return nil, err
	}

	host := normalizedHostFacts(options.Host)
	if shouldUseWailsBuild(host) {
		args := []string{
			"build",
			"-nopackage",
			"-nosyncgomod",
			"-skipbindings",
			"-m",
			"-nocolour",
			"-platform", desktopPlatform(host),
			"-o", BinaryName("pm-desktop", host.GOOS),
		}
		if options.SkipFrontend {
			args = append(args, "-s")
		}
		return []CommandSpec{
			{
				Dir:  ResolveOutputPath(root, "desktop"),
				Name: "wails",
				Args: args,
			},
		}, nil
	}

	output := options.Output
	if output == "" {
		output = DefaultDesktopOutput(root, host.GOOS)
	} else {
		output = ResolveOutputPath(root, output)
	}

	commands := make([]CommandSpec, 0, 3)
	frontendDir := ResolveOutputPath(root, "desktop/frontend")
	if !options.SkipFrontend {
		commands = append(commands,
			CommandSpec{
				Dir:  frontendDir,
				Name: "npm",
				Args: []string{"ci"},
			},
			CommandSpec{
				Dir:  frontendDir,
				Name: "npm",
				Args: []string{"run", "build"},
			},
		)
	}

	tags := []string{"production"}
	env := make([]string, 0, 3)
	if host.GOOS == "linux" {
		tags = append(tags, "webkit2_41")
		env = append(env, "CGO_ENABLED=1", "GOOS=linux")
		if goarch := EffectiveLinuxDesktopGOARCH(host, options.ForceGoHost); goarch != "" {
			env = append(env, "GOARCH="+goarch)
		}
	}

	commands = append(commands, CommandSpec{
		Dir:  ResolveOutputPath(root, "desktop"),
		Name: "go",
		Args: []string{"build", "-tags", strings.Join(tags, ","), "-o", output, "."},
		Env:  env,
	})

	return commands, nil
}

func BuildDesktop(ctx context.Context, options DesktopBuildOptions, runner Runner) error {
	root, err := ResolveRoot(options.Root)
	if err != nil {
		return err
	}
	host := normalizedHostFacts(options.Host)
	output := options.Output
	if output == "" {
		output = DefaultDesktopOutput(root, host.GOOS)
	} else {
		output = ResolveOutputPath(root, output)
	}

	commands, err := DesktopBuildCommands(DesktopBuildOptions{
		Root:         root,
		Output:       output,
		SkipFrontend: options.SkipFrontend,
		ForceGoHost:  options.ForceGoHost,
		Host:         host,
	})
	if err != nil {
		return err
	}
	if shouldUseWailsBuild(host) {
		if err := EnsureTool("wails", "wails CLI was not found on PATH; install it with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`"); err != nil {
			return err
		}
	}
	if err := runner.RunAll(ctx, commands); err != nil {
		return err
	}
	if !shouldUseWailsBuild(host) {
		return nil
	}

	source := WailsDesktopArtifact(root, host.GOOS)
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if err := copyFile(source, output, 0755); err != nil {
		return fmt.Errorf("copy Wails artifact: %w", err)
	}
	return nil
}

func EffectiveLinuxDesktopGOARCH(host HostFacts, forceGoHost bool) string {
	goarch := strings.TrimSpace(host.GoEnvGOARCH)
	if goarch == "" {
		goarch = strings.TrimSpace(host.RuntimeGOARCH)
	}
	if forceGoHost {
		return goarch
	}
	if goarch == "386" && isX8664Machine(host.Machine) {
		return "amd64"
	}
	return goarch
}

func normalizedHostFacts(host HostFacts) HostFacts {
	if host.GOOS == "" {
		host.GOOS = runtime.GOOS
	}
	if host.RuntimeGOARCH == "" {
		host.RuntimeGOARCH = runtime.GOARCH
	}
	if host.GoEnvGOARCH == "" {
		host.GoEnvGOARCH = host.RuntimeGOARCH
	}
	if host.Machine == "" {
		host.Machine = host.RuntimeGOARCH
	}
	return host
}

func shouldUseWailsBuild(host HostFacts) bool {
	return host.GOOS == "darwin" || host.GOOS == "windows"
}

func desktopPlatform(host HostFacts) string {
	return host.GOOS + "/" + desktopGOARCH(host)
}

func desktopGOARCH(host HostFacts) string {
	goarch := strings.TrimSpace(host.GoEnvGOARCH)
	if goarch == "" {
		goarch = strings.TrimSpace(host.RuntimeGOARCH)
	}
	return goarch
}

func WailsDesktopArtifact(root string, goos string) string {
	return filepath.Join(root, "desktop", "build", "bin", BinaryName("pm-desktop", goos))
}

func detectMachine(ctx context.Context, root string) string {
	if runtime.GOOS == "windows" {
		return runtime.GOARCH
	}
	out := commandOutput(ctx, root, "uname", "-m")
	if strings.TrimSpace(out) == "" {
		return runtime.GOARCH
	}
	return firstLine(out)
}

func isX8664Machine(machine string) bool {
	switch strings.ToLower(strings.TrimSpace(machine)) {
	case "x86_64", "amd64":
		return true
	default:
		return false
	}
}

func commandOutput(ctx context.Context, dir string, name string, args ...string) string {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	line, _, _ := strings.Cut(s, "\n")
	return strings.TrimSpace(line)
}
