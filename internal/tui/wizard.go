package tui

import (
	"fmt"
	"io"
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
// When isCategory is true, the option acts as a non-selectable group header.
type option struct {
	label      string
	value      string
	isCategory bool
	disabled   bool
}

// Title, Description, FilterValue implement list.DefaultItem for use with bubbles list.
func (o option) Title() string       { return o.label }
func (o option) Description() string { return "" }
func (o option) FilterValue() string { return o.label }

// wizardDelegate is a custom list delegate that renders category headers
// differently from selectable items.
type wizardDelegate struct {
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	categoryStyle lipgloss.Style
	disabledStyle lipgloss.Style
}

func (d wizardDelegate) Height() int                             { return 1 }
func (d wizardDelegate) Spacing() int                            { return 0 }
func (d wizardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d wizardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	opt, ok := item.(option)
	if !ok {
		return
	}

	if opt.isCategory {
		fmt.Fprint(w, d.categoryStyle.Render(opt.label))
		return
	}

	if opt.disabled {
		fmt.Fprint(w, d.disabledStyle.Render("   "+opt.label))
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, d.selectedStyle.Render("  "+opt.label))
	} else {
		fmt.Fprint(w, d.normalStyle.Render("   "+opt.label))
	}
}

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
		colWidth := max(m.width/3-4, 16)

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

		// Jump to field by ctrl+number.
		switch msg.String() {
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
		case "ctrl+0":
			m.focus = focusName
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

	// Match project name box width to the 3-column grid.
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

// NewWizardModel creates and returns a fully initialised wizard model
// with a text input for the project name and a bubbles list for each setup field.
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

		// Strip all chrome — we only want title + items.
		l.SetShowFilter(false)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetShowPagination(false)
		l.SetFilteringEnabled(false)
		l.DisableQuitKeybindings()
		l.InfiniteScrolling = true

		// Skip initial category header so cursor starts on a selectable item.
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

// setUpView renders the setup fields as a 3-column grid of bordered lists.
// Each cell contains a bubbles list with a label title and selectable options.
// The focused cell's border is highlighted with ColorAccent.
func (m wizardModel) setUpView() string {
	colWidth := max(m.width/3, 20)

	var boxes []string
	for i := range m.fields {
		focused := m.focus == uint(i+1)

		borderColor := ColorMuted
		if focused {
			borderColor = ColorAccent
		}

		style := lipgloss.NewStyle().
			Width(colWidth-2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

		boxes = append(boxes, style.Render(m.lists[i].View()))
	}

	// Arrange boxes into rows of 3 columns.
	var rows []string
	for i := 0; i < len(boxes); i += 3 {
		end := min(i+3, len(boxes))
		row := lipgloss.JoinHorizontal(lipgloss.Top, boxes[i:end]...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// buildConfig assembles a ProjectConfig from the current wizard selections.
// updateCompat marks formatter and linter options as disabled/enabled based on:
// 1. Base framework group exclusions (e.g. oxfmt/oxlint excluded for astro)
// 2. Cross-field: biome formatter → only biome linter, and vice versa
func (m wizardModel) updateCompat() wizardModel {
	reg := pkg.GetRegistry()
	base := m.selectedValue(0)
	baseEntry := reg.GetBase(base.value)
	if baseEntry == nil {
		return m
	}

	selectedFmt := m.selectedValue(int(focusFmt) - 1)
	selectedLinter := m.selectedValue(int(focusLinter) - 1)

	// Update formatter list
	fmtIdx := int(focusFmt) - 1
	var fmtItems []list.Item
	for _, f := range reg.Formatters {
		disabled := f.ExcludesGroup(baseEntry.Group)
		// If linter is biome, only biome formatter is allowed
		if selectedLinter.value == "biome" && f.Value != "biome" {
			disabled = true
		}
		fmtItems = append(fmtItems, option{label: f.Label, value: f.Value, disabled: disabled})
	}
	m.lists[fmtIdx].SetItems(fmtItems)
	m = m.fixSelection(fmtIdx, fmtItems)

	// Update linter list
	linterIdx := int(focusLinter) - 1
	var linterItems []list.Item
	for _, l := range reg.Linters {
		disabled := l.ExcludesGroup(baseEntry.Group)
		// If formatter is biome, only biome linter is allowed
		if selectedFmt.value == "biome" && l.Value != "biome" {
			disabled = true
		}
		linterItems = append(linterItems, option{label: l.Label, value: l.Value, disabled: disabled})
	}
	m.lists[linterIdx].SetItems(linterItems)
	m = m.fixSelection(linterIdx, linterItems)

	return m
}

// fixSelection moves the cursor to the first enabled item if the current selection is disabled.
func (m wizardModel) fixSelection(listIdx int, items []list.Item) wizardModel {
	if sel, ok := m.lists[listIdx].SelectedItem().(option); ok && sel.disabled {
		for i, item := range items {
			if opt, ok := item.(option); ok && !opt.disabled {
				m.lists[listIdx].Select(i)
				break
			}
		}
	}
	return m
}

func (m wizardModel) buildConfig() pkg.ProjectConfig {
	name := m.projectName()
	return pkg.ProjectConfig{
		ProjectName: name,
		Base:        pkg.BaseFramework(m.selectedValue(0).value),
		CSS:         pkg.CSSFramework(m.selectedValue(1).value),
		Fmt:         pkg.Formatter(m.selectedValue(2).value),
		Linter:      pkg.Linter(m.selectedValue(3).value),
		PM:          pkg.PackageManager(m.selectedValue(4).value),
		NoGit:       m.selectedValue(5).value == "no",
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
// Skips category headers and returns the first selectable option as fallback.
func (m wizardModel) selectedValue(listIdx int) option {
	item := m.lists[listIdx].SelectedItem()
	if opt, ok := item.(option); ok && !opt.isCategory {
		return opt
	}
	for _, o := range m.fields[listIdx].options {
		if !o.isCategory {
			return o
		}
	}
	return m.fields[listIdx].options[0]
}

// summaryView renders a confirmation screen with all selections and install instructions.
func (m wizardModel) summaryView() string {
	var b strings.Builder

	name := m.projectName()
	base := m.selectedValue(0)
	css := m.selectedValue(1)
	fmtOpt := m.selectedValue(2)
	linter := m.selectedValue(3)
	pm := m.selectedValue(4)
	git := m.selectedValue(5)

	label := MutedStyle.Width(20)
	value := ActiveStyle

	b.WriteString(TitleStyle.Render("bungkus-cli") + "\n\n")
	b.WriteString(AccentStyle.Render("  Summary") + "\n\n")

	rows := []struct{ l, v string }{
		{"Project Name", name},
		{"Base", base.label},
		{"CSS", css.label},
		{"Formatter", fmtOpt.label},
		{"Linter", linter.label},
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
	return AccentStyle.Margin(1, 0).Render(art) + "\n"
}

func (m wizardModel) footerView() string {
	return "\n" + MutedStyle.Render("  tab/shift+tab navigate  ↑/↓ select  enter confirm  esc quit")
}
