package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// field represents a single selectable config field.
type field struct {
	label   string
	options []option
	cursor  int
}

func (f *field) next() {
	if f.cursor < len(f.options)-1 {
		f.cursor++
	} else {
		f.cursor = 0
	}
}

func (f *field) prev() {
	if f.cursor > 0 {
		f.cursor--
	} else {
		f.cursor = len(f.options) - 1
	}
}

func (f field) selected() string {
	return f.options[f.cursor].value
}

func (f field) selectedLabel() string {
	return f.options[f.cursor].label
}

type option struct {
	label string
	value string
}

// focusIndex tracks which field is focused.
const (
	focusName = iota
	focusBase
	focusCSS
	focusFmt
	focusPM
	focusGit
	fieldCount
)

type wizardModel struct {
	focus     int
	textInput textinput.Model
	fields    [fieldCount - 1]field // all fields except name
	cfg       pkg.ProjectConfig
}

type WizardResultMsg struct {
	Cfg      pkg.ProjectConfig
	Canceled bool
}

type WizardFinalModel struct {
	Cfg      pkg.ProjectConfig
	Canceled bool
}

func (m WizardFinalModel) Init() tea.Cmd                       { return nil }
func (m WizardFinalModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m WizardFinalModel) View() string                        { return "" }

func NewWizardModel() wizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-app"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 20
	ti.PromptStyle = ActiveStyle
	ti.TextStyle = BoldStyle

	fields := [fieldCount - 1]field{
		{label: "Base", options: []option{
			{label: "Astro", value: "astro"},
			{label: "Vite", value: "vite"},
		}},
		{label: "CSS", options: []option{
			{label: "Vanilla", value: "vanilla"},
			{label: "Tailwind", value: "tailwindcss"},
		}},
		{label: "Formatter", options: []option{
			{label: "Prettier", value: "prettier"},
			{label: "Biome", value: "biome"},
		}},
		{label: "Package Manager", options: []option{
			{label: "bun", value: "bun"},
			{label: "npm", value: "npm"},
			{label: "yarn", value: "yarn"},
			{label: "pnpm", value: "pnpm"},
		}},
		{label: "Git", options: []option{
			{label: "Yes", value: "yes"},
			{label: "No", value: "no"},
		}},
	}

	return wizardModel{
		focus:     focusName,
		textInput: ti,
		fields:    fields,
		cfg:       pkg.NewProjectConfig(),
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return WizardFinalModel{Canceled: true}, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % fieldCount
			m.updateTextInputFocus()
			return m, nil

		case "shift+tab":
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			m.updateTextInputFocus()
			return m, nil

		case "up", "k":
			if m.focus != focusName {
				m.currentField().prev()
			}
			return m, nil

		case "down", "j":
			if m.focus != focusName {
				m.currentField().next()
			}
			return m, nil

		case "enter":
			cfg := m.buildConfig()
			return WizardFinalModel{Cfg: cfg}, tea.Quit
		}
	}

	if m.focus == focusName {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *wizardModel) updateTextInputFocus() {
	if m.focus == focusName {
		m.textInput.Focus()
	} else {
		m.textInput.Blur()
	}
}

func (m *wizardModel) currentField() *field {
	return &m.fields[m.focus-1]
}

func (m wizardModel) buildConfig() pkg.ProjectConfig {
	name := strings.TrimSpace(m.textInput.Value())
	if name == "" {
		name = "my-app"
	}

	return pkg.ProjectConfig{
		ProjectName: name,
		Base:        pkg.BaseFramework(m.fields[focusBase-1].selected()),
		CSS:         pkg.CSSFramework(m.fields[focusCSS-1].selected()),
		Fmt:         pkg.Formatter(m.fields[focusFmt-1].selected()),
		PM:          pkg.PackageManager(m.fields[focusPM-1].selected()),
		NoGit:       m.fields[focusGit-1].selected() == "no",
	}
}

func (m wizardModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("bungkus-cli") + "\n\n")

	// Name field
	nameLabel := HintStyle.Render("  Name ")
	if m.focus == focusName {
		nameLabel = AccentStyle.Render("▸ Name ")
	}
	b.WriteString(nameLabel + m.textInput.View() + "\n\n")

	// Selection fields
	for i, f := range m.fields {
		focused := m.focus == i+1

		label := fmt.Sprintf("  %-17s", f.label)
		if focused {
			label = AccentStyle.Render(fmt.Sprintf("▸ %-17s", f.label))
		} else {
			label = HintStyle.Render(label)
		}

		var opts []string
		for j, opt := range f.options {
			if j == f.cursor {
				if focused {
					opts = append(opts, ActiveStyle.Render("[ "+opt.label+" ]"))
				} else {
					opts = append(opts, AccentStyle.Render(opt.label))
				}
			} else {
				if focused {
					opts = append(opts, HintStyle.Render("  "+opt.label+"  "))
				} else {
					opts = append(opts, HintStyle.Render(opt.label))
				}
			}
		}

		b.WriteString(label + strings.Join(opts, HintStyle.Render(" / ")) + "\n")
	}

	b.WriteString("\n" + HintStyle.Render("  tab/shift+tab navigate • ↑/↓ change • enter confirm • esc quit") + "\n")

	return b.String()
}
