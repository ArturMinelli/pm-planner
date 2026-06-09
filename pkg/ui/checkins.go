package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pm-cli/pkg/plan"
)

type PlanModel struct {
	in1                textinput.Model
	out1               textinput.Model
	in2                textinput.Model
	first              time.Duration
	second             time.Duration
	total              time.Duration
	extra              time.Duration
	out2Str            string
	alternativeOut2Str string
	originals          []time.Time
	baseTarget         time.Duration
	balance            *plan.BalanceAdjustment
	balanceError       string
	date               time.Time
	focus              int
}

func NewPlanModel(date time.Time, baseTarget time.Duration, originals []time.Time, in1, out1, in2 string, balance *plan.BalanceAdjustment, balanceError string) PlanModel {
	mk := func(val string) textinput.Model {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 5
		t.Placeholder = "HH:MM"
		t.SetValue(val)
		t.CursorStyle = lipgloss.NewStyle().Foreground(colors.Cursor)
		t.TextStyle = lipgloss.NewStyle().Foreground(colors.InputText)
		t.PlaceholderStyle = lipgloss.NewStyle().Foreground(colors.Placeholder)
		return t
	}
	m := PlanModel{
		in1:          mk(in1),
		out1:         mk(out1),
		in2:          mk(in2),
		originals:    originals,
		baseTarget:   baseTarget,
		balance:      balance,
		balanceError: balanceError,
		date:         date,
		focus:        0,
	}
	m.in1.Focus()
	m.recalc()
	return m
}

func (m *PlanModel) parse(s string) time.Time {
	t, err := time.ParseInLocation("15:04", s, m.date.Location())
	if err != nil {
		return time.Time{}
	}
	return time.Date(m.date.Year(), m.date.Month(), m.date.Day(), t.Hour(), t.Minute(), 0, 0, m.date.Location())
}

func (m *PlanModel) recalc() {
	in1 := m.parse(m.in1.Value())
	out1 := m.parse(m.out1.Value())
	in2 := m.parse(m.in2.Value())

	m.first = durBetween(in1, out1)
	need := m.baseTarget - m.first
	if need < 0 {
		need = 0
	}
	if !in2.IsZero() {
		out2 := in2.Add(need)
		m.second = durBetween(in2, out2)
		m.out2Str = out2.Format("15:04")
		m.alternativeOut2Str = ""
		if m.balance != nil && m.balance.AppliesToday && m.balance.TargetAdjustmentSecs != 0 {
			alternativeNeed := time.Duration(m.balance.AdjustedTargetSecs)*time.Second - m.first
			if alternativeNeed < 0 {
				alternativeNeed = 0
			}
			m.alternativeOut2Str = in2.Add(alternativeNeed).Format("15:04")
		}
	} else {
		m.second = 0
		m.out2Str = "--:--"
		m.alternativeOut2Str = ""
	}
	m.total = m.first + m.second
	extra := m.total - m.baseTarget
	if extra < 0 {
		extra = 0
	}
	m.extra = extra
}

func durBetween(a, b time.Time) time.Duration {
	if a.IsZero() || b.IsZero() || b.Before(a) {
		return 0
	}
	return b.Sub(a)
}

func (m PlanModel) Init() tea.Cmd { return textinput.Blink }

func (m PlanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case tea.KeyMsg:
		switch k.String() {
		case "up":
			m.adjustFocused(15 * time.Minute)
			m.recalc()
			return m, nil
		case "down":
			m.adjustFocused(-15 * time.Minute)
			m.recalc()
			return m, nil
		case "tab":
			m.focus = (m.focus + 1) % 3
			m.syncFocus()
			m.recalc()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + 2) % 3
			m.syncFocus()
			m.recalc()
			return m, nil
		}
		if k.Type == tea.KeyCtrlC || k.Type == tea.KeyEsc || k.Type == tea.KeyEnter {
			return m, tea.Quit
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.in1, cmd = m.in1.Update(msg)
	cmds = append(cmds, cmd)
	m.out1, cmd = m.out1.Update(msg)
	cmds = append(cmds, cmd)
	m.in2, cmd = m.in2.Update(msg)
	cmds = append(cmds, cmd)
	m.normalizeInputs()
	m.recalc()
	return m, tea.Batch(cmds...)
}

func (m *PlanModel) normalizeInputs() {
	n := func(t *textinput.Model) {
		v := t.Value()
		if len(v) == 0 {
			return
		}
		if !containsColon(v) {
			d := countDigits(v)
			if d == 4 {
				t.SetValue(formatDigitsToHHMM(v))
			}
		}
		if len(v) > 5 {
			t.SetValue(v[:5])
		}
	}
	n(&m.in1)
	n(&m.out1)
	n(&m.in2)
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
	c := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			c++
		}
	}
	return c
}

