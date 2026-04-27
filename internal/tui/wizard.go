package tui

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

const (
	screenWizard  = iota // (Main screen for project config)
	screenSummary        // (Summary screen for scaffold confirmation)
)

const (
	focusProjectName = iota // 0 (Top input text field)
	focusBase               // 1 (Left panel: base framework list)
	focusTooling            // 2 (Middle panel: CSS, Formatter, Linter)
	focusLibraries          // 3 (Right panel: Validation, Form, Query, CMS)
	focusPM                 // 4 (Package manager horizontal selector)
	focusLen
)

const (
	panelBoxWidth   = 30 // outer width passed to border style
	panelInnerWidth = 26 // content width inside border(2) + padding(2)
	fullRowWidth    = panelBoxWidth * 3
)

type PMModel struct {
	options  []pkg.PMEntry
	cursor   int
	selected int
}

func (p *PMModel) CursorRight() {
	p.cursor = (p.cursor + 1) % len(p.options)
}

func (p *PMModel) CursorLeft() {
	p.cursor = (p.cursor - 1 + len(p.options)) % len(p.options)
}

func (p *PMModel) Select() {
	p.selected = p.cursor
}

func (p *PMModel) View(active bool) string {
	var parts []string
	for i, opt := range p.options {
		style := lipgloss.NewStyle().Padding(0, 2)

		isSelected := i == p.selected
		switch {
		case active && i == p.cursor && isSelected:
			style = style.Background(ColorGreen).Foreground(ColorGray1).Bold(true)
		case active && i == p.cursor:
			style = style.Background(ColorGray3).Foreground(ColorLuster).Bold(true)
		case isSelected:
			style = style.Background(ColorGreen).Foreground(ColorGray1)
		default:
			style = style.Foreground(ColorLack)
		}

		parts = append(parts, style.Render(opt.Label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

type WizardModel struct {
	Cfg         pkg.ProjectConfig
	Canceled    bool
	screen      uint
	focus       uint
	width       int
	height      int
	BaseList    list.Model
	tooling     AddOnsModel
	libraries   AddOnsModel
	pm          PMModel
	projectName textinput.Model
}

type Option struct {
	label string
	value string
}

func (o Option) FilterValue() string { return o.label }

type optionDelegate struct{}

func (d optionDelegate) Height() int                             { return 1 }
func (d optionDelegate) Spacing() int                            { return 0 }
func (d optionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d optionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	o, ok := listItem.(Option)
	if !ok {
		return
	}

	width := m.Width()
	var line string
	if index == m.Index() {
		line = lipgloss.NewStyle().Width(width).Background(ColorGray3).Foreground(ColorLuster).Bold(true).Render(" " + o.label)
	} else {
		line = lipgloss.NewStyle().Width(width).Foreground(ColorLack).Render(" " + o.label)
	}

	fmt.Fprint(w, line)
}

// AddOnsModel holds multiple radio groups, each with independent selection.
type AddOnsModel struct {
	groups []RadioGroup
	cursor int
}

type RadioGroup struct {
	name     string
	options  []RadioOption
	selected int
}

type RadioOption struct {
	label string
	value string
}

func (a *AddOnsModel) totalItems() int {
	n := 0
	for _, g := range a.groups {
		n += len(g.options)
	}
	return n
}

// cursorPos maps the flat cursor to (group index, item index).
func (a *AddOnsModel) cursorPos() (int, int) {
	offset := 0
	for gi, g := range a.groups {
		if a.cursor < offset+len(g.options) {
			return gi, a.cursor - offset
		}
		offset += len(g.options)
	}
	return 0, 0
}

func (a *AddOnsModel) CursorDown() {
	total := a.totalItems()
	if total == 0 {
		return
	}
	a.cursor = (a.cursor + 1) % total
}

func (a *AddOnsModel) CursorUp() {
	total := a.totalItems()
	if total == 0 {
		return
	}
	a.cursor = (a.cursor - 1 + total) % total
}

func (a *AddOnsModel) Select() {
	gi, ii := a.cursorPos()
	a.groups[gi].selected = ii
}

func (a *AddOnsModel) View(active bool, width int) string {
	var s strings.Builder

	flatIdx := 0
	for i, g := range a.groups {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString(AccentStyle.Render(g.name) + "\n")

		for j, opt := range g.options {
			marker := "◦"
			isSelected := j == g.selected
			if isSelected {
				marker = "•"
			}

			text := " " + marker + " " + opt.label
			style := lipgloss.NewStyle().Width(width)

			switch {
			case active && flatIdx == a.cursor && isSelected:
				style = style.Background(ColorGreen).Foreground(ColorGray1).Bold(true)
			case active && flatIdx == a.cursor:
				style = style.Background(ColorGray3).Foreground(ColorLuster).Bold(true)
			case isSelected:
				style = style.Background(ColorGreen).Foreground(ColorGray1)
			default:
				style = style.Foreground(ColorLack)
			}

			s.WriteString(style.Render(text) + "\n")
			flatIdx++
		}
	}

	return s.String()
}

func NewWizardModel() WizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-app"
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(30)

	registry := pkg.GetRegistry()

	baseItems := make([]list.Item, len(registry.Bases))
	for i, b := range registry.Bases {
		baseItems[i] = Option{
			label: b.Label,
			value: b.Value,
		}
	}

	baseList := list.New(baseItems, optionDelegate{}, panelInnerWidth, len(registry.Bases)+2)
	baseList.Title = "BASES"
	baseList.Styles.TitleBar = lipgloss.NewStyle()
	baseList.Styles.Title = PanelTitleStyle
	baseList.Styles.NoItems = MutedStyle
	baseList.SetShowStatusBar(false)
	baseList.SetFilteringEnabled(false)
	baseList.SetShowHelp(false)
	baseList.SetShowPagination(false)

	first := registry.Bases[0]
	tooling, libraries := buildAddOnPanels(registry, first.Group, first.Integration)

	pm := PMModel{options: registry.PackageManagers}

	return WizardModel{
		projectName: ti,
		BaseList:    baseList,
		tooling:     tooling,
		libraries:   libraries,
		pm:          pm,
	}
}

func buildAddOnPanels(reg *pkg.Registry, group string, integration string) (AddOnsModel, AddOnsModel) {
	// Resolve effective integration: nuxt is implicitly vue
	effectiveInt := integration
	if group == "nuxt" {
		effectiveInt = "vue"
	}

	toolingCats := []struct {
		name    string
		entries []pkg.OptionEntry
	}{
		{"CSS", reg.CSS},
		{"Formatter", reg.Formatters},
		{"Linter", reg.Linters},
	}

	libraryCats := []struct {
		name    string
		entries []pkg.OptionEntry
	}{
		{"Validation", reg.Validation},
		{"Form", reg.Form},
		{"Query", reg.Query},
		{"State", reg.State},
		{"CMS", reg.CMS},
	}

	build := func(cats []struct {
		name    string
		entries []pkg.OptionEntry
	}) AddOnsModel {
		var groups []RadioGroup
		for _, cat := range cats {
			var opts []RadioOption
			for _, e := range cat.entries {
				if e.ExcludesGroup(group) {
					continue
				}
				if len(e.RequiresIntegration) > 0 {
					if effectiveInt == "" || !slices.Contains(e.RequiresIntegration, effectiveInt) {
						continue
					}
				}
				opts = append(opts, RadioOption{label: e.Label, value: e.Value})
			}
			if len(opts) > 0 {
				groups = append(groups, RadioGroup{name: cat.name, options: opts})
			}
		}
		return AddOnsModel{groups: groups}
	}

	return build(toolingCats), build(libraryCats)
}

func (m WizardModel) Init() tea.Cmd {
	m.focus = focusProjectName
	return textinput.Blink
}

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		// Summary screen key handling
		if m.screen == screenSummary {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.Canceled = true
				return m, tea.Quit
			case "enter":
				// Confirm — quit with config ready
				return m, tea.Quit
			case "backspace":
				// Go back to wizard
				m.screen = screenWizard
				return m, nil
			}
			return m, nil
		}

		// Wizard screen key handling
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Canceled = true
			return m, tea.Quit
		case "enter":
			// Collect config and move to summary screen
			if m.focus != focusProjectName {
				m.collectConfig()
				m.screen = screenSummary
				return m, nil
			}
		case "tab":
			m.focus = m.focusNext()
		case "shift+tab":
			m.focus = m.focusPrev()
		case "down", "j", "right", "l":
			if m.focus == focusBase {
				var cmd tea.Cmd
				m.BaseList, cmd = m.BaseList.Update(msg)
				m.rebuildAddOns()
				return m, cmd
			}
			if m.focus == focusTooling {
				m.tooling.CursorDown()
				return m, nil
			}
			if m.focus == focusLibraries {
				m.libraries.CursorDown()
				return m, nil
			}
			if m.focus == focusPM {
				m.pm.CursorRight()
				return m, nil
			}
		case "up", "k", "left", "h":
			if m.focus == focusBase {
				var cmd tea.Cmd
				m.BaseList, cmd = m.BaseList.Update(msg)
				m.rebuildAddOns()
				return m, cmd
			}
			if m.focus == focusTooling {
				m.tooling.CursorUp()
				return m, nil
			}
			if m.focus == focusLibraries {
				m.libraries.CursorUp()
				return m, nil
			}
			if m.focus == focusPM {
				m.pm.CursorLeft()
				return m, nil
			}
		case "space":
			if m.focus == focusTooling {
				m.tooling.Select()
				return m, nil
			}
			if m.focus == focusLibraries {
				m.libraries.Select()
				return m, nil
			}
			if m.focus == focusPM {
				m.pm.Select()
				return m, nil
			}
		}
	}

	// Update text input field
	if m.focus == focusProjectName {
		var cmd tea.Cmd
		m.projectName, cmd = m.projectName.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd

	return m, cmd
}

