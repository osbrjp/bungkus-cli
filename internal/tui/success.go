package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// PrintSkippedIntegration prints a styled warning that a library is skipped
// because it's not compatible with the chosen base framework.
func PrintSkippedIntegration(lib, base string) {
	tag := WarnStyle.Render(" WARN ")
	msg := fmt.Sprintf(
		"%s %s is not supported on %s — skipping %s",
		tag,
		AccentStyle.Render(lib),
		AccentStyle.Render(base),
		SkipStyle.Render(lib),
	)
	fmt.Println(msg)
}

// PrintSuccess prints a styled success box with get-started instructions.
func PrintSuccess(cfg pkg.ProjectConfig) {
	header := PrimaryStyle.Render("✔ ") + "Project scaffolded at " + AccentStyle.Render(cfg.ProjectName)

	// Only show "cd <name>" when scaffolded into a new subfolder, not when using ".".
	var cdLine string
	if cfg.DestDir != "." {
		cdLine = "\n    " + lipgloss.NewStyle().Foreground(ColorOrange).Render("cd "+cfg.ProjectName)
	}

	cmds := fmt.Sprintf(
		"\n\n  %s%s\n    %s\n    %s",
		AccentStyle.Render("Get started:"),
		cdLine,
		lipgloss.NewStyle().Foreground(ColorOrange).Render(cfg.PM.InstallCmd()),
		lipgloss.NewStyle().Foreground(ColorOrange).Render(cfg.PM.RunCmd()),
	)

	if cfg.Deployment == "cloudflare-pages" {
		cmds += fmt.Sprintf(
			"\n\n  %s\n\n    %s\n    %s\n    %s",
			AccentStyle.Render("Deploy to Cloudflare Pages:"),
			MutedStyle.Render("1. Push repo to GitHub"),
			MutedStyle.Render("2. dash.cloudflare.com → Workers & Pages → Create → Pages"),
			MutedStyle.Render("3. Build: "+string(cfg.PM)+" run build  |  Output: "+cfg.Base.OutputDir()),
		)
	} else if cfg.Deployment == "cloudflare-workers" {
		cmds += fmt.Sprintf(
			"\n\n  %s\n\n    %s\n    %s\n    %s",
			AccentStyle.Render("Deploy to Cloudflare Workers:"),
			MutedStyle.Render("Local:  wrangler login (once) → "+string(cfg.PM)+" run deploy"),
			MutedStyle.Render("CI/CD:  gh secret set CLOUDFLARE_API_TOKEN"),
			MutedStyle.Render("        gh secret set CLOUDFLARE_ACCOUNT_ID  ← wrangler whoami"),
		)
	}

	fmt.Println(BoxStyle.Render(header + cmds))
}
