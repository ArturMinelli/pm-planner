package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlanModel struct {
	in1   textinput.Model
	out1  textinput.Model
	in2   textinput.Model
	first time.Duration
	second time.Duration
	total  time.Duration
	extra  time.Duration
	out2Str string
	arget time.Duration
	date  time.Time
	focus int
}

func NewPlanModel(date time.Time, target time.Duration, in1, out1, in2, _ string) PlanModel {
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
		in1:   mk(in1),
		out1:  mk(out1),
		in2:   mk(in2),
		arget: target,
		date:  date,
		focus: 0,
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
	need := m.arget - m.first
	if need < 0 {
		need = 0
	}
	if !in2.IsZero() {
		out2 := in2.Add(need)
		m.second = durBetween(in2, out2)
		m.out2Str = out2.Format("15:04")
	} else {
		m.second = 0
		m.out2Str = "--:--"
	}
	m.total = m.first + m.second
	extra := m.total - m.arget
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focus = (m.focus + 1) % 3
			m.syncFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + 2) % 3
			m.syncFocus()
			return m, nil
		}
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter {
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
	m.recalc()
	return m, tea.Batch(cmds...)
}

func (m *PlanModel) syncFocus() {
	m.in1.Blur(); m.out1.Blur(); m.in2.Blur()
	switch m.focus { case 0: m.in1.Focus(); case 1: m.out1.Focus(); case 2: m.in2.Focus() }
}

func (m PlanModel) View() string {
	header := styles.Title.Render("Ponto Planner")
	sub := styles.Subtitle.Render(m.date.Format("2006-01-02") + "  •  Meta " + fmtDur(m.arget))

	inputs := lipgloss.JoinVertical(lipgloss.Left,
		row("Entrada 1", m.in1.View()),
		row("Saída 1", m.out1.View()),
		row("Entrada 2", m.in2.View()),
		row("Saída 2", styles.Calculated.Render(m.out2Str)+" "+styles.Muted.Render("(calculado)")),
	)
	inputs = styles.Panel.Render(inputs)

	resume := lipgloss.JoinVertical(lipgloss.Left,
		row("1ª Jornada", fmtDur(m.first)),
		row("2ª Jornada", fmtDur(m.second)),
		row("Total", fmtDur(m.total)),
		row("Hora Extra", fmtDur(m.extra)),
	)
	resume = styles.Panel.Render(resume)

	content := lipgloss.JoinHorizontal(lipgloss.Top, inputs, styles.Spacer, resume)
	keys := styles.Keys.Render("Tab/Shift+Tab alternar • Enter/Esc sair")
	return lipgloss.JoinVertical(lipgloss.Left, header, sub, styles.Gap, content, styles.Gap, keys)
}

func fmtDur(d time.Duration) string {
	if d < 0 { d = 0 }
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
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
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Panel     lipgloss.Style
	RowLabel  lipgloss.Style
	RowValue  lipgloss.Style
	Calculated lipgloss.Style
	Muted     lipgloss.Style
	Spacer    string
	Gap       string
	Keys      lipgloss.Style
} {
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colors.PanelBorder).Padding(1, 2)
	return struct {
		Title     lipgloss.Style
		Subtitle  lipgloss.Style
		Panel     lipgloss.Style
		RowLabel  lipgloss.Style
		RowValue  lipgloss.Style
		Calculated lipgloss.Style
		Muted     lipgloss.Style
		Spacer    string
		Gap       string
		Keys      lipgloss.Style
	}{
		Title:     lipgloss.NewStyle().Foreground(colors.Title).Bold(true).MarginBottom(1),
		Subtitle:  lipgloss.NewStyle().Foreground(colors.Subtitle),
		Panel:     panel,
		RowLabel:  lipgloss.NewStyle().Foreground(colors.Label),
		RowValue:  lipgloss.NewStyle().Foreground(colors.Value).Bold(true),
		Calculated: lipgloss.NewStyle().Foreground(colors.Calculated).Bold(true),
		Muted:     lipgloss.NewStyle().Foreground(colors.Muted),
		Spacer:    lipgloss.NewStyle().Width(4).Render(" "),
		Gap:       lipgloss.NewStyle().Height(1).Render(""),
		Keys:      lipgloss.NewStyle().Foreground(colors.Keys),
	}
}()

func row(label, value string) string {
	l := styles.RowLabel.Render(label + ":")
	v := styles.RowValue.Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Top, l, lipgloss.NewStyle().Width(2).Render(" "), v)
}
