package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/osbrjp/bungkus-cli/pkg"
)

// Layout constants of the bento screen.
const (
	minWidth    = 60         // narrower than this, only a notice is drawn
	minHeight   = 20         // shorter than this, only a notice is drawn
	wideWidth   = 100        // from this width all three panes show side by side
	stepsWidth  = 36         // outer width of the steps pane
	optionsMin  = 34         // narrowest outer width of the options pane
	optionsMax  = 48         // widest outer width of the options pane
	headerRows  = mascotRows // the header text sits beside the mascot
	statusRows  = 1
	maxPkgLines = 3 // packages listed under an option before "…"
)

// quotes are the header's line-4 texts when no update is announced; one is
// picked at random per launch.
var quotes = []string{
	"Your code, bungkus'd to go.",
	"Rice to meet you. Let's scaffold something.",
	"Tapau your worries, scaffold your project.",
	"Wrap it up, ship it out. Bungkus!",
	"Bungkus satu! One project, to go.",
	"Extra rice, extra tests.",
	"No sambal, no deploy.",
	"It's not a bug, it's ikan bilis — tiny, crunchy, everywhere.",
	"Nasi sudah menjadi bubur — commit before you refactor.",
	"Egg, sambal, ikan bilis, kacang, timun. Five ingredients, zero config.",
	"Keep dependencies like sambal: fresh, and not too much.",
	"Lemak is great in rice, not in bundles.",
	"Banana leaf: the original zero-waste packaging. Keep your bundle small too.",
	"Wrap your code like nasi bungkus: tight, tidy, nothing leaks.",
	"Nasi lemak: full-stack since forever.",
	"Teh tarik is pulled, not force-pushed.",
	"The mamak is open 24/7. Your on-call shouldn't be.",
	"Too many cooks spoil the nasi. One task per branch.",
	"Don't let your merge conflicts go cold like yesterday's rice.",
	"Sambal on the side, secrets in .env.",
}

var (
	headingStyle = lipgloss.NewStyle().Foreground(ColorMuted).Bold(true)
	quoteStyle   = FooterDescStyle.Italic(true)
	okStyle      = lipgloss.NewStyle().Foreground(ColorOK)
	cmdStyle     = lipgloss.NewStyle().Foreground(ColorAccent)
)

// frame wraps a rendered screen in the view every screen uses: alternate
// screen, painted background.
func frame(s string) tea.View {
	v := tea.NewView(s)
	v.AltScreen = true
	return paint(v)
}

// View renders the whole screen at the current terminal size: the header,
// the panes (or the help overlay) and the status bar, exactly height rows.
// Under 100 columns only the focused pane shows; under 60×20 a notice does.
func (m WizardModel) View() tea.View {
	if m.width == 0 {
		return frame("")
	}
	if m.width < minWidth || m.height < minHeight {
		return frame(fit(fmt.Sprintf("terminal too small: need %d×%d, have %d×%d", minWidth, minHeight, m.width, m.height), m.width))
	}

	rows := m.header()
	if m.help {
		_, h := m.paneBox(paneSteps)
		rows = append(rows, box("help", helpLines(), 0, m.width, h, true)...)
	} else {
		var cols [][]string
		for p := range paneCount {
			if m.width < wideWidth && p != m.focus {
				continue
			}
			w, h := m.paneBox(p)
			title, lines, off := m.paneContent(p, w-2, h-2)
			cols = append(cols, box(title, lines, off, w, h, p == m.focus))
		}
		for i := range cols[0] {
			var b strings.Builder
			for _, c := range cols {
				b.WriteString(c[i])
			}
			rows = append(rows, b.String())
		}
	}
	rows = append(rows, m.statusBar())
	return frame(strings.Join(rows, "\n"))
}

// paneBox returns the outer size of pane p, or the full width in single-pane
// mode. The steps pane is fixed; the options pane gets 40% of the rest,
// between 34 and 48 columns; the preview takes what remains, which is at
// least 30 from the 100-column wide layout up.
// Every pane is as tall as the space between header and status bar.
func (m WizardModel) paneBox(p pane) (w, h int) {
	h = m.height - headerRows - statusRows
	rest := m.width - stepsWidth
	options := min(optionsMax, max(optionsMin, (rest*4+5)/10))
	switch {
	case m.width < wideWidth:
		return m.width, h
	case p == paneSteps:
		return stepsWidth, h
	case p == paneOptions:
		return options, h
	}
	return rest - options, h
}

