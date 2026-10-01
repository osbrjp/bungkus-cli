package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/osbrjp/bungkus-cli/config"
)

func monoCfg() ProjectConfig {
	c := NewProjectConfig()
	c.ProjectName = "test-app"
	c.Base = "astro-react"
	c.PM = "pnpm"
	c.Layout = LayoutMonorepo
	c.Backend = "hono"
	c.ORM = "drizzle"
	c.Database = "postgres"
	c.Validation = "zod"
	c.Form = "react-hook-form"
	return c
}

func TestMonorepoFrontendPackageOmitsBackend(t *testing.T) {
	setupRegistry(t)
	p := buildAndParse(t, monoCfg())

	// backend/orm belong to apps/api, not the frontend
	for _, forbidden := range []string{"hono", "@hono/node-server", "drizzle-orm", "pg", "better-sqlite3"} {
		if has(p.Dependencies, forbidden) || has(p.DevDependencies, forbidden) {
			t.Errorf("frontend package should not contain %q in monorepo mode", forbidden)
		}
	}
	// but it should depend on the shared domain package
	if p.Dependencies[DomainPackage] != "workspace:*" {
		t.Errorf("frontend should depend on domain workspace:*, got %q", p.Dependencies[DomainPackage])
	}
	// and keep its own FE deps
	if !has(p.Dependencies, "react-hook-form") {
		t.Error("frontend should still carry its own deps")
	}
}

func TestChannelLatestPreservesWorkspaceDeps(t *testing.T) {
	setupRegistry(t)
	c := monoCfg()
	c.Channel = ChannelLatest
	p := buildAndParse(t, c)
	if p.Dependencies[DomainPackage] != "workspace:*" {
		t.Errorf("channel=latest must not rewrite workspace deps, got domain=%q", p.Dependencies[DomainPackage])
	}
	if p.Dependencies["astro"] != "latest" {
		t.Errorf("channel=latest should still pin normal deps to latest, got astro=%q", p.Dependencies["astro"])
	}
}

func TestBuildAPIPackageJSON(t *testing.T) {
	setupRegistry(t)
	data, err := BuildAPIPackageJSON(monoCfg())
	if err != nil {
		t.Fatalf("BuildAPIPackageJSON: %v", err)
	}
	var p struct {
		Name         string            `json:"name"`
		Scripts      map[string]string `json:"scripts"`
		Dependencies map[string]string `json:"dependencies"`
		DevDeps      map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Name != "api" {
		t.Errorf("api name = %q, want api", p.Name)
	}
	for _, dep := range []string{"hono", "drizzle-orm", "pg"} {
		if _, ok := p.Dependencies[dep]; !ok {
			t.Errorf("api missing dependency %q", dep)
		}
	}
	if p.Dependencies[DomainPackage] != "workspace:*" {
		t.Error("api should depend on domain workspace:*")
	}
	if p.Scripts["dev"] == "" {
		t.Error("api should have a dev script aliased from the backend watcher")
	}
	if _, ok := p.DevDeps["typescript"]; !ok {
		t.Error("api should carry typescript devDep")
	}
}

func TestBuildDomainPackageJSON(t *testing.T) {
	setupRegistry(t)

	withZod := monoCfg()
	data, _ := BuildDomainPackageJSON(withZod)
	var p struct {
		Name         string            `json:"name"`
		Main         string            `json:"main"`
		Dependencies map[string]string `json:"dependencies"`
	}
	json.Unmarshal(data, &p)
	if p.Name != DomainPackage || p.Main == "" {
		t.Errorf("domain pkg name/main wrong: %q %q", p.Name, p.Main)
	}
	if _, ok := p.Dependencies["zod"]; !ok {
		t.Error("domain should carry zod when validation=zod")
	}

	noZod := monoCfg()
	noZod.Validation = "none"
	data2, _ := BuildDomainPackageJSON(noZod)
	var p2 struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	json.Unmarshal(data2, &p2)
	if _, ok := p2.Dependencies["zod"]; ok {
		t.Error("domain should not carry zod without zod validation")
	}
}

func TestApplyDefaultLayout(t *testing.T) {
	cases := []struct {
		name    string
		backend BackendLib
		pm      PackageManager
		start   Layout
		want    Layout
	}{
		{"backend+pnpm upgrades", "hono", "pnpm", LayoutFlat, LayoutMonorepo},
		{"no backend stays flat", "none", "pnpm", LayoutFlat, LayoutFlat},
		{"backend+bun upgrades", "hono", "bun", LayoutFlat, LayoutMonorepo},
		{"backend+npm upgrades", "hono", "npm", LayoutFlat, LayoutMonorepo},
		{"backend+yarn upgrades", "elysia", "yarn", LayoutFlat, LayoutMonorepo},
		{"no backend on npm stays flat", "none", "npm", LayoutFlat, LayoutFlat},
		{"already monorepo untouched", "hono", "pnpm", LayoutMonorepo, LayoutMonorepo},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := ProjectConfig{Backend: tc.backend, PM: tc.pm, Layout: tc.start}
			c.ApplyDefaultLayout()
			if c.Layout != tc.want {
				t.Errorf("layout = %q, want %q", c.Layout, tc.want)
			}
		})
	}
}