func formatDigitsToHHMM(s string) string {
	digits := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) >= 4 {
		h, _ := strconv.Atoi(string(digits[:2]))
		m, _ := strconv.Atoi(string(digits[2:4]))
		return fmt.Sprintf("%02d:%02d", h%24, clampMinute(m))
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

func (m *PlanModel) adjustFocused(delta time.Duration) {
	apply := func(t *textinput.Model) {
		base := m.parse(t.Value())
		if base.IsZero() {
			return
		}
		t.SetValue(base.Add(delta).Format("15:04"))
	}
	switch m.focus {
	case 0:
		apply(&m.in1)
	case 1:
		apply(&m.out1)
	case 2:
		apply(&m.in2)
	}
}

func (m *PlanModel) syncFocus() {
	m.in1.Blur()
	m.out1.Blur()
	m.in2.Blur()
	switch m.focus {
	case 0:
		m.in1.Focus()
	case 1:
		m.out1.Focus()
	case 2:
		m.in2.Focus()
	}
}

func (m PlanModel) View() string {
	header := styles.Title.Render("Ponto Planner")
	sub := styles.Subtitle.Render(m.date.Format("2006-01-02") + "  •  Meta " + fmtDur(m.baseTarget))
	orig := styles.Muted.Render(originalsLine(m.originals))

	inputs := lipgloss.JoinVertical(lipgloss.Left,
		row("Entrada 1", m.in1.View()),
		row("Saída 1", m.out1.View()),
		row("Entrada 2", m.in2.View()),
		row("Saída 2", styles.Calculated.Render(m.out2Str)+" "+styles.Muted.Render("(calculado)")),
	)
	if m.alternativeOut2Str != "" && m.balance != nil {
		inputs = lipgloss.JoinVertical(lipgloss.Left, inputs,
			styles.Muted.Render("Saída 2 alternativa: "+m.alternativeOut2Str+" (banco "+fmtSignedMinutes(m.balance.BalanceSecs)+")"),
		)
	} else if m.balanceError != "" {
		inputs = lipgloss.JoinVertical(lipgloss.Left, inputs, styles.Muted.Render(m.balanceError))
	}
	inputs = styles.Panel.Render(inputs)

	resume := lipgloss.JoinVertical(lipgloss.Left,
		row("Meta do Dia", fmtDur(m.baseTarget)),
		row("1ª Jornada", fmtDur(m.first)),
		row("2ª Jornada", fmtDur(m.second)),
		row("Total", fmtDur(m.total)),
		row("Hora Extra", fmtDur(m.extra)),
	)
	resume = styles.Panel.Render(resume)

	content := lipgloss.JoinHorizontal(lipgloss.Top, inputs, styles.Spacer, resume)
	keys := styles.Keys.Render("↑/↓ ±15min • Tab/Shift+Tab alternar • Enter/Esc sair")
	return lipgloss.JoinVertical(lipgloss.Left, header, sub, orig, styles.Gap, content, styles.Gap, keys)
}

func originalsLine(ts []time.Time) string {
	if len(ts) == 0 {
		return "Registros originais: (nenhum)"
	}
	s := "Registros originais: "
	for i, t := range ts {
		if i > 0 {
			s += ", "
		}
		s += t.Format("15:04")
	}
	return s
}

func fmtDur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
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

var colors = struct {
	Title       lipgloss.Color
	Subtitle    lipgloss.Color
	PanelBorder lipgloss.Color
	Label       lipgloss.Color
	Value       lipgloss.Color
	Calculated  lipgloss.Color
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
	Calculated lipgloss.Style
	Muted      lipgloss.Style
	Spacer     string
	Gap        string
	Keys       lipgloss.Style
} {
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.PanelBorder).Padding(1, 2)
	return struct {
		Title      lipgloss.Style
		Subtitle   lipgloss.Style
		Panel      lipgloss.Style
		RowLabel   lipgloss.Style
		RowValue   lipgloss.Style
		Calculated lipgloss.Style
		Muted      lipgloss.Style
		Spacer     string
		Gap        string
		Keys       lipgloss.Style
	}{
		Title:      lipgloss.NewStyle().Foreground(colors.Title).Bold(true).MarginBottom(1),
		Subtitle:   lipgloss.NewStyle().Foreground(colors.Subtitle),
		Panel:      panel,
		RowLabel:   lipgloss.NewStyle().Foreground(colors.Label),
		RowValue:   lipgloss.NewStyle().Foreground(colors.Value).Bold(true),
		Calculated: lipgloss.NewStyle().Foreground(colors.Calculated).Bold(true),
		Muted:      lipgloss.NewStyle().Foreground(colors.Muted),
		Spacer:     lipgloss.NewStyle().Width(4).Render(" "),
		Gap:        lipgloss.NewStyle().Height(1).Render(""),
		Keys:       lipgloss.NewStyle().Foreground(colors.Keys),
	}
}()

func row(label, value string) string {
	l := styles.RowLabel.Render(label + ":")
	v := styles.RowValue.Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Top, l, lipgloss.NewStyle().Width(2).Render(" "), v)
}