// paneContent returns pane p's title, body lines for an inner width iw, and
// the first line to show so the cursor stays inside a body of bh rows.
func (m WizardModel) paneContent(p pane, iw, bh int) (title string, lines []string, off int) {
	switch p {
	case paneSteps:
		lines, from, to := m.stepsLines(iw)
		return "steps", lines, follow(from, to, bh, len(lines))
	case paneOptions:
		lines, from, to := m.optionsLines(iw)
		return steps[m.step].name, lines, follow(from, to, bh, len(lines))
	}
	lines = m.previewLines(iw)
	return "preview", lines, min(m.scroll, max(0, len(lines)-bh))
}

// follow returns the scroll offset that shows lines from..to (inclusive) of n
// in a body of bh rows, scrolling as little as possible from the top.
func follow(from, to, bh, n int) int {
	off := min(max(0, to+1-bh), from)
	return max(0, min(off, n-bh))
}

// box draws a pane in the border of InactiveBorder, or ActiveBorder when
// focused, with the title set into the top edge (plus " ↕" when the body
// does not fit) and bh = h-2 body rows starting at lines[off], each cut or
// padded to the inner width w-2. It returns exactly h rows of width w.
func box(title string, lines []string, off, w, h int, focused bool) []string {
	frameStyle := InactiveBorder
	if focused {
		frameStyle = ActiveBorder
	}
	b, st := frameStyle.GetBorderStyle(), lipgloss.NewStyle().Foreground(frameStyle.GetBorderTopForeground())
	iw, bh := w-2, h-2
	if len(lines) > bh {
		title += " ↕"
	}
	t := " " + title + " "
	tst := st
	if focused {
		tst = tst.Bold(true)
	}
	out := make([]string, 0, h)
	out = append(out, st.Render(b.TopLeft+b.Top)+tst.Render(t)+st.Render(strings.Repeat(b.Top, max(0, iw-1-ansi.StringWidth(t)))+b.TopRight))
	for i := range bh {
		l := ""
		if off+i < len(lines) {
			l = lines[off+i]
		}
		out = append(out, st.Render(b.Left)+fit(l, iw)+st.Render(b.Right))
	}
	return append(out, st.Render(b.BottomLeft+strings.Repeat(b.Bottom, iw)+b.BottomRight))
}

// fit cuts s to w columns (ending in "…" when cut) and pads it with spaces
// to exactly w columns. ANSI styling does not count toward the width.
func fit(s string, w int) string {
	s = ansi.Truncate(s, w, "…")
	return s + strings.Repeat(" ", max(0, w-ansi.StringWidth(s)))
}

// spread returns left and right separated by enough spaces to fill w
// columns; left is cut first when both do not fit.
func spread(left, right string, w int) string {
	lw := w - ansi.StringWidth(right) - 1
	if lw < 1 {
		return fit(right, w)
	}
	return fit(left, lw) + " " + right
}

// header renders the 4 header rows: the mascot, then beside it the name and
// version (or, in single-pane mode, a breadcrumb of the focused pane), the
// author, the working directory with $HOME as "~", and the update notice or
// the launch's quote. Each row is cut to the terminal width.
func (m WizardModel) header() []string {
	muted := FooterDescStyle
	first := BoldStyle.Render("bungkus-cli") + muted.Render(" · "+versionLabel(Version))
	if m.width < wideWidth {
		crumb := "create"
		switch m.focus {
		case paneOptions:
			crumb += " › " + steps[m.step].name
		case panePreview:
			crumb += " › preview"
		}
		first += "  " + cmdStyle.Render(crumb)
	}
	last := quoteStyle.Render(`"` + quotes[m.quote%len(quotes)] + `"`)
	if UpdateAvailable != "" {
		last = WarnStyle.Render("update available: " + UpdateAvailable + " · bungkus-cli update")
	}
	text := []string{first, muted.Render("by spencer · osbr"), muted.Render(tildePath(m.wd, m.home)), last}
	sprite := mascot()
	if m.phase == phaseCreating {
		sprite = mascotFrame(duckLoop[m.duck])
	}
	rows := make([]string, headerRows)
	for i := range rows {
		rows[i] = fit(sprite[i]+"    "+text[i], m.width)
	}
	return rows
}

