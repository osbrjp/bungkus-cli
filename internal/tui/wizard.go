package tui

import (
	"encoding/json"
	"io/fs"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/osbrjp/bungkus-cli/pkg"
)

// Version is the bungkus-cli version shown in the header; cmd.SetVersion
// sets it at startup.
var Version = pkg.DevVersion

// UpdateAvailable is the newer release tag (e.g. "v1.9.0") the header
// announces, or "" when no update is known at startup.
var UpdateAvailable string

// pane identifies one of the three bento panes.
type pane int

const (
	paneSteps pane = iota
	paneOptions
	panePreview
	paneCount
)

// phase is where the wizard is in its life: editing the config, running the
// scaffold, or showing the scaffold's result.
type phase int

const (
	phaseEdit phase = iota
	phaseCreating
	phaseDone
)

// advItem is one row of the advanced step: a label and the values ←/→ cycle
// through. The first option is always the default.
type advItem struct {
	name    string
	options []string
	cursor  int
}

// advancedModel holds the low-frequency settings of the advanced step:
// version channel, pin strategy, install, git init and the node engine.
type advancedModel struct {
	items []advItem
	row   int
}

// newAdvancedModel seeds the advanced rows so each starts on cfg's value.
func newAdvancedModel(cfg pkg.ProjectConfig) advancedModel {
	boolOpts := func(def bool) []string {
		if def {
			return []string{"yes", "no"}
		}
		return []string{"no", "yes"}
	}
	return advancedModel{items: []advItem{
		{name: "channel", options: []string{string(pkg.ChannelPinned), string(pkg.ChannelLatest)}},
		{name: "pin", options: []string{string(pkg.PinDefault), string(pkg.PinCaret), string(pkg.PinTilde), string(pkg.PinExact)}},
		{name: "install", options: boolOpts(cfg.Install)},
		{name: "git init", options: boolOpts(cfg.GitInit)},
		{name: "node", options: []string{cfg.NodeEngine, ">=20.11.0", ">=18.18.0"}},
	}}
}

// shift moves the focused row's value by d (±1), wrapping around.
func (a *advancedModel) shift(d int) {
	it := &a.items[a.row]
	it.cursor = (it.cursor + d + len(it.options)) % len(it.options)
}

// value returns the selected value of the named row.
func (a advancedModel) value(name string) string {
	for _, it := range a.items {
		if it.name == name {
			return it.options[it.cursor]
		}
	}
	return ""
}

// changed counts the rows moved off their default.
func (a advancedModel) changed() int {
	n := 0
	for _, it := range a.items {
		if it.cursor != 0 {
			n++
		}
	}
	return n
}

// apply writes the advanced settings into cfg.
func (a advancedModel) apply(cfg *pkg.ProjectConfig) {
	cfg.Channel = pkg.VersionChannel(a.value("channel"))
	cfg.Pin = pkg.PinStrategy(a.value("pin"))
	cfg.Install = a.value("install") == "yes"
	cfg.GitInit = a.value("git init") == "yes"
	cfg.NodeEngine = a.value("node")
}

// WizardModel is the interactive create wizard: a steps pane listing every
// choice, an options pane editing the selected step, and a preview pane with
// the equivalent command, layout and dependencies. Scaffolding runs in the
// preview pane; after the program exits the caller reads Cfg, Canceled,
// Created and Err and runs the post-scaffold steps (install, git init),
// whose subprocess output needs the normal terminal.
type WizardModel struct {
	// Cfg is the configuration being built; it is always consistent
	// (see normalize).
	Cfg pkg.ProjectConfig
	// Canceled is true when the user quit before scaffolding finished.
	Canceled bool
	// Created is true when the project was scaffolded without error.
	Created bool
	// Err is the scaffolding error, if any.
	Err error

	templates     fs.FS
	width, height int
	focus         pane
	step          int // index into steps
	opt           int // cursor in the options pane
	scroll        int // first visible line of the preview pane
	help          bool
	phase         phase
	blocked       string // why the last create was refused
	note          string // why the previous step kept its value; shown on the next step
	name          textinput.Model
	advanced      advancedModel
	spin          spinner.Model
	deps          depsMsg // package.json preview; stale unless deps.cfg == Cfg
	wd, home      string
	quote         int // index into quotes shown in the header
	duck          int // position in duckLoop while scaffolding
}

