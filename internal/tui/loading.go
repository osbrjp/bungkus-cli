package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

type model struct {
	spinner spinner.Model
	done    bool
	cfg     pkg.ProjectConfig
	err     error
}

type DoneMsg struct{ Err error }

func NewSpinnerModel(cfg pkg.ProjectConfig) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = PrimaryStyle
	return model{spinner: s, cfg: cfg}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DoneMsg:
		m.done = true
		m.err = msg.Err
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.done {
		if m.err != nil {
			return ErrorStyle.Render("✘ "+m.err.Error()) + "\n"
		}
		header := PrimaryStyle.Render("✔ " + "Project ready at " + AccentStyle.Render(m.cfg.ProjectName))
		details := fmt.Sprintf(
			"\n\n  %s %s\n  %s %s\n  %s %s",
			MutedStyle.Render("Base: "),
			PrimaryStyle.Render(string(m.cfg.Base)),
			MutedStyle.Render("CSS: "),
			PrimaryStyle.Render(string(m.cfg.CSS)),
			MutedStyle.Render("Formatter: "),
			PrimaryStyle.Render(string(m.cfg.Fmt)),
		)
		return BoxStyle.Render(header+details) + "\n"
	}

	return m.spinner.View() + "Scaffolding project...\n"
}