// versionLabel formats a version for the header with one leading "v", or
// "dev" for builds without a version.
func versionLabel(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == pkg.DevVersion {
		return pkg.DevVersion
	}
	return "v" + v
}

// tildePath shows path with the home directory prefix replaced by "~".
func tildePath(path, home string) string {
	if home != "" && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + path[len(home):]
	}
	return path
}

// stepValue is the value a step shows in the steps pane and review: the
// project name, the choice's registry value, or a summary of the advanced
// settings. Review has none.
func (m WizardModel) stepValue(s stepDef) string {
	switch s.kind {
	case kindName:
		if m.Cfg.DestDir == "." {
			return "."
		}
		return m.Cfg.ProjectName
	case kindAdvanced:
		if n := m.advanced.changed(); n > 0 {
			return strconv.Itoa(n) + " changed"
		}
		return "default"
	case kindChoice:
		return s.get(m.Cfg)
	}
	return ""
}

// setCount returns how many frontend and backend steps have a value, and how
// many such steps exist, for the status bar's "N/M set".
func (m WizardModel) setCount() (n, total int) {
	for _, s := range steps {
		if s.group == "frontend" || s.group == "backend" {
			total++
			if isSet(m.stepValue(s)) {
				n++
			}
		}
	}
	return n, total
}

// stepsLines renders the steps pane for inner width iw: group headings and
// one "name   value" row per step ("-" for none), the value right-aligned and
// cut before the name is.
// from and to are the selected row's line index.
func (m WizardModel) stepsLines(iw int) (lines []string, from, to int) {
	group := "-"
	for i, s := range steps {
		if s.group != group {
			if i > 0 {
				lines = append(lines, "")
			}
			if s.group != "" {
				lines = append(lines, " "+headingStyle.Render(s.group))
			}
			group = s.group
		}
		v := m.stepValue(s)
		if s.kind != kindReview && !isSet(v) {
			v = "-"
		}
		// The index fills columns 1-2 beside the cursor marker in column 0,
		// right-aligned so 9 and 16 line up; digits jump to it. Review is 0.
		n := i + 1
		if s.kind == kindReview {
			n = 0
		}
		idx := fmt.Sprintf("%2d  ", n)
		row := spread(s.name, ansi.Truncate(v, iw-2-len(idx)-ansi.StringWidth(s.name), "…"), iw-1-len(idx))
		switch {
		case i == m.step && m.focus == paneSteps:
			from, to = len(lines), len(lines)
			row = CursorStyle.Render(fit(">"+idx+row, iw))
		case i == m.step:
			from, to = len(lines), len(lines)
			row = okStyle.Render(">"+idx) + BoldStyle.Render(row)
		default:
			row = " " + okStyle.Render(idx) + row
		}
		lines = append(lines, row)
	}
	return lines, from, to
}

// optionsLines renders the options pane for the selected step at inner width
// iw, and the line range of the cursor.
// nameCallToAction is the line under the name input that says how to move on:
// a highlighted enter badge and the next step while a valid name is being
// typed, a hint to open the field when it isn't focused, or a reminder to fix
// an invalid name first.
func nameCallToAction(valid, focused bool, next string) string {
	switch {
	case !valid:
		return FooterDescStyle.Render("fix the name to continue")
	case !focused:
		return FooterKeyStyle.Render("enter") + FooterDescStyle.Render(" to edit")
	default:
		return StatusModeStyle.Render(" ↵ enter ") + FooterDescStyle.Render(" next: "+next)
	}
}

