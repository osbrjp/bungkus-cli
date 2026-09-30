package tui

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/pkg"
)

// newTestModel returns a wizard sized w×h with a fixed working directory,
// home and quote, so renders are deterministic.
func newTestModel(t *testing.T, w, h int) WizardModel {
	t.Helper()
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatal(err)
	}
	m := NewWizardModel(nil)
	m.wd, m.home, m.quote = "/home/u/work/app", "/home/u", 0
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(WizardModel)
}

// press sends keys to m: named keys (enter, esc, tab, space, up, down, left,
// right) or single printable characters.
func press(m WizardModel, keys ...string) WizardModel {
	named := map[string]rune{
		"enter": tea.KeyEnter, "esc": tea.KeyEscape, "tab": tea.KeyTab, "space": tea.KeySpace,
		"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
	}
	for _, k := range keys {
		msg := tea.KeyPressMsg{Code: []rune(k)[0], Text: k}
		if c, ok := named[k]; ok {
			msg = tea.KeyPressMsg{Code: c}
		}
		next, _ := m.Update(msg)
		m = next.(WizardModel)
	}
	return m
}

// screen returns the rendered frame without ANSI styling.
func screen(m WizardModel) string { return ansi.Strip(m.View().Content) }

// stepIndex returns the index of the step called name.
func stepIndex(t *testing.T, name string) int {
	t.Helper()
	i := slices.IndexFunc(steps, func(s stepDef) bool { return s.name == name })
	if i < 0 {
		t.Fatalf("no step %q", name)
	}
	return i
}

func TestStepList(t *testing.T) {
	want := []struct{ name, group, flag string }{
		{"name", "frontend", ""},
		{"base", "frontend", "base"},
		{"styling", "frontend", "css"},
		{"format", "frontend", "fmt"},
		{"lint", "frontend", "linter"},
		{"test", "frontend", "test"},
		{"audit", "frontend", "audit"},
		{"validation", "frontend", "validation"},
		{"form", "frontend", "form"},
		{"query", "frontend", "query"},
		{"state", "frontend", "state"},
		{"cms", "frontend", "cms"},
		{"deploy", "frontend", "deploy"},
		{"ci/cd", "frontend", "cicd"},
		{"desktop", "frontend", "desktop"},
		{"framework", "backend", "backend"},
		{"orm", "backend", "orm"},
		{"database", "backend", "db"},
		{"package mgr", "setup", "pm"},
		{"advanced", "setup", ""},
		{"review", "", ""},
	}
	if len(steps) != len(want) {
		t.Fatalf("got %d steps, want %d", len(steps), len(want))
	}
	m := newTestModel(t, 110, 40)
	for i, w := range want {
		s := steps[i]
		if s.name != w.name || s.group != w.group || s.flag != w.flag {
			t.Errorf("step %d = {%s %s %s}, want %+v", i, s.name, s.group, s.flag, w)
		}
		if s.kind == kindChoice && len(choices(s, m.Cfg)) == 0 {
			t.Errorf("step %s has no options in the registry", s.name)
		}
	}
	if n, total := m.setCount(); n != 5 || total != 18 {
		t.Errorf("defaults count %d/%d set, want 5/18", n, total)
	}
}

