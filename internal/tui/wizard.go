package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	focusTooling            // 2 (Middle panel: CSS, Formatter, Linter, Test, Audit)
	focusLibraries          // 3 (Right panel: Validation, Form, Query, State, CMS)
	focusPM                 // 4 (Package manager horizontal selector)
	focusAdvanced           // 5 (Advanced options dropdown row)
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

// advItem is one row in the advanced fold: a labeled horizontal value picker.
type advItem struct {
	name    string
	options []Option // label/value pairs
	cursor  int      // selected option index
}

// AdvancedModel is a collapsible set of low-frequency settings (version
// channel, pin strategy, install, git, node engine). It renders as a single
// muted line unless focused, keeping the common path uncluttered.
type AdvancedModel struct {
	items    []advItem
	row      int
	expanded bool
}

func (a *AdvancedModel) Toggle()  { a.expanded = !a.expanded }
func (a *AdvancedModel) RowDown() { a.row = (a.row + 1) % len(a.items) }
func (a *AdvancedModel) RowUp()   { a.row = (a.row - 1 + len(a.items)) % len(a.items) }
func (a *AdvancedModel) ValueRight() {
	it := &a.items[a.row]
	it.cursor = (it.cursor + 1) % len(it.options)
}
func (a *AdvancedModel) ValueLeft() {
	it := &a.items[a.row]
	it.cursor = (it.cursor - 1 + len(it.options)) % len(it.options)
}

// value returns the selected value for the named item.
func (a *AdvancedModel) value(name string) string {
	for _, it := range a.items {
		if it.name == name {
			return it.options[it.cursor].value
		}
	}
	return ""
}

func (a AdvancedModel) View(active bool) string {
	caret := "▸"
	label := "Advanced options"
	head := MutedStyle.Render(label)
	if active {
		head = AccentStyle.Render(label)
	}
	if !a.expanded {
		hint := FooterDescStyle.Render("  (space to expand)")
		if !active {
			hint = ""
		}
		return caret + " " + head + hint
	}

	var b strings.Builder
	b.WriteString("▾ " + head + "\n")
	for i, it := range a.items {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  %-13s", it.name+":")))
		for j, opt := range it.options {
			style := lipgloss.NewStyle().Padding(0, 1)
			selected := j == it.cursor
			focused := active && i == a.row
			switch {
			case focused && selected:
				style = style.Background(ColorGreen).Foreground(ColorGray1).Bold(true)
			case selected:
				style = style.Background(ColorGreen).Foreground(ColorGray1)
			case focused:
				style = style.Foreground(ColorLuster)
			default:
				style = style.Foreground(ColorLack)
			}
			b.WriteString(style.Render(opt.label))
		}
	}
	return b.String()
}

