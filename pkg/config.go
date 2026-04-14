package pkg

import "errors"

type (
	BaseFramework  string
	CSSFramework   string
	Formatter      string
	Linter         string
	CMS            string
	PackageManager string
	BaseGroup      string
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

func (c CSSFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCSS(string(c))
}

func (f Formatter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasFormatter(string(f))
}

func (l Linter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasLinter(string(l))
}

func (b BaseFramework) GetGroup(base string) (BaseGroup, error) {
	if globalRegistry == nil {
		return "", errors.New("unable to read registry")
	}

	entry := globalRegistry.GetBase(base)
	return BaseGroup(entry.Group), nil
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
	CMS         CMS
	PM          PackageManager
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		CSS:         "vanilla",
		Fmt:         "prettier",
		Linter:      "eslint",
		CMS:         "none",
		PM:          "bun",
	}
}
