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

// field represents a single selectable config field.
type field struct {
	label   string
	options []option
	cursor  int
}

// option represents a single selectable value within a field.
type option struct {
	label string
	value string
}

// Title, Description, FilterValue implement list.DefaultItem for use with bubbles list.
func (o option) Title() string       { return o.label }
func (o option) Description() string { return "" }
func (o option) FilterValue() string { return o.label }

// focusIndex tracks which field is focused.
const (
	focusName uint = iota
	focusBase
	focusCSS
	focusFmt
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
	lists     [fieldCount - 1]list.Model // bubbles list for each setup field
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

		// Resize each list column to fit the 3-column grid.
		colWidth := m.width/3 - 4
		if colWidth < 16 {
			colWidth = 16
		}
		for i := range m.lists {
			m.lists[i].SetWidth(colWidth)
		}

	case tea.KeyPressMsg:
		// Summary screen: enter to confirm and scaffold, esc to go back.
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

		// Wizard screen key handling.
		switch msg.String() {
		case "ctrl+c", "esc":
			return WizardFinalModel{Canceled: true}, tea.Quit

		case "enter":
			m.screen = screenSummary
			return m, nil

		// Navigate between fields.
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

		// Delegate key events to the focused component.
		if m.focus == focusName {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Forward remaining keys to the focused list (up/down/j/k navigation).
		idx := m.focus - 1
		var cmd tea.Cmd
		m.lists[idx], cmd = m.lists[idx].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m wizardModel) View() tea.View {
	if m.screen == screenSummary {
		return tea.NewView(m.summaryView())
	}

	var s strings.Builder

	// Match project name box width to the 3-column grid.
	colWidth := m.width / 3
	if colWidth < 20 {
		colWidth = 20
	}
	gridWidth := colWidth*3 - 2

	borderColor := ColorMuted
	if m.focus == focusName {
		borderColor = ColorAccent
	}
	pn := lipgloss.NewStyle().
		Width(gridWidth).
		BorderStyle(lipgloss.ASCIIBorder()).
		BorderForeground(borderColor)

	layout := lipgloss.JoinVertical(lipgloss.Top,
		m.headerView(),
		pn.Render(AccentStyle.Render("Project Name")+"\n"+m.textInput.View()),
		m.setUpView(),
		m.footerView(),
	)

	s.Write([]byte(layout))
	return tea.NewView(s.String())
}

// NewWizardModel creates and returns a fully initialised wizard model
// with a text input for the project name and a bubbles list for each setup field.
func NewWizardModel() wizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-app"
	ti.Prompt = "  "
	ti.Focus()
	ti.CharLimit = 64
	ti.SetWidth(40)

	fields := [fieldCount - 1]field{
		{label: "Base", options: []option{
			{label: "Astro", value: "astro"},
			{label: "Astro Vue", value: "astro-vue"},
			{label: "Astro React", value: "astro-react"},
			{label: "Vite", value: "vite"},
		}},
		{label: "CSS", options: []option{
			{label: "Tailwind", value: "tailwindcss"},
			{label: "Vanilla", value: "vanilla"},
		}},
		{label: "Formatter", options: []option{
			{label: "Biome [Recommended]", value: "biome"},
			{label: "Prettier", value: "prettier"},
		}},
		{label: "Package Manager", options: []option{
			{label: "pnpm [Recommended]", value: "pnpm"},
			{label: "bun", value: "bun"},
			{label: "npm", value: "npm"},
			{label: "yarn", value: "yarn"},
		}},
		{label: "Git", options: []option{
			{label: "Yes", value: "yes"},
			{label: "No", value: "no"},
		}},
	}

	// Find the tallest field so all lists share the same height.
	maxItems := 0
	for _, f := range fields {
		if len(f.options) > maxItems {
			maxItems = len(f.options)
		}
	}
	listHeight := maxItems + 2 // items + title bar overhead

	// Initialise a bubbles list for each field with compact, label-only rendering.
	var lists [fieldCount - 1]list.Model
	for i, f := range fields {
		items := make([]list.Item, len(f.options))
		for j, opt := range f.options {
			items[j] = opt
		}

		// Compact delegate: single-line items, no description, no spacing.
		delegate := list.NewDefaultDelegate()
		delegate.ShowDescription = false
		delegate.SetSpacing(0)
		delegate.Styles.NormalTitle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 0, 0, 4)
		delegate.Styles.SelectedTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			PaddingLeft(1).
			SetString("▸ ")

		// Uniform height across all grid cells.
		l := list.New(items, delegate, 20, listHeight)
		l.Title = f.label
		l.Styles.Title = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
		l.Styles.TitleBar = lipgloss.NewStyle().
			Padding(0, 0, 1, 0)

		// Strip all chrome — we only want title + items.
		l.SetShowFilter(false)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetShowPagination(false)
		l.SetFilteringEnabled(false)
		l.DisableQuitKeybindings()
		l.InfiniteScrolling = true

		lists[i] = l
	}

	return wizardModel{
		textInput: ti,
		fields:    fields,
		lists:     lists,
	}
}

