package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spencer-osbrjp/bungkus-cli/config"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectProject(t *testing.T) {
	setupRegistry(t)

	astroPkg := `{"name": "site", "dependencies": {"astro": "^5.0.0"}}`

	cases := []struct {
		name     string
		pkgJSON  string
		files    map[string]string // relative to the project dir
		subdir   string            // detect from this subdir if set
		wantPM   PackageManager
		wantBase BaseFramework
		wantCSS  CSSFramework
		wantName string
	}{
		{
			name:    "pnpm lockfile, astro base",
			pkgJSON: astroPkg, files: map[string]string{"pnpm-lock.yaml": ""},
			wantPM: "pnpm", wantBase: "astro", wantCSS: "vanilla", wantName: "site",
		},
		{
			name:    "bun binary lockfile",
			pkgJSON: astroPkg, files: map[string]string{"bun.lockb": ""},
			wantPM: "bun", wantBase: "astro",
		},
		{
			name:    "packageManager field beats lockfile",
			pkgJSON: `{"name": "site", "packageManager": "yarn@4.1.0", "dependencies": {"astro": "^5.0.0"}}`,
			files:   map[string]string{"pnpm-lock.yaml": ""},
			wantPM:  "yarn", wantBase: "astro",
		},
		{
			name:    "no lockfile means no PM",
			pkgJSON: astroPkg, files: map[string]string{".git/HEAD": ""},
			wantPM: "", wantBase: "astro",
		},
		{
			name:    "two lockfiles are ambiguous",
			pkgJSON: astroPkg, files: map[string]string{"pnpm-lock.yaml": "", "package-lock.json": ""},
			wantPM: "", wantBase: "astro",
		},
		{
			name:    "lockfile found in a parent workspace root",
			pkgJSON: astroPkg,
			files:   map[string]string{"../../pnpm-lock.yaml": "", "../../.git/HEAD": ""},
			subdir:  "apps/web",
			wantPM:  "pnpm", wantBase: "astro",
		},
		{
			name: "astro-react beats astro",
			pkgJSON: `{"dependencies": {"astro": "1", "@astrojs/react": "1", "react": "1", "react-dom": "1"},
			           "devDependencies": {"@types/react": "1", "@types/react-dom": "1"}}`,
			wantBase: "astro-react",
		},
		{
			name:     "nuxt detected",
			pkgJSON:  `{"dependencies": {"nuxt": "1", "vue": "1", "vue-router": "1"}, "devDependencies": {"@types/node": "1"}}`,
			wantBase: "nuxt",
		},
		{
			name:     "unrecognized deps mean no base",
			pkgJSON:  `{"dependencies": {"express": "^4.0.0"}}`,
			wantBase: "",
		},
		{
			name:     "tailwind detected",
			pkgJSON:  `{"dependencies": {"astro": "1", "tailwindcss": "^4.0.0"}}`,
			wantBase: "astro", wantCSS: "tailwindcss",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := root
			if tc.subdir != "" {
				dir = filepath.Join(root, tc.subdir)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			writeFiles(t, dir, tc.files)
			// Confine PM detection: without a .git the walk-up would escape the
			// temp dir into the developer's own filesystem.
			if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
				writeFiles(t, root, map[string]string{".git/HEAD": ""})
			}

			cfg, err := DetectProject(dir, []byte(tc.pkgJSON))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.PM != tc.wantPM {
				t.Errorf("PM = %q, want %q", cfg.PM, tc.wantPM)
			}
			if cfg.Base != tc.wantBase {
				t.Errorf("Base = %q, want %q", cfg.Base, tc.wantBase)
			}
			if tc.wantCSS != "" && cfg.CSS != tc.wantCSS {
				t.Errorf("CSS = %q, want %q", cfg.CSS, tc.wantCSS)
			}
			if tc.wantName != "" && cfg.ProjectName != tc.wantName {
				t.Errorf("ProjectName = %q, want %q", cfg.ProjectName, tc.wantName)
			}
		})
	}
}

func addCfg(base BaseFramework, css CSSFramework) ProjectConfig {
	cfg := NewProjectConfig()
	cfg.Base, cfg.PM, cfg.CSS = base, "pnpm", css
	cfg.Fmt, cfg.Linter = "", ""
	return cfg
}

