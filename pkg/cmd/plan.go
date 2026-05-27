package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
	"pm-cli/pkg/ui"
)

var planDate string
var live bool

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan your day with interactive suggestions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if planDate == "" {
			planDate = time.Now().Format("2006-01-02")
		}

		client := api.New()
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		st, err := app.FetchPlannerPayload(ctx, client, planDate)
		if err != nil {
			return err
		}

		loc := time.Now().Location()
		date, _ := time.ParseInLocation("2006-01-02", planDate, loc)

		stamps := make([]time.Time, 0, len(st.OriginalTimes))
		for _, s := range st.OriginalTimes {
			if t, err := api.ParseHHMMOnDate(s, date); err == nil {
				stamps = append(stamps, t)
			}
		}
		sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

		target := time.Duration(st.TargetSecs) * time.Second

		in1Str := st.In1
		out1Str := st.Out1
		in2Str := st.In2

		if live {
			m := ui.NewPlanModel(date, target, stamps, in1Str, out1Str, in2Str)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		}

		out2Str := st.Out2
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Registros originais").Description(formatStamps(stamps)),
				huh.NewInput().Title("Entrada 1").Value(&in1Str),
				huh.NewInput().Title("Saída 1").Value(&out1Str),
				huh.NewInput().Title("Entrada 2").Value(&in2Str),
				huh.NewNote().Title("Saída 2 (calculado)").Description(out2Str),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}

		if sum, err := app.RecalculatePlanner(planDate, st.TargetSecs, in1Str, out1Str, in2Str); err == nil {
			out2Str = sum.Out2
		}

		fmt.Println()
		fmt.Printf("Entrada 1: %s\nSaída 1: %s\nEntrada 2: %s\nSaída 2: %s\n", in1Str, out1Str, in2Str, out2Str)
		return nil
	},
}

func formatStamps(stamps []time.Time) string {
	if len(stamps) == 0 {
		return "(nenhum)"
	}
	s := ""
	for i, t := range stamps {
		if i > 0 {
			s += ", "
		}
		s += t.Format("15:04")
	}
	return s
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVar(&planDate, "date", "", "Date in YYYY-MM-DD (default: today)")
	planCmd.Flags().BoolVar(&live, "live", true, "Live mode (Bubble Tea) for auto-updating summary")
}
