package pkg

import (
	"errors"
	"slices"
)

type (
	BaseFramework   string
	CSSFramework    string
	Formatter       string
	Linter          string
	ValidationLib   string
	FormLib         string
	QueryLib        string
	StateLib        string
	CMS             string
	PackageManager  string
	BaseGroup       string
	BaseIntegration string
)

func (b BaseFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasBase(string(b))
}

func (b BaseFramework) IsAstro() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "astro"
}

func (b BaseFramework) IsNuxt() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "nuxt"
}

func (b BaseFramework) IsVite() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "vite"
}

func (b BaseFramework) GetIntegration() (BaseIntegration, error) {
	if globalRegistry != nil {
		return BaseIntegration(""), errors.New("unable to read registry")
	}

	entry := globalRegistry.GetBase(string(b))
	return BaseIntegration(entry.Integration), nil
}

// IsReactInt Check if integration is react
func (b BaseFramework) IsReactInt() bool {
	integration, err := b.GetIntegration()
	if err != nil {
		return false
	}

	return integration == "react"
}

// IsVueInt Check if integration is vue
func (b BaseFramework) IsVueInt() bool {
	integration, err := b.GetIntegration()
	if err != nil {
		return false
	}

	return integration == "react"
}

func (c CSSFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCSS(string(c))
}

func (f Formatter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasFormatter(string(f))
}

func (l Linter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasLinter(string(l))
}

func (f FormLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasForm(string(f))
}

func (f FormLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	form := globalRegistry.GetForm(string(f))
	if form == nil {
		return false
	}
	if len(form.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(form.RequiresIntegration, effective)
}

func (b BaseFramework) GetGroup(base string) (BaseGroup, error) {
	if globalRegistry == nil {
		return "", errors.New("unable to read registry")
	}

	entry := globalRegistry.GetBase(base)
	return BaseGroup(entry.Group), nil
}

func (v ValidationLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasValidation(string(v))
}

func (q QueryLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasQuery(string(q))
}

func (q QueryLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	query := globalRegistry.GetQuery(string(q))
	if query == nil {
		return false
	}
	if len(query.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(query.RequiresIntegration, effective)
}

func (s StateLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasState(string(s))
}

func (s StateLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	state := globalRegistry.GetState(string(s))
	if state == nil {
		return false
	}
	if len(state.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(state.RequiresIntegration, effective)
}

func (c CMS) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCMS(string(c))
}

func (p PackageManager) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasPM(string(p))
}

func (p PackageManager) InstallCmd() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.InstallCmd
		}
	}
	return string(p) + " install"
}

func (p PackageManager) Exec() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.ExecCmd
		}
	}
	return "npx"
}

func (p PackageManager) RunCmd() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.RunCmd
		}
	}
	return string(p) + " dev"
}

type ProjectConfig struct {
	ProjectName string
	Site        string
	Base        BaseFramework
	CSS         CSSFramework
	Fmt         Formatter
	Linter      Linter
	Validation  ValidationLib
	Form        FormLib
	Query       QueryLib
	State       StateLib
	CMS         CMS
	PM          PackageManager
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		Base:        "astro",
		CSS:         "vanilla",
		Fmt:         "biome",
		Linter:      "biome",
		Validation:  "none",
		Form:        "none",
		Query:       "none",
		State:       "none",
		CMS:         "none",
		PM:          "pnpm",
	}
}