// collectConfig gathers all current selections into the ProjectConfig.
func (m *WizardModel) collectConfig() {
	name := m.projectName.Value()
	if name == "" {
		name = "my-app"
	}
	m.Cfg.ProjectName = name

	if sel, ok := m.BaseList.SelectedItem().(Option); ok {
		m.Cfg.Base = pkg.BaseFramework(sel.value)
	}

	for _, g := range m.tooling.groups {
		selected := g.options[g.selected]
		switch g.name {
		case "CSS":
			m.Cfg.CSS = pkg.CSSFramework(selected.value)
		case "Formatter":
			m.Cfg.Fmt = pkg.Formatter(selected.value)
		case "Linter":
			m.Cfg.Linter = pkg.Linter(selected.value)
		}
	}

	for _, g := range m.libraries.groups {
		selected := g.options[g.selected]
		switch g.name {
		case "Validation":
			m.Cfg.Validation = pkg.ValidationLib(selected.value)
		case "Form":
			m.Cfg.Form = pkg.FormLib(selected.value)
		case "Query":
			m.Cfg.Query = pkg.QueryLib(selected.value)
		case "State":
			m.Cfg.State = pkg.StateLib(selected.value)
		case "CMS":
			m.Cfg.CMS = pkg.CMS(selected.value)
		}
	}

	m.Cfg.PM = pkg.PackageManager(m.pm.options[m.pm.selected].Value)
}

