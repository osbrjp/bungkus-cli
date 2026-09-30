package tui

import (
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/osbrjp/bungkus-cli/pkg"
)

// stepKind says how a step is edited in the options pane.
type stepKind int

const (
	kindChoice   stepKind = iota // radio list of registry entries
	kindName                     // free-text project name
	kindAdvanced                 // low-frequency settings changed with ←/→
	kindReview                   // summary and create
)

// stepDef is one row of the steps pane. Choice steps carry the registry
// category they pick from, the `create` flag that sets the same field, and
// accessors for that field of pkg.ProjectConfig.
type stepDef struct {
	name    string // row label
	group   string // heading the row sits under; "" for review
	flag    string // `create` flag for the field; "" for non-choice steps
	kind    stepKind
	entries func(*pkg.Registry) []pkg.OptionEntry
	get     func(pkg.ProjectConfig) string
	set     func(*pkg.ProjectConfig, string)
	// fits is the integration check pkg defines for the field, if any.
	fits func(value, base string) bool
	// rule returns why value cannot be picked given the rest of the config,
	// for constraints `create` enforces between fields; "" when it can.
	rule func(cfg pkg.ProjectConfig, value string) string
}

// steps is every row of the steps pane, in wizard order. The order matters
// twice: enter walks it, and normalize fixes fields in it, so a field must
// come after the fields its rule reads (ci/cd after deploy, database after
// orm, everything after base).
var steps = []stepDef{
	{name: "name", group: "frontend", kind: kindName},
	{name: "base", group: "frontend", flag: "base", entries: baseEntries,
		get: func(c pkg.ProjectConfig) string { return string(c.Base) },
		set: func(c *pkg.ProjectConfig, v string) { c.Base = pkg.BaseFramework(v) }},
	{name: "styling", group: "frontend", flag: "css",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.CSS },
		get:     func(c pkg.ProjectConfig) string { return string(c.CSS) },
		set:     func(c *pkg.ProjectConfig, v string) { c.CSS = pkg.CSSFramework(v) }},
	{name: "format", group: "frontend", flag: "fmt",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Formatters },
		get:     func(c pkg.ProjectConfig) string { return string(c.Fmt) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Fmt = pkg.Formatter(v) }},
	{name: "lint", group: "frontend", flag: "linter",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Linters },
		get:     func(c pkg.ProjectConfig) string { return string(c.Linter) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Linter = pkg.Linter(v) }},
	{name: "test", group: "frontend", flag: "test",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Test },
		get:     func(c pkg.ProjectConfig) string { return string(c.Test) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Test = pkg.TestingFramework(v) }},
	{name: "audit", group: "frontend", flag: "audit",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Audit },
		get:     func(c pkg.ProjectConfig) string { return string(c.Audit) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Audit = pkg.AuditTool(v) }},
	{name: "validation", group: "frontend", flag: "validation",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Validation },
		get:     func(c pkg.ProjectConfig) string { return string(c.Validation) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Validation = pkg.ValidationLib(v) }},
	{name: "form", group: "frontend", flag: "form",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Form },
		get:     func(c pkg.ProjectConfig) string { return string(c.Form) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Form = pkg.FormLib(v) },
		fits:    func(v, base string) bool { return pkg.FormLib(v).IsValidIntegration(base) }},
	{name: "query", group: "frontend", flag: "query",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Query },
		get:     func(c pkg.ProjectConfig) string { return string(c.Query) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Query = pkg.QueryLib(v) },
		fits:    func(v, base string) bool { return pkg.QueryLib(v).IsValidIntegration(base) }},
	{name: "state", group: "frontend", flag: "state",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.State },
		get:     func(c pkg.ProjectConfig) string { return string(c.State) },
		set:     func(c *pkg.ProjectConfig, v string) { c.State = pkg.StateLib(v) },
		fits:    func(v, base string) bool { return pkg.StateLib(v).IsValidIntegration(base) }},
	{name: "cms", group: "frontend", flag: "cms",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.CMS },
		get:     func(c pkg.ProjectConfig) string { return string(c.CMS) },
		set:     func(c *pkg.ProjectConfig, v string) { c.CMS = pkg.CMS(v) }},
	{name: "deploy", group: "frontend", flag: "deploy",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Deployment },
		get:     func(c pkg.ProjectConfig) string { return string(c.Deployment) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Deployment = pkg.DeployTarget(v) }},
	{name: "ci/cd", group: "frontend", flag: "cicd",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.CICD },
		get:     func(c pkg.ProjectConfig) string { return string(c.CICD) },
		set:     func(c *pkg.ProjectConfig, v string) { c.CICD = pkg.CICDProvider(v) },
		rule: func(c pkg.ProjectConfig, v string) string {
			if v != "none" && c.Deployment == "none" {
				return "needs a deploy target"
			}
			return ""
		}},
	{name: "desktop", group: "frontend", flag: "desktop",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Desktop },
		get:     func(c pkg.ProjectConfig) string { return string(c.Desktop) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Desktop = pkg.DesktopTarget(v) }},
	{name: "framework", group: "backend", flag: "backend",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Backend },
		get:     func(c pkg.ProjectConfig) string { return string(c.Backend) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Backend = pkg.BackendLib(v) }},
	{name: "orm", group: "backend", flag: "orm",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.ORM },
		get:     func(c pkg.ProjectConfig) string { return string(c.ORM) },
		set:     func(c *pkg.ProjectConfig, v string) { c.ORM = pkg.ORMLib(v) }},
	{name: "database", group: "backend", flag: "db",
		entries: func(r *pkg.Registry) []pkg.OptionEntry { return r.Database },
		get:     func(c pkg.ProjectConfig) string { return string(c.Database) },
		set:     func(c *pkg.ProjectConfig, v string) { c.Database = pkg.Database(v) },
		rule: func(c pkg.ProjectConfig, v string) string {
			switch {
			case v != "none" && c.ORM == "none":
				return "needs an ORM"
			case v == "d1" && c.ORM == "prisma":
				return "needs Drizzle ORM"
			}
			return ""
		}},
	{name: "package mgr", group: "setup", flag: "pm", entries: pmEntries,
		get: func(c pkg.ProjectConfig) string { return string(c.PM) },
		set: func(c *pkg.ProjectConfig, v string) { c.PM = pkg.PackageManager(v) }},
	{name: "advanced", group: "setup", kind: kindAdvanced},
	{name: "review", kind: kindReview},
}