func TestOptionRendering(t *testing.T) {
	cases := []struct {
		name    string
		base    string
		step    string
		want    []string
		notWant []string
	}{
		{"react libs on a plain base", "vite", "form",
			[]string{"React Hook Form", "needs React", "VeeValidate", "needs Vue", "TanStack Form", "needs React or Vue"}, nil},
		{"react libs on a react base", "astro-react", "form",
			[]string{"React Hook Form", "+ react-hook-form", "VeeValidate", "needs Vue"}, []string{"needs React"}},
		{"nuxt counts as vue", "nuxt", "state",
			[]string{"Pinia", "+ pinia", "Jotai", "needs React"}, []string{"needs Vue"}},
		{"excluded group", "astro", "format",
			[]string{"OxFmt", "not on Astro", "Prettier", "+ prettier"}, nil},
		{"packages, cut and recommended", "astro", "styling",
			[]string{"(*) TailwindCSS", "( ) Vanilla", "+ @tailwindcss/vite", "…", "* recommended"}, []string{"TailwindCSS *"}},
		{"ci/cd needs a deploy target", "astro", "ci/cd",
			[]string{"GitHub Actions", "needs a deploy target"}, nil},
		{"database needs an orm", "astro", "database",
			[]string{"PostgreSQL", "needs an ORM"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(t, 110, 40)
			m.Cfg.Base = pkg.BaseFramework(tc.base)
			normalize(&m.Cfg)
			m.setStep(stepIndex(t, tc.step))
			m.focus = paneOptions
			ow, _ := m.paneBox(paneOptions)
			lines, _, _ := m.optionsLines(ow - 2)
			out := ansi.Strip(strings.Join(lines, "\n"))
			for _, w := range tc.want {
				if !strings.Contains(out, w) {
					t.Errorf("missing %q in\n%s", w, out)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(out, w) {
					t.Errorf("unexpected %q in\n%s", w, out)
				}
			}
		})
	}
}

func TestIncompatibleOptionCannotBePicked(t *testing.T) {
	m := newTestModel(t, 110, 40)
	m.Cfg.Base = "vite"
	normalize(&m.Cfg)
	form := stepIndex(t, "form")
	m.setStep(form)
	m.focus = paneOptions
	m = press(m, "j") // React Hook Form, which needs React
	if m = press(m, "space"); m.Cfg.Form != "none" || m.step != form {
		t.Errorf("space picked an incompatible option: form=%s step=%d", m.Cfg.Form, m.step)
	}
	if m = press(m, "enter"); m.Cfg.Form != "none" || m.step != form+1 {
		t.Errorf("enter should keep form=none and move on: form=%s step=%d", m.Cfg.Form, m.step)
	}
	if want := "React Hook Form needs React · kept None"; m.note != want {
		t.Errorf("note = %q, want %q", m.note, want)
	}
	if !strings.Contains(screen(m), "needs React · kept") {
		t.Error("the note is not shown on the next step")
	}
	if m.setStep(m.step + 1); m.note != "" {
		t.Error("the note should clear when the step changes")
	}
}

func TestBaseChangeResetsIncompatiblePicks(t *testing.T) {
	m := newTestModel(t, 110, 40)
	m.Cfg.Base, m.Cfg.Form, m.Cfg.Deployment, m.Cfg.CICD = "astro-react", "react-hook-form", "cloudflare-pages", "github-actions"
	normalize(&m.Cfg)
	m.Cfg.Base, m.Cfg.Deployment = "astro-vue", "none"
	normalize(&m.Cfg)
	if m.Cfg.Form != "none" || m.Cfg.CICD != "none" {
		t.Errorf("form=%s cicd=%s, want both none", m.Cfg.Form, m.Cfg.CICD)
	}
}

func TestCommandPreview(t *testing.T) {
	cases := []struct {
		name string
		edit func(*pkg.ProjectConfig)
		want string
	}{
		{"defaults", func(*pkg.ProjectConfig) {},
			"create my-app --base astro --css vanilla --fmt biome --linter biome --pm pnpm"},
		{"current directory", func(c *pkg.ProjectConfig) { c.ProjectName, c.DestDir = "app", "." },
			"create . --base astro"},
		{"full stack", func(c *pkg.ProjectConfig) {
			c.Base, c.Validation, c.Form, c.Backend, c.ORM, c.Database = "astro-react", "zod", "react-hook-form", "hono", "drizzle", "postgres"
		}, "--validation zod --form react-hook-form --backend hono --orm drizzle --db postgres --pm pnpm"},
		{"advanced", func(c *pkg.ProjectConfig) {
			c.Channel, c.Install, c.GitInit, c.NodeEngine = "latest", true, false, ">=20.11.0"
		}, "--pm pnpm --channel latest --install=true --git=false --node-engine >=20.11.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
				t.Fatal(err)
			}
			cfg := pkg.NewProjectConfig()
			tc.edit(&cfg)
			normalize(&cfg)
			args := CreateArgs(cfg)
			if got := strings.Join(args, " "); !strings.Contains(got, tc.want) {
				t.Errorf("args %q do not contain %q", got, tc.want)
			}
			for _, w := range []int{30, 47, 80} {
				lines := commandLines(args, w)
				for i, l := range lines {
					if ansi.StringWidth(l) > w {
						t.Errorf("width %d: line %q is too wide", w, l)
					}
					if cont := strings.HasSuffix(l, " \\"); cont != (i < len(lines)-1) {
						t.Errorf("width %d: line %d %q continuation = %v", w, i, l, cont)
					}
				}
				joined := strings.ReplaceAll(strings.Join(lines, "\n"), " \\\n  ", " ")
				if !strings.HasPrefix(joined, "bungkus-cli create ") {
					t.Errorf("width %d: command starts %q", w, joined)
				}
				if strings.Contains(joined, ">=") && !strings.Contains(joined, "'>=20.11.0'") {
					t.Errorf("width %d: node engine not quoted: %q", w, joined)
				}
			}
		})
	}
}

