package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrUnknownAddOption is returned when the option matches no addable category.
var ErrUnknownAddOption = errors.New("unknown add option")

// AddReport records what Add did (and refused to do). File paths are relative
// to the project directory.
type AddReport struct {
	CreatedFiles, SkippedFiles   []string
	DepsAdded, DepsSkipped       []string
	ScriptsAdded, ScriptsSkipped []string
	PkgJSONChanged               bool
	Categories                   []string
}

// addCategory wires one registry category into `add`. Options inside each
// category are 100% registry-driven; only the category list is Go-known,
// mirroring the per-category branches in Scaffold/BuildPackageJSON. The
// supported set is the standalone categories (config files + deps, no source
// edits inside the user's app).
type addCategory struct {
	name   string // registry/template dir name
	lookup func(*Registry, string) *OptionEntry
	set    func(*ProjectConfig, string)
}

var addCategories = []addCategory{
	{"fmt", (*Registry).GetFormatter, func(c *ProjectConfig, v string) { c.Fmt = Formatter(v) }},
	{"linter", (*Registry).GetLinter, func(c *ProjectConfig, v string) { c.Linter = Linter(v) }},
	{"test", (*Registry).GetTestingFramework, func(c *ProjectConfig, v string) { c.Test = TestingFramework(v) }},
	{"audit", (*Registry).GetAudit, func(c *ProjectConfig, v string) { c.Audit = AuditTool(v) }},
	{"deploy", (*Registry).GetDeployment, func(c *ProjectConfig, v string) { c.Deployment = DeployTarget(v) }},
	{"cicd", (*Registry).GetCICD, func(c *ProjectConfig, v string) { c.CICD = CICDProvider(v) }},
}

// AddableCategory is one row of the `add` help listing.
type AddableCategory struct {
	Name    string
	Options []string
}

// AddableOptions lists what `add` accepts, generated from the live registry so
// the listing can never go stale.
func AddableOptions() []AddableCategory {
	r := GetRegistry()
	cats := []struct {
		name    string
		entries []OptionEntry
	}{
		{"fmt", r.Formatters}, {"linter", r.Linters}, {"test", r.Test},
		{"audit", r.Audit}, {"deploy", r.Deployment}, {"cicd", r.CICD},
	}
	out := make([]AddableCategory, 0, len(cats))
	for _, c := range cats {
		var opts []string
		for _, e := range c.entries {
			if e.Value != "none" {
				opts = append(opts, e.Value)
			}
		}
		out = append(out, AddableCategory{c.name, opts})
	}
	return out
}

// lockfileNames maps each package manager to its lockfile names. These are npm
// ecosystem facts, not bungkus choices, so they live here, not in the registry.
var lockfileNames = map[string][]string{
	"pnpm": {"pnpm-lock.yaml"},
	"bun":  {"bun.lock", "bun.lockb"},
	"npm":  {"package-lock.json"},
	"yarn": {"yarn.lock"},
}