// NewWizardModel returns the wizard on its first step with a random header
// quote, starting from pkg.NewProjectConfig with every choice on its
// recommended or first option (see initialPicks). templates is the template
// tree pkg.Scaffold renders from when the user creates the project.
func NewWizardModel(templates fs.FS) WizardModel {
	ti := textinput.New()
	ti.Placeholder = "my-app"
	ti.CharLimit = 214
	ti.Prompt = ""

	cfg := pkg.NewProjectConfig()
	wd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	m := WizardModel{
		Cfg:       cfg,
		templates: templates,
		name:      ti,
		advanced:  newAdvancedModel(cfg),
		spin:      spinner.New(spinner.WithSpinner(spinner.Dot)),
		wd:        wd,
		home:      home,
		quote:     rand.IntN(len(quotes)),
	}
	initialPicks(&m.Cfg)
	return m
}

// Init starts loading the dependency preview.
func (m WizardModel) Init() tea.Cmd {
	return loadDeps(m.Cfg)
}

// appDeps is one package.json of the preview: its app name and its
// dependencies then devDependencies, each sorted by name, as "name version".
type appDeps struct {
	name string
	pkgs []string
}

// depsMsg carries the package.json preview built for cfg.
type depsMsg struct {
	cfg    pkg.ProjectConfig
	apps   []appDeps
	extras []string // web packages only a combination of picks adds
	err    error
}

// scaffoldedMsg reports the end of pkg.Scaffold.
type scaffoldedMsg struct{ err error }

// duckMsg advances the header mascot's duck animation by one frame.
type duckMsg struct{}

// duckTick schedules the next duck frame, or nothing without a terminal.
func duckTick() tea.Cmd {
	if !animate {
		return nil
	}
	return tea.Tick(duckInterval, func(time.Time) tea.Msg { return duckMsg{} })
}

// pkgBuilder names one package.json the preview shows and builds it.
type pkgBuilder struct {
	name  string
	build func(pkg.ProjectConfig) ([]byte, error)
}

// loadDeps builds the package.json files cfg would produce: web alone in the
// flat layout, web, api and domain in the monorepo. It runs off the update
// loop because the builder shells out to the package manager for its version.
func loadDeps(cfg pkg.ProjectConfig) tea.Cmd {
	return func() tea.Msg {
		msg := depsMsg{cfg: cfg}
		builders := []pkgBuilder{{"web", pkg.BuildPackageJSON}}
		if cfg.Layout.IsMonorepo() {
			builders = append(builders, pkgBuilder{"api", pkg.BuildAPIPackageJSON}, pkgBuilder{"domain", pkg.BuildDomainPackageJSON})
		}
		for _, b := range builders {
			raw, err := b.build(cfg)
			if err != nil {
				msg.err = err
				return msg
			}
			var p struct {
				Dependencies, DevDependencies map[string]string
			}
			if err := json.Unmarshal(raw, &p); err != nil {
				msg.err = err
				return msg
			}
			app := appDeps{name: b.name}
			for _, set := range []map[string]string{p.Dependencies, p.DevDependencies} {
				for _, n := range slices.Sorted(maps.Keys(set)) {
					app.pkgs = append(app.pkgs, n+" "+set[n])
				}
			}
			msg.apps = append(msg.apps, app)
			if b.name == "web" {
				msg.extras = comboExtras(cfg, p.Dependencies, p.DevDependencies)
			}
		}
		return msg
	}
}

// comboExtras returns the names in sets that neither the registry's common
// packages nor any single pick lists as its own: what the package.json
// builder adds for a combination (prettier with tailwind, react-hook-form
// with zod, a database driver, …). The result is sorted.
func comboExtras(cfg pkg.ProjectConfig, sets ...map[string]string) []string {
	common := pkg.GetRegistry().CommonPackages
	known := map[string]bool{"domain": true}
	for n := range common.Dependencies {
		known[n] = true
	}
	for n := range common.DevDependencies {
		known[n] = true
	}
	for _, s := range steps {
		if s.kind != kindChoice {
			continue
		}
		v := s.get(cfg)
		for _, c := range choices(s, cfg) {
			if c.value == v {
				for _, n := range c.adds {
					known[n] = true
				}
			}
		}
	}
	var out []string
	for _, set := range sets {
		for n := range set {
			if !known[n] {
				out = append(out, n)
			}
		}
	}
	slices.Sort(out)
	return out
}