// baseEntries lists the registry's bases in the OptionEntry shape the
// options pane renders.
func baseEntries(r *pkg.Registry) []pkg.OptionEntry {
	out := make([]pkg.OptionEntry, len(r.Bases))
	for i, b := range r.Bases {
		out[i] = pkg.OptionEntry{Value: b.Value, Label: b.Label, Packages: b.Packages}
	}
	return out
}

// pmEntries lists the registry's package managers in the OptionEntry shape
// the options pane renders; they add no packages.
func pmEntries(r *pkg.Registry) []pkg.OptionEntry {
	out := make([]pkg.OptionEntry, len(r.PackageManagers))
	for i, p := range r.PackageManagers {
		out[i] = pkg.OptionEntry{Value: p.Value, Label: p.Label}
	}
	return out
}

// choice is one option of a choice step as the options pane shows it.
type choice struct {
	value       string
	label       string   // registry label without the recommendation mark
	recommended bool     // the registry label ends in "*"
	adds        []string // package names the option brings, sorted
	reason      string   // why it cannot be picked; "" when it can
}

// choices returns the options of choice step s for cfg, in registry order.
// Incompatible options are kept, with the reason set, so the pane can show
// them greyed instead of hiding them. The packages listed include the
// entry's integration packages for cfg's base (Nuxt falls back to Vue, as
// the package.json builder does).
func choices(s stepDef, cfg pkg.ProjectConfig) []choice {
	reg := pkg.GetRegistry()
	base := reg.GetBase(string(cfg.Base))
	var out []choice
	for _, e := range s.entries(reg) {
		label := strings.TrimSpace(e.Label)
		c := choice{value: e.Value, recommended: strings.HasSuffix(label, "*")}
		c.label = strings.TrimSpace(strings.TrimSuffix(label, "*"))

		names := map[string]bool{}
		add := func(p pkg.Packages) {
			for n := range p.Dependencies {
				names[n] = true
			}
			for n := range p.DevDependencies {
				names[n] = true
			}
		}
		add(e.Packages)
		if base != nil {
			key := base.Integration
			if base.Group == "nuxt" {
				key = "nuxt"
				if _, ok := e.IntegrationPackages[key]; !ok {
					key = "vue"
				}
			}
			add(e.IntegrationPackages[key])
		}
		c.adds = slices.Sorted(maps.Keys(names))

		switch {
		case base != nil && e.ExcludesGroup(base.Group):
			c.reason = "not on " + titleCase(base.Group)
		case s.fits != nil && !s.fits(e.Value, string(cfg.Base)):
			reqs := make([]string, len(e.RequiresIntegration))
			for i, r := range e.RequiresIntegration {
				reqs[i] = titleCase(r)
			}
			c.reason = "needs " + strings.Join(reqs, " or ")
		case s.rule != nil:
			c.reason = s.rule(cfg, e.Value)
		}
		out = append(out, c)
	}
	return out
}

