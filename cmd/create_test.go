package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// --dry-run prints what a preset plus flags resolves to and writes nothing.
// A preset sets several options at once, so the output must show the preset's
// values and the flags typed over them.
func TestCreateDryRun(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	c := &cobra.Command{RunE: createCmd.RunE, SilenceUsage: true}
	addCreateFlags(c)
	c.SetOut(&out)
	c.SetArgs([]string{"demo", "-t", "astro-react", "--css", "vanilla", "--dry-run"})
	if err := c.Execute(); err != nil {
		t.Fatalf("create --dry-run: %v", err)
	}
	for _, want := range []string{"base         astro-react", "css          vanilla", "form         react-hook-form", "state        nanostores"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("dry-run output missing %q:\n%s", want, out.String())
		}
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry run wrote %d entries, want none", len(entries))
	}
}

// Each flag's help text must name every value the registry accepts for it, and
// -t every preset: --help is what agents and users read to learn the options.
func TestCreateHelpListsEveryValue(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	values := func(entries []pkg.OptionEntry) []string {
		out := make([]string, 0, len(entries))
		for _, e := range entries {
			out = append(out, e.Value)
		}
		return out
	}
	r := pkg.GetRegistry()
	cases := map[string][]string{
		"css":        values(r.CSS),
		"fmt":        values(r.Formatters),
		"linter":     values(r.Linters),
		"validation": values(r.Validation),
		"form":       values(r.Form),
		"query":      values(r.Query),
		"state":      values(r.State),
		"cms":        values(r.CMS),
		"test":       values(r.Test),
		"audit":      values(r.Audit),
		"desktop":    values(r.Desktop),
		"deploy":     values(r.Deployment),
		"cicd":       values(r.CICD),
		"backend":    values(r.Backend),
		"orm":        values(r.ORM),
		"db":         values(r.Database),
	}
	for _, b := range r.Bases {
		cases["base"] = append(cases["base"], b.Value)
	}
	for _, pm := range r.PackageManagers {
		cases["pm"] = append(cases["pm"], pm.Value)
	}
	for name := range templates {
		cases["template"] = append(cases["template"], name)
	}

	c := &cobra.Command{}
	addCreateFlags(c)
	for flag, want := range cases {
		usage := c.Flags().Lookup(flag).Usage
		listed := strings.FieldsFunc(usage, func(r rune) bool { return strings.ContainsRune(" ,().", r) })
		for _, v := range want {
			if !slices.Contains(listed, v) {
				t.Errorf("--%s help does not list %q: %s", flag, v, usage)
			}
		}
	}
}