// runScaffold renders the project into dest.
func runScaffold(dest string, templates fs.FS, cfg pkg.ProjectConfig) tea.Cmd {
	return func() tea.Msg {
		return scaffoldedMsg{err: pkg.Scaffold(dest, templates, cfg)}
	}
}

// Update handles window size, async results and keys.
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case depsMsg:
		if msg.cfg == m.Cfg {
			m.deps = msg
		}
		return m, nil
	case spinner.TickMsg:
		if m.phase != phaseCreating {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case duckMsg:
		if m.phase != phaseCreating {
			return m, nil
		}
		m.duck = (m.duck + 1) % len(duckLoop)
		return m, duckTick()
	case scaffoldedMsg:
		m.phase = phaseDone
		m.duck = 0
		m.Err = msg.err
		m.Created = msg.err == nil
		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg)
	}
	if m.editingName() {
		var cmd tea.Cmd
		m.name, cmd = m.name.Update(msg)
		return m, cmd
	}
	return m, nil
}

// editingName reports whether printable keys go to the name field.
func (m WizardModel) editingName() bool {
	return m.phase == phaseEdit && !m.help && m.focus == paneOptions && steps[m.step].kind == kindName
}

// key dispatches one key press. The order encodes precedence: ctrl+c always
// quits, then the scaffold phases, the help overlay, the name field, and
// finally the pane keys.
func (m WizardModel) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" {
		m.Canceled = m.phase != phaseDone
		return m, tea.Quit
	}
	switch m.phase {
	case phaseCreating:
		return m, nil
	case phaseDone:
		switch k {
		case "enter", "q", "esc":
			return m, tea.Quit
		}
		return m, nil
	}
	if m.help {
		m.help = false
		return m, nil
	}

	if m.editingName() {
		switch k {
		case "tab", "shift+tab", "enter", "esc":
		default:
			var cmd tea.Cmd
			m.name, cmd = m.name.Update(msg)
			return m, tea.Batch(cmd, m.syncName())
		}
	}

	switch k {
	case "q":
		if m.focus != panePreview {
			m.Canceled = true
			return m, tea.Quit
		}
	case "?":
		m.help = true
	case "r":
		m.setStep(len(steps) - 1)
		return m.focusPane(paneOptions)
	case "esc":
		return m.focusPane(paneSteps)
	case "tab":
		return m.focusPane((m.focus + 1) % paneCount)
	case "shift+tab":
		return m.focusPane((m.focus + paneCount - 1) % paneCount)
	case "left", "h", "right", "l":
		d := 1
		if k == "left" || k == "h" {
			d = -1
		}
		// In the advanced step h/l and ←/→ change the row's value, like vim
		// motions everywhere else; esc and tab still leave the pane.
		if m.focus == paneOptions && steps[m.step].kind == kindAdvanced {
			m.advanced.shift(d)
			return m, m.refresh()
		}
		return m.focusPane(pane(min(max(int(m.focus)+d, 0), int(paneCount)-1)))
	case "down", "j":
		m.move(1)
	case "up", "k":
		m.move(-1)
	case "space":
		if m.focus == paneOptions {
			return m, m.pick()
		}
	case "enter":
		return m.enter()
	}
	return m, nil
}

// focusPane moves focus to p and focuses or blurs the name field to match.
func (m WizardModel) focusPane(p pane) (tea.Model, tea.Cmd) {
	m.focus = p
	if m.editingName() {
		return m, m.name.Focus()
	}
	m.name.Blur()
	return m, nil
}

// enter opens the options of the selected step from the steps pane. In the
// options pane it picks the option under the cursor and moves to the next
// step, or creates the project on the review step; an option that cannot be
// picked leaves everything as it is.
func (m WizardModel) enter() (tea.Model, tea.Cmd) {
	switch m.focus {
	case paneSteps:
		return m.focusPane(paneOptions)
	case paneOptions:
		s := steps[m.step]
		if s.kind == kindReview {
			return m.create()
		}
		var cmd tea.Cmd
		note := ""
		if s.kind == kindChoice {
			if c := choices(s, m.Cfg)[m.opt]; c.reason != "" {
				note = keptNote(s, m.Cfg, c)
			} else {
				cmd = m.pick()
			}
		}
		m.setStep(m.step + 1)
		m.note = note
		next, focusCmd := m.focusPane(paneOptions)
		return next, tea.Batch(cmd, focusCmd)
	}
	return m, nil
}

