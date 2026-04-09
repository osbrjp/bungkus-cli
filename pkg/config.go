package pkg

type BaseFramework string
type CSSFramework string
type Formatter string
type PackageManager string

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

func (c CSSFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCSS(string(c))
}

func (f Formatter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasFormatter(string(f))
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
	PM          PackageManager
	NoGit       bool
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		CSS:         "vanilla",
		Fmt:         "prettier",
		PM:          "bun",
	}
}