func TestNameValidationDisplay(t *testing.T) {
	cases := []struct {
		input string
		ok    bool
		want  string
	}{
		{"", true, "empty uses my-app"},
		{".", true, "✔ scaffolds into this directory"},
		{"my-app", true, "✔ valid npm package name"},
		{"My App", false, "✘ invalid project name"},
		{"../escape", false, "✘ invalid project name"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			m := newTestModel(t, 110, 40)
			m = press(m, "enter") // focus the name field
			for _, r := range tc.input {
				m = press(m, string(r))
			}
			if got := m.name.Value(); got != tc.input {
				t.Fatalf("typed %q, field holds %q", tc.input, got)
			}
			if ok, _ := nameStatus(tc.input); ok != tc.ok {
				t.Errorf("nameStatus(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}
			if out := screen(m); !strings.Contains(out, tc.want) {
				t.Errorf("screen lacks %q:\n%s", tc.want, out)
			}
		})
	}

	m := newTestModel(t, 110, 40)
	m = press(m, "enter", "B", "a", "d", "esc", "r", "enter")
	if m.phase != phaseEdit || !strings.Contains(screen(m), "✘ invalid project name") {
		t.Error("review must refuse to create with an invalid name")
	}
}

func TestFrameFitsTerminal(t *testing.T) {
	scenarios := []struct {
		name   string
		setup  func(WizardModel) WizardModel
		cursor string
	}{
		{"steps cursor on review", func(m WizardModel) WizardModel {
			m.setStep(len(steps) - 1)
			return m
		}, ">21  review"},
		{"options cursor on the last base", func(m WizardModel) WizardModel {
			m.setStep(stepIndex(t, "base"))
			m.focus = paneOptions
			m.opt = len(choices(steps[m.step], m.Cfg)) - 1
			return m
		}, "Vite + Vue"},
		{"advanced last row", func(m WizardModel) WizardModel {
			m.setStep(stepIndex(t, "advanced"))
			m.focus = paneOptions
			m.advanced.row = len(m.advanced.items) - 1
			return m
		}, "> node"},
	}
	sizes := [][2]int{{120, 40}, {110, 40}, {100, 30}, {80, 24}, {60, 20}}
	for _, sz := range sizes {
		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				m := sc.setup(newTestModel(t, sz[0], sz[1]))
				lines := strings.Split(screen(m), "\n")
				if len(lines) != sz[1] {
					t.Fatalf("%dx%d: %d rows", sz[0], sz[1], len(lines))
				}
				for i, l := range lines {
					if w := ansi.StringWidth(l); w > sz[0] {
						t.Errorf("%dx%d: row %d is %d wide", sz[0], sz[1], i, w)
					}
				}
				if last := lines[len(lines)-1]; !strings.Contains(last, "CREATE") || !strings.Contains(last, "set") {
					t.Errorf("%dx%d: status bar missing: %q", sz[0], sz[1], last)
				}
				if !strings.Contains(strings.Join(lines, "\n"), sc.cursor) {
					t.Errorf("%dx%d: cursor %q not visible:\n%s", sz[0], sz[1], sc.cursor, strings.Join(lines, "\n"))
				}
			})
		}
	}

	m := newTestModel(t, 110, 30)
	m.setStep(len(steps) - 1)
	if !strings.Contains(screen(m), "steps ↕") {
		t.Error("a clipped steps pane must mark its title with ↕")
	}
}

