package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// PrintSuccess prints a styled success box with get-started instructions.
func PrintSuccess(cfg pkg.ProjectConfig) {
	header := PrimaryStyle.Render("✔ ") + "Project scaffolded at " + AccentStyle.Render(cfg.ProjectName)
	cmds := fmt.Sprintf(
		"\n\n  %s\n\n    %s\n    %s\n    %s",
		AccentStyle.Render("Get started:"),
		lipgloss.NewStyle().Foreground(ColorAccent).Render("cd "+cfg.ProjectName),
		lipgloss.NewStyle().Foreground(ColorAccent).Render(cfg.PM.InstallCmd()),
		lipgloss.NewStyle().Foreground(ColorAccent).Render(cfg.PM.RunCmd()),
	)
	fmt.Println(BoxStyle.Render(header + cmds))
}
