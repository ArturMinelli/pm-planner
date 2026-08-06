package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"pm-cli/pkg/message"
	"pm-cli/pkg/update"
)

const checkTimeout = 60 * time.Second

func newUpdateCommand() *cobra.Command {
	var checkOnly bool
	var applyOnly bool
	var asJSON bool
	var relaunch bool

	command := &cobra.Command{
		Use:   "update",
		Short: "Update this PM Planner installation",
		Long: "Fetches the latest code and rebuilds the CLI and the desktop app.\n" +
			"Without flags it checks for updates first and only rebuilds when there is something new.",
		RunE: func(command *cobra.Command, args []string) error {
			if checkOnly && applyOnly {
				return errors.New("use --check ou --apply, não ambos")
			}
			if applyOnly {
				return applyUpdate(command, relaunch)
			}

			status, err := checkUpdate(command.Context())
			if err != nil {
				return err
			}
			if checkOnly {
				return printStatus(command.OutOrStdout(), status, asJSON)
			}

			_ = printStatus(command.OutOrStdout(), status, false)
			if len(status.Blockers) > 0 {
				return errors.New("atualização bloqueada")
			}
			if !status.UpdateAvailable {
				return nil
			}
			return applyUpdate(command, relaunch)
		},
	}

	command.Flags().BoolVar(&checkOnly, "check", false, "Only report whether an update is available")
	command.Flags().BoolVar(&applyOnly, "apply", false, "Apply the update without checking first")
	command.Flags().BoolVar(&asJSON, "json", false, "Print the check result as JSON")
	command.Flags().BoolVar(&relaunch, "relaunch", false, "Reopen the desktop app once the update finishes")
	return command
}

func checkUpdate(ctx context.Context) (*update.Status, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	return update.Check(ctx)
}

func applyUpdate(command *cobra.Command, relaunch bool) error {
	result, err := update.Apply(command.Context(), update.ApplyOptions{
		Output:   command.OutOrStdout(),
		Relaunch: relaunch,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(command.OutOrStdout(), message.ResultEnglish(result.Message))
	if !result.OK {
		return errors.New("atualização falhou")
	}
	return nil
}

func printStatus(out io.Writer, status *update.Status, asJSON bool) error {
	if asJSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}

	if status.Root != "" {
		fmt.Fprintf(out, "Instalação: %s\n", status.Root)
	}
	switch {
	case !status.IsGit && status.Root != "":
		fmt.Fprintln(out, "Versão atual: desconhecida (instalação sem git)")
	case status.CommitSHA != "":
		fmt.Fprintf(out, "Versão atual: %s (%s)\n", status.CommitSHA, status.CommitDate)
	}

	for _, blocker := range status.Blockers {
		fmt.Fprintf(out, "! %s\n", message.BlockerEnglish(blocker))
	}
	if len(status.Blockers) > 0 {
		return nil
	}

	switch {
	case !status.IsGit:
		fmt.Fprintln(out, "Não é possível comparar versões — atualizar reinstala a partir do código mais recente.")
	case status.Behind == 1:
		fmt.Fprintln(out, "1 atualização disponível.")
	case status.Behind > 1:
		fmt.Fprintf(out, "%d atualizações disponíveis.\n", status.Behind)
	default:
		fmt.Fprintln(out, "PM Planner está atualizado.")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(newUpdateCommand())
}
