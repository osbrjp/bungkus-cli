package cmd

import (
	"fmt"

	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/internal/tui"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

// templates are preset factories. Each starts from NewProjectConfig() so
// new ProjectConfig fields inherit sane defaults without editing every
// template.
var templates = map[string]func() pkg.ProjectConfig{
	"astro": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "astro"
		c.CSS = "tailwindcss"
		c.Fmt = "prettier"
		c.Linter = "eslint"
		return c
	},
	"astro-react": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "astro-react"
		c.CSS = "tailwindcss"
		c.Validation = "zod"
		c.Form = "react-hook-form"
		c.Query = "tanstack-query"
		c.State = "nanostores"
		return c
	},
	"astro-vue": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "astro-vue"
		c.CSS = "tailwindcss"
		c.Fmt = "prettier"
		c.Linter = "eslint"
		c.Validation = "zod"
		c.Form = "veevalidate"
		c.Query = "tanstack-query"
		c.State = "pinia"
		return c
	},
	"nuxt": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "nuxt"
		c.CSS = "tailwindcss"
		c.Fmt = "prettier"
		c.Linter = "eslint"
		c.Validation = "zod"
		c.Form = "veevalidate"
		c.Query = "tanstack-query"
		c.State = "pinia"
		return c
	},
	"vite": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "vite"
		c.CSS = "tailwindcss"
		return c
	},
	"vite-react": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "vite-react"
		c.CSS = "tailwindcss"
		c.Validation = "zod"
		c.Form = "react-hook-form"
		c.Query = "tanstack-query"
		c.State = "zustand"
		return c
	},
	"vite-vue": func() pkg.ProjectConfig {
		c := pkg.NewProjectConfig()
		c.Base = "vite-vue"
		c.CSS = "tailwindcss"
		c.Fmt = "prettier"
		c.Linter = "prettier"
		c.Validation = "zod"
		c.Form = "veevalidate"
		c.Query = "tanstack-query"
		c.State = "pinia"
		return c
	},
}

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new frontend project.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.NewProjectConfig()

		if name, _ := cmd.Flags().GetString("template"); name != "" {
			tmpl, ok := templates[name]
			if !ok {
				return fmt.Errorf("invalid template: %s", name)
			}
			fmt.Printf("Template selected: %s\n", name)
			cfg = tmpl()
		}

		if len(args) > 0 {
			cfg.ProjectName = args[0]
		}

		if cmd.Flags().Changed("base") {
			v, _ := cmd.Flags().GetString("base")
			cfg.Base = pkg.BaseFramework(v)
		}
		if cmd.Flags().Changed("css") {
			v, _ := cmd.Flags().GetString("css")
			cfg.CSS = pkg.CSSFramework(v)
		}
		if cmd.Flags().Changed("fmt") {
			v, _ := cmd.Flags().GetString("fmt")
			cfg.Fmt = pkg.Formatter(v)
		}
		if cmd.Flags().Changed("linter") {
			v, _ := cmd.Flags().GetString("linter")
			cfg.Linter = pkg.Linter(v)
		}
		if cmd.Flags().Changed("pm") {
			v, _ := cmd.Flags().GetString("pm")
			cfg.PM = pkg.PackageManager(v)
		}
		if cmd.Flags().Changed("validation") {
			v, _ := cmd.Flags().GetString("validation")
			cfg.Validation = pkg.ValidationLib(v)
		}
		if cmd.Flags().Changed("form") {
			v, _ := cmd.Flags().GetString("form")
			cfg.Form = pkg.FormLib(v)
		}
		if cmd.Flags().Changed("query") {
			v, _ := cmd.Flags().GetString("query")
			cfg.Query = pkg.QueryLib(v)
		}
		if cmd.Flags().Changed("state") {
			v, _ := cmd.Flags().GetString("state")
			cfg.State = pkg.StateLib(v)
		}
		if cmd.Flags().Changed("cms") {
			v, _ := cmd.Flags().GetString("cms")
			cfg.CMS = pkg.CMS(v)
		}
		if cmd.Flags().Changed("deploy") {
			v, _ := cmd.Flags().GetString("deploy")
			cfg.Deployment = pkg.DeployTarget(v)
		}
		if cmd.Flags().Changed("test") {
			v, _ := cmd.Flags().GetString("test")
			cfg.Test = pkg.TestingFramework(v)
		}
		if cmd.Flags().Changed("audit") {
			v, _ := cmd.Flags().GetString("audit")
			cfg.Audit = pkg.AuditTool(v)
		}

		if !cfg.Base.IsValid() {
			return fmt.Errorf("invalid base framework: %s", cfg.Base)
		}
		if !cfg.CSS.IsValid() {
			return fmt.Errorf("invalid css framework: %s", cfg.CSS)
		}
		if !cfg.Fmt.IsValid() {
			return fmt.Errorf("invalid formatter: %s", cfg.Fmt)
		}
		if !cfg.Linter.IsValid() {
			return fmt.Errorf("invalid linter: %s", cfg.Linter)
		}
		if !cfg.PM.IsValid() {
			return fmt.Errorf("invalid package manager: %s", cfg.PM)
		}
		if !cfg.Validation.IsValid() {
			return fmt.Errorf("invalid validation library: %s", cfg.Validation)
		}
		if !cfg.Form.IsValid() {
			return fmt.Errorf("invalid form library: %s", cfg.Form)
		}
		if !cfg.Query.IsValid() {
			return fmt.Errorf("invalid query library: %s", cfg.Query)
		}
		if !cfg.State.IsValid() {
			return fmt.Errorf("invalid state library: %s", cfg.State)
		}
		if !cfg.Deployment.IsValid() {
			return fmt.Errorf("invalid deployment target: %s", cfg.Deployment)
		}

		if cfg.Form != "none" && !cfg.Form.IsValidIntegration(string(cfg.Base)) {
			tui.PrintSkippedIntegration(string(cfg.Form), string(cfg.Base))
			cfg.Form = "none"
		}
		if cfg.Query != "none" && !cfg.Query.IsValidIntegration(string(cfg.Base)) {
			tui.PrintSkippedIntegration(string(cfg.Query), string(cfg.Base))
			cfg.Query = "none"
		}
		if cfg.State != "none" && !cfg.State.IsValidIntegration(string(cfg.Base)) {
			tui.PrintSkippedIntegration(string(cfg.State), string(cfg.Base))
			cfg.State = "none"
		}

		if err := pkg.Scaffold(cfg.ProjectName, config.Templates, cfg); err != nil {
			return fmt.Errorf("scaffold failed: %w", err)
		}

		tui.PrintSuccess(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("base", "astro", "Base framework (astro, vite)")
	createCmd.Flags().String("css", "vanilla", "CSS framework (vanilla, tailwindcss)")
	createCmd.Flags().String("fmt", "biome", "Formatter (prettier, biome, oxfmt)")
	createCmd.Flags().String("linter", "biome", "Linter (biome, eslint, oxlint)")
	createCmd.Flags().String("validation", "none", "Validation library (none, zod)")
	createCmd.Flags().String("form", "none", "Form library (none, tanstack-form)")
	createCmd.Flags().String("query", "none", "Query library (none, tanstack-query)")
	createCmd.Flags().String("state", "none", "State management library (none, jotai, zustand, pinia, nanostores)")
	createCmd.Flags().String("pm", "pnpm", "Package manager (bun, npm, yarn, pnpm)")
	createCmd.Flags().String("cms", "none", "CMS (none, microcms)")
	createCmd.Flags().String("test", "none", "Testing library (none, playwright)")
	createCmd.Flags().String("audit", "none", "Audit / performance tool (none, lhci)")
	createCmd.Flags().StringP("template", "t", "", "Predefined template (astro, astro-react, astro-vue, nuxt, vite, vite-react, vite-vue)")
	createCmd.Flags().String("deploy", "none", "Deployment target (none, cloudflare-pages, cloudflare-workers)")
}
