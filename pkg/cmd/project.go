package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"pm-cli/internal/project"
)

func newProjectCommand() *cobra.Command {
	var root string

	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Maintain and build this project",
		Long:  "Project maintenance commands for building, checking, installing, and cleaning pm-planner.",
	}
	projectCmd.PersistentFlags().StringVar(&root, "root", "", "Project root (default: auto-detect)")

	projectCmd.AddCommand(newProjectDoctorCommand(&root))
	projectCmd.AddCommand(newProjectBuildCommand(&root))
	projectCmd.AddCommand(newProjectDevCommand(&root))
	projectCmd.AddCommand(newProjectInstallCommand(&root))
	projectCmd.AddCommand(newProjectCleanCommand(&root))
	return projectCmd
}

func newProjectDoctorCommand(root *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check local build dependencies",
		RunE: func(command *cobra.Command, args []string) error {
			return project.Doctor(command.Context(), *root, command.OutOrStdout())
		},
	}
}

func newProjectBuildCommand(root *string) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build project artifacts",
	}
	buildCmd.AddCommand(newProjectBuildCLICommand(root))
	buildCmd.AddCommand(newProjectBuildDesktopCommand(root))
	return buildCmd
}

func newProjectBuildCLICommand(root *string) *cobra.Command {
	var output string
	var targetGOOS string
	var targetGOARCH string

	command := &cobra.Command{
		Use:   "cli",
		Short: "Build the pm CLI",
		RunE: func(command *cobra.Command, args []string) error {
			commands, err := project.CLIBuildCommands(project.CLIBuildOptions{
				Root:   *root,
				Output: output,
				GOOS:   targetGOOS,
				GOARCH: targetGOARCH,
			})
			if err != nil {
				return err
			}
			return runnerFor(command).RunAll(command.Context(), commands)
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "Output binary path (default: bin/pm)")
	command.Flags().StringVar(&targetGOOS, "goos", "", "Target GOOS for CLI cross-builds")
	command.Flags().StringVar(&targetGOARCH, "goarch", "", "Target GOARCH for CLI cross-builds")
	return command
}

func newProjectBuildDesktopCommand(root *string) *cobra.Command {
	var output string
	var skipFrontend bool
	var forceGoHost bool

	command := &cobra.Command{
		Use:   "desktop",
		Short: "Build the Wails desktop app for this platform",
		RunE: func(command *cobra.Command, args []string) error {
			resolvedRoot, err := project.ResolveRoot(*root)
			if err != nil {
				return err
			}
			return project.BuildDesktop(command.Context(), project.DesktopBuildOptions{
				Root:         resolvedRoot,
				Output:       output,
				SkipFrontend: skipFrontend,
				ForceGoHost:  forceGoHost,
				Host:         project.DetectHostFacts(command.Context(), resolvedRoot),
			}, runnerFor(command))
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "Output binary path (default: bin/pm-desktop)")
	command.Flags().BoolVar(&skipFrontend, "skip-frontend", false, "Skip npm install/build and only build the Go desktop shell")
	command.Flags().BoolVar(&forceGoHost, "force-go-host", false, "Do not override linux/386 Go on x86_64 Linux")
	return command
}

func newProjectDevCommand(root *string) *cobra.Command {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Run development workflows",
	}
	devCmd.AddCommand(&cobra.Command{
		Use:   "desktop",
		Short: "Run Wails desktop dev mode",
		RunE: func(command *cobra.Command, args []string) error {
			if err := project.EnsureTool("wails", "wails CLI was not found on PATH; install it with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`"); err != nil {
				return err
			}
			resolvedRoot, err := project.ResolveRoot(*root)
			if err != nil {
				return err
			}
			return runnerFor(command).Run(command.Context(), project.CommandSpec{
				Dir:  filepath.Join(resolvedRoot, "desktop"),
				Name: "wails",
				Args: []string{"dev"},
			})
		},
	})
	return devCmd
}

func newProjectInstallCommand(root *string) *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install project artifacts",
	}

	var skipBuildMenu bool
	desktopMenuCmd := &cobra.Command{
		Use:   "desktop-menu",
		Short: "Install the Linux desktop menu entry for the current user (alias for install desktop)",
		RunE: func(command *cobra.Command, args []string) error {
			return project.InstallDesktopMenu(command.Context(), project.DesktopMenuOptions{
				Root:      *root,
				SkipBuild: skipBuildMenu,
			}, runnerFor(command))
		},
	}
	desktopMenuCmd.Flags().BoolVar(&skipBuildMenu, "skip-build", false, "Install existing bin/pm-desktop without rebuilding")
	installCmd.AddCommand(desktopMenuCmd)
	installCmd.AddCommand(newProjectInstallDesktopCommand(root))
	return installCmd
}

func newProjectInstallDesktopCommand(root *string) *cobra.Command {
	var skipBuild bool
	var systemInstall bool
	var desktopShortcut bool

	command := &cobra.Command{
		Use:   "desktop",
		Short: "Build and install the desktop app for this platform",
		Long:  "Builds and installs PM Planner as a native application: .app bundle on macOS, XDG launcher on Linux, Start Menu entry on Windows.",
		RunE: func(command *cobra.Command, args []string) error {
			resolvedRoot, err := project.ResolveRoot(*root)
			if err != nil {
				return err
			}
			return project.InstallDesktop(command.Context(), project.DesktopInstallOptions{
				Root:            resolvedRoot,
				SkipBuild:       skipBuild,
				System:          systemInstall,
				DesktopShortcut: desktopShortcut,
				Host:            project.DetectHostFacts(command.Context(), resolvedRoot),
			}, runnerFor(command))
		},
	}
	command.Flags().BoolVar(&skipBuild, "skip-build", false, "Install from existing build artifacts without rebuilding")
	command.Flags().BoolVar(&systemInstall, "system", false, "macOS: install to /Applications instead of ~/Applications (may require sudo)")
	command.Flags().BoolVar(&desktopShortcut, "desktop-shortcut", false, "Windows: also create a shortcut on the Desktop")
	return command
}

func newProjectCleanCommand(root *string) *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove generated build outputs",
		RunE: func(command *cobra.Command, args []string) error {
			if err := project.Clean(*root, command.OutOrStdout()); err != nil {
				return err
			}
			fmt.Fprintln(command.OutOrStdout(), "clean complete")
			return nil
		},
	}
}

func runnerFor(command *cobra.Command) project.Runner {
	return project.Runner{
		Stdout: command.OutOrStdout(),
		Stderr: command.ErrOrStderr(),
	}
}

func init() {
	rootCmd.AddCommand(newProjectCommand())
}
