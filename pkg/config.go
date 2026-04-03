package pkg

// Base framework
type BaseFramework string

const (
	ViteBase  BaseFramework = "vite"
	AstroBase BaseFramework = "astro"
)

// CSSFramework
type CSSFramework string

const (
	VanillaCSS  CSSFramework = "vanilla"
	TailwindCSS CSSFramework = "tailwindcss"
)

// Formatter
type Formatter string

const (
	PrettierFmt Formatter = "prettier"
	BiomeFmt    Formatter = "biome"
	OxFmt       Formatter = "oxfmt"
)

func (c CSSFramework) IsValid() bool {
	switch c {
	case VanillaCSS, TailwindCSS:
		return true
	default:
		return false
	}
}

func (f Formatter) IsValid() bool {
	switch f {
	case PrettierFmt, BiomeFmt, OxFmt:
		return true
	default:
		return false
	}
}

func (b BaseFramework) IsValid() bool {
	switch b {
	case ViteBase, AstroBase:
		return true
	default:
		return false
	}
}

type ProjectConfig struct {
	ProjectName string
	Site        string
	Base        BaseFramework
	CSS         CSSFramework
	Fmt         Formatter
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		CSS:         VanillaCSS,
		Fmt:         PrettierFmt,
	}
}