func TestAltScreenOnEveryScreen(t *testing.T) {
	cases := map[string]func(*testing.T) WizardModel{
		"unsized": func(t *testing.T) WizardModel {
			m := newTestModel(t, 110, 40)
			m.width, m.height = 0, 0
			return m
		},
		"too small": func(t *testing.T) WizardModel { return newTestModel(t, 59, 19) },
		"wide":      func(t *testing.T) WizardModel { return newTestModel(t, 120, 40) },
		"narrow":    func(t *testing.T) WizardModel { return newTestModel(t, 80, 24) },
		"help":      func(t *testing.T) WizardModel { return press(newTestModel(t, 110, 40), "?") },
		"review":    func(t *testing.T) WizardModel { return press(newTestModel(t, 110, 40), "r") },
		"creating": func(t *testing.T) WizardModel {
			m := newTestModel(t, 110, 40)
			m.phase = phaseCreating
			return m
		},
		"done": func(t *testing.T) WizardModel {
			next, _ := newTestModel(t, 110, 40).Update(scaffoldedMsg{})
			return next.(WizardModel)
		},
		"failed": func(t *testing.T) WizardModel {
			next, _ := newTestModel(t, 110, 40).Update(scaffoldedMsg{err: errors.New("disk full")})
			return next.(WizardModel)
		},
	}
	for name, mk := range cases {
		t.Run(name, func(t *testing.T) {
			if !mk(t).View().AltScreen {
				t.Error("AltScreen is off")
			}
		})
	}
	if out := screen(newTestModel(t, 59, 19)); !strings.Contains(out, "terminal too small") || strings.Contains(out, "\n") {
		t.Errorf("too-small notice must be one line, got %q", out)
	}
}

func TestSinglePaneUnder100Columns(t *testing.T) {
	cases := []struct {
		name   string
		width  int
		keys   []string
		titles []string
		crumb  string
	}{
		{"wide shows all panes", 100, nil, []string{"steps", "name", "preview"}, ""},
		{"narrow steps", 80, nil, []string{"steps"}, "create"},
		{"narrow options", 80, []string{"j", "j", "enter"}, []string{"styling"}, "create › styling"},
		{"narrow preview", 80, []string{"tab", "tab"}, []string{"preview"}, "create › preview"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := press(newTestModel(t, tc.width, 24), tc.keys...)
			lines := strings.Split(screen(m), "\n")
			var titles []string
			for _, f := range strings.Fields(lines[headerRows]) {
				if f != "" && !strings.ContainsAny(f, "─━┌┏┐┓↕") {
					titles = append(titles, f)
				}
			}
			if !slices.Equal(titles, tc.titles) {
				t.Errorf("pane titles %q, want %q", titles, tc.titles)
			}
			if tc.crumb != "" && !strings.Contains(lines[0], tc.crumb) {
				t.Errorf("header %q lacks breadcrumb %q", lines[0], tc.crumb)
			}
			if tc.crumb == "" && strings.Contains(lines[0], "create") {
				t.Errorf("wide header shows a breadcrumb: %q", lines[0])
			}
		})
	}
}

