package cmd

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
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

		payload, warning, err := loadPlanPayload(ctx, client, planDate)
		if err != nil {
			return err
		}
		if warning != "" {
			fmt.Fprintln(cmd.ErrOrStderr(), warning)
		}

		return runPlanWithPayload(planDate, payload, live)
	},
}

const planDefaultsWarning = "aviso: não foi possível carregar o dia; usando padrões do planner"

func loadPlanPayload(
	ctx context.Context,
	client *api.Client,
	dateStr string,
) (*app.PlannerPayload, string, error) {
	payload, err := app.FetchPlannerPayload(ctx, client, dateStr)
	if err == nil {
		return payload, "", nil
	}

	defaultsPayload, defaultsErr := app.BuildDefaultsPlannerPayload(dateStr)
	if defaultsErr != nil {
		return nil, "", err
	}
	return defaultsPayload, planDefaultsWarning, nil
}

func runPlanWithPayload(dateStr string, payload *app.PlannerPayload, liveMode bool) error {
	loc := time.Now().Location()
	date, _ := time.ParseInLocation("2006-01-02", dateStr, loc)
	baseTarget := time.Duration(payload.BaseTargetSecs) * time.Second

	if liveMode {
		model := ui.NewPlanModel(
			date,
			baseTarget,
			payload.OriginalTimes,
			payload.Journeys,
			payload.SolvedSlot,
			payload.Balance,
			balanceUnavailableToday(date, payload.BalanceError),
		)
		program := tea.NewProgram(model)
		if _, runErr := program.Run(); runErr != nil {
			return runErr
		}
		return nil
	}

	return runFormFlow(date, payload)
}

// journeyEditInputs holds mutable string fields for a single journey's form inputs.
type journeyEditInputs struct {
	entryTime string
	exitTime  string
}

func runFormFlow(date time.Time, payload *app.PlannerPayload) error {
	editableJourneys := buildEditableJourneys(payload.Journeys)

	fields := buildFormFields(payload, editableJourneys)
	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		return err
	}

	updatedJourneys := applyEditsToJourneys(payload.Journeys, editableJourneys)
	summary, err := app.RecalculatePlanner(date, payload.BaseTargetSecs, payload.Balance, updatedJourneys, payload.SolvedSlot)
	if err != nil {
		return err
	}

	fmt.Println()
	for index, journey := range summary.Journeys {
		fmt.Printf("Entrada %d: %s\nSaída %d: %s\n", index+1, journey.Entry.Time, index+1, journey.Exit.Time)
	}
	if line := formatAlternativeTime(summary.AlternativeTime, payload.Balance); line != "" {
		fmt.Println(line)
	} else if warning := balanceUnavailableToday(date, payload.BalanceError); warning != "" {
		fmt.Println(warning)
	}
	return nil
}

func buildEditableJourneys(journeys []plan.Journey) []journeyEditInputs {
	editable := make([]journeyEditInputs, len(journeys))
	for i, journey := range journeys {
		editable[i] = journeyEditInputs{
			entryTime: journey.Entry.Time,
			exitTime:  journey.Exit.Time,
		}
	}
	return editable
}

func buildFormFields(payload *app.PlannerPayload, editableJourneys []journeyEditInputs) []huh.Field {
	fields := []huh.Field{
		huh.NewNote().Title("Registros originais").Description(formatStamps(payload.OriginalTimes)),
	}
	for i := range editableJourneys {
		entrySolved := payload.SolvedSlot.Valid() && payload.SolvedSlot.JourneyIndex == i && payload.SolvedSlot.IsEntry
		exitSolved := payload.SolvedSlot.Valid() && payload.SolvedSlot.JourneyIndex == i && !payload.SolvedSlot.IsEntry

		if entrySolved {
			fields = append(fields,
				huh.NewNote().Title(fmt.Sprintf("Entrada %d (calculado)", i+1)).Description(editableJourneys[i].entryTime),
			)
		} else {
			fields = append(fields,
				huh.NewInput().Title(fmt.Sprintf("Entrada %d", i+1)).Value(&editableJourneys[i].entryTime),
			)
		}
		if exitSolved {
			fields = append(fields,
				huh.NewNote().Title(fmt.Sprintf("Saída %d (calculado)", i+1)).Description(editableJourneys[i].exitTime),
			)
		} else {
			fields = append(fields,
				huh.NewInput().Title(fmt.Sprintf("Saída %d", i+1)).Value(&editableJourneys[i].exitTime),
			)
		}
	}
	if warning := balanceUnavailableToday(time.Now(), payload.BalanceError); warning != "" {
		fields = append(fields, huh.NewNote().Title("Banco de horas").Description(warning))
	}
	return fields
}

func applyEditsToJourneys(original []plan.Journey, editable []journeyEditInputs) []plan.Journey {
	updated := make([]plan.Journey, len(original))
	for i, journey := range original {
		updated[i] = plan.Journey{
			Entry: plan.ClockSlot{
				Time:       editable[i].entryTime,
				Registered: journey.Entry.Registered,
			},
			Exit: plan.ClockSlot{
				Time:       editable[i].exitTime,
				Registered: journey.Exit.Registered,
			},
		}
	}
	return updated
}

func formatAlternativeTime(alternativeTime string, balance *plan.BalanceAdjustment) string {
	if alternativeTime == "" || balance == nil || !balance.AppliesToday || balance.TargetAdjustmentSecs == 0 {
		return ""
	}
	return fmt.Sprintf("Horário alternativo: %s (banco %s)", alternativeTime, formatSignedMinutes(balance.BalanceSecs))
}

func formatSignedMinutes(seconds int64) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	if seconds == 0 {
		sign = ""
	}
	minutes := (seconds + 30) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, minutes/60, minutes%60)
}

func balanceUnavailableToday(date time.Time, balanceError string) string {
	now := time.Now().In(date.Location())
	if balanceError == "" || date.Year() != now.Year() || date.YearDay() != now.YearDay() {
		return ""
	}
	return "Saldo indisponível; Saída normal mantida."
}

func formatStamps(stamps []string) string {
	if len(stamps) == 0 {
		return "(nenhum)"
	}
	result := ""
	for i, stamp := range stamps {
		if i > 0 {
			result += ", "
		}
		result += stamp
	}
	return result
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVar(&planDate, "date", "", "Date in YYYY-MM-DD (default: today)")
	planCmd.Flags().BoolVar(&live, "live", true, "Live mode (Bubble Tea) for auto-updating summary")
}
