package pkg

import (
	"encoding/json"
	"testing"
)

type parsedPkg struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func buildAndParse(t *testing.T, cfg ProjectConfig) parsedPkg {
	t.Helper()
	data, err := BuildPackageJSON(cfg)
	if err != nil {
		t.Fatalf("BuildPackageJSON: %v", err)
	}
	var p parsedPkg
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return p
}

func has(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func TestBuildPackageJSON(t *testing.T) {
	setupRegistry(t)

	baseCfg := func(base BaseFramework) ProjectConfig {
		cfg := NewProjectConfig()
		cfg.ProjectName = "test-app"
		cfg.Base = base
		cfg.PM = "pnpm"
		return cfg
	}

	tests := []struct {
		name       string
		cfg        func() ProjectConfig
		wantDep    []string // must be in dependencies
		wantDev    []string // must be in devDependencies
		forbidDep  []string // must NOT be in dependencies
		forbidDev  []string // must NOT be in devDependencies
	}{
		{
			name: "astro-react + react-hook-form + zod installs resolver",
			cfg: func() ProjectConfig {
				c := baseCfg("astro-react")
				c.Form = "react-hook-form"
				c.Validation = "zod"
				return c
			},
			wantDep: []string{"react-hook-form", "zod", "@hookform/resolvers", "react"},
		},
		{
			name: "nuxt + veevalidate + zod installs vee-validate/zod and nuxt module",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.Form = "veevalidate"
				c.Validation = "zod"
				return c
			},
			wantDep: []string{"@vee-validate/nuxt", "@vee-validate/zod", "zod", "nuxt"},
		},
		{
			name: "nuxt + react-hook-form is skipped at merge time",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.Form = "react-hook-form"
				return c
			},
			forbidDep: []string{"react-hook-form", "@hookform/resolvers"},
		},
		{
			name: "plain vite + tanstack-form is skipped (no integration)",
			cfg: func() ProjectConfig {
				c := baseCfg("vite")
				c.Form = "tanstack-form"
				return c
			},
			forbidDep: []string{"@tanstack/react-form", "@tanstack/vue-form"},
		},
		{
			name: "astro-vue + tanstack-form installs vue binding",
			cfg: func() ProjectConfig {
				c := baseCfg("astro-vue")
				c.Form = "tanstack-form"
				return c
			},
			wantDep:   []string{"@tanstack/vue-form"},
			forbidDep: []string{"@tanstack/react-form"},
		},
		{
			name: "nuxt + pinia installs @pinia/nuxt adapter",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.State = "pinia"
				return c
			},
			wantDep: []string{"pinia", "@pinia/nuxt"},
		},
		{
			name: "astro-vue + pinia has pinia but no nuxt adapter",
			cfg: func() ProjectConfig {
				c := baseCfg("astro-vue")
				c.State = "pinia"
				return c
			},
			wantDep:   []string{"pinia"},
			forbidDep: []string{"@pinia/nuxt"},
		},
		{
			name: "astro-react + nanostores installs react binding",
			cfg: func() ProjectConfig {
				c := baseCfg("astro-react")
				c.State = "nanostores"
				return c
			},
			wantDep:   []string{"nanostores", "@nanostores/react"},
			forbidDep: []string{"@nanostores/vue"},
		},
		{
			name: "nuxt + nanostores falls back to vue binding",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.State = "nanostores"
				return c
			},
			wantDep:   []string{"nanostores", "@nanostores/vue"},
			forbidDep: []string{"@nanostores/react"},
		},
		{
			name: "plain vite + nanostores is skipped",
			cfg: func() ProjectConfig {
				c := baseCfg("vite")
				c.State = "nanostores"
				return c
			},
			forbidDep: []string{"nanostores", "@nanostores/react", "@nanostores/vue"},
		},
		{
			name: "nuxt + jotai is skipped (react-only)",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.State = "jotai"
				return c
			},
			forbidDep: []string{"jotai"},
		},
		{
			name: "vite-react + zustand installs zustand",
			cfg: func() ProjectConfig {
				c := baseCfg("vite-react")
				c.State = "zustand"
				return c
			},
			wantDep: []string{"zustand", "react"},
		},
		{
			name: "prettier + tailwindcss adds prettier-plugin-tailwindcss",
			cfg: func() ProjectConfig {
				c := baseCfg("astro-react")
				c.CSS = "tailwindcss"
				c.Fmt = "prettier"
				return c
			},
			wantDev: []string{"prettier", "prettier-plugin-tailwindcss"},
			wantDep: []string{"tailwindcss"},
		},
		{
			name: "prettier + astro adds prettier-plugin-astro",
			cfg: func() ProjectConfig {
				c := baseCfg("astro")
				c.Fmt = "prettier"
				return c
			},
			wantDev: []string{"prettier", "prettier-plugin-astro"},
		},
		{
			name: "tanstack-query on nuxt installs vue binding",
			cfg: func() ProjectConfig {
				c := baseCfg("nuxt")
				c.Query = "tanstack-query"
				return c
			},
			wantDep:   []string{"@tanstack/vue-query"},
			forbidDep: []string{"@tanstack/react-query"},
		},
		{
			name: "tanstack-query on plain vite is skipped",
			cfg: func() ProjectConfig {
				c := baseCfg("vite")
				c.Query = "tanstack-query"
				return c
			},
			forbidDep: []string{"@tanstack/react-query", "@tanstack/vue-query"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := buildAndParse(t, tt.cfg())
			for _, dep := range tt.wantDep {
				if !has(p.Dependencies, dep) {
					t.Errorf("dependencies missing %q\ngot: %v", dep, p.Dependencies)
				}
			}
			for _, dep := range tt.wantDev {
				if !has(p.DevDependencies, dep) {
					t.Errorf("devDependencies missing %q\ngot: %v", dep, p.DevDependencies)
				}
			}
			for _, dep := range tt.forbidDep {
				if has(p.Dependencies, dep) {
					t.Errorf("dependencies should not contain %q\ngot: %v", dep, p.Dependencies)
				}
			}
			for _, dep := range tt.forbidDev {
				if has(p.DevDependencies, dep) {
					t.Errorf("devDependencies should not contain %q\ngot: %v", dep, p.DevDependencies)
				}
			}
		})
	}
}