func TestHeader(t *testing.T) {
	defer func(v, u string) { Version, UpdateAvailable = v, u }(Version, UpdateAvailable)
	Version = "1.2.3"

	cases := []struct {
		name   string
		update string
		quote  int
		width  int
		last   string
	}{
		{"quote", "", 0, 110, `"` + quotes[0] + `"`},
		{"another quote", "", 7, 110, `"` + quotes[7] + `"`},
		{"update replaces the quote", "v1.9.0", 0, 110, "update available: v1.9.0 · bungkus-cli update"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			UpdateAvailable = tc.update
			m := newTestModel(t, tc.width, 40)
			m.quote = tc.quote
			rows := m.header()
			if len(rows) != headerRows {
				t.Fatalf("header has %d rows, want %d", len(rows), headerRows)
			}
			text := make([]string, len(rows))
			for i, r := range rows {
				text[i] = strings.TrimSpace(string([]rune(ansi.Strip(r))[len(mascotPixels[0])+4:]))
			}
			want := []string{"bungkus-cli · v1.2.3", "by spencer · osbr", "~/work/app", tc.last}
			if !slices.Equal(text, want) {
				t.Errorf("header text %q, want %q", text, want)
			}
			if lines := strings.Split(screen(m), "\n"); len(lines) != 40 {
				t.Errorf("frame is %d rows, want 40", len(lines))
			}
		})
	}

	UpdateAvailable = ""
	m := newTestModel(t, 60, 20)
	m.quote = slices.IndexFunc(quotes, func(q string) bool { return len(q) > 60 })
	last := ansi.Strip(m.header()[3])
	if ansi.StringWidth(last) != 60 || !strings.HasSuffix(last, "…") {
		t.Errorf("long quote not cut to 60 columns with …: %q", last)
	}
}

func TestVersionAndPathLabels(t *testing.T) {
	for in, want := range map[string]string{"1.2.3": "v1.2.3", "v1.2.3": "v1.2.3", "dev": "dev", "": "dev"} {
		if got := versionLabel(in); got != want {
			t.Errorf("versionLabel(%q) = %q, want %q", in, got, want)
		}
	}
	cases := []struct{ path, home, want string }{
		{"/home/u/work", "/home/u", "~/work"},
		{"/home/u", "/home/u", "~"},
		{"/home/user2/x", "/home/u", "/home/user2/x"},
		{"/srv/x", "", "/srv/x"},
	}
	for _, tc := range cases {
		if got := tildePath(tc.path, tc.home); got != tc.want {
			t.Errorf("tildePath(%q, %q) = %q, want %q", tc.path, tc.home, got, tc.want)
		}
	}
}

func TestMascot(t *testing.T) {
	rows := mascot()
	want := []string{
		"   ▄██▄",
		"  ▄▀██▀▄",
		" ▀▀▀▀▀▀▀▀",
		"   ┛  ┛",
	}
	if len(rows) != len(want) {
		t.Fatalf("mascot has %d rows, want %d", len(rows), len(want))
	}
	for i, r := range rows {
		if got := strings.TrimRight(ansi.Strip(r), " "); got != want[i] {
			t.Errorf("row %d = %q, want %q", i, got, want[i])
		}
		if w := lipgloss.Width(r); w != len(mascotPixels[0]) {
			t.Errorf("row %d is %d wide, want %d", i, w, len(mascotPixels[0]))
		}
	}
	legs := lipgloss.NewStyle().Foreground(tokMascotLegs.color()).Render("┛  ┛")
	if rows[3] != "   "+legs+"   " {
		t.Errorf("legs row %q is not U+251B in the leg colour", rows[3])
	}
}

func TestEnterWalksStepsToReview(t *testing.T) {
	m := newTestModel(t, 110, 40)
	m = press(m, "enter") // steps → options (name)
	for range len(steps) - 1 {
		m = press(m, "enter")
	}
	if steps[m.step].kind != kindReview || m.focus != paneOptions {
		t.Fatalf("enter did not reach review: step %s focus %d", steps[m.step].name, m.focus)
	}
	m = press(m, "esc")
	if m.focus != paneSteps {
		t.Error("esc must return to the steps pane")
	}
	if m = press(m, "q"); !m.Canceled {
		t.Error("q must quit from the steps pane")
	}
}

