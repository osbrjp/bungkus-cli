/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/internal/tui"
	"github.com/osbrjp/bungkus-cli/pkg"
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

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	notify := startUpdateCheck()
	err := rootCmd.Execute()
	notify()
	if err != nil {
		os.Exit(1)
	}
}

// startUpdateCheck looks for a newer release alongside the real command and
// returns a func that prints a one-line hint once that command is done. The
// check never delays or fails the command it rides along with: it is skipped
// unless stderr is a terminal, and a network error stays silent.
func startUpdateCheck() func() {
	silent := func() {}
	if os.Getenv("BUNGKUS_NO_UPDATE_CHECK") != "" || !isTerminal(os.Stderr) {
		return silent
	}
	if len(os.Args) > 1 && os.Args[1] == "update" {
		return silent
	}

	current := rootCmd.Version
	latest := make(chan string, 1)
	go func() { latest <- pkg.AvailableUpdate(current) }()

	return func() {
		select {
		case tag := <-latest:
			if tag != "" {
				fmt.Fprintf(os.Stderr, "\na newer version is available (%s → %s) — run: bungkus-cli update\n",
					pkg.NormalizeVersion(current), tag)
			}
		case <-time.After(500 * time.Millisecond):
			// Still in flight; not worth holding the shell for.
		}
	}
}
