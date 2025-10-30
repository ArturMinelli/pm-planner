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
	"pm-cli/pkg/plan"
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

		wd, err := client.FetchWorkDay(ctx, planDate)
		if err != nil {
			return err
		}

		loc := time.Now().Location()
		date, _ := time.ParseInLocation("2006-01-02", planDate, loc)

		stamps := make([]time.Time, 0, len(wd.TimeCards))
		for _, c := range wd.TimeCards {
			if t, err := api.ParseHHMMOnDate(c.Time, date); err == nil {
				stamps = append(stamps, t)
			}
		}
		sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

		periods := make([]plan.Period, 0, len(wd.ShiftDay.Periods))
		for _, p := range wd.ShiftDay.Periods {
			enter, err1 := api.ParseHHMMOnDate(p.EnterTime, date)
			leave, err2 := api.ParseHHMMOnDate(p.LeaveTime, date)
			if err1 == nil && err2 == nil {
				periods = append(periods, plan.Period{Enter: enter, Leave: leave})
			}
		}

		target := 8*time.Hour + 30*time.Minute
		if wd.ShiftTime > 0 {
			target = time.Duration(wd.ShiftTime * float64(time.Second))
		}

		sug, err := plan.Suggest(date, stamps, periods, target)
		if err != nil {
			return err
		}

		in1Str := sug.In1.Format("15:04")
		out1Str := sug.Out1.Format("15:04")
		in2Str := sug.In2.Format("15:04")

		if live {
			m := ui.NewPlanModel(date, target, stamps, in1Str, out1Str, in2Str)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		}

		// non-live: last field is computed and displayed as a note
		out2Str := computeOut2(in1Str, out1Str, in2Str, target, date)
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

		out2Str = computeOut2(in1Str, out1Str, in2Str, target, date)

		fmt.Println()
		fmt.Printf("Entrada 1: %s\nSaída 1: %s\nEntrada 2: %s\nSaída 2: %s\n", in1Str, out1Str, in2Str, out2Str)
		return nil
	},
}

func computeOut2(in1, out1, in2 string, target time.Duration, date time.Time) string {
	p := func(s string) time.Time { t, _ := time.ParseInLocation("15:04", s, date.Location()); return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()) }
	in1T, out1T, in2T := p(in1), p(out1), p(in2)
	first := out1T.Sub(in1T)
	need := target - first
	if need < 0 { need = 0 }
	if in2T.IsZero() { return "--:--" }
	return in2T.Add(need).Format("15:04")
}

func dur(d time.Duration) string { if d < 0 { d = 0 } ; h := int(d.Hours()) ; m := int(d.Minutes()) % 60 ; s := int(d.Seconds()) % 60 ; return fmt.Sprintf("%02d:%02d:%02d", h, m, s) }

func formatStamps(stamps []time.Time) string {
	if len(stamps) == 0 { return "(nenhum)" }
	s := ""
	for i, t := range stamps { if i > 0 { s += ", " } ; s += t.Format("15:04") }
	return s
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVar(&planDate, "date", "", "Date in YYYY-MM-DD (default: today)")
	planCmd.Flags().BoolVar(&live, "live", true, "Live mode (Bubble Tea) for auto-updating summary")
}
