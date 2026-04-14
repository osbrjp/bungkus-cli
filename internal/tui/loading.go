package tui

import (
	"fmt"
	"io/fs"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

type step int

const (
	stepScaffold step = iota
	stepDone
)

type model struct {
	spinner   spinner.Model
	step      step
	cfg       pkg.ProjectConfig
	templates fs.FS
	err       error
}

type ScaffoldDoneMsg struct{ Err error }

func NewSpinnerModel(cfg pkg.ProjectConfig, templates fs.FS) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = PrimaryStyle
	return model{spinner: s, cfg: cfg, templates: templates, step: stepScaffold}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runScaffold())
}

func (m model) runScaffold() tea.Cmd {
	return func() tea.Msg {
		err := pkg.Scaffold(m.cfg.ProjectName, m.templates, m.cfg)
		return ScaffoldDoneMsg{Err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScaffoldDoneMsg:
		m.err = msg.Err
		m.step = stepDone
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() tea.View {
	if m.step == stepDone {
		if m.err != nil {
			return tea.NewView(ErrorStyle.Render("✘ "+m.err.Error()) + "\n")
		}

		header := PrimaryStyle.Render("✔ ") + "Project scaffolded at " + AccentStyle.Render(m.cfg.ProjectName)
		hint := fmt.Sprintf(
			"\n\n  %s\n\n    %s\n    %s\n    %s",
			AccentStyle.Render("Get started:"),
			lipgloss.NewStyle().Foreground(ColorOrange).Render("cd "+m.cfg.ProjectName),
			lipgloss.NewStyle().Foreground(ColorOrange).Render(m.cfg.PM.InstallCmd()),
			lipgloss.NewStyle().Foreground(ColorOrange).Render(m.cfg.PM.RunCmd()),
		)
		return tea.NewView(BoxStyle.Render(header+hint) + "\n")
	}

	return tea.NewView(m.spinner.View() + " Scaffolding project...\n")
}