// newAdvancedModel seeds the fold's selectors from the config defaults so the
// highlighted option matches what scaffolding would use if left untouched.
func newAdvancedModel(cfg pkg.ProjectConfig) AdvancedModel {
	boolOpts := func(trueFirst bool) []Option {
		yes := Option{label: "yes", value: "true"}
		no := Option{label: "no", value: "false"}
		if trueFirst {
			return []Option{yes, no}
		}
		return []Option{no, yes}
	}
	return AdvancedModel{items: []advItem{
		{name: "Channel", options: []Option{{"pinned", "pinned"}, {"latest", "latest"}}},
		{name: "Pin", options: []Option{{"default", "default"}, {"caret", "caret"}, {"tilde", "tilde"}, {"exact", "exact"}}},
		{name: "Install", options: boolOpts(cfg.Install)},
		{name: "Git init", options: boolOpts(cfg.GitInit)},
		{name: "Node", options: []Option{{cfg.NodeEngine, cfg.NodeEngine}, {">=20.11.0", ">=20.11.0"}, {">=18.18.0", ">=18.18.0"}}},
	}}
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
	advanced    AdvancedModel
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
	disabled bool
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

func (a *AddOnsModel) groupIndex(name string) int {
	for i, g := range a.groups {
		if g.name == name {
			return i
		}
	}
	return -1
}

func (a *AddOnsModel) CursorDown() {
	total := a.totalItems()
	if total == 0 {
		return
	}
	for range total {
		a.cursor = (a.cursor + 1) % total
		gi, _ := a.cursorPos()
		if !a.groups[gi].disabled {
			return
		}
	}
}

func (a *AddOnsModel) CursorUp() {
	total := a.totalItems()
	if total == 0 {
		return
	}
	for range total {
		a.cursor = (a.cursor - 1 + total) % total
		gi, _ := a.cursorPos()
		if !a.groups[gi].disabled {
			return
		}
	}
}

func (a *AddOnsModel) Select() {
	gi, ii := a.cursorPos()
	if a.groups[gi].disabled {
		return
	}
	a.groups[gi].selected = ii
}

func (a *AddOnsModel) View(active bool, width int) string {
	var s strings.Builder

	flatIdx := 0
	for i, g := range a.groups {
		if i > 0 {
			s.WriteString("\n")
		}

		if g.disabled {
			s.WriteString(MutedStyle.Render(g.name) + "\n")
			for _, opt := range g.options {
				text := " ◦ " + opt.label
				s.WriteString(lipgloss.NewStyle().Width(width).Foreground(ColorGray3).Render(text) + "\n")
				flatIdx++
			}
			continue
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

	defaults := pkg.NewProjectConfig()
	m := WizardModel{
		projectName: ti,
		BaseList:    baseList,
		tooling:     tooling,
		libraries:   libraries,
		pm:          pm,
		advanced:    newAdvancedModel(defaults),
		Cfg:         defaults,
	}
	m.syncLibraryConstraints()
	return m
}

// syncLibraryConstraints disables the CI/CD group when no deploy target is selected.
func (m *WizardModel) syncLibraryConstraints() {
	deployIdx := m.libraries.groupIndex("Deploy")
	cicdIdx := m.libraries.groupIndex("CI/CD")
	if deployIdx < 0 || cicdIdx < 0 {
		return
	}
	deploy := &m.libraries.groups[deployIdx]
	cicd := &m.libraries.groups[cicdIdx]
	noDeploySelected := deploy.options[deploy.selected].value == "none"
	if noDeploySelected {
		cicd.disabled = true
		cicd.selected = 0
	} else {
		cicd.disabled = false
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
		{"Test", reg.Test},
		{"Audit", reg.Audit},
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
		{"Deploy", reg.Deployment},
		{"CI/CD", reg.CICD},
		{"Backend", reg.Backend},
		{"ORM", reg.ORM},
		{"Database", reg.Database},
	}

	build := func(cats []struct {
		name    string
		entries []pkg.OptionEntry
	},
	) AddOnsModel {
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
			if m.focus == focusAdvanced && m.advanced.expanded {
				if k := msg.String(); k == "down" || k == "j" {
					m.advanced.RowDown()
				} else {
					m.advanced.ValueRight()
				}
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
			if m.focus == focusAdvanced && m.advanced.expanded {
				if k := msg.String(); k == "up" || k == "k" {
					m.advanced.RowUp()
				} else {
					m.advanced.ValueLeft()
				}
				return m, nil
			}
		case "space":
			if m.focus == focusTooling {
				m.tooling.Select()
				return m, nil
			}
			if m.focus == focusLibraries {
				m.libraries.Select()
				m.syncLibraryConstraints()
				return m, nil
			}
			if m.focus == focusPM {
				m.pm.Select()
				return m, nil
			}
			if m.focus == focusAdvanced {
				m.advanced.Toggle()
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
	if name == "." {
		cwd, _ := os.Getwd()
		m.Cfg.DestDir = "."
		name = filepath.Base(cwd)
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
		case "Test":
			m.Cfg.Test = pkg.TestingFramework(selected.value)
		case "Audit":
			m.Cfg.Audit = pkg.AuditTool(selected.value)
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
		case "Deploy":
			m.Cfg.Deployment = pkg.DeployTarget(selected.value)
		case "CI/CD":
			m.Cfg.CICD = pkg.CICDProvider(selected.value)
		case "Backend":
			m.Cfg.Backend = pkg.BackendLib(selected.value)
		case "ORM":
			m.Cfg.ORM = pkg.ORMLib(selected.value)
		case "Database":
			m.Cfg.Database = pkg.Database(selected.value)
		}
	}

	m.Cfg.PM = pkg.PackageManager(m.pm.options[m.pm.selected].Value)

	m.Cfg.Channel = pkg.VersionChannel(m.advanced.value("Channel"))
	m.Cfg.Pin = pkg.PinStrategy(m.advanced.value("Pin"))
	m.Cfg.Install = m.advanced.value("Install") == "true"
	m.Cfg.GitInit = m.advanced.value("Git init") == "true"
	m.Cfg.NodeEngine = m.advanced.value("Node")

	// A selected backend implies the monorepo layout (pnpm only).
	m.Cfg.ApplyDefaultLayout()
}

func (m WizardModel) View() tea.View {
	var c *tea.Cursor
	if !m.projectName.VirtualCursor() {
		c = m.projectName.Cursor()
		c.Y += lipgloss.Height("10")
	}

	var s strings.Builder

	lists := m.middleRow()
	layout := lipgloss.JoinVertical(lipgloss.Top, m.projectNameInputView(), lists, m.pmView(), m.advancedView(), m.footerView())

	s.Write([]byte(layout))

	if m.screen == screenSummary {
		return tea.NewView(m.centered(m.summaryPopup()))
	}

	return tea.NewView(s.String())
}

// centered places an overlay in the middle of the terminal.
func (m WizardModel) centered(overlay string) string {
	w, h := m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, overlay)
}

func (m WizardModel) summaryPopup() string {
	title := AccentStyle.Render("  Confirm Project") + "\n\n"

	row := func(label, value string) string {
		if value == "none" {
			return ""
		}
		return MutedStyle.Render("  "+label) + PrimaryStyle.Render(value) + "\n"
	}

	yesno := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	body := row("Project:    ", m.Cfg.ProjectName) +
		row("Base:       ", string(m.Cfg.Base)) +
		row("CSS:        ", string(m.Cfg.CSS)) +
		row("Formatter:  ", string(m.Cfg.Fmt)) +
		row("Linter:     ", string(m.Cfg.Linter)) +
		row("Test:       ", string(m.Cfg.Test)) +
		row("Audit:      ", string(m.Cfg.Audit)) +
		row("Validation: ", string(m.Cfg.Validation)) +
		row("Form:       ", string(m.Cfg.Form)) +
		row("Query:      ", string(m.Cfg.Query)) +
		row("State:      ", string(m.Cfg.State)) +
		row("CMS:        ", string(m.Cfg.CMS)) +
		row("Deploy:     ", string(m.Cfg.Deployment)) +
		row("CI/CD:      ", string(m.Cfg.CICD)) +
		row("Backend:    ", string(m.Cfg.Backend)) +
		row("ORM:        ", string(m.Cfg.ORM)) +
		row("Database:   ", string(m.Cfg.Database)) +
		row("Layout:     ", string(m.Cfg.Layout)) +
		row("PM:         ", string(m.Cfg.PM)) +
		row("Channel:    ", string(m.Cfg.Channel)) +
		row("Pin:        ", string(m.Cfg.Pin)) +
		row("Install:    ", yesno(m.Cfg.Install)) +
		row("Git init:   ", yesno(m.Cfg.GitInit)) +
		row("Node:       ", m.Cfg.NodeEngine)

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

func (m WizardModel) advancedView() string {
	box := m.borderFor(focusAdvanced).Width(fullRowWidth)
	return box.Render(m.advanced.View(m.focus == focusAdvanced))
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
		m.syncLibraryConstraints()
	}
}

func (m WizardModel) borderFor(section uint) lipgloss.Style {
	if m.focus == section {
		return ActiveBorder
	}
	return InactiveBorder
}

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
	if m.focus == focusAdvanced {
		if m.advanced.expanded {
			bindings = append(bindings, key("↑/↓", "row"), key("←/→", "change"), key("space", "collapse"))
		} else {
			bindings = append(bindings, key("space", "expand"))
		}
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
