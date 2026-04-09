package pkg

// Base framework
type BaseFramework string

type AstroFramework struct {
	base        string
	integration string
}

const (
	ViteBase       BaseFramework = "vite"
	AstroBase      BaseFramework = "astro"
	AstroReactBase BaseFramework = "astro-react"
	AstroVueBase   BaseFramework = "astro-vue"
	NuxtBase       BaseFramework = "nuxt"
)

// CSSFramework
type CSSFramework string

const (
	VanillaCSS  CSSFramework = "vanilla"
	TailwindCSS CSSFramework = "tailwindcss"
)

// Formatter
type Formatter string

// PackageManager
type PackageManager string

const (
	Bun  PackageManager = "bun"
	Npm  PackageManager = "npm"
	Yarn PackageManager = "yarn"
	Pnpm PackageManager = "pnpm"
)

const (
	PrettierFmt Formatter = "prettier"
	BiomeFmt    Formatter = "biome"
	OxFmt       Formatter = "oxfmt"
)

func (b BaseFramework) IsValid() bool {
	switch b {

	case ViteBase, AstroBase, AstroVueBase, AstroReactBase, NuxtBase:
		return true
	default:
		return false
	}
}

func (b BaseFramework) IsAstro() bool {
	switch b {
	case AstroBase, AstroReactBase, AstroVueBase:
		return true
	default:
		return false
	}
}

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

func (p PackageManager) IsValid() bool {
	switch p {
	case Bun, Npm, Yarn, Pnpm:
		return true
	default:
		return false
	}
}

func (p PackageManager) InstallCmd() string {
	return string(p) + " install"
}

func (p PackageManager) Exec() string {
	switch p {
	case Yarn:
		return "yarn dlx"
	case Pnpm:
		return "pnpx"
	case Bun:
		return "bunx"
	default:
		return "npx"
	}
}

func (p PackageManager) RunCmd() string {
	if p == Npm || p == Yarn {
		return string(p) + " run dev"
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
		CSS:         VanillaCSS,
		Fmt:         PrettierFmt,
		PM:          Bun,
	}
}
