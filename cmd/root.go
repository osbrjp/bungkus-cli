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
		wizardResult, err := tea.NewProgram(tui.NewWizardModel()).Run()
		if err != nil {
			return err
		}

		wm, ok := wizardResult.(tui.WizardModel)
		if !ok || wm.Canceled {
			return nil
		}
		cfg := wm.Cfg

		if cfg.CICD != "none" && cfg.Deployment == "none" {
			tui.PrintCICDSkipped()
			cfg.CICD = "none"
		}

		if cfg.DestDir != "." {
			if err := pkg.ValidateProjectName(cfg.ProjectName); err != nil {
				return err
			}
		}

		destDir := cfg.ProjectName
		if cfg.DestDir != "" {
			destDir = cfg.DestDir
		}
		if err := pkg.ValidateDest(destDir); err != nil {
			return err
		}

		// Scaffold project files.
		if err := pkg.Scaffold(destDir, config.Templates, cfg); err != nil {
			return fmt.Errorf("scaffold failed: %w", err)
		}

		// Optional post-steps (install, git init) gated by advanced config.
		if err := pkg.PostScaffold(destDir, cfg); err != nil {
			return err
		}

		tui.PrintSuccess(cfg)
		return nil
	},
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