func (m WizardModel) optionsLines(iw int) (lines []string, from, to int) {
	s := steps[m.step]
	focused := m.focus == paneOptions
	head := map[stepKind]string{kindChoice: "pick one", kindName: "type a name", kindAdvanced: "h/l ←/→ change", kindReview: "check and create"}
	lines = []string{" " + FooterDescStyle.Render(s.name+" · "+head[s.kind]), ""}
	if m.note != "" {
		for _, l := range strings.Split(ansi.Wrap(m.note, iw-2, ""), "\n") {
			lines = append(lines, " "+WarnStyle.Render(l))
		}
		lines = append(lines, "")
	}

	switch s.kind {
	case kindName:
		ti := m.name
		ti.SetWidth(iw - 4)
		lines = append(lines, " › "+ti.View(), "")
		from, to = 2, 2
		ok, msg := nameStatus(m.name.Value())
		switch {
		case !ok:
			for _, l := range strings.Split(ansi.Wrap("✘ "+msg, iw-2, ""), "\n") {
				lines = append(lines, " "+ErrorStyle.Render(l))
			}
		case m.name.Value() == "":
			lines = append(lines, " "+FooterDescStyle.Render(msg))
		default:
			lines = append(lines, " "+okStyle.Render("✔ "+msg))
		}
		lines = append(lines, "", " "+nameCallToAction(ok, focused, steps[m.step+1].name))

	case kindChoice:
		cur := s.get(m.Cfg)
		rec := false
		for i, c := range choices(s, m.Cfg) {
			radio := "( )"
			if c.value == cur {
				radio = okStyle.Render("(*)")
			}
			mark := ""
			if c.recommended {
				mark, rec = "*", true
			}
			row := spread(" "+radio+" "+c.label, mark+" ", iw)
			if c.reason != "" {
				row = lipgloss.NewStyle().Foreground(ColorDim).Render(ansi.Strip(row))
			}
			if focused && i == m.opt {
				row = CursorStyle.Render(ansi.Strip(row))
				from = len(lines)
			}
			lines = append(lines, row)
			if c.reason != "" {
				lines = append(lines, "     "+FooterDescStyle.Render(c.reason))
			} else {
				for j, n := range c.adds {
					if j == maxPkgLines {
						lines = append(lines, "     "+FooterDescStyle.Render("…"))
						break
					}
					lines = append(lines, "     "+FooterDescStyle.Render("+ "+n))
				}
			}
			if focused && i == m.opt {
				to = len(lines) - 1
			}
			lines = append(lines, "")
		}
		if rec {
			lines = append(lines, " "+FooterDescStyle.Render("* recommended"))
		}
		if m.deps.cfg == m.Cfg && len(m.deps.extras) > 0 {
			lines = append(lines, "", " "+headingStyle.Render("these picks together also add"))
			for _, n := range m.deps.extras {
				lines = append(lines, "   "+FooterDescStyle.Render("+ "+n))
			}
		}

	case kindAdvanced:
		for i, it := range m.advanced.items {
			v := it.options[it.cursor]
			row := fmt.Sprintf("%-10s", it.name)
			if focused && i == m.advanced.row {
				from, to = len(lines), len(lines)
				lines = append(lines, CursorStyle.Render(fit(" > "+row+"‹ "+v+" ›", iw)))
				continue
			}
			if it.cursor != 0 {
				v = okStyle.Render(v)
			}
			lines = append(lines, "   "+row+"  "+v)
		}

	case kindReview:
		for _, st := range steps {
			if v := m.stepValue(st); st.kind != kindReview && isSet(v) {
				lines = append(lines, " "+FooterDescStyle.Render(fmt.Sprintf("%-12s", st.name))+v)
			}
		}
		lines = append(lines,
			" "+FooterDescStyle.Render(fmt.Sprintf("%-12s", "layout"))+string(m.Cfg.Layout),
			" "+FooterDescStyle.Render(fmt.Sprintf("%-12s", "into"))+"./"+strings.TrimPrefix(m.dest(), "./"),
			"")
		if m.blocked != "" {
			for _, l := range strings.Split(ansi.Wrap("✘ "+m.blocked, iw-2, ""), "\n") {
				lines = append(lines, " "+ErrorStyle.Render(l))
			}
			lines = append(lines, "")
		}
		lines = append(lines, " "+keyHint("enter", "create")+FooterSepStyle.Render(" · ")+keyHint("esc", "back"))
		from, to = len(lines)-1, len(lines)-1
	}
	return lines, from, to
}

