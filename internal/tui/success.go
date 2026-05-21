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
	cmds := fmt.Sprintf(
		"\n\n  %s\n\n    %s\n    %s\n    %s",
		AccentStyle.Render("Get started:"),
		lipgloss.NewStyle().Foreground(ColorOrange).Render("cd "+cfg.ProjectName),
		lipgloss.NewStyle().Foreground(ColorOrange).Render(cfg.PM.InstallCmd()),
		lipgloss.NewStyle().Foreground(ColorOrange).Render(cfg.PM.RunCmd()),
	)

	if cfg.Deployment == "cloudflare" {
		cmds += fmt.Sprintf(
			"\n\n  %s\n\n    %s\n    %s\n    %s",
			AccentStyle.Render("Deploy to Cloudflare Pages:"),
			MutedStyle.Render("1. Push repo to GitHub"),
			MutedStyle.Render("2. dash.cloudflare.com → Workers & Pages → Create → Pages"),
			MutedStyle.Render("3. Build: "+string(cfg.PM)+" run build  |  Output: "+cfg.Base.OutputDir()),
		)
	}

	fmt.Println(BoxStyle.Render(header + cmds))
}
