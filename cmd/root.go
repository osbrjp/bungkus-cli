/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bungkus-cli",
	Short: "A frontend scaffolding cli tool.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.NewProjectConfig()

		css, _ := cmd.Flags().GetString("css")

		cssFramework := pkg.CSSFramework(css)
		if !cssFramework.IsValid() {
			return fmt.Errorf("invalid css framework: %s", css)
		}

		cfg.CSS = pkg.CSSFrameworkConfig{
			UseTailwindCSS: cssFramework == pkg.TailwindCSS,
		}

		destDir := cfg.ProjectName

		fmt.Printf("Scaffolding project in %s...\n", destDir)

		if err := pkg.Scaffold(destDir, config.Templates, cfg); err != nil {
			return err
		}

		fmt.Println("Done!")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("css", "vanilla", "Css Framework")
}
