package pkg

import (
	"encoding/json"
	"testing"
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
	if p.Dependencies["domain"] != "workspace:*" {
		t.Errorf("frontend should depend on domain workspace:*, got %q", p.Dependencies["domain"])
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
	if p.Dependencies["domain"] != "workspace:*" {
		t.Errorf("channel=latest must not rewrite workspace deps, got domain=%q", p.Dependencies["domain"])
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
	if p.Dependencies["domain"] != "workspace:*" {
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
	if p.Name != "domain" || p.Main == "" {
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
		{"backend+bun stays flat", "hono", "bun", LayoutFlat, LayoutFlat},
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
