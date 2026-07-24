package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spencer-osbrjp/bungkus-cli/config"
)

// goTemplateMarkers are fragments that only appear in an *unrendered* Go
// template. They are chosen not to collide with Vue's `{{ count }}`
// interpolation, which is legitimately present in scaffolded .vue files.
var goTemplateMarkers = []string{
	"{{ .", "{{.", "{{-", "-}}",
	"{{ if", "{{if", "{{ end", "{{end", "{{ range", "{{ else",
	"{{ printf", "{{ eq", "{{ ne", "{{ or", "{{ and",
}

// TestScaffoldRenders scaffolds representative combos in-process and asserts
// two universal invariants on every emitted file — no unrendered template
// residue, and valid JSON — plus a few targeted content checks. This is the
// guard that would have caught the sqlite `file:` runtime bug, which passed
// `go test` but broke at runtime.
func TestScaffoldRenders(t *testing.T) {
	setupRegistry(t)

	fullstack := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Backend, c.ORM, c.Database = "astro-react", "hono", "drizzle", "sqlite"
		c.PM, c.Layout = "pnpm", LayoutMonorepo
		return c
	}
	prismaPostgres := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Backend, c.ORM, c.Database = "nuxt", "elysia", "prisma", "postgres"
		c.PM, c.Layout = "pnpm", LayoutMonorepo
		return c
	}
	vitePlaywright := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Test, c.CSS = "vite-react", "playwright", "tailwindcss"
		return c
	}
	plainAstro := func() ProjectConfig { return NewProjectConfig() }

	cases := []struct {
		name string
		cfg  ProjectConfig
		// present/absent are project-root-relative paths that must / must not exist.
		present []string
		absent  []string
		// contains maps a root-relative path to substrings it must contain.
		contains map[string][]string
	}{
		{
			name:    "fullstack_drizzle_sqlite",
			cfg:     fullstack(),
			present: []string{"apps/api/db/seed.ts", ".claude/settings.json", ".claude/commands/verify.md"},
			absent:  []string{".mcp.json"},
			contains: map[string][]string{
				"apps/api/server/index.ts": {"/health-check"},
				"apps/api/db/index.ts":     {"replace(/^file:"},
				"apps/api/package.json":    {"db:seed"},
				"AGENTS.md":                {"localhost:8000", "db:seed"},
			},
		},
		{
			name:    "fullstack_prisma_postgres",
			cfg:     prismaPostgres(),
			present: []string{"apps/api/prisma/seed.ts", "docker-compose.yml"},
			contains: map[string][]string{
				".claude/settings.json":    {"docker compose"},
				"apps/api/server/index.ts": {"/health-check"},
			},
		},
		{
			name:    "vite_playwright_mcp",
			cfg:     vitePlaywright(),
			present: []string{".mcp.json"},
			contains: map[string][]string{
				".mcp.json": {"playwright", "@playwright/mcp"},
			},
		},
		{
			name:   "plain_astro",
			cfg:    plainAstro(),
			absent: []string{".mcp.json", "docker-compose.yml", "apps/api/db/seed.ts"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := Scaffold(dir, config.Templates, tc.cfg); err != nil {
				t.Fatalf("Scaffold: %v", err)
			}

			// Universal invariants across every emitted file.
			err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, _ := filepath.Rel(dir, path)
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				content := string(b)
				for _, m := range goTemplateMarkers {
					if strings.Contains(content, m) {
						t.Errorf("unrendered template marker %q in %s", m, rel)
					}
				}
				// tsconfig*.json and biome.json are JSONC (they carry comments);
				// the scaffold leaves them unformatted, so don't strict-parse them.
				base := filepath.Base(path)
				isJSONC := strings.HasPrefix(base, "tsconfig") || base == "biome.json"
				if strings.HasSuffix(path, ".json") && !isJSONC {
					if err := json.Unmarshal(b, new(json.RawMessage)); err != nil {
						t.Errorf("invalid JSON in %s: %v", rel, err)
					}
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}

			for _, p := range tc.present {
				if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
					t.Errorf("expected %s to exist: %v", p, err)
				}
			}
			for _, p := range tc.absent {
				if _, err := os.Stat(filepath.Join(dir, p)); err == nil {
					t.Errorf("expected %s NOT to exist", p)
				}
			}
			for p, subs := range tc.contains {
				b, err := os.ReadFile(filepath.Join(dir, p))
				if err != nil {
					t.Errorf("read %s: %v", p, err)
					continue
				}
				for _, s := range subs {
					if !strings.Contains(string(b), s) {
						t.Errorf("%s: expected to contain %q", p, s)
					}
				}
			}
		})
	}
}
