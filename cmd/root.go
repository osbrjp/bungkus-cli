/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/internal/tui"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bungkus-cli",
	Short: "A frontend scaffolding cli tool.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the wizard.
		wizardProgram := tea.NewProgram(tui.NewWizardModel())
		wizardResult, err := wizardProgram.Run()
		if err != nil {
			return err
		}

		wm, ok := wizardResult.(tui.WizardFinalModel)
		if !ok || wm.Canceled {
			return nil
		}
		cfg := wm.Cfg

		// Scaffold the project files.
		if err := pkg.Scaffold(cfg.ProjectName, config.Templates, cfg); err != nil {
			return fmt.Errorf("scaffold failed: %w", err)
		}

		// Print success and install instructions in a styled box.
		header := tui.PrimaryStyle.Render("✔ ") + "Project scaffolded at " + tui.AccentStyle.Render(cfg.ProjectName)
		cmds := fmt.Sprintf(
			"\n\n  %s\n\n    %s\n    %s\n    %s",
			tui.AccentStyle.Render("Get started:"),
			lipgloss.NewStyle().Foreground(tui.ColorAccent).Render("cd "+cfg.ProjectName),
			lipgloss.NewStyle().Foreground(tui.ColorAccent).Render(cfg.PM.InstallCmd()),
			lipgloss.NewStyle().Foreground(tui.ColorAccent).Render(cfg.PM.RunCmd()),
		)
		fmt.Println(tui.BoxStyle.Render(header + cmds))
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