func TestDuckAnimation(t *testing.T) {
	strip := func(rows []string) []string {
		out := make([]string, len(rows))
		for i, r := range rows {
			out[i] = strings.TrimRight(ansi.Strip(r), " ")
		}
		return out
	}
	poses := []struct {
		pose int
		want []string
	}{
		{poseStand, []string{"   ▄██▄", "  ▄▀██▀▄", " ▀▀▀▀▀▀▀▀", "   ┛  ┛"}},
		{poseHalf, []string{"    ▄▄", "   ████", " ▄▀▀▀▀▀▀▄", "   ┛  ┛"}},
		{poseFull, []string{"", "   ▄██▄", "  ▄▀██▀▄", " ▀▀▀▀▀▀▀▀"}},
	}
	for _, tc := range poses {
		rows := mascotFrame(tc.pose)
		if got := strip(rows); !slices.Equal(got, tc.want) {
			t.Errorf("pose %d = %q, want %q", tc.pose, got, tc.want)
		}
		for i, r := range rows {
			if w := lipgloss.Width(r); w != len(mascotPixels[0]) {
				t.Errorf("pose %d row %d is %d wide", tc.pose, i, w)
			}
		}
	}

	defer func(a bool) { animate = a }(animate)
	animate = true
	m := newTestModel(t, 110, 40)
	if _, cmd := m.Update(duckMsg{}); cmd != nil {
		t.Error("duck ticks while not scaffolding")
	}
	m.phase = phaseCreating
	next, cmd := m.Update(duckMsg{})
	if cmd == nil || next.(WizardModel).duck != 1 {
		t.Error("duck must advance and reschedule while scaffolding")
	}
	if got := strip(next.(WizardModel).header())[0]; !strings.HasPrefix(got, "    ▄▄ ") {
		t.Errorf("header while scaffolding does not duck: %q", got)
	}
	next, _ = next.Update(scaffoldedMsg{})
	done := next.(WizardModel)
	if _, cmd := done.Update(duckMsg{}); cmd != nil {
		t.Error("duck keeps ticking after scaffolding")
	}
	if got := strip(done.header())[0]; !strings.HasPrefix(got, "   ▄██▄ ") {
		t.Errorf("header after scaffolding is not the stand pose: %q", got)
	}
}

func TestMascotOnInstructionScreens(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatal(err)
	}
	cfg := pkg.NewProjectConfig()
	out := ansi.Strip(successText(cfg))
	for _, want := range []string{"▄▀██▀▄", "┛  ┛", "Wrapped!", "my-app is ready.", "cd my-app", cfg.PM.InstallCmd(), cfg.PM.RunCmd()} {
		if !strings.Contains(out, want) {
			t.Errorf("success output lacks %q:\n%s", want, out)
		}
	}

	next, _ := newTestModel(t, 110, 40).Update(scaffoldedMsg{})
	lines := strings.Split(screen(next.(WizardModel)), "\n")
	i := slices.IndexFunc(lines, func(l string) bool { return strings.Contains(l, "Wrapped!") })
	if i < headerRows || !strings.Contains(lines[i], "▄██▄") || !strings.Contains(lines[i+1], "my-app is ready.") ||
		!strings.Contains(lines[i+3], "┛  ┛") || !strings.Contains(lines[i+3], "cd my-app") {
		t.Errorf("preview pane lacks the mascot beside the success block:\n%s", strings.Join(lines, "\n"))
	}

	help := strings.Split(screen(press(newTestModel(t, 110, 40), "?")), "\n")
	if body := strings.Join(help[headerRows:], "\n"); !strings.Contains(body, "▄▀██▀▄") || !strings.Contains(body, "switch pane") {
		t.Errorf("help overlay lacks the mascot or the keys:\n%s", body)
	}
}

