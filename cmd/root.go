/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
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

		tui.PrintSuccess(cfg)
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
