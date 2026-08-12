package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spencer-osbrjp/bungkus-cli/config"
)

var importSpecRE = regexp.MustCompile(`(?m)^\s*import\s+(?:[^'"]*?\sfrom\s+)?['"]([^'"]+)['"]`)

// packageName extracts the npm package a specifier resolves to
// ("eslint/config" -> "eslint", "@eslint/js" -> "@eslint/js").
func packageName(spec string) string {
	parts := strings.Split(spec, "/")
	if strings.HasPrefix(spec, "@") && len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// TestGeneratedConfigImportsAreInstalled guards the issue #118 class of bug:
// every package imported by a generated top-level config file (eslint.config,
// vite.config, astro.config, ...) must actually be installed by the generated
// package.json. A config template importing a tool the registry entry does
// not ship breaks the project's very first lint/build.
func TestGeneratedConfigImportsAreInstalled(t *testing.T) {
	setupRegistry(t)

	combo := func(base BaseFramework, mutate func(*ProjectConfig)) ProjectConfig {
		c := NewProjectConfig()
		c.Base = base
		if mutate != nil {
			mutate(&c)
		}
		return c
	}
	eslint := func(c *ProjectConfig) { c.Fmt, c.Linter, c.CSS = "prettier", "eslint", "tailwindcss" }

	combos := map[string]ProjectConfig{
		"astro_eslint":       combo("astro", eslint), // the issue #118 repro
		"astro_react_eslint": combo("astro-react", eslint),
		"astro_vue_eslint":   combo("astro-vue", eslint),
		"vite_vue_eslint":    combo("vite-vue", eslint),
		"vite_react_eslint":  combo("vite-react", eslint),
		"nuxt_eslint":        combo("nuxt", eslint),
		"vite_ox":            combo("vite", func(c *ProjectConfig) { c.Fmt, c.Linter = "oxfmt", "oxlint" }),
		"astro_defaults":     combo("astro", nil),
		"vite_playwright":    combo("vite-react", func(c *ProjectConfig) { c.Test = "playwright" }),
	}

	for name, cfg := range combos {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := Scaffold(dir, config.Templates, cfg); err != nil {
				t.Fatalf("Scaffold: %v", err)
			}

			raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			var pj struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if err := json.Unmarshal(raw, &pj); err != nil {
				t.Fatal(err)
			}
			installed := func(name string) bool {
				_, d := pj.Dependencies[name]
				_, dd := pj.DevDependencies[name]
				return d || dd
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				n := e.Name()
				if e.IsDir() || !strings.Contains(n, ".config.") {
					continue
				}
				b, err := os.ReadFile(filepath.Join(dir, n))
				if err != nil {
					t.Fatal(err)
				}
				for _, m := range importSpecRE.FindAllStringSubmatch(string(b), -1) {
					spec := m[1]
					if strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") ||
						strings.HasPrefix(spec, "node:") || strings.Contains(spec, ":") {
						continue // relative, absolute, builtin, or virtual module
					}
					if pkg := packageName(spec); !installed(pkg) {
						t.Errorf("%s imports %q but package.json does not install %q", n, spec, pkg)
					}
				}
			}
		})
	}
}