// The wizard's untouched config must match what the tabbed wizard of
// v1.8.0 produced (NewWizardModel + collectConfig on origin/main): the first
// option of every group, which is the recommended one where the registry
// marks one.
func TestWizardDefaultsMatchPreviousWizard(t *testing.T) {
	m := newTestModel(t, 110, 40)
	want := pkg.NewProjectConfig()
	want.Base, want.CSS, want.Fmt, want.Linter = "astro", "tailwindcss", "biome", "biome"
	want.Test, want.Audit = "none", "none"
	want.Validation, want.Form, want.Query, want.State, want.CMS = "none", "none", "none", "none", "none"
	want.Deployment, want.CICD, want.Desktop = "none", "none", "none"
	want.Backend, want.ORM, want.Database = "none", "none", "none"
	want.PM, want.Layout = "pnpm", pkg.LayoutFlat
	want.Channel, want.Pin, want.Install, want.GitInit, want.NodeEngine = pkg.ChannelPinned, pkg.PinDefault, false, true, pkg.DefaultNodeEngine
	got := m.Cfg
	got.Date, want.Date = "", ""
	if got != want {
		t.Errorf("wizard defaults\n got %+v\nwant %+v", got, want)
	}
	if cmd := strings.Join(CreateArgs(m.Cfg), " "); cmd != "create my-app --base astro --css tailwindcss --fmt biome --linter biome --pm pnpm" {
		t.Errorf("default command = %q", cmd)
	}
	if n, total := m.setCount(); n != 5 || total != 18 {
		t.Errorf("defaults count %d/%d set, want 5/18", n, total)
	}
}

func TestNameStepShowsHowToContinue(t *testing.T) {
	cases := []struct {
		name           string
		valid, focused bool
		want           string
	}{
		{"typing a valid name", true, true, "↵ enter  next: base"},
		{"not yet opened", true, false, "enter to edit"},
		{"invalid name", false, true, "fix the name to continue"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ansi.Strip(nameCallToAction(tc.valid, tc.focused, "base")); !strings.Contains(got, tc.want) {
				t.Errorf("nameCallToAction(%v, %v) = %q, want it to contain %q", tc.valid, tc.focused, got, tc.want)
			}
		})
	}
}

func TestAdvancedStepHLChangeValue(t *testing.T) {
	m := newTestModel(t, 110, 40)
	m.setStep(slices.IndexFunc(steps, func(s stepDef) bool { return s.kind == kindAdvanced }))
	next, _ := m.focusPane(paneOptions)
	m = next.(WizardModel)
	before := m.advanced.items[m.advanced.row].cursor
	for _, k := range []string{"l", "h", "l"} {
		m = press(m, k)
		if m.focus != paneOptions {
			t.Fatalf("%q left the options pane on the advanced step", k)
		}
	}
	if m.advanced.items[m.advanced.row].cursor == before {
		t.Error("h/l did not change the advanced value")
	}
	if m = press(m, "esc"); m.focus != paneSteps {
		t.Error("esc should return to the steps pane")
	}
}

// The monorepo success box names the chosen package manager and prints its
// own workspace commands.
func TestSuccessTextMonorepoPerPM(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatal(err)
	}
	cases := map[pkg.PackageManager]string{
		"pnpm": "pnpm --filter api run db:migrate",
		"bun":  "bun run --filter api db:migrate",
		"npm":  "npm run db:migrate -w api",
		"yarn": "yarn workspace api run db:migrate",
	}
	for pm, migrate := range cases {
		cfg := pkg.NewProjectConfig()
		cfg.PM, cfg.Backend, cfg.ORM, cfg.Database = pm, "hono", "drizzle", "sqlite"
		cfg.ApplyDefaultLayout()
		out := ansi.Strip(successText(cfg))
		for _, want := range []string{"Workspace (" + string(pm) + "):", migrate} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: success output lacks %q:\n%s", pm, want, out)
			}
		}
	}
}

