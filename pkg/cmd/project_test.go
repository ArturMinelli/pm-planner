package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProjectCommandTree(t *testing.T) {
	command := newProjectCommand()

	assertCommand(t, command, "doctor")
	assertCommand(t, command, "clean")

	build := assertCommand(t, command, "build")
	buildCLI := assertCommand(t, build, "cli")
	assertFlag(t, buildCLI, "output")
	assertFlag(t, buildCLI, "goos")
	assertFlag(t, buildCLI, "goarch")

	buildDesktop := assertCommand(t, build, "desktop")
	assertFlag(t, buildDesktop, "output")
	assertFlag(t, buildDesktop, "skip-frontend")
	assertFlag(t, buildDesktop, "force-go-host")

	dev := assertCommand(t, command, "dev")
	assertCommand(t, dev, "desktop")

	install := assertCommand(t, command, "install")
	desktopMenu := assertCommand(t, install, "desktop-menu")
	assertFlag(t, desktopMenu, "skip-build")

	desktopInstall := assertCommand(t, install, "desktop")
	assertFlag(t, desktopInstall, "skip-build")
	assertFlag(t, desktopInstall, "system")
	assertFlag(t, desktopInstall, "desktop-shortcut")
}

func assertCommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, command := range parent.Commands() {
		if command.Name() == name {
			return command
		}
	}
	t.Fatalf("missing command %q under %q", name, parent.Name())
	return nil
}

func assertFlag(t *testing.T, command *cobra.Command, name string) {
	t.Helper()
	if command.Flags().Lookup(name) == nil {
		t.Fatalf("missing flag %q on %q", name, command.Name())
	}
}