// commandLines renders `bungkus-cli create …` for args (see CreateArgs) as
// shell lines of at most w columns: each flag stays with its value, every
// line but the last ends in " \", and continuation lines are indented, so
// the block pastes into a shell as one command.
func commandLines(args []string, w int) []string {
	toks := []string{"bungkus-cli", args[0], shellQuote(args[1])}
	for i := 2; i < len(args); i++ {
		if strings.Contains(args[i], "=") || i+1 == len(args) {
			toks = append(toks, args[i])
			continue
		}
		toks = append(toks, args[i]+" "+shellQuote(args[i+1]))
		i++
	}
	cur := toks[0]
	var out []string
	for _, t := range toks[1:] {
		if ansi.StringWidth(cur)+1+ansi.StringWidth(t)+2 > w {
			out = append(out, cur+" \\")
			cur = "  " + t
			continue
		}
		cur += " " + t
	}
	return append(out, cur)
}

// previewLines renders the preview pane at inner width iw: the equivalent
// command, the layout and the dependencies per app while editing; the
// spinner while scaffolding; the result and next commands afterwards.
func (m WizardModel) previewLines(iw int) []string {
	switch m.phase {
	case phaseCreating:
		return []string{"", " " + m.spin.View() + " wrapping " + m.dest() + "…"}
	case phaseDone:
		return m.doneLines(iw)
	}

	lines := []string{" " + headingStyle.Render("command")}
	for _, l := range commandLines(CreateArgs(m.Cfg), iw-1) {
		lines = append(lines, " "+cmdStyle.Render(l))
	}

	lines = append(lines, "", " "+headingStyle.Render("layout")+"  "+string(m.Cfg.Layout))
	if m.Cfg.Layout.IsMonorepo() {
		tree := [][2]string{{"apps/web", string(m.Cfg.Base)}}
		if m.Cfg.Backend != "none" {
			tree = append(tree, [2]string{"apps/api", string(m.Cfg.Backend)})
		}
		tree = append(tree, [2]string{"packages/domain", "shared types"})
		for _, t := range tree {
			lines = append(lines, fmt.Sprintf("   %-17s", t[0])+FooterDescStyle.Render(t[1]))
		}
	}

	lines = append(lines, "")
	switch {
	case m.deps.cfg != m.Cfg:
		lines = append(lines, " "+headingStyle.Render("dependencies"), "   "+FooterDescStyle.Render("resolving…"))
	case m.deps.err != nil:
		lines = append(lines, " "+ErrorStyle.Render("✘ "+m.deps.err.Error()))
	default:
		for i, a := range m.deps.apps {
			if i > 0 {
				lines = append(lines, "")
			}
			title := "dependencies"
			if m.Cfg.Layout.IsMonorepo() {
				title = a.name + " dependencies"
			}
			lines = append(lines, " "+headingStyle.Render(title)+FooterDescStyle.Render(fmt.Sprintf("  (%d)", len(a.pkgs))))
			lines = append(lines, depLines(a.pkgs, iw-3, "   ")...)
		}
	}
	return lines
}

// depLines renders one line per package, prefixed with indent and at most w
// columns wide: the name padded to the group's longest, then the version in
// an aligned column. Names too long to leave room for the longest version
// are truncated with "…".
func depLines(pkgs [][2]string, w int, indent string) []string {
	nameW, verW := 0, 0
	for _, p := range pkgs {
		nameW = max(nameW, ansi.StringWidth(p[0]))
		verW = max(verW, ansi.StringWidth(p[1]))
	}
	nameW = max(1, min(nameW, w-2-verW))
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		name := ansi.Truncate(p[0], nameW, "…")
		name += strings.Repeat(" ", nameW-ansi.StringWidth(name))
		out[i] = indent + ansi.Truncate(name+"  "+FooterDescStyle.Render(p[1]), w, "…")
	}
	return out
}