// DetectProject inspects dir and rawPkg (the project's package.json bytes) and
// returns a partial ProjectConfig. Base and PM are left empty when they cannot
// be determined unambiguously, for the caller to fill from flags or reject.
// Fmt/Linter are left empty so only the option being added triggers
// cross-cutting package rules.
func DetectProject(dir string, rawPkg []byte) (ProjectConfig, error) {
	cfg := NewProjectConfig()
	cfg.Base, cfg.PM, cfg.Fmt, cfg.Linter = "", "", "", ""

	var pj struct {
		Name            string            `json:"name"`
		PackageManager  string            `json:"packageManager"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(rawPkg, &pj); err != nil {
		return cfg, fmt.Errorf("could not parse package.json: %w", err)
	}
	if pj.Name != "" {
		cfg.ProjectName = pj.Name
	} else if abs, err := filepath.Abs(dir); err == nil {
		cfg.ProjectName = filepath.Base(abs)
	}

	names := map[string]bool{}
	for k := range pj.Dependencies {
		names[k] = true
	}
	for k := range pj.DevDependencies {
		names[k] = true
	}

	cfg.PM = detectPM(dir, pj.PackageManager)
	cfg.Base = BaseFramework(detectBase(names))
	if names["tailwindcss"] {
		cfg.CSS = "tailwindcss"
	} else {
		cfg.CSS = "vanilla"
	}
	return cfg, nil
}

// detectPM resolves the package manager: package.json's packageManager field
// wins; otherwise walk up from dir looking for exactly one PM's lockfile,
// stopping at the repo root (a directory containing .git). Zero or several
// matching lockfiles yield "" so the caller can require --pm.
func detectPM(dir, pmField string) PackageManager {
	if name, _, _ := strings.Cut(pmField, "@"); name != "" && GetRegistry().HasPM(name) {
		return PackageManager(name)
	}
	d, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		var found []string
		for pm, files := range lockfileNames {
			for _, f := range files {
				if _, err := os.Stat(filepath.Join(d, f)); err == nil {
					found = append(found, pm)
					break
				}
			}
		}
		if len(found) == 1 {
			return PackageManager(found[0])
		}
		if len(found) > 1 {
			return "" // ambiguous
		}
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return "" // repo root reached, no lockfile anywhere
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}

// detectBase matches the dependency-name signature of every registry base: a
// base is a candidate when all of its packages appear among the project's
// dependency names. The candidate with the most packages wins (astro-react
// beats astro); a tie is ambiguous and yields "".
func detectBase(names map[string]bool) string {
	best, bestCount, tie := "", 0, false
	for _, b := range GetRegistry().Bases {
		count := len(b.Packages.Dependencies) + len(b.Packages.DevDependencies)
		if count == 0 {
			continue
		}
		ok := true
		for k := range b.Packages.Dependencies {
			if !names[k] {
				ok = false
				break
			}
		}
		for k := range b.Packages.DevDependencies {
			if !ok || !names[k] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		switch {
		case count > bestCount:
			best, bestCount, tie = b.Value, count, false
		case count == bestCount:
			tie = true
		}
	}
	if tie {
		return ""
	}
	return best
}

// Add applies option opt to the existing project at dir. cfg is the detected/
// overridden project config (Base and PM must be valid). It renders each
// matched category's templates without ever overwriting an existing file,
// merges the option's packages into package.json additively, and returns a
// report of everything it did. The only file ever rewritten is package.json.
func Add(dir string, templates fs.FS, cfg ProjectConfig, opt string) (*AddReport, error) {
	reg := GetRegistry()
	baseEntry := reg.GetBase(string(cfg.Base))
	if baseEntry == nil {
		return nil, fmt.Errorf("unknown base framework: %s", cfg.Base)
	}
	if opt == "none" {
		return nil, ErrUnknownAddOption
	}

	rep := &AddReport{}
	var entries []*OptionEntry
	var matched []addCategory
	for _, cat := range addCategories {
		if e := cat.lookup(reg, opt); e != nil {
			entries = append(entries, e)
			matched = append(matched, cat)
			rep.Categories = append(rep.Categories, cat.name)
		}
	}
	if len(entries) == 0 {
		return nil, ErrUnknownAddOption
	}

	for _, e := range entries {
		if e.ExcludesGroup(baseEntry.Group) {
			return nil, fmt.Errorf("%s does not support %s projects", opt, baseEntry.Group)
		}
	}
	for _, cat := range matched {
		if cat.name == "cicd" && cfg.Deployment == "none" {
			return nil, fmt.Errorf("%s needs a deploy target: pass --deploy (cloudflare-pages, cloudflare-workers)", opt)
		}
		cat.set(&cfg, opt)
	}

	for _, cat := range matched {
		tmplDir := "templates/" + cat.name + "/" + opt
		if cat.name == "cicd" {
			tmplDir = "templates/cicd/" + opt + "/" + string(cfg.Deployment)
		}
		if _, err := fs.Stat(templates, tmplDir); err != nil {
			continue // deps-only option, no template dir
		}
		sub, err := fs.Sub(templates, tmplDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s templates: %w", cat.name, err)
		}
		if err := copyDir(sub, dir, cfg, "", rep); err != nil {
			return nil, err
		}
	}
	// Playwright ships the agent MCP config too, same as create.
	if cfg.Test == "playwright" {
		if sub, err := fs.Sub(templates, "templates/agent/mcp"); err == nil {
			if err := copyDir(sub, dir, cfg, "", rep); err != nil {
				return nil, err
			}
		}
	}

	pkgPath := filepath.Join(dir, "package.json")
	raw, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("could not read package.json: %w", err)
	}
	merged, err := MergeAddPackages(raw, collectAddPackages(cfg, entries), rep)
	if err != nil {
		return nil, err
	}
	if rep.PkgJSONChanged {
		if err := os.WriteFile(pkgPath, merged, 0o644); err != nil {
			return nil, err
		}
	}

	rel := func(paths []string) []string {
		out := make([]string, 0, len(paths))
		for _, p := range paths {
			if r, err := filepath.Rel(dir, p); err == nil {
				out = append(out, r)
			} else {
				out = append(out, p)
			}
		}
		return out
	}
	rep.CreatedFiles, rep.SkippedFiles = rel(rep.CreatedFiles), rel(rep.SkippedFiles)
	return rep, nil
}

// collectAddPackages builds the packages an add contributes: each matched
// entry's own packages plus the cross-cutting rules evaluated against the
// detected config (e.g. `add prettier` on a tailwind project also brings
// prettier-plugin-tailwindcss). Reuses the create path's rule code so version
// literals are never duplicated.
func collectAddPackages(cfg ProjectConfig, entries []*OptionEntry) Packages {
	p := packageJSON{Scripts: map[string]string{}, Dependencies: map[string]string{}, DevDependencies: map[string]string{}}
	for _, e := range entries {
		mergePackages(&p, e.Packages)
	}
	applyCrossCuttingRules(&p, cfg)
	return Packages{Scripts: p.Scripts, Dependencies: p.Dependencies, DevDependencies: p.DevDependencies}
}
