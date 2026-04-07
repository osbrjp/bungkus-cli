/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/internal/tui"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bungkus-cli",
	Short: "A frontend scaffolding cli tool.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the wizard
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

		// Run the scaffold spinner
		p := tea.NewProgram(tui.NewSpinnerModel(cfg))
		go func() {
			err := pkg.Scaffold(cfg.ProjectName, config.Templates, cfg)
			p.Send(tui.ScaffoldDoneMsg{Err: err})
		}()

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