func (m WizardModel) View() tea.View {
	var c *tea.Cursor
	if !m.projectName.VirtualCursor() {
		c = m.projectName.Cursor()
		c.Y += lipgloss.Height("10")
	}

	var s strings.Builder

	lists := m.middleRow()
	layout := lipgloss.JoinVertical(lipgloss.Top, m.projectNameInputView(), lists, m.pmView(), m.footerView())

	s.Write([]byte(layout))

	if m.screen == screenSummary {
		overlay := m.summaryPopup()
		w := m.width
		h := m.height
		if w == 0 {
			w = 80
		}
		if h == 0 {
			h = 24
		}
		placed := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, overlay)
		return tea.NewView(placed)
	}

	return tea.NewView(s.String())
}

func (m WizardModel) summaryPopup() string {
	title := AccentStyle.Render("  Confirm Project") + "\n\n"

	row := func(label, value string) string {
		if value == "none" {
			return ""
		}
		return MutedStyle.Render("  "+label) + PrimaryStyle.Render(value) + "\n"
	}

	body := row("Project:    ", m.Cfg.ProjectName) +
		row("Base:       ", string(m.Cfg.Base)) +
		row("CSS:        ", string(m.Cfg.CSS)) +
		row("Formatter:  ", string(m.Cfg.Fmt)) +
		row("Linter:     ", string(m.Cfg.Linter)) +
		row("Validation: ", string(m.Cfg.Validation)) +
		row("Form:       ", string(m.Cfg.Form)) +
		row("Query:      ", string(m.Cfg.Query)) +
		row("State:      ", string(m.Cfg.State)) +
		row("CMS:        ", string(m.Cfg.CMS)) +
		row("PM:         ", string(m.Cfg.PM))

	key := func(k, desc string) string {
		return FooterKeyStyle.Render(k) + FooterDescStyle.Render(" "+desc)
	}

	footer := "\n" +
		key("enter", "confirm") +
		FooterSepStyle.Render("  •  ") +
		key("backspace", "back") +
		FooterSepStyle.Render("  •  ") +
		key("esc", "quit")

	popup := lipgloss.NewStyle().
		Border(lipgloss.ASCIIBorder()).
		BorderForeground(ColorLack).
		Padding(1, 2).
		Render(title + body + footer)

	return popup
}