// keptNote explains why enter on an option the stack can't use moved on
// without picking it, e.g. "React Hook Form needs React · kept None".
func keptNote(s stepDef, cfg pkg.ProjectConfig, c choice) string {
	kept := s.get(cfg)
	for _, o := range choices(s, cfg) {
		if o.value == kept {
			kept = o.label
			break
		}
	}
	return c.label + " " + c.reason + " · kept " + kept
}

// setStep selects step i (clamped) and puts the options cursor on that
// step's current value.
func (m *WizardModel) setStep(i int) {
	m.step = min(max(i, 0), len(steps)-1)
	m.opt = 0
	m.advanced.row = 0
	m.blocked = ""
	m.note = ""
	if s := steps[m.step]; s.kind == kindChoice {
		v := s.get(m.Cfg)
		m.opt = max(0, slices.IndexFunc(choices(s, m.Cfg), func(c choice) bool { return c.value == v }))
	}
}

// move handles j/k in the focused pane: the step cursor, the option or
// advanced-row cursor, or the preview scroll offset.
func (m *WizardModel) move(d int) {
	switch m.focus {
	case paneSteps:
		m.setStep(m.step + d)
	case paneOptions:
		switch s := steps[m.step]; s.kind {
		case kindChoice:
			m.opt = min(max(m.opt+d, 0), len(choices(s, m.Cfg))-1)
		case kindAdvanced:
			m.advanced.row = min(max(m.advanced.row+d, 0), len(m.advanced.items)-1)
		}
	case panePreview:
		w, h := m.paneBox(panePreview)
		m.scroll = min(max(m.scroll+d, 0), max(0, len(m.previewLines(w-2))-(h-2)))
	}
}

// pick applies the option under the cursor of the options pane: a choice
// step's option when it can be picked, or the next value of the focused
// advanced row. It returns the command reloading the dependency preview.
func (m *WizardModel) pick() tea.Cmd {
	switch s := steps[m.step]; s.kind {
	case kindChoice:
		c := choices(s, m.Cfg)[m.opt]
		if c.reason != "" {
			return nil
		}
		s.set(&m.Cfg, c.value)
	case kindAdvanced:
		m.advanced.shift(1)
	default:
		return nil
	}
	return m.refresh()
}

// refresh re-derives Cfg after an edit and reloads the dependency preview.
func (m *WizardModel) refresh() tea.Cmd {
	m.advanced.apply(&m.Cfg)
	normalize(&m.Cfg)
	return loadDeps(m.Cfg)
}

// syncName copies the name field into Cfg as `create` reads the same
// argument: empty means "my-app" and "." means the current directory.
func (m *WizardModel) syncName() tea.Cmd {
	switch v := m.name.Value(); v {
	case "":
		m.Cfg.ProjectName, m.Cfg.DestDir = "my-app", ""
	case ".":
		m.Cfg.ProjectName, m.Cfg.DestDir = filepath.Base(m.wd), "."
	default:
		m.Cfg.ProjectName, m.Cfg.DestDir = v, ""
	}
	return loadDeps(m.Cfg)
}

// dest is the directory the project is scaffolded into.
func (m WizardModel) dest() string {
	if m.Cfg.DestDir != "" {
		return m.Cfg.DestDir
	}
	return m.Cfg.ProjectName
}

// create validates the name and destination with the checks `create`
// applies, then starts scaffolding in the preview pane. A refusal is shown on
// the review step instead.
func (m WizardModel) create() (tea.Model, tea.Cmd) {
	if ok, msg := nameStatus(m.name.Value()); !ok {
		m.blocked = msg
		return m, nil
	}
	if err := pkg.ValidateDest(m.dest()); err != nil {
		m.blocked = err.Error()
		return m, nil
	}
	m.phase = phaseCreating
	m.focus = panePreview
	m.name.Blur()
	m.duck = 0
	return m, tea.Batch(m.spin.Tick, duckTick(), runScaffold(m.dest(), m.templates, m.Cfg))
}