// expire ends the pending digit window of m as its jumpWindow tick would.
func expire(m WizardModel) WizardModel {
	next, _ := m.Update(jumpExpiredMsg{seq: m.jumpSeq})
	return next.(WizardModel)
}

// Digits jump to a step by its 1-based number and focus the steps pane.
func TestJumpByNumber(t *testing.T) {
	cases := []struct {
		name  string
		start pane
		keys  []string // "wait" lets the digit window expire
		want  int      // 1-based step number
	}{
		{"single digit", paneSteps, []string{"3"}, 3},
		{"from the options pane", paneOptions, []string{"5"}, 5},
		{"two digits within the window", paneSteps, []string{"1", "6"}, 16},
		{"two digits after the window", paneSteps, []string{"1", "wait", "6"}, 6},
		{"out of range keeps the last jump", paneSteps, []string{"4", "9"}, 4},
		{"zero alone does nothing", paneSteps, []string{"0"}, 1},
		{"another key ends the number", paneSteps, []string{"1", "k", "2"}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(t, 120, 40)
			if tc.start == paneOptions {
				m.setStep(stepIndex(t, "styling"))
			}
			m.focus = tc.start
			for _, k := range tc.keys {
				if k == "wait" {
					m = expire(m)
					continue
				}
				m = press(m, k)
			}
			if m.step+1 != tc.want || m.focus != paneSteps {
				t.Errorf("step %d focus %d, want step %d in the steps pane", m.step+1, m.focus, tc.want)
			}
		})
	}
}

// While the name field has focus digits are part of the name.
func TestJumpIgnoredWhileEditingName(t *testing.T) {
	m := press(newTestModel(t, 120, 40), "enter")
	if !m.editingName() {
		t.Fatal("enter on the name step should focus the name field")
	}
	m = press(m, "a", "1", "2")
	if m.step != 0 || m.focus != paneOptions || m.name.Value() != "a12" {
		t.Errorf("step %d focus %d name %q, want name a12 on step 1", m.step+1, m.focus, m.name.Value())
	}
}

// A tick from an earlier digit must not cut the window of a later one short.
func TestJumpStaleTickIgnored(t *testing.T) {
	m := press(newTestModel(t, 120, 40), "1")
	stale := m.jumpSeq
	m = press(expire(m), "1")
	next, _ := m.Update(jumpExpiredMsg{seq: stale})
	m = press(next.(WizardModel), "7")
	if m.step+1 != 17 {
		t.Errorf("step %d, want 17: a stale tick reset the pending digit", m.step+1)
	}
}

// The steps pane is fixed, the preview shrinks from 40 toward 30 and the
// options pane takes the rest; the three always fill the width.
func TestPaneWidths(t *testing.T) {
	cases := []struct{ width, steps, options, preview int }{
		{140, 36, 64, 40},
		{120, 36, 44, 40},
		{110, 36, 38, 36},
		{100, 36, 34, 30},
	}
	for _, tc := range cases {
		m := newTestModel(t, tc.width, 40)
		s, _ := m.paneBox(paneSteps)
		o, _ := m.paneBox(paneOptions)
		p, _ := m.paneBox(panePreview)
		if s != tc.steps || o != tc.options || p != tc.preview {
			t.Errorf("width %d: steps/options/preview = %d/%d/%d, want %d/%d/%d", tc.width, s, o, p, tc.steps, tc.options, tc.preview)
		}
	}
}

// Every step row carries its right-aligned number; group headings do not.
func TestStepsNumbered(t *testing.T) {
	m := newTestModel(t, 120, 40)
	lines, _, _ := m.stepsLines(stepsWidth - 2)
	out := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"> 1  name", "  9  form", " 16  framework", "\n frontend\n"} {
		if !strings.Contains("\n"+out+"\n", want) {
			t.Errorf("steps pane lacks %q:\n%s", want, out)
		}
	}
}
