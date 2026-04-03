package pkg

// CSSFramework
type CSSFrameworkConfig struct {
	UseTailwindCSS bool
}

type CSSFramework string

const (
	VanillaCSS  CSSFramework = "vanilla"
	TailwindCSS CSSFramework = "tailwindcss"
)

// Formatter
type FormatterConfig struct {
	UsePrettier bool
	UseBiome    bool
}
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

func (c Formatter) IsValid() bool {
	switch c {
	case PrettierFmt, BiomeFmt, OxFmt:
		return true
	default:
		return false
	}
}

type ProjectConfig struct {
	ProjectName string
	Site        string
	CSS         CSSFrameworkConfig
	Fmt         FormatterConfig
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		CSS: CSSFrameworkConfig{
			UseTailwindCSS: false,
		},
		Fmt: FormatterConfig{
			UsePrettier: true,
		},
	}
}