// setUpView renders the setup fields as a 3-column grid of bordered lists.
// Each cell contains a bubbles list with a label title and selectable options.
// The focused cell's border is highlighted with ColorAccent.
func (m wizardModel) setUpView() string {
	colWidth := m.width / 3
	if colWidth < 20 {
		colWidth = 20
	}

	// Build a bordered box for each field list.
	// Unfocused boxes are muted (dim border + faint content).
	var boxes []string
	for i := range m.fields {
		focused := m.focus == uint(i+1)

		borderColor := ColorMuted
		if focused {
			borderColor = ColorAccent
		}

		style := lipgloss.NewStyle().
			Width(colWidth - 2).
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(borderColor)

		content := m.lists[i].View()
		if !focused {
			content = lipgloss.NewStyle().Faint(true).Render(content)
		}

		boxes = append(boxes, style.Render(content))
	}

	// Arrange boxes into rows of 3 columns.
	var rows []string
	for i := 0; i < len(boxes); i += 3 {
		end := i + 3
		if end > len(boxes) {
			end = len(boxes)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, boxes[i:end]...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// buildConfig assembles a ProjectConfig from the current wizard selections.
func (m wizardModel) buildConfig() pkg.ProjectConfig {
	name := m.projectName()
	return pkg.ProjectConfig{
		ProjectName: name,
		Base:        pkg.BaseFramework(m.selectedValue(0).value),
		CSS:         pkg.CSSFramework(m.selectedValue(1).value),
		Fmt:         pkg.Formatter(m.selectedValue(2).value),
		PM:          pkg.PackageManager(m.selectedValue(3).value),
		NoGit:       m.selectedValue(4).value == "no",
	}
}

// projectName returns the entered name or the placeholder default.
func (m wizardModel) projectName() string {
	name := strings.TrimSpace(m.textInput.Value())
	if name == "" {
		name = "my-app"
	}
	return name
}

// selectedValue returns the selected option's value for a given list index.
func (m wizardModel) selectedValue(listIdx int) option {
	item := m.lists[listIdx].SelectedItem()
	if item == nil {
		return m.fields[listIdx].options[0]
	}
	return item.(option)
}

// summaryView renders a confirmation screen with all selections and install instructions.
func (m wizardModel) summaryView() string {
	var b strings.Builder

	name := m.projectName()
	base := m.selectedValue(0)
	css := m.selectedValue(1)
	fmtOpt := m.selectedValue(2)
	pm := m.selectedValue(3)
	git := m.selectedValue(4)

	label := lipgloss.NewStyle().Foreground(ColorMuted).Width(20)
	value := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)

	b.WriteString(TitleStyle.Render("bungkus-cli") + "\n\n")
	b.WriteString(AccentStyle.Render("  Summary") + "\n\n")

	rows := []struct{ l, v string }{
		{"Project Name", name},
		{"Base", base.label},
		{"CSS", css.label},
		{"Formatter", fmtOpt.label},
		{"Package Manager", pm.label},
		{"Git", git.label},
	}
	for _, r := range rows {
		b.WriteString("  " + label.Render(r.l) + value.Render(r.v) + "\n")
	}

	b.WriteString("\n" + HintStyle.Render("  enter scaffold • esc go back"))

	return b.String()
}

func (m wizardModel) headerView() string {
	art := `88""Yb 88   88 88b 88  dP""b8 88  dP 88   88 .dP"Y8      dP""b8 88     88
88__dP 88   88 88Yb88 dP   ` + "`" + `" 88odP  88   88 ` + "`" + `Ybo."     dP   ` + "`" + `" 88     88
88""Yb Y8   8P 88 Y88 Yb  "88 88"Yb  Y8   8P o.` + "`" + `Y8b     Yb      88  .o 88
88oodP ` + "`" + `YbodP' 88  Y8  YboodP 88  Yb ` + "`" + `YbodP' 8bodP'      YboodP 88ood8 88`
	return PrimaryStyle.Margin(1, 0).Render(art) + "\n"
}

func (m wizardModel) footerView() string {
	return "\n" + HintStyle.Render("  tab/shift+tab navigate • ↑/↓ select • enter confirm • esc quit")
}