func (m WizardModel) projectNameInputView() string {
	label := AccentStyle.Render("Project Name:")
	box := m.borderFor(focusProjectName).Width(fullRowWidth)
	return box.Render(label + "\n" + m.projectName.View())
}

func (m WizardModel) middleRow() string {
	baseContent := m.BaseList.View()
	toolingContent := PanelTitleStyle.Render("TOOLING") + "\n" + m.tooling.View(m.focus == focusTooling, panelInnerWidth)
	librariesContent := PanelTitleStyle.Render("LIBRARIES") + "\n" + m.libraries.View(m.focus == focusLibraries, panelInnerWidth)

	// Trim trailing newlines so lipgloss.Height counts consistently
	baseContent = strings.TrimRight(baseContent, "\n")
	toolingContent = strings.TrimRight(toolingContent, "\n")
	librariesContent = strings.TrimRight(librariesContent, "\n")

	h := max(
		lipgloss.Height(baseContent),
		lipgloss.Height(toolingContent),
		lipgloss.Height(librariesContent),
	)

	box := func(focus uint) lipgloss.Style {
		return m.borderFor(focus).Width(panelBoxWidth).Height(h)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		box(focusBase).Render(baseContent),
		box(focusTooling).Render(toolingContent),
		box(focusLibraries).Render(librariesContent),
	)
}

func (m WizardModel) pmView() string {
	label := AccentStyle.Render("Package Manager:")
	box := m.borderFor(focusPM).Width(fullRowWidth)
	return box.Render(label + "\n" + m.pm.View(m.focus == focusPM))
}

func (m *WizardModel) rebuildAddOns() {
	sel, ok := m.BaseList.SelectedItem().(Option)
	if !ok {
		return
	}
	reg := pkg.GetRegistry()
	base := reg.GetBase(sel.value)
	if base != nil {
		m.tooling, m.libraries = buildAddOnPanels(reg, base.Group, base.Integration)
	}
}

func (m WizardModel) borderFor(section uint) lipgloss.Style {
	if m.focus == section {
		return ActiveBorder
	}
	return InactiveBorder
}

func (m WizardModel) headerView() string { return "Bungkus-cli" }

func (m WizardModel) footerView() string {
	key := func(k, desc string) string {
		return FooterKeyStyle.Render(k) + FooterDescStyle.Render(" "+desc)
	}

	bindings := []string{
		key("tab/shift+tab", "navigate"),
	}

	switch m.focus {
	case focusBase, focusTooling, focusLibraries:
		bindings = append(bindings, key("↑/↓", "move"))
	case focusPM:
		bindings = append(bindings, key("←/→/↑/↓", "move"))
	}
	if m.focus == focusTooling || m.focus == focusLibraries || m.focus == focusPM {
		bindings = append(bindings, key("space", "select"))
	}

	bindings = append(bindings, key("enter", "confirm"), key("esc", "quit"))

	line := strings.Join(bindings, FooterSepStyle.Render("  •  "))
	line += "\n" + FooterDescStyle.Render("* recommended")

	return FooterBarStyle.Render(line)
}

func (m WizardModel) focusNext() uint {
	if m.focus == focusLen-1 {
		return focusProjectName
	}

	return m.focus + 1
}

func (m WizardModel) focusPrev() uint {
	if m.focus == 0 {
		return focusLen - 1
	}

	return m.focus - 1
}

// Helper functions
func cursorNext(current, total uint) uint {
	if current == total-1 {
		return 0
	}

	return current + 1
}

func cursorPrev(current, total uint) uint {
	if current == 0 {
		return total - 1
	}

	return current - 1
}
