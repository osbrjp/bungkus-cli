package pkg

import (
	"encoding/json"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/osbrjp/bungkus-cli/config"
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
	lhciAstro := func() ProjectConfig {
		c := NewProjectConfig()
		c.Audit = "lhci"
		return c
	}
	lhciNuxt := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Audit = "nuxt", "lhci"
		return c
	}
	tauriAstroReact := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Desktop = "astro-react", "tauri"
		return c
	}
	tauriNuxt := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Desktop = "nuxt", "tauri"
		return c
	}
	tauriMonorepo := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Backend, c.ORM, c.Database = "astro-react", "hono", "drizzle", "sqlite"
		c.PM, c.Layout, c.Desktop = "pnpm", LayoutMonorepo, "tauri"
		return c
	}
	lhciMonorepo := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Backend, c.ORM, c.Database = "astro-react", "hono", "drizzle", "sqlite"
		c.PM, c.Layout, c.Audit = "pnpm", LayoutMonorepo, "lhci"
		return c
	}
	ciMonorepo := func() ProjectConfig {
		c := NewProjectConfig()
		c.Base, c.Backend, c.ORM, c.Database = "astro-react", "hono", "drizzle", "sqlite"
		c.PM, c.Layout = "pnpm", LayoutMonorepo
		c.Test, c.Deployment, c.CICD = "playwright", "cloudflare-pages", "github-actions"
		return c
	}

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
			absent: []string{".mcp.json", "docker-compose.yml", "apps/api/db/seed.ts", "src-tauri"},
		},
		{
			name:    "lhci_astro",
			cfg:     lhciAstro(),
			present: []string{"lighthouserc.json"},
			contains: map[string][]string{
				".github/workflows/lhci.yml": {`pages_dir="src/pages"`, "steps.affected.outputs.urls"},
			},
		},
		{
			name: "lhci_nuxt",
			cfg:  lhciNuxt(),
			contains: map[string][]string{
				".github/workflows/lhci.yml": {`pages_dir="app/pages"`, "steps.affected.outputs.urls"},
			},
		},
		{
			name: "tauri_astro_react",
			cfg:  tauriAstroReact(),
			present: []string{
				"src-tauri/Cargo.toml", "src-tauri/build.rs", "src-tauri/src/main.rs",
				"src-tauri/src/lib.rs", "src-tauri/capabilities/default.json",
				"src-tauri/.gitignore", "src-tauri/icons/icon.ico", "src-tauri/icons/icon.icns",
				"src-tauri/icons/32x32.png",
			},
			contains: map[string][]string{
				"src-tauri/tauri.conf.json": {
					"http://localhost:3000", `"frontendDist": "../dist"`,
					"pnpm run dev", "pnpm run build", "com.example.my-app",
				},
				"src-tauri/Cargo.toml": {`name = "my-app"`},
				"package.json":         {"@tauri-apps/cli", `"tauri": "tauri"`},
				"README.md":            {"tauri dev"},
			},
		},
		{
			name: "tauri_nuxt",
			cfg:  tauriNuxt(),
			contains: map[string][]string{
				"src-tauri/tauri.conf.json": {"pnpm run generate"},
				"nuxt.config.ts":            {"ssr: false"},
			},
		},
		{
			// #115: the workflow must land at the workspace root (GitHub reads
			// nothing else), run lhci inside apps/web, and diff apps/web paths.
			name:    "lhci_monorepo",
			cfg:     lhciMonorepo(),
			present: []string{".github/workflows/lhci.yml", "apps/web/lighthouserc.json"},
			absent:  []string{"apps/web/.github"},
			contains: map[string][]string{
				".github/workflows/lhci.yml": {
					"working-directory: apps/web",
					`pages_dir="apps/web/src/pages"`,
					"- 'apps/web/**'",
				},
			},
		},
		{
			// #125: same class as #115 — the playwright and deploy workflows
			// must land at the workspace root (GitHub reads nothing else) with
			// their app-scoped steps running inside apps/web, while
			// playwright.config.ts stays in the web app.
			name: "ci_monorepo",
			cfg:  ciMonorepo(),
			present: []string{
				".github/workflows/playwright.yml",
				".github/workflows/deploy.yml",
				"apps/web/playwright.config.ts",
			},
			absent: []string{"apps/web/.github"},
			contains: map[string][]string{
				".github/workflows/playwright.yml": {
					"working-directory: apps/web",
					"path: apps/web/playwright-report/",
					"- 'apps/web/**'",
				},
				".github/workflows/deploy.yml": {
					"workingDirectory: apps/web",
					"working-directory: apps/web",
					"- 'apps/web/**'",
				},
			},
		},
		{
			name: "tauri_monorepo",
			cfg:  tauriMonorepo(),
			present: []string{
				"apps/web/src-tauri/tauri.conf.json",
			},
			contains: map[string][]string{
				"apps/web/package.json": {"@tauri-apps/cli"},
				"README.md":             {"--filter web tauri dev"},
			},
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

// TestTauriIconsRGBA guards the Tauri icon footgun: PNG icons must be square
// and RGBA (tauri-build rejects palette-based PNGs at compile time), and the
// committed binary assets must decode at all.
func TestTauriIconsRGBA(t *testing.T) {
	for _, name := range []string{"32x32.png", "128x128.png", "128x128@2x.png"} {
		f, err := config.Templates.Open("templates/desktop/tauri/src-tauri/icons/" + name)
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		b := img.Bounds()
		if b.Dx() != b.Dy() {
			t.Errorf("%s is not square: %dx%d", name, b.Dx(), b.Dy())
		}
		switch img.ColorModel() {
		case color.RGBAModel, color.NRGBAModel, color.RGBA64Model, color.NRGBA64Model:
		default:
			t.Errorf("%s is not an RGBA png (tauri build would reject it)", name)
		}
	}
	for _, name := range []string{"icon.icns", "icon.ico"} {
		f, err := config.Templates.Open("templates/desktop/tauri/src-tauri/icons/" + name)
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		f.Close()
	}
}

// TestAstroConfigCommas guards the astro.config.mjs template's comma wiring:
// astro-react/astro-vue used to render the env block's `}` with no comma
// before `integrations:`, and the tailwind branch appended a second comma
// after the integrations list (`react()],,`), so every scaffolded astro
// integration project failed its very first build.
func TestAstroConfigCommas(t *testing.T) {
	setupRegistry(t)
	missingComma := regexp.MustCompile(`\}\n\s*(integrations|vite):`)
	for _, base := range []BaseFramework{"astro", "astro-react", "astro-vue"} {
		for _, css := range []CSSFramework{"vanilla", "tailwindcss"} {
			t.Run(string(base)+"_"+string(css), func(t *testing.T) {
				c := NewProjectConfig()
				c.Base, c.CSS = base, css
				dir := t.TempDir()
				if err := Scaffold(dir, config.Templates, c); err != nil {
					t.Fatalf("Scaffold: %v", err)
				}
				b, err := os.ReadFile(filepath.Join(dir, "astro.config.mjs"))
				if err != nil {
					t.Fatal(err)
				}
				s := string(b)
				if strings.Contains(s, ",,") {
					t.Errorf("doubled comma in astro.config.mjs:\n%s", s)
				}
				if missingComma.MatchString(s) {
					t.Errorf("missing comma before top-level key in astro.config.mjs:\n%s", s)
				}
			})
		}
	}
}
