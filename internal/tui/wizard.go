package tui

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// screen tracks which view is currently displayed.
const (
	screenWizard uint = iota
	screenSummary
)

// WizardFinalModel is returned when the wizard completes.
// The caller type-asserts the result to access Cfg or check Canceled.
type WizardFinalModel struct {
	Cfg      pkg.ProjectConfig
	Canceled bool
}

func (m WizardFinalModel) Init() tea.Cmd                       { return nil }
func (m WizardFinalModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m WizardFinalModel) View() tea.View                      { return tea.NewView("") }

// focusIndex tracks which field is focused.
const (
	focusName uint = iota
	focusBase
	focusCSS
	focusFmt
	focusLinter
	focusPM
	focusGit
	fieldCount
)

// wizardModel is the main TUI model for the setup wizard.
type wizardModel struct {
	screen    uint
	focus     uint
	textInput textinput.Model
	fields    [fieldCount - 1]field
	lists     [fieldCount - 1]list.Model
	width     int
	height    int
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		colWidth := max(m.width/3-4, 16)
		for i := range m.lists {
			m.lists[i].SetWidth(colWidth)
		}

	case tea.KeyPressMsg:
		if m.screen == screenSummary {
			switch msg.String() {
			case "ctrl+c":
				return WizardFinalModel{Canceled: true}, tea.Quit
			case "esc":
				m.screen = screenWizard
				return m, nil
			case "enter":
				return WizardFinalModel{Cfg: m.buildConfig()}, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return WizardFinalModel{Canceled: true}, tea.Quit
		case "enter":
			m.screen = screenSummary
			return m, nil
		case "tab":
			if m.focus == fieldCount-1 {
				m.focus = 0
			} else {
				m.focus++
			}
			return m, nil
		case "shift+tab":
			if m.focus == 0 {
				m.focus = fieldCount - 1
			} else {
				m.focus--
			}
			return m, nil
		}

		// Jump to field by ctrl+number.
		switch msg.String() {
		case "ctrl+0":
			m.focus = focusName
			return m, nil
		case "ctrl+1":
			m.focus = focusBase
			return m, nil
		case "ctrl+2":
			m.focus = focusCSS
			return m, nil
		case "ctrl+3":
			m.focus = focusFmt
			return m, nil
		case "ctrl+4":
			m.focus = focusLinter
			return m, nil
		case "ctrl+5":
			m.focus = focusPM
			return m, nil
		case "ctrl+6":
			m.focus = focusGit
			return m, nil
		}

		// Delegate key events to the focused component.
		if m.focus == focusName {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Forward remaining keys to the focused list.
		idx := m.focus - 1
		var cmd tea.Cmd
		m.lists[idx], cmd = m.lists[idx].Update(msg)

		// Skip category headers and disabled items.
		if sel, ok := m.lists[idx].SelectedItem().(option); ok && (sel.isCategory || sel.disabled) {
			switch msg.String() {
			case "up", "k":
				m.lists[idx].CursorUp()
			default:
				m.lists[idx].CursorDown()
			}
		}

		// Update compat when base, formatter, or linter changes.
		if idx == 0 || m.focus == focusFmt || m.focus == focusLinter {
			m = m.updateCompat()
		}

		return m, cmd
	}
	return m, nil
}

func (m wizardModel) View() tea.View {
	if m.screen == screenSummary {
		return tea.NewView(m.summaryView())
	}

	var s strings.Builder

	colWidth := max(m.width/3, 20)
	gridWidth := colWidth*3 - 2

	borderColor := ColorMuted
	if m.focus == focusName {
		borderColor = ColorAccent
	}
	pn := lipgloss.NewStyle().
		Width(gridWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	layout := lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		pn.Render(AccentStyle.Render("[0] Project Name")+"\n"+m.textInput.View()),
		m.setUpView(),
		m.footerView(),
	)

	s.Write([]byte(layout))
	return tea.NewView(s.String())
}

// NewWizardModel creates and returns a fully initialised wizard model.
func NewWizardModel() wizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-app"
	ti.Prompt = "  "
	ti.Focus()
	ti.CharLimit = 64
	ti.SetWidth(40)

	reg := pkg.GetRegistry()

	// Build Base options with category headers from registry.
	var baseOpts []option
	lastCat := ""
	for _, b := range reg.Bases {
		if b.Category != lastCat {
			baseOpts = append(baseOpts, option{label: b.Category, isCategory: true})
			lastCat = b.Category
		}
		baseOpts = append(baseOpts, option{label: b.Label, value: b.Value})
	}

	// Build flat option lists from registry.
	var cssOpts []option
	for _, c := range reg.CSS {
		cssOpts = append(cssOpts, option{label: c.Label, value: c.Value})
	}
	var fmtOpts []option
	for _, f := range reg.Formatters {
		fmtOpts = append(fmtOpts, option{label: f.Label, value: f.Value})
	}
	var linterOpts []option
	for _, l := range reg.Linters {
		linterOpts = append(linterOpts, option{label: l.Label, value: l.Value})
	}
	var pmOpts []option
	for _, p := range reg.PackageManagers {
		pmOpts = append(pmOpts, option{label: p.Label, value: p.Value})
	}

	fields := [fieldCount - 1]field{
		{label: "[1] Base", options: baseOpts},
		{label: "[2] CSS", options: cssOpts},
		{label: "[3] Formatter", options: fmtOpts},
		{label: "[4] Linter", options: linterOpts},
		{label: "[5] Package Manager", options: pmOpts},
		{label: "[6] Git", options: []option{
			{label: "Yes", value: "yes"},
			{label: "No", value: "no"},
		}},
	}

	maxItems := 0
	for _, f := range fields {
		if len(f.options) > maxItems {
			maxItems = len(f.options)
		}
	}
	listHeight := maxItems + 2

	var lists [fieldCount - 1]list.Model
	for i, f := range fields {
		items := make([]list.Item, len(f.options))
		for j, opt := range f.options {
			items[j] = opt
		}

		delegate := wizardDelegate{
			normalStyle: lipgloss.NewStyle().
				Foreground(ColorMuted),
			selectedStyle: lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true),
			categoryStyle: lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true).
				PaddingLeft(1),
			disabledStyle: lipgloss.NewStyle().
				Foreground(ColorMuted).
				Faint(true).
				Strikethrough(true),
		}

		l := list.New(items, delegate, 20, listHeight)
		l.Title = f.label
		l.Styles.Title = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			PaddingLeft(1)
		l.Styles.TitleBar = lipgloss.NewStyle().
			Padding(0, 0, 0, 0)

		l.SetShowFilter(false)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetShowPagination(false)
		l.SetFilteringEnabled(false)
		l.DisableQuitKeybindings()
		l.InfiniteScrolling = true

		if len(f.options) > 0 && f.options[0].isCategory {
			l.Select(1)
		}

		lists[i] = l
	}

	m := wizardModel{
		textInput: ti,
		fields:    fields,
		lists:     lists,
	}
	return m.updateCompat()
}
