package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/internal/tui"
	"github.com/osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

// TestTemplatePresetsAreValid runs every preset through the same enum checks
// `create` applies to flags. A typo in a preset (e.g. vite-vue once set
// Linter "prettier") otherwise only surfaces when a user runs -t <name>.
func TestTemplatePresetsAreValid(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	for name, factory := range templates {
		c := factory()
		checks := map[string]bool{
			"base":       c.Base.IsValid(),
			"css":        c.CSS.IsValid(),
			"fmt":        c.Fmt.IsValid(),
			"linter":     c.Linter.IsValid(),
			"validation": c.Validation.IsValid(),
			"form":       c.Form.IsValid(),
			"query":      c.Query.IsValid(),
			"state":      c.State.IsValid(),
			"cms":        c.CMS.IsValid(),
			"pm":         c.PM.IsValid(),
		}
		for field, ok := range checks {
			if !ok {
				t.Errorf("template %q: invalid %s", name, field)
			}
		}
	}
}

// The wizard's command preview must be exact: feeding its arguments back
// through `create`'s flag parsing rebuilds the same ProjectConfig.
func TestWizardCommandReproducesConfig(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*pkg.ProjectConfig){
		"defaults": func(*pkg.ProjectConfig) {},
		"frontend": func(c *pkg.ProjectConfig) {
			c.ProjectName, c.Base, c.CSS, c.Fmt, c.Linter = "site", "astro-react", "tailwindcss", "prettier", "eslint"
			c.Validation, c.Form, c.Query, c.State = "zod", "react-hook-form", "tanstack-query", "jotai"
			c.CMS, c.Test, c.Audit, c.Deployment, c.CICD = "microcms", "playwright", "lhci", "cloudflare-pages", "github-actions"
		},
		"monorepo": func(c *pkg.ProjectConfig) {
			c.Base, c.Backend, c.ORM, c.Database, c.Layout = "nuxt", "hono", "drizzle", "postgres", pkg.LayoutMonorepo
		},
		"backend on bun": func(c *pkg.ProjectConfig) {
			c.Base, c.Backend, c.ORM, c.Database, c.PM, c.Layout = "vite-vue", "elysia", "prisma", "sqlite", "bun", pkg.LayoutMonorepo
		},
		"backend on npm": func(c *pkg.ProjectConfig) {
			c.Base, c.Backend, c.PM, c.Layout = "astro-react", "hono", "npm", pkg.LayoutMonorepo
		},
		"backend on yarn": func(c *pkg.ProjectConfig) {
			c.Base, c.Backend, c.PM, c.Layout = "astro-react", "hono", "yarn", pkg.LayoutMonorepo
		},
		"advanced": func(c *pkg.ProjectConfig) {
			c.Channel, c.Pin, c.Install, c.GitInit, c.NodeEngine = pkg.ChannelLatest, pkg.PinExact, true, false, ">=20.11.0"
		},
		"current directory": func(c *pkg.ProjectConfig) {
			c.ProjectName, c.DestDir, c.Desktop = filepath.Base(wd), ".", "tauri"
		},
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			want := pkg.NewProjectConfig()
			edit(&want)

			args := tui.CreateArgs(want)
			c := &cobra.Command{}
			addCreateFlags(c)
			if err := c.ParseFlags(args[1:]); err != nil {
				t.Fatalf("parse %q: %v", args, err)
			}
			got, err := configFromFlags(c, c.Flags().Args())
			if err != nil {
				t.Fatalf("configFromFlags(%q): %v", args, err)
			}
			got.Date, want.Date = "", ""
			if got != want {
				t.Errorf("args %q\n got %+v\nwant %+v", args, got, want)
			}
		})
	}
}

// create accepts --layout monorepo for every package manager and scaffolds
// the workspace pieces.
func TestCreateMonorepoAllPMs(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	for _, pm := range []string{"pnpm", "bun", "npm", "yarn"} {
		t.Run(pm, func(t *testing.T) {
			t.Chdir(t.TempDir())
			c := &cobra.Command{RunE: createCmd.RunE, SilenceUsage: true}
			addCreateFlags(c)
			c.SetArgs([]string{"demo", "--base", "astro-react", "--backend", "hono", "--layout", "monorepo", "--pm", pm, "--git=false"})
			if err := c.Execute(); err != nil {
				t.Fatalf("create --pm %s --layout monorepo: %v", pm, err)
			}
			for _, f := range []string{"package.json", "apps/web/package.json", "apps/api/package.json", "packages/domain/package.json"} {
				if _, err := os.Stat(filepath.Join("demo", f)); err != nil {
					t.Errorf("missing %s: %v", f, err)
				}
			}
		})
	}
}
