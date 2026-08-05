package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pm-cli/pkg/app"
	"pm-cli/pkg/plan"
)

// PlanModel holds the TUI state for the interactive planner.
type PlanModel struct {
	journeys        []plan.Journey
	solvedSlot      plan.SolvedSlot
	entryInputs     []textinput.Model
	exitInputs      []textinput.Model
	journeySpanSecs []int64
	totalSpanSecs   int64
	overtimeSecs    int64
	alternativeTime string
	warnings        []string
	originals       []string
	baseTarget      time.Duration
	balance         *plan.BalanceAdjustment
	balanceError    string
	date            time.Time
	focusedSlot     int
}

func NewPlanModel(
	date time.Time,
	baseTarget time.Duration,
	originals []string,
	journeys []plan.Journey,
	solvedSlot plan.SolvedSlot,
	balance *plan.BalanceAdjustment,
	balanceError string,
) PlanModel {
	m := PlanModel{
		journeys:     journeys,
		solvedSlot:   solvedSlot,
		originals:    originals,
		baseTarget:   baseTarget,
		balance:      balance,
		balanceError: balanceError,
		date:         date,
		focusedSlot:  0,
	}
	m.entryInputs = make([]textinput.Model, len(journeys))
	m.exitInputs = make([]textinput.Model, len(journeys))
	for i, journey := range journeys {
		m.entryInputs[i] = newTimeInput(journey.Entry.Time)
		m.exitInputs[i] = newTimeInput(journey.Exit.Time)
	}
	m.recalc()
	m.syncFocus()
	return m
}

func newTimeInput(value string) textinput.Model {
	t := textinput.New()
	t.Prompt = ""
	t.CharLimit = 5
	t.Placeholder = "HH:MM"
	t.SetValue(value)
	t.CursorStyle = lipgloss.NewStyle().Foreground(colors.Cursor)
	t.TextStyle = lipgloss.NewStyle().Foreground(colors.InputText)
	t.PlaceholderStyle = lipgloss.NewStyle().Foreground(colors.Placeholder)
	return t
}

func (m *PlanModel) parse(s string) time.Time {
	t, err := time.ParseInLocation("15:04", s, m.date.Location())
	if err != nil {
		return time.Time{}
	}
	return time.Date(m.date.Year(), m.date.Month(), m.date.Day(), t.Hour(), t.Minute(), 0, 0, m.date.Location())
}

func (m *PlanModel) buildJourneysFromInputs() []plan.Journey {
	journeys := make([]plan.Journey, len(m.journeys))
	for i := range journeys {
		journeys[i] = plan.Journey{
			Entry: plan.ClockSlot{
				Time:       m.entryInputs[i].Value(),
				Registered: m.journeys[i].Entry.Registered,
			},
			Exit: plan.ClockSlot{
				Time:       m.exitInputs[i].Value(),
				Registered: m.journeys[i].Exit.Registered,
			},
		}
	}
	return journeys
}

func (m *PlanModel) isSolved(journeyIndex int, isEntry bool) bool {
	return m.solvedSlot.Valid() && m.solvedSlot.JourneyIndex == journeyIndex && m.solvedSlot.IsEntry == isEntry
}

func (m *PlanModel) recalc() {
	journeys := m.buildJourneysFromInputs()
	summary, err := app.RecalculatePlanner(m.date, int64(m.baseTarget.Seconds()), m.balance, journeys, m.solvedSlot)
	if err != nil {
		return
	}
	m.journeys = summary.Journeys
	m.journeySpanSecs = summary.JourneySpanSecs
	m.totalSpanSecs = summary.TotalSpanSecs
	m.overtimeSecs = summary.OvertimeSecs
	m.alternativeTime = summary.AlternativeTime
	m.warnings = summary.Warnings
	if m.solvedSlot.Valid() && m.solvedSlot.JourneyIndex < len(m.journeys) {
		if m.solvedSlot.IsEntry {
			m.entryInputs[m.solvedSlot.JourneyIndex].SetValue(m.journeys[m.solvedSlot.JourneyIndex].Entry.Time)
		} else {
			m.exitInputs[m.solvedSlot.JourneyIndex].SetValue(m.journeys[m.solvedSlot.JourneyIndex].Exit.Time)
		}
	}
}

// focusableSlots returns flat slot indexes that can receive focus (solved slot excluded).
func (m *PlanModel) focusableSlots() []int {
	slots := make([]int, 0, len(m.journeys)*2)
	for i := range m.journeys {
		if !m.isSolved(i, true) {
			slots = append(slots, i*2)
		}
		if !m.isSolved(i, false) {
			slots = append(slots, i*2+1)
		}
	}
	return slots
}

