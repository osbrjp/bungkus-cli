package tui

import (
	"fmt"
	"os"
	"os/exec"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

type step int

const (
	stepScaffold step = iota
	stepInstall
	stepGitInit
	stepDone
)

type model struct {
	spinner spinner.Model
	step    step
	cfg     pkg.ProjectConfig
	err     error
}

type (
	ScaffoldDoneMsg struct{ Err error }
	InstallDoneMsg  struct{ Err error }
	GitInitDoneMsg  struct{ Err error }
)

func NewSpinnerModel(cfg pkg.ProjectConfig) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = PrimaryStyle
	return model{spinner: s, cfg: cfg, step: stepScaffold}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScaffoldDoneMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.step = stepDone
			return m, tea.Quit
		}
		m.step = stepInstall
		return m, m.runInstall()
	case InstallDoneMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.step = stepDone
			return m, tea.Quit
		}
		m.step = stepGitInit
		return m, m.runGitInit()
	case GitInitDoneMsg:
		m.err = msg.Err
		m.step = stepDone
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) runInstall() tea.Cmd {
	return func() tea.Msg {
		pm := string(m.cfg.PM)
		cmd := exec.Command(pm, "install")
		cmd.Dir = m.cfg.ProjectName
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return InstallDoneMsg{Err: fmt.Errorf("%s install failed", pm)}
		}
		return InstallDoneMsg{}
	}
}

func (m model) runGitInit() tea.Cmd {
	return func() tea.Msg {
		dir := m.cfg.ProjectName
		cmds := [][]string{
			{"git", "init"},
			{"git", "add", "."},
			{"git", "commit", "-m", "initial commit"},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = dir
			if err := cmd.Run(); err != nil {
				return GitInitDoneMsg{Err: err}
			}
		}
		return GitInitDoneMsg{}
	}
}

func (m model) View() tea.View {
	if m.step == stepDone {
		if m.err != nil {
			return tea.NewView(ErrorStyle.Render("✘ "+m.err.Error()) + "\n")
		}

		header := PrimaryStyle.Render("✔ ") + "Project ready at " + AccentStyle.Render(m.cfg.ProjectName)
		details := fmt.Sprintf(
			"\n\n  %s %s\n  %s %s\n  %s %s",
			MutedStyle.Render("Base:"),
			PrimaryStyle.Render(string(m.cfg.Base)),
			MutedStyle.Render("CSS: "),
			PrimaryStyle.Render(string(m.cfg.CSS)),
			MutedStyle.Render("Fmt: "),
			PrimaryStyle.Render(string(m.cfg.Fmt)),
		)
		hint := fmt.Sprintf(
			"\n\n  %s %s",
			MutedStyle.Render("→"),
			AccentStyle.Render("cd "+m.cfg.ProjectName+" && "+m.cfg.PM.RunCmd()),
		)
		return tea.NewView(BoxStyle.Render(header+details+hint) + "\n")
	}

	check := PrimaryStyle.Render("✔ ")
	switch m.step {
	case stepScaffold:
		return tea.NewView(m.spinner.View() + " Scaffolding project...\n")
	case stepInstall:
		return tea.NewView(check + "Project scaffolded\n" +
			m.spinner.View() + " Installing dependencies...\n")
	case stepGitInit:
		return tea.NewView(check + "Project scaffolded\n" +
			check + "Dependencies installed\n" +
			m.spinner.View() + " Initializing git...\n")
	}

	return tea.NewView("")
}
