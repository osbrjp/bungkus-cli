package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
)

// packageJSON defines the structure and field order for generated package.json files.
// Field order in the struct controls key order in the JSON output.
type packageJSON struct {
	Name            string            `json:"name"`
	Private         bool              `json:"private,omitempty"`
	Version         string            `json:"version"`
	Type            string            `json:"type"`
	Engines         map[string]string `json:"engines"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	PNPM            *pnpmSection      `json:"pnpm,omitempty"`
}

type pnpmSection struct {
	Overrides map[string]string `json:"overrides"`
}

// BuildPackageJSON constructs a package.json by merging packages from the
// registry based on the user's selections, then marshals to formatted JSON.
func BuildPackageJSON(cfg ProjectConfig) ([]byte, error) {
	reg := GetRegistry()

	base := reg.GetBase(string(cfg.Base))
	if base == nil {
		return nil, fmt.Errorf("unknown base framework: %s", cfg.Base)
	}

	pkg := packageJSON{
		Name:            cfg.ProjectName,
		Private:         base.Private,
		Version:         "0.0.1",
		Type:            "module",
		Engines:         map[string]string{"node": ">=22.12.0"},
		Scripts:         make(map[string]string),
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
	}

	// Merge in order: common → base → css → formatter → linter → cms
	mergePackages(&pkg, reg.CommonPackages)
	mergePackages(&pkg, base.Packages)

	if css := reg.GetCSS(string(cfg.CSS)); css != nil {
		mergePackages(&pkg, css.Packages)
	}

	if fmtEntry := reg.GetFormatter(string(cfg.Fmt)); fmtEntry != nil {
		mergePackages(&pkg, fmtEntry.Packages)
	}

	if linter := reg.GetLinter(string(cfg.Linter)); linter != nil {
		mergePackages(&pkg, linter.Packages)
	}

	if cfg.Validation != "none" {
		if val := reg.GetValidation(string(cfg.Validation)); val != nil {
			mergePackages(&pkg, val.Packages)
		}
	}

	integration := base.Integration
	if cfg.Base.IsNuxt() {
		integration = "nuxt"
	}

	if cfg.Form != "none" {
		if form := reg.GetForm(string(cfg.Form)); form != nil {
			mergePackages(&pkg, form.Packages)
			mergeIntegrationPackages(&pkg, form, integration)
		}
	}

	if cfg.Query != "none" {
		if query := reg.GetQuery(string(cfg.Query)); query != nil {
			mergePackages(&pkg, query.Packages)
			mergeIntegrationPackages(&pkg, query, integration)
		}
	}

	if cfg.CMS != "none" {
		if cms := reg.GetCMS(string(cfg.CMS)); cms != nil {
			mergePackages(&pkg, cms.Packages)
		}
	}

	applyCrossCuttingRules(&pkg, cfg)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(pkg); err != nil {
		return nil, fmt.Errorf("failed to marshal package.json: %w", err)
	}

	return buf.Bytes(), nil
}

func mergeIntegrationPackages(pkg *packageJSON, entry *OptionEntry, integration string) {
	if integration == "" || entry.IntegrationPackages == nil {
		return
	}
	if pkgs, ok := entry.IntegrationPackages[integration]; ok {
		mergePackages(pkg, pkgs)
		return
	}
	// nuxt falls back to "vue" when no "nuxt" key exists
	if integration == "nuxt" {
		if pkgs, ok := entry.IntegrationPackages["vue"]; ok {
			mergePackages(pkg, pkgs)
		}
	}
}

func mergePackages(pkg *packageJSON, src Packages) {
	maps.Copy(pkg.Scripts, src.Scripts)
	maps.Copy(pkg.Dependencies, src.Dependencies)
	maps.Copy(pkg.DevDependencies, src.DevDependencies)
}

func applyCrossCuttingRules(pkg *packageJSON, cfg ProjectConfig) {
	// prettier + tailwindcss → prettier-plugin-tailwindcss
	if cfg.Fmt == "prettier" && cfg.CSS == "tailwindcss" {
		pkg.DevDependencies["prettier-plugin-tailwindcss"] = "0.7.2"
	}

	// prettier + astro → prettier-plugin-astro
	if cfg.Fmt == "prettier" && cfg.Base.IsAstro() {
		pkg.DevDependencies["prettier-plugin-astro"] = "0.14.1"
	}

	// eslint + vite-react → extra React eslint plugins
	if cfg.Linter == "eslint" && cfg.Base == "vite-react" {
		pkg.DevDependencies["globals"] = "^17.4.0"
		pkg.DevDependencies["eslint-plugin-react-hooks"] = "^7.0.1"
		pkg.DevDependencies["eslint-plugin-react-refresh"] = "^0.5.2"
		pkg.DevDependencies["typescript-eslint"] = "^8.58.0"
	}

	// veevalidate + zod → @vee-validate/zod adapter
	if cfg.Form == "veevalidate" && cfg.Validation == "zod" {
		pkg.Dependencies["@vee-validate/zod"] = "^4.15.1"
	}

	// pnpm + astro → vite as devDep
	if cfg.PM == "pnpm" && cfg.Base.IsAstro() {
		pkg.DevDependencies["vite"] = "^6.3.5"
	}

	// pnpm + (astro integration or nuxt) → pnpm overrides
	if cfg.PM == "pnpm" {
		base := GetRegistry().GetBase(string(cfg.Base))
		needsOverrides := cfg.Base.IsNuxt() || (cfg.Base.IsAstro() && base != nil && base.Integration != "")
		if needsOverrides {
			pkg.PNPM = &pnpmSection{
				Overrides: map[string]string{
					"semver@6.3.1": "^7.7.1",
				},
			}
		}
	}
}
