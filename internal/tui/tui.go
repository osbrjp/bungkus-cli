package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InitOptions struct {
	Path      string
	Base      string
	CSS       string
	Formatter string
	Linter    string
}

type phase int

const (
	phaseSelecting phase = iota
	phaseConfirm
	phaseDone
)

// focus indices: 0 = path input, 1-4 = panels
const (
	focusPath      = 0
	focusBase      = 1
	focusCSS       = 2
	focusFormatter = 3
	focusLinter    = 4
	focusCount     = 5
)

type Model struct {
	pathInput textinput.Model
	panels    [4]Panel
	focus     int
	phase     phase
	width     int
	height    int
	result    *InitOptions
	quitting  bool
}

func initialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "./my-project"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	panels := [4]Panel{
		NewPanel("Base", []string{"astro", "vite"}, 0),
		NewPanel("CSS", []string{"vanilla", "tailwindcss"}, 1),
		NewPanel("Formatter", []string{"prettier", "biome"}, 0),
		NewPanel("Linter", []string{"eslint", "biome"}, 0),
	}

	return Model{
		pathInput: ti,
		panels:    panels,
		focus:     focusPath,
		phase:     phaseSelecting,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if key.Matches(msg, keys.Quit) && m.focus != focusPath {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.phase {
		case phaseSelecting:
			return m.updateSelecting(msg)
		case phaseConfirm:
			return m.updateConfirm(msg)
		}
	}

	return m, nil
}

func (m Model) updateSelecting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
		case key.Matches(msg, keys.Tab):
		m = m.blur()
		m.focus = (m.focus + 1) % focusCount
		m = m.applyFocus()
		return m, nil

	case key.Matches(msg, keys.ShiftTab):
		m = m.blur()
		m.focus = (m.focus - 1 + focusCount) % focusCount
		m = m.applyFocus()
		return m, nil
	}

	// When path input is focused, pass keys to it
	if m.focus == focusPath {
		// Enter on path input moves to confirm if path is non-empty
		if key.Matches(msg, keys.Confirm) {
			return m.tryConfirm()
		}
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	// Panel is focused
	panelIdx := m.focus - 1

	// Enter on a panel also moves to confirm
	if msg.String() == "enter" {
		m.panels[panelIdx] = m.panels[panelIdx].Update(msg)
		return m.tryConfirm()
	}

	m.panels[panelIdx] = m.panels[panelIdx].Update(msg)
	return m, nil
}

func (m Model) tryConfirm() (tea.Model, tea.Cmd) {
	path := m.pathInput.Value()
	if path == "" {
		path = m.pathInput.Placeholder
	}
	_ = path
	m.phase = phaseConfirm
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.phase = phaseSelecting
		return m, nil
	case key.Matches(msg, keys.Confirm):
		path := m.pathInput.Value()
		if path == "" {
			path = m.pathInput.Placeholder
		}
		m.result = &InitOptions{
			Path:      path,
			Base:      m.panels[0].Selected(),
			CSS:       m.panels[1].Selected(),
			Formatter: m.panels[2].Selected(),
			Linter:    m.panels[3].Selected(),
		}
		m.phase = phaseDone
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) blur() Model {
	if m.focus == focusPath {
		m.pathInput.Blur()
	}
	return m
}

func (m Model) applyFocus() Model {
	if m.focus == focusPath {
		m.pathInput.Focus()
	}
	return m
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	w := m.width
	if w == 0 {
		w = 80
	}

	switch m.phase {
	case phaseConfirm:
		return m.confirmView(w)
	default:
		return m.selectingView(w)
	}
}

func (m Model) selectingView(w int) string {
	// Path input
	pathLabel := "  Project Path: "
	if m.focus == focusPath {
		pathLabel = titleStyle.Render("  Project Path: ")
	}

	pathBorder := unfocusedBorder
	if m.focus == focusPath {
		pathBorder = focusedBorder
	}
	pathView := pathBorder.Width(w - 2).Render(pathLabel + m.pathInput.View())

	// Panel rows
	row1 := renderPanelRow(
		[]Panel{m.panels[0], m.panels[1]},
		m.focus-1, 0, w,
	)
	row2 := renderPanelRow(
		[]Panel{m.panels[2], m.panels[3]},
		m.focus-1, 2, w,
	)

	// Help
	help := helpStyle.Render(
		"  Tab: switch panel  |  j/k: move  |  Enter/Space: select  |  q: quit",
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		pathView,
		row1,
		row2,
		help,
	)
}

func (m Model) confirmView(w int) string {
	return confirmView(m.pathInput.Value(), m.panels, w)
}

func Run() (*InitOptions, error) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m, ok := finalModel.(Model)
	if !ok {
		return nil, nil
	}

	return m.result, nil
}