// doneLines renders the scaffold result: the error, or the mascot beside
// "Wrapped!" and the commands to run next, then the post-scaffold steps that
// run on exit.
func (m WizardModel) doneLines(iw int) []string {
	if m.Err != nil {
		lines := []string{"", " " + ErrorStyle.Render("✘ scaffold failed")}
		for _, l := range strings.Split(ansi.Wrap(m.Err.Error(), iw-2, ""), "\n") {
			lines = append(lines, " "+l)
		}
		return append(lines, "", " "+keyHint("enter", "exit"))
	}
	lines := []string{""}
	for _, l := range besideMascot(successLines(m.Cfg)) {
		lines = append(lines, " "+l)
	}
	var post []string
	if m.Cfg.Install {
		post = append(post, "install")
	}
	if m.Cfg.GitInit && m.Cfg.DestDir != "." {
		post = append(post, "git init")
	}
	if len(post) > 0 {
		lines = append(lines, "", " "+FooterDescStyle.Render("on exit: "+strings.Join(post, " · ")))
	}
	return append(lines, "", " "+keyHint("enter", "exit"))
}

// keyHint renders a key and what it does, as in the status bar.
func keyHint(key, desc string) string {
	return FooterKeyStyle.Render(key) + FooterDescStyle.Render(" "+desc)
}

// statusBar renders the bottom row: the CREATE badge, the keys that apply to
// the focused pane and state, and "N/M set" at the right edge.
func (m WizardModel) statusBar() string {
	var keys [][2]string
	switch {
	case m.help:
		keys = [][2]string{{"any key", "close"}}
	case m.phase == phaseCreating:
		keys = [][2]string{{"ctrl+c", "abort"}}
	case m.phase == phaseDone:
		keys = [][2]string{{"enter", "exit"}}
	case m.focus == paneSteps:
		keys = [][2]string{{"j/k", "move"}, {"l/enter", "open"}, {jumpKeys(), "jump"}, {"0/r", "review"}, {"?", "help"}, {"q", "quit"}}
	case m.focus == panePreview:
		keys = [][2]string{{"j/k", "scroll"}, {"h", "back"}, {"?", "help"}}
	default:
		switch steps[m.step].kind {
		case kindName:
			keys = [][2]string{{"type", "name"}, {"enter", "next"}, {"esc", "back"}, {"tab", "pane"}}
		case kindAdvanced:
			keys = [][2]string{{"j/k", "row"}, {"h/l ←/→", "change"}, {"enter", "next"}, {"esc", "back"}, {"?", "help"}}
		case kindReview:
			keys = [][2]string{{"enter", "create"}, {"esc", "back"}, {"?", "help"}, {"q", "quit"}}
		default:
			keys = [][2]string{{"j/k", "move"}, {"space", "pick"}, {"enter", "next"}, {"h", "back"}, {"r", "review"}, {"?", "help"}, {"q", "quit"}}
		}
	}
	hints := make([]string, len(keys))
	for i, k := range keys {
		hints[i] = keyHint(k[0], k[1])
	}
	n, total := m.setCount()
	left := StatusModeStyle.Render(" CREATE ") + " " + strings.Join(hints, FooterSepStyle.Render(" · "))
	return spread(left, FooterDescStyle.Render(fmt.Sprintf("%d/%d set", n, total)), m.width)
}

// jumpKeys names the digit keys that jump to a step, e.g. "1–20"; review
// is 0 and listed separately.
func jumpKeys() string {
	return fmt.Sprintf("1–%d", len(steps)-1)
}

// helpLines lists every key of the wizard for the help overlay, beside the
// mascot.
func helpLines() []string {
	keys := [][2]string{
		{"j/k  ↓/↑", "move within the pane"},
		{"h/l  ←/→  tab", "switch pane"},
		{"h/l ←/→", "change a value (advanced step; esc/tab leave)"},
		{"space", "pick the option under the cursor"},
		{"enter", "steps: open · options: pick and next · review: create"},
		{jumpKeys(), "jump to a step (steps and options panes)"},
		{"0  r", "jump to review"},
		{"esc", "back one level"},
		{"?", "this help"},
		{"q", "quit (steps and options panes)"},
		{"ctrl+c", "quit"},
	}
	var lines []string
	for _, k := range keys {
		lines = append(lines, FooterKeyStyle.Render(fmt.Sprintf("%-16s", k[0]))+FooterDescStyle.Render(k[1]))
	}
	lines = append(lines, "", FooterDescStyle.Render("any key closes"))
	out := []string{""}
	for _, l := range besideMascot(lines) {
		out = append(out, " "+l)
	}
	return out
}
