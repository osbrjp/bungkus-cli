package pkg

import "testing"

func TestBackendORMPackageJSON(t *testing.T) {
	setupRegistry(t)

	baseCfg := func() ProjectConfig {
		c := NewProjectConfig()
		c.ProjectName = "test-app"
		c.Base = "astro"
		c.PM = "pnpm"
		return c
	}

	tests := []struct {
		name      string
		cfg       func() ProjectConfig
		wantDep   []string
		wantDev   []string
		forbidDep []string
	}{
		{
			name:    "hono backend adds server deps",
			cfg:     func() ProjectConfig { c := baseCfg(); c.Backend = "hono"; return c },
			wantDep: []string{"hono", "@hono/node-server"},
			wantDev: []string{"tsx"},
		},
		{
			name:    "elysia backend adds elysia",
			cfg:     func() ProjectConfig { c := baseCfg(); c.Backend = "elysia"; return c },
			wantDep: []string{"elysia"},
		},
		{
			name:    "drizzle + sqlite picks better-sqlite3 driver",
			cfg:     func() ProjectConfig { c := baseCfg(); c.ORM = "drizzle"; c.Database = "sqlite"; return c },
			wantDep: []string{"drizzle-orm", "better-sqlite3"},
			wantDev: []string{"drizzle-kit", "@types/better-sqlite3"},
		},
		{
			name:      "drizzle + postgres picks pg driver, not sqlite",
			cfg:       func() ProjectConfig { c := baseCfg(); c.ORM = "drizzle"; c.Database = "postgres"; return c },
			wantDep:   []string{"drizzle-orm", "pg"},
			wantDev:   []string{"@types/pg"},
			forbidDep: []string{"better-sqlite3", "mysql2"},
		},
		{
			name:      "drizzle + mysql picks mysql2 driver",
			cfg:       func() ProjectConfig { c := baseCfg(); c.ORM = "drizzle"; c.Database = "mysql"; return c },
			wantDep:   []string{"drizzle-orm", "mysql2"},
			forbidDep: []string{"better-sqlite3", "pg"},
		},
		{
			name:    "drizzle without db defaults to sqlite driver",
			cfg:     func() ProjectConfig { c := baseCfg(); c.ORM = "drizzle"; return c },
			wantDep: []string{"drizzle-orm", "better-sqlite3"},
		},
		{
			name:      "prisma bundles its own engine, no drizzle driver",
			cfg:       func() ProjectConfig { c := baseCfg(); c.ORM = "prisma"; c.Database = "postgres"; return c },
			wantDep:   []string{"@prisma/client"},
			wantDev:   []string{"prisma"},
			forbidDep: []string{"pg", "better-sqlite3", "drizzle-orm"},
		},
		{
			name:      "defaults add no backend/orm packages",
			cfg:       baseCfg,
			forbidDep: []string{"hono", "elysia", "drizzle-orm", "@prisma/client"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := buildAndParse(t, tt.cfg())
			for _, d := range tt.wantDep {
				if !has(p.Dependencies, d) {
					t.Errorf("missing dependency %q", d)
				}
			}
			for _, d := range tt.wantDev {
				if !has(p.DevDependencies, d) {
					t.Errorf("missing devDependency %q", d)
				}
			}
			for _, d := range tt.forbidDep {
				if has(p.Dependencies, d) {
					t.Errorf("unexpected dependency %q", d)
				}
			}
		})
	}
}

func TestBackendORMDatabaseIsValid(t *testing.T) {
	setupRegistry(t)

	valid := func(ok bool, name string) {
		if !ok {
			t.Errorf("%s should be valid", name)
		}
	}
	invalid := func(ok bool, name string) {
		if ok {
			t.Errorf("%s should be invalid", name)
		}
	}

	valid(BackendLib("none").IsValid(), "backend none")
	valid(BackendLib("hono").IsValid(), "backend hono")
	valid(BackendLib("elysia").IsValid(), "backend elysia")
	invalid(BackendLib("express").IsValid(), "backend express")

	valid(ORMLib("none").IsValid(), "orm none")
	valid(ORMLib("drizzle").IsValid(), "orm drizzle")
	valid(ORMLib("prisma").IsValid(), "orm prisma")
	invalid(ORMLib("typeorm").IsValid(), "orm typeorm")

	valid(Database("none").IsValid(), "db none")
	valid(Database("sqlite").IsValid(), "db sqlite")
	valid(Database("postgres").IsValid(), "db postgres")
	valid(Database("mysql").IsValid(), "db mysql")
	invalid(Database("mongodb").IsValid(), "db mongodb")
}