func (m *PlanModel) moveFocus(delta int) {
	slots := m.focusableSlots()
	if len(slots) == 0 {
		return
	}
	currentPosition := 0
	for index, slot := range slots {
		if slot == m.focusedSlot {
			currentPosition = index
			break
		}
	}
	newPosition := (currentPosition + delta + len(slots)) % len(slots)
	m.focusedSlot = slots[newPosition]
}

func (m *PlanModel) syncFocus() {
	for i := range m.entryInputs {
		m.entryInputs[i].Blur()
	}
	for i := range m.exitInputs {
		m.exitInputs[i].Blur()
	}
	journeyIndex := m.focusedSlot / 2
	isExitSlot := m.focusedSlot%2 == 1
	if journeyIndex >= len(m.journeys) {
		return
	}
	if isExitSlot {
		m.exitInputs[journeyIndex].Focus()
		return
	}
	m.entryInputs[journeyIndex].Focus()
}

func (m *PlanModel) adjustFocused(delta time.Duration) {
	journeyIndex := m.focusedSlot / 2
	isExitSlot := m.focusedSlot%2 == 1
	if journeyIndex >= len(m.journeys) {
		return
	}
	if isExitSlot {
		m.adjustInput(&m.exitInputs[journeyIndex], delta)
		return
	}
	m.adjustInput(&m.entryInputs[journeyIndex], delta)
}

func (m *PlanModel) adjustInput(input *textinput.Model, delta time.Duration) {
	base := m.parse(input.Value())
	if base.IsZero() {
		return
	}
	input.SetValue(base.Add(delta).Format("15:04"))
}

func (m *PlanModel) addJourney() {
	entryTime := ""
	if len(m.journeys) > 0 {
		lastJourney := m.journeys[len(m.journeys)-1]
		previousExitTime := m.parse(lastJourney.Exit.Time)
		if !previousExitTime.IsZero() {
			entryTime = previousExitTime.Add(time.Hour).Format("15:04")
		}
	}
	newJourney := plan.Journey{
		Entry: plan.ClockSlot{Time: entryTime},
		Exit:  plan.ClockSlot{},
	}
	m.journeys = append(m.journeys, newJourney)
	m.entryInputs = append(m.entryInputs, newTimeInput(entryTime))
	m.exitInputs = append(m.exitInputs, newTimeInput(""))
	m.solvedSlot = plan.SolvedSlot{JourneyIndex: len(m.journeys) - 1, IsEntry: false}
	m.focusedSlot = (len(m.journeys) - 1) * 2
}

func (m *PlanModel) removeLastJourney() {
	if len(m.journeys) <= 1 {
		return
	}
	lastIndex := len(m.journeys) - 1
	lastJourney := m.journeys[lastIndex]
	if lastJourney.Entry.Registered || lastJourney.Exit.Registered {
		return
	}
	m.journeys = m.journeys[:lastIndex]
	m.entryInputs = m.entryInputs[:lastIndex]
	m.exitInputs = m.exitInputs[:lastIndex]
	if m.solvedSlot.Valid() && m.solvedSlot.JourneyIndex >= lastIndex {
		m.solvedSlot = plan.SolvedSlot{JourneyIndex: lastIndex - 1, IsEntry: false}
	}
	maxFocusableSlot := (lastIndex-1)*2 + 1
	if m.focusedSlot > maxFocusableSlot {
		m.focusedSlot = lastIndex * 2
	}
}

func (m *PlanModel) normalizeInputs() {
	for i := range m.entryInputs {
		if m.isSolved(i, true) {
			continue
		}
		normalizeTimeInput(&m.entryInputs[i])
	}
	for i := range m.exitInputs {
		if m.isSolved(i, false) {
			continue
		}
		normalizeTimeInput(&m.exitInputs[i])
	}
}

func normalizeTimeInput(input *textinput.Model) {
	value := input.Value()
	if len(value) == 0 {
		return
	}
	if !containsColon(value) && countDigits(value) == 4 {
		input.SetValue(formatDigitsToHHMM(value))
	}
	if len(value) > 5 {
		input.SetValue(value[:5])
	}
}

func (m PlanModel) Init() tea.Cmd { return textinput.Blink }

