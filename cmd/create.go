/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/internal/tui"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new frontend project.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.NewProjectConfig()

		if len(args) > 0 {
			cfg.ProjectName = args[0]
		}

		base, _ := cmd.Flags().GetString("base")
		cfg.Base = pkg.BaseFramework(base)
		if !cfg.Base.IsValid() {
			return fmt.Errorf("invalid base framework: %s", base)
		}

		css, _ := cmd.Flags().GetString("css")
		cfg.CSS = pkg.CSSFramework(css)
		if !cfg.CSS.IsValid() {
			return fmt.Errorf("invalid css framework: %s", css)
		}

		fmtFlag, _ := cmd.Flags().GetString("fmt")
		cfg.Fmt = pkg.Formatter(fmtFlag)
		if !cfg.Fmt.IsValid() {
			return fmt.Errorf("invalid formatter: %s", fmtFlag)
		}

		p := tea.NewProgram(tui.NewSpinnerModel(cfg))
		go func() {
			err := pkg.Scaffold(cfg.ProjectName, config.Templates, cfg)
			p.Send(tui.DoneMsg{Err: err})
		}()

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("base", "astro", "Base framework (astro, vite)")
	createCmd.Flags().String("css", "vanilla", "CSS framework (vanilla, tailwindcss)")
	createCmd.Flags().String("fmt", "prettier", "Formatter (prettier, biome)")
}
