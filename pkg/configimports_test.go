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

// TestEslintReactPluginsWired asserts the positive direction the import-vs-
// deps guard can't see: react bases must actually get the react eslint
// plugins (both the config imports and the installed packages), and
// non-react bases must not. Before the GetIntegration nil-check fix,
// IsReactInt always returned false and this wiring silently never fired.
func TestEslintReactPluginsWired(t *testing.T) {
	setupRegistry(t)
	scaffoldEslint := func(base BaseFramework) (conf, pkgJSON string) {
		c := NewProjectConfig()
		c.Base, c.Fmt, c.Linter = base, "prettier", "eslint"
		dir := t.TempDir()
		if err := Scaffold(dir, config.Templates, c); err != nil {
			t.Fatalf("Scaffold(%s): %v", base, err)
		}
		cb, err := os.ReadFile(filepath.Join(dir, "eslint.config.mjs"))
		if err != nil {
			t.Fatalf("read eslint config for %s: %v", base, err)
		}
		pb, err := os.ReadFile(filepath.Join(dir, "package.json"))
		if err != nil {
			t.Fatal(err)
		}
		return string(cb), string(pb)
	}

	for _, base := range []BaseFramework{"astro-react", "vite-react"} {
		conf, pkgJSON := scaffoldEslint(base)
		for _, want := range []string{"eslint-plugin-react-hooks", "eslint-plugin-react-refresh"} {
			if !strings.Contains(conf, want) {
				t.Errorf("%s: eslint.config.mjs missing %s", base, want)
			}
			if !strings.Contains(pkgJSON, want) {
				t.Errorf("%s: package.json missing %s", base, want)
			}
		}
	}
	for _, base := range []BaseFramework{"astro", "astro-vue", "vite-vue", "nuxt"} {
		conf, pkgJSON := scaffoldEslint(base)
		if strings.Contains(conf, "eslint-plugin-react") || strings.Contains(pkgJSON, "eslint-plugin-react") {
			t.Errorf("%s: react eslint plugins leaked into a non-react base", base)
		}
	}
}