func (m PlanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKeyMsg := msg.(tea.KeyMsg)
	if !isKeyMsg {
		return m.updateInputs(msg)
	}

	switch keyMsg.String() {
	case "up":
		m.adjustFocused(15 * time.Minute)
		m.recalc()
		return m, nil
	case "down":
		m.adjustFocused(-15 * time.Minute)
		m.recalc()
		return m, nil
	case "tab":
		m.moveFocus(1)
		m.syncFocus()
		return m, nil
	case "shift+tab":
		m.moveFocus(-1)
		m.syncFocus()
		return m, nil
	case "+", "ctrl+a":
		m.addJourney()
		m.recalc()
		m.syncFocus()
		return m, nil
	case "-", "ctrl+d":
		m.removeLastJourney()
		m.recalc()
		m.syncFocus()
		return m, nil
	}

	if keyMsg.Type == tea.KeyCtrlC || keyMsg.Type == tea.KeyEsc || keyMsg.Type == tea.KeyEnter {
		return m, tea.Quit
	}

	return m.updateInputs(msg)
}

func (m PlanModel) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	for i := range m.entryInputs {
		if m.isSolved(i, true) {
			continue
		}
		m.entryInputs[i], cmd = m.entryInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	for i := range m.exitInputs {
		if m.isSolved(i, false) {
			continue
		}
		m.exitInputs[i], cmd = m.exitInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	m.normalizeInputs()
	m.recalc()
	return m, tea.Batch(cmds...)
}

func (m PlanModel) View() string {
	header := styles.Title.Render("Ponto Planner")
	sub := styles.Subtitle.Render(m.date.Format("2006-01-02") + "  •  Meta " + fmtDur(m.baseTarget))
	orig := styles.Muted.Render(originalsLine(m.originals))

	journeyLines := make([]string, 0, len(m.journeys)*3)
	for i, journey := range m.journeys {
		journeyLines = append(journeyLines, styles.Subtitle.Render(fmt.Sprintf("Jornada %d", i+1)))

		entryLabel := fmt.Sprintf("  Entrada %d", i+1)
		if m.isSolved(i, true) {
			solvedValue := styles.Calculated.Render(journey.Entry.Time) + " " + styles.Muted.Render("◆ (calculado)")
			journeyLines = append(journeyLines, row(entryLabel, solvedValue))
		} else {
			journeyLines = append(journeyLines, row(entryLabel, slotValue(m.entryInputs[i].View(), journey.Entry.Registered)))
		}

		exitLabel := fmt.Sprintf("  Saída %d", i+1)
		if m.isSolved(i, false) {
			solvedValue := styles.Calculated.Render(journey.Exit.Time) + " " + styles.Muted.Render("◆ (calculado)")
			journeyLines = append(journeyLines, row(exitLabel, solvedValue))
		} else {
			journeyLines = append(journeyLines, row(exitLabel, slotValue(m.exitInputs[i].View(), journey.Exit.Registered)))
		}
	}

	inputs := lipgloss.JoinVertical(lipgloss.Left, journeyLines...)
	if m.alternativeTime != "" && m.balance != nil {
		inputs = lipgloss.JoinVertical(lipgloss.Left, inputs,
			styles.Muted.Render("Horário alternativo: "+m.alternativeTime+" (banco "+fmtSignedMinutes(m.balance.BalanceSecs)+")"),
		)
	} else if m.balanceError != "" {
		inputs = lipgloss.JoinVertical(lipgloss.Left, inputs, styles.Muted.Render(m.balanceError))
	}
	inputs = styles.Panel.Render(inputs)

	summaryLines := make([]string, 0, len(m.journeys)+3)
	summaryLines = append(summaryLines, row("Meta do Dia", fmtDur(m.baseTarget)))
	for i := range m.journeys {
		journeyLabel := fmt.Sprintf("%dª Jornada", i+1)
		if i < len(m.journeySpanSecs) {
			spanDuration := time.Duration(m.journeySpanSecs[i]) * time.Second
			summaryLines = append(summaryLines, row(journeyLabel, fmtDur(spanDuration)))
		}
	}
	summaryLines = append(summaryLines,
		row("Total", fmtDur(time.Duration(m.totalSpanSecs)*time.Second)),
		row("Hora Extra", fmtDur(time.Duration(m.overtimeSecs)*time.Second)),
	)
	resume := styles.Panel.Render(lipgloss.JoinVertical(lipgloss.Left, summaryLines...))

	content := lipgloss.JoinHorizontal(lipgloss.Top, inputs, styles.Spacer, resume)
	keys := styles.Keys.Render("↑/↓ ±15min • Tab/Shift+Tab alternar • +/Ctrl+A adicionar • -/Ctrl+D remover • Enter/Esc sair")
	return lipgloss.JoinVertical(lipgloss.Left, header, sub, orig, styles.Gap, content, styles.Gap, keys)
}

func originalsLine(originals []string) string {
	if len(originals) == 0 {
		return "Registros originais: (nenhum)"
	}
	return "Registros originais: " + app.FormatOriginalStampStrings(originals)
}

func durBetween(a, b time.Time) time.Duration {
	if a.IsZero() || b.IsZero() || b.Before(a) {
		return 0
	}
	return b.Sub(a)
}

func fmtDur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalHours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", totalHours, minutes, seconds)
}