func TestAdd(t *testing.T) {
	setupRegistry(t)

	astroPkg := `{
  "name": "site",
  "author": "someone",
  "scripts": {"dev": "astro dev"},
  "dependencies": {"astro": "^5.0.0"}
}
`

	cases := []struct {
		name       string
		files      map[string]string
		cfg        ProjectConfig
		opt        string
		deploy     DeployTarget
		wantErr    string
		created    []string
		skipped    []string
		pkgHas     []string
		pkgMissing []string
	}{
		{
			name: "add lhci",
			cfg:  addCfg("astro", "vanilla"), opt: "lhci",
			created: []string{"lighthouserc.json", ".github/workflows/lhci.yml"},
			pkgHas:  []string{"@lhci/cli", "\"lhci\": \"lhci autorun\"", "\"author\": \"someone\""},
		},
		{
			name:  "add lhci keeps an existing lighthouserc",
			files: map[string]string{"lighthouserc.json": `{"sentinel": true}`},
			cfg:   addCfg("astro", "vanilla"), opt: "lhci",
			created: []string{".github/workflows/lhci.yml"},
			skipped: []string{"lighthouserc.json"},
		},
		{
			name: "add biome matches formatter and linter",
			cfg:  addCfg("astro", "vanilla"), opt: "biome",
			created: []string{"biome.json"},
			pkgHas:  []string{"@biomejs/biome", "\"format\"", "\"lint\""},
		},
		{
			name: "add playwright ships the mcp config too",
			cfg:  addCfg("astro", "vanilla"), opt: "playwright",
			created: []string{"playwright.config.ts", ".mcp.json"},
			pkgHas:  []string{"@playwright/test", "test:e2e"},
		},
		{
			name: "add prettier pulls cross-cutting plugins",
			cfg:  addCfg("astro", "tailwindcss"), opt: "prettier",
			pkgHas: []string{"prettier-plugin-tailwindcss", "prettier-plugin-astro"},
		},
		{
			name: "add github-actions needs a deploy target",
			cfg:  addCfg("astro", "vanilla"), opt: "github-actions",
			wantErr: "needs a deploy target",
		},
		{
			name: "add github-actions with deploy",
			cfg:  addCfg("astro", "vanilla"), opt: "github-actions", deploy: "cloudflare-pages",
			pkgHas: []string{"\"name\": \"site\""},
		},
		{
			name: "add oxfmt on astro is rejected",
			cfg:  addCfg("astro", "vanilla"), opt: "oxfmt",
			wantErr: "does not support astro",
		},
		{
			name: "unknown option",
			cfg:  addCfg("astro", "vanilla"), opt: "unknown-thing",
			wantErr: ErrUnknownAddOption.Error(),
		},
		{
			name: "none is rejected",
			cfg:  addCfg("astro", "vanilla"), opt: "none",
			wantErr: ErrUnknownAddOption.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{"package.json": astroPkg})
			writeFiles(t, dir, tc.files)

			cfg := tc.cfg
			if tc.deploy != "" {
				cfg.Deployment = tc.deploy
			}
			rep, err := Add(dir, config.Templates, cfg, tc.opt)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Add: %v", err)
			}

			for _, f := range tc.created {
				if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
					t.Errorf("expected %s to be created: %v", f, err)
				}
			}
			for _, f := range tc.skipped {
				found := false
				for _, s := range rep.SkippedFiles {
					if filepath.ToSlash(s) == f {
						found = true
					}
				}
				if !found {
					t.Errorf("expected %s in SkippedFiles, got %v", f, rep.SkippedFiles)
				}
			}

			raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, new(json.RawMessage)); err != nil {
				t.Fatalf("package.json is invalid JSON after add: %v\n%s", err, raw)
			}
			for _, s := range tc.pkgHas {
				if !strings.Contains(string(raw), s) {
					t.Errorf("package.json missing %q:\n%s", s, raw)
				}
			}
			for _, s := range tc.pkgMissing {
				if strings.Contains(string(raw), s) {
					t.Errorf("package.json should not contain %q", s)
				}
			}
		})
	}
}

func TestAddPreservesExistingFileContent(t *testing.T) {
	setupRegistry(t)
	dir := t.TempDir()
	sentinel := `{"my": "custom lighthouse config"}`
	writeFiles(t, dir, map[string]string{
		"package.json":      `{"name": "site", "dependencies": {"astro": "1"}}`,
		"lighthouserc.json": sentinel,
	})
	if _, err := Add(dir, config.Templates, addCfg("astro", "vanilla"), "lhci"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "lighthouserc.json"))
	if string(got) != sentinel {
		t.Errorf("existing file was modified:\n%s", got)
	}
}

func TestAddIsIdempotent(t *testing.T) {
	setupRegistry(t)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"package.json": `{"name": "site", "dependencies": {"astro": "1"}}`,
	})
	if _, err := Add(dir, config.Templates, addCfg("astro", "vanilla"), "lhci"); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(dir, "package.json"))

	rep, err := Add(dir, config.Templates, addCfg("astro", "vanilla"), "lhci")
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.CreatedFiles) != 0 {
		t.Errorf("second run created files: %v", rep.CreatedFiles)
	}
	if rep.PkgJSONChanged {
		t.Error("second run reported package.json changes")
	}
	after, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if string(before) != string(after) {
		t.Errorf("package.json changed on the second run:\nbefore: %s\nafter:  %s", before, after)
	}
}
