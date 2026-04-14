/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

		// Scaffold project files.
		if err := pkg.Scaffold(cfg.ProjectName, config.Templates, cfg); err != nil {
			return fmt.Errorf("scaffold failed: %w", err)
		}

		// Initialize git repository.
		for _, args := range [][]string{
			{"git", "init"},
			{"git", "add", "."},
			{"git", "commit", "-m", "initial commit"},
		} {
			c := exec.Command(args[0], args[1:]...)
			c.Dir = cfg.ProjectName
			if err := c.Run(); err != nil {
				return fmt.Errorf("git init failed: %w", err)
			}
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