func TestLayoutIsValid(t *testing.T) {
	if !Layout("flat").IsValid() || !Layout("monorepo").IsValid() {
		t.Error("flat and monorepo should be valid")
	}
	if Layout("polyrepo").IsValid() {
		t.Error("polyrepo should be invalid")
	}
	if !LayoutMonorepo.IsMonorepo() || LayoutFlat.IsMonorepo() {
		t.Error("IsMonorepo mismatch")
	}
}

// TestMonorepoPerPM scaffolds the monorepo with each package manager and
// checks the workspace list, the domain link and the root scripts.
func TestMonorepoPerPM(t *testing.T) {
	setupRegistry(t)
	cases := []struct {
		pm            PackageManager
		dep           string
		workspaces    []string
		pnpmWorkspace bool
		dev, build    string
		concurrently  bool
	}{
		{"pnpm", "workspace:*", nil, true,
			"pnpm --recursive --parallel run dev", "pnpm --recursive run build", false},
		{"bun", "workspace:*", []string{"apps/*", "packages/*"}, false,
			"bun run --filter '*' dev", "bun run --filter '*' build", false},
		{"npm", "*", []string{"apps/*", "packages/*"}, false,
			`concurrently -n web,api "npm run dev -w web" "npm run dev -w api"`,
			"npm run build --workspaces --if-present", true},
		{"yarn", "*", []string{"apps/*", "packages/*"}, false,
			`concurrently -n web,api "yarn workspace web run dev" "yarn workspace api run dev"`,
			"yarn workspace @repo/domain run build && yarn workspace api run build && yarn workspace web run build", true},
	}
	type pj struct {
		Workspaces      []string          `json:"workspaces"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	read := func(t *testing.T, path string) pj {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var p pj
		if err := json.Unmarshal(data, &p); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		return p
	}
	for _, tc := range cases {
		t.Run(string(tc.pm), func(t *testing.T) {
			cfg := monoCfg()
			cfg.PM = tc.pm
			dir := t.TempDir()
			if err := Scaffold(dir, config.Templates, cfg); err != nil {
				t.Fatalf("Scaffold: %v", err)
			}

			root := read(t, filepath.Join(dir, "package.json"))
			if !slices.Equal(root.Workspaces, tc.workspaces) {
				t.Errorf("workspaces = %v, want %v", root.Workspaces, tc.workspaces)
			}
			if root.Scripts["dev"] != tc.dev {
				t.Errorf("dev = %q, want %q", root.Scripts["dev"], tc.dev)
			}
			if root.Scripts["build"] != tc.build {
				t.Errorf("build = %q, want %q", root.Scripts["build"], tc.build)
			}
			if _, ok := root.DevDependencies["concurrently"]; ok != tc.concurrently {
				t.Errorf("concurrently devDep present = %v, want %v", ok, tc.concurrently)
			}

			for _, app := range []string{"apps/web", "apps/api"} {
				if got := read(t, filepath.Join(dir, app, "package.json")).Dependencies[DomainPackage]; got != tc.dep {
					t.Errorf("%s domain dep = %q, want %q", app, got, tc.dep)
				}
			}

			_, err := os.Stat(filepath.Join(dir, "pnpm-workspace.yaml"))
			if exists := err == nil; exists != tc.pnpmWorkspace {
				t.Errorf("pnpm-workspace.yaml exists = %v, want %v", exists, tc.pnpmWorkspace)
			}
		})
	}
}

// Without a backend there is only one dev server, so npm/yarn need no
// concurrently and the build skips the absent api.
func TestRootScriptsWithoutBackend(t *testing.T) {
	setupRegistry(t)
	cfg := monoCfg()
	cfg.PM, cfg.Backend, cfg.ORM, cfg.Database = "yarn", "none", "none", "none"
	data, err := BuildRootPackageJSON(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Scripts         map[string]string `json:"scripts"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	if p.Scripts["dev"] != "yarn workspace web run dev" {
		t.Errorf("dev = %q", p.Scripts["dev"])
	}
	if p.Scripts["build"] != "yarn workspace @repo/domain run build && yarn workspace web run build" {
		t.Errorf("build = %q", p.Scripts["build"])
	}
	if _, ok := p.DevDependencies["concurrently"]; ok {
		t.Error("a single dev server needs no concurrently")
	}
}

// The latest channel must not turn npm/yarn's "*" domain link into "latest",
// which would resolve it from the registry instead of the workspace.
func TestChannelLatestKeepsDomainLink(t *testing.T) {
	setupRegistry(t)
	for _, pm := range []PackageManager{"npm", "yarn"} {
		c := monoCfg()
		c.PM, c.Channel = pm, ChannelLatest
		if got := buildAndParse(t, c).Dependencies[DomainPackage]; got != "*" {
			t.Errorf("%s: domain = %q, want *", pm, got)
		}
	}
}

func TestWorkspaceRun(t *testing.T) {
	cases := map[PackageManager]string{
		"pnpm": "pnpm --filter api run db:migrate",
		"bun":  "bun run --filter api db:migrate",
		"npm":  "npm run db:migrate -w api",
		"yarn": "yarn workspace api run db:migrate",
	}
	for pm, want := range cases {
		if got := pm.WorkspaceRun("api", "db:migrate"); got != want {
			t.Errorf("%s: %q, want %q", pm, got, want)
		}
	}
}