func fmtSignedMinutes(seconds int64) string {
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

func containsColon(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return true
		}
	}
	return false
}

func countDigits(s string) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			count++
		}
	}
	return count
}

func formatDigitsToHHMM(s string) string {
	digits := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) >= 4 {
		hours, _ := strconv.Atoi(string(digits[:2]))
		minutes, _ := strconv.Atoi(string(digits[2:4]))
		return fmt.Sprintf("%02d:%02d", hours%24, clampMinute(minutes))
	}
	return string(digits)
}

func clampMinute(m int) int {
	if m < 0 {
		return 0
	}
	if m > 59 {
		return 59
	}
	return m
}

var colors = struct {
	Title       lipgloss.Color
	Subtitle    lipgloss.Color
	PanelBorder lipgloss.Color
	Label       lipgloss.Color
	Value       lipgloss.Color
	Calculated  lipgloss.Color
	Registered  lipgloss.Color
	Muted       lipgloss.Color
	Spacer      lipgloss.Color
	Keys        lipgloss.Color
	Cursor      lipgloss.Color
	InputText   lipgloss.Color
	Placeholder lipgloss.Color
}{
	Title:       lipgloss.Color("63"),
	Subtitle:    lipgloss.Color("245"),
	PanelBorder: lipgloss.Color("60"),
	Label:       lipgloss.Color("245"),
	Value:       lipgloss.Color("255"),
	Calculated:  lipgloss.Color("42"),
	Registered:  lipgloss.Color("34"),
	Muted:       lipgloss.Color("241"),
	Spacer:      lipgloss.Color("237"),
	Keys:        lipgloss.Color("244"),
	Cursor:      lipgloss.Color("69"),
	InputText:   lipgloss.Color("252"),
	Placeholder: lipgloss.Color("240"),
}

var styles = func() struct {
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Panel      lipgloss.Style
	RowLabel   lipgloss.Style
	RowValue   lipgloss.Style
	Calculated       lipgloss.Style
	RegisteredBorder lipgloss.Style
	Muted            lipgloss.Style
	Spacer           string
	Gap              string
	Keys             lipgloss.Style
} {
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.PanelBorder).Padding(1, 2)
	return struct {
		Title            lipgloss.Style
		Subtitle         lipgloss.Style
		Panel            lipgloss.Style
		RowLabel         lipgloss.Style
		RowValue         lipgloss.Style
		Calculated       lipgloss.Style
		RegisteredBorder lipgloss.Style
		Muted            lipgloss.Style
		Spacer           string
		Gap              string
		Keys             lipgloss.Style
	}{
		Title:      lipgloss.NewStyle().Foreground(colors.Title).Bold(true).MarginBottom(1),
		Subtitle:   lipgloss.NewStyle().Foreground(colors.Subtitle),
		Panel:      panel,
		RowLabel:   lipgloss.NewStyle().Foreground(colors.Label),
		RowValue:   lipgloss.NewStyle().Foreground(colors.Value).Bold(true),
		Calculated: lipgloss.NewStyle().Foreground(colors.Calculated).Bold(true),
		RegisteredBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Registered).
			Padding(0, 1),
		Muted: lipgloss.NewStyle().Foreground(colors.Muted),
		Spacer:     lipgloss.NewStyle().Width(4).Render(" "),
		Gap:        lipgloss.NewStyle().Height(1).Render(""),
		Keys:       lipgloss.NewStyle().Foreground(colors.Keys),
	}
}()

func slotValue(value string, registered bool) string {
	if registered {
		return styles.RegisteredBorder.Render(value)
	}
	return value
}

func row(label, value string) string {
	labelRendered := styles.RowLabel.Render(label + ":")
	valueRendered := styles.RowValue.Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Top, labelRendered, lipgloss.NewStyle().Width(2).Render(" "), valueRendered)
}
