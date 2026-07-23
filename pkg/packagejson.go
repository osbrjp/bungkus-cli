package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os/exec"
	"strings"
)

// DefaultNodeEngine is the engines.node constraint used unless overridden.
const DefaultNodeEngine = ">=22.12.0"

// packageJSON defines the structure and field order for generated package.json files.
// Field order in the struct controls key order in the JSON output.
type packageJSON struct {
	Name            string            `json:"name"`
	Private         bool              `json:"private,omitempty"`
	Version         string            `json:"version"`
	Type            string            `json:"type"`
	PackageManager  string            `json:"packageManager,omitempty"`
	Engines         map[string]string `json:"engines"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// BuildPackageJSON constructs a package.json by merging packages from the
// registry based on the user's selections, then marshals to formatted JSON.
func BuildPackageJSON(cfg ProjectConfig) ([]byte, error) {
	reg := GetRegistry()

	base := reg.GetBase(string(cfg.Base))
	if base == nil {
		return nil, fmt.Errorf("unknown base framework: %s", cfg.Base)
	}

	nodeEngine := cfg.NodeEngine
	if nodeEngine == "" {
		nodeEngine = DefaultNodeEngine
	}

	pkg := packageJSON{
		Name:            cfg.ProjectName,
		Private:         base.Private,
		Version:         "0.0.1",
		Type:            "module",
		Engines:         map[string]string{"node": nodeEngine},
		Scripts:         make(map[string]string),
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
	}

	// Merge in order: common → base → css → formatter → linter → validation
	// → form → query → state → cms → test → audit
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

	if cfg.Form != "none" && cfg.Form.IsValidIntegration(string(cfg.Base)) {
		if form := reg.GetForm(string(cfg.Form)); form != nil {
			mergePackages(&pkg, form.Packages)
			mergeIntegrationPackages(&pkg, form, integration)
		}
	}

	if cfg.Query != "none" && cfg.Query.IsValidIntegration(string(cfg.Base)) {
		if query := reg.GetQuery(string(cfg.Query)); query != nil {
			mergePackages(&pkg, query.Packages)
			mergeIntegrationPackages(&pkg, query, integration)
		}
	}

	if cfg.State != "none" && cfg.State.IsValidIntegration(string(cfg.Base)) {
		if state := reg.GetState(string(cfg.State)); state != nil {
			mergePackages(&pkg, state.Packages)
			mergeIntegrationPackages(&pkg, state, integration)
		}
	}

	if cfg.CMS != "none" {
		if cms := reg.GetCMS(string(cfg.CMS)); cms != nil {
			mergePackages(&pkg, cms.Packages)
		}
	}

	if cfg.Test != "none" {
		if test := reg.GetTestingFramework(string(cfg.Test)); test != nil {
			mergePackages(&pkg, test.Packages)
		}
	}

	if cfg.Audit != "none" {
		if audit := reg.GetAudit(string(cfg.Audit)); audit != nil {
			mergePackages(&pkg, audit.Packages)
		}
	}

	if cfg.Deployment != "none" {
		if dt := reg.GetDeployment(string(cfg.Deployment)); dt != nil {
			mergePackages(&pkg, dt.Packages)
		}
	}

	if v, err := pmVersion(string(cfg.PM)); err == nil {
		pkg.PackageManager = string(cfg.PM) + "@" + v
	}

	applyCrossCuttingRules(&pkg, cfg)

	// On the "latest" channel the user opts out of the vetted registry pins;
	// every dep resolves to whatever npm serves at install time. Otherwise an
	// explicit pin strategy rewrites every range operator.
	switch {
	case cfg.Channel == ChannelLatest:
		for name := range pkg.Dependencies {
			pkg.Dependencies[name] = "latest"
		}
		for name := range pkg.DevDependencies {
			pkg.DevDependencies[name] = "latest"
		}
	case cfg.Pin != "" && cfg.Pin != PinDefault:
		for name, v := range pkg.Dependencies {
			pkg.Dependencies[name] = applyPinStrategy(v, cfg.Pin)
		}
		for name, v := range pkg.DevDependencies {
			pkg.DevDependencies[name] = applyPinStrategy(v, cfg.Pin)
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(pkg); err != nil {
		return nil, fmt.Errorf("failed to marshal package.json: %w", err)
	}

	return buf.Bytes(), nil
}

func pmVersion(pm string) (string, error) {
	out, err := exec.Command(pm, "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// applyPinStrategy rewrites a version's range operator. It strips any existing
// operator (reusing bump's rangePrefix) and applies the strategy's. Values
// that aren't a plain version (e.g. "latest", "workspace:*") are left as-is.
func applyPinStrategy(version string, s PinStrategy) string {
	bare := rangePrefix.ReplaceAllString(version, "")
	if bare == "" || bare[0] < '0' || bare[0] > '9' {
		return version // not a plain semver (e.g. "latest", "workspace:*")
	}
	switch s {
	case PinCaret:
		return "^" + bare
	case PinTilde:
		return "~" + bare
	case PinExact:
		return bare
	}
	return version
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

	// eslint → globals and typescript-eslint always required by eslint.config.mjs
	if cfg.Linter == "eslint" {
		pkg.DevDependencies["globals"] = "^17.4.0"
		pkg.DevDependencies["typescript-eslint"] = "^8.58.0"
	}

	// eslint + react integration → React eslint plugins
	if cfg.Linter == "eslint" && cfg.Base.IsReactInt() {
		pkg.DevDependencies["eslint-plugin-react-hooks"] = "^7.0.1"
		pkg.DevDependencies["eslint-plugin-react-refresh"] = "^0.5.2"
	}

	// veevalidate + zod → @vee-validate/zod adapter
	if cfg.Form == "veevalidate" && cfg.Validation == "zod" && cfg.Form.IsValidIntegration(string(cfg.Base)) {
		pkg.Dependencies["@vee-validate/zod"] = "^4.15.1"
	}

	// react-hook-form + zod → @hookform/resolvers
	if cfg.Form == "react-hook-form" && cfg.Validation == "zod" && cfg.Form.IsValidIntegration(string(cfg.Base)) {
		pkg.Dependencies["@hookform/resolvers"] = "^5.2.2"
	}

	// pnpm + astro → vite as devDep
	if cfg.PM == "pnpm" && cfg.Base.IsAstro() {
		pkg.DevDependencies["vite"] = "^6.3.5"
	}

	// astro + cloudflare -> @astrojs/cloudflare adapter
	if cfg.Base.IsAstro() && (cfg.Deployment == "cloudflare-pages" || cfg.Deployment == "cloudflare-workers") {
		pkg.DevDependencies["@astrojs/cloudflare"] = "^13.0.0"
	}

}