// titleCase upper-cases the first letter of an ASCII word ("react" → "React").
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// normalize makes cfg consistent after a pick: every choice field whose value
// has become incompatible (or was never offered) moves to the first option
// that can be picked, and the layout is re-derived from the backend and
// package manager as `create` derives it when --layout is not given.
func normalize(cfg *pkg.ProjectConfig) {
	for _, s := range steps {
		if s.kind != kindChoice {
			continue
		}
		cs := choices(s, *cfg)
		cur := s.get(*cfg)
		if i := slices.IndexFunc(cs, func(c choice) bool { return c.value == cur }); i >= 0 && cs[i].reason == "" {
			continue
		}
		if i := slices.IndexFunc(cs, func(c choice) bool { return c.reason == "" }); i >= 0 {
			s.set(cfg, cs[i].value)
		}
	}
	cfg.Layout = pkg.LayoutFlat
	cfg.ApplyDefaultLayout()
}

// initialPicks sets every choice field of cfg to the option the wizard
// starts on: the registry's recommended option when it can be picked,
// otherwise the first option that can. Fields are visited in step order, so
// later options are judged against the earlier picks. The layout is then
// derived as normalize does.
func initialPicks(cfg *pkg.ProjectConfig) {
	for _, s := range steps {
		if s.kind != kindChoice {
			continue
		}
		cs := choices(s, *cfg)
		i := slices.IndexFunc(cs, func(c choice) bool { return c.recommended && c.reason == "" })
		if i < 0 {
			i = slices.IndexFunc(cs, func(c choice) bool { return c.reason == "" })
		}
		if i >= 0 {
			s.set(cfg, cs[i].value)
		}
	}
	normalize(cfg)
}

// isSet reports whether a step's value counts toward the status bar's
// "N/M set": anything but empty or "none".
func isSet(v string) bool { return v != "" && v != "none" }

// CreateArgs returns the `bungkus-cli create` arguments (without the program
// name) that rebuild cfg. Choice fields are passed when not "none", with base
// and package manager always present; advanced settings only when they differ
// from pkg.NewProjectConfig. The layout is left out because `create` derives
// it the same way the wizard does. Values are unquoted; see shellQuote.
func CreateArgs(cfg pkg.ProjectConfig) []string {
	name := cfg.ProjectName
	if cfg.DestDir == "." {
		name = "."
	}
	args := []string{"create", name}
	for _, s := range steps {
		if s.flag == "" {
			continue
		}
		if v := s.get(cfg); v != "none" {
			args = append(args, "--"+s.flag, v)
		}
	}
	def := pkg.NewProjectConfig()
	if cfg.Channel != def.Channel {
		args = append(args, "--channel", string(cfg.Channel))
	}
	if cfg.Pin != def.Pin {
		args = append(args, "--pin", string(cfg.Pin))
	}
	if cfg.Install != def.Install {
		args = append(args, "--install="+strconv.FormatBool(cfg.Install))
	}
	if cfg.GitInit != def.GitInit {
		args = append(args, "--git="+strconv.FormatBool(cfg.GitInit))
	}
	if cfg.NodeEngine != def.NodeEngine {
		args = append(args, "--node-engine", cfg.NodeEngine)
	}
	return args
}

var shellSafe = regexp.MustCompile(`^[A-Za-z0-9._/@:=+,-]+$`)

// shellQuote returns s as a single POSIX shell word: unchanged when it holds
// only characters the shell treats literally, single-quoted otherwise (so
// ">=22.12.0" does not become a redirection).
func shellQuote(s string) string {
	if shellSafe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// nameStatus describes the project name typed into the name step. ok is false
// when `create` would refuse the name; msg is the line shown under the input.
// Empty falls back to "my-app", and "." scaffolds into the current directory.
func nameStatus(v string) (ok bool, msg string) {
	switch v {
	case "":
		return true, "empty uses my-app"
	case ".":
		return true, "scaffolds into this directory"
	}
	if err := pkg.ValidateProjectName(v); err != nil {
		return false, err.Error()
	}
	return true, "valid npm package name"
}
