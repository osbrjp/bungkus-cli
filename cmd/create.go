package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	Long: `Create a new frontend project.

project-name must be a valid npm package name: lowercase letters, digits,
'.', '-', '_', starting with a letter or digit (max 214 chars). It is used
verbatim as the destination directory and the package.json "name", so paths
that escape the current directory (absolute paths, "..") are rejected. Use "."
to scaffold into the current directory.`,
	Args: cobra.MaximumNArgs(1),
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
			if args[0] == "." {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current working directory: %w", err)
				}
				cfg.ProjectName = filepath.Base(cwd)
				cfg.DestDir = "."
			} else {
				if err := pkg.ValidateProjectName(args[0]); err != nil {
					return err
				}
				cfg.ProjectName = args[0]
			}
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
		if cmd.Flags().Changed("cicd") {
			v, _ := cmd.Flags().GetString("cicd")
			cfg.CICD = pkg.CICDProvider(v)
		}
		if cmd.Flags().Changed("test") {
			v, _ := cmd.Flags().GetString("test")
			cfg.Test = pkg.TestingFramework(v)
		}
		if cmd.Flags().Changed("audit") {
			v, _ := cmd.Flags().GetString("audit")
			cfg.Audit = pkg.AuditTool(v)
		}
		if cmd.Flags().Changed("backend") {
			v, _ := cmd.Flags().GetString("backend")
			cfg.Backend = pkg.BackendLib(v)
		}
		if cmd.Flags().Changed("orm") {
			v, _ := cmd.Flags().GetString("orm")
			cfg.ORM = pkg.ORMLib(v)
		}
		if cmd.Flags().Changed("db") {
			v, _ := cmd.Flags().GetString("db")
			cfg.Database = pkg.Database(v)
		}
		if cmd.Flags().Changed("layout") {
			v, _ := cmd.Flags().GetString("layout")
			cfg.Layout = pkg.Layout(v)
		}
		if cmd.Flags().Changed("channel") {
			v, _ := cmd.Flags().GetString("channel")
			cfg.Channel = pkg.VersionChannel(v)
		}
		if cmd.Flags().Changed("pin") {
			v, _ := cmd.Flags().GetString("pin")
			cfg.Pin = pkg.PinStrategy(v)
		}
		if cmd.Flags().Changed("install") {
			cfg.Install, _ = cmd.Flags().GetBool("install")
		}
		if cmd.Flags().Changed("git") {
			cfg.GitInit, _ = cmd.Flags().GetBool("git")
		}
		if cmd.Flags().Changed("node-engine") {
			cfg.NodeEngine, _ = cmd.Flags().GetString("node-engine")
		}

		// A selected backend defaults to the monorepo layout unless the user
		// chose one explicitly (and only when pnpm, which monorepo requires).
		if !cmd.Flags().Changed("layout") {
			cfg.ApplyDefaultLayout()
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
		if !cfg.CICD.IsValid() {
			return fmt.Errorf("invalid cicd provider: %s", cfg.CICD)
		}
		if !cfg.Channel.IsValid() {
			return fmt.Errorf("invalid version channel: %s (pinned, latest)", cfg.Channel)
		}
		if !cfg.Pin.IsValid() {
			return fmt.Errorf("invalid pin strategy: %s (default, caret, tilde, exact)", cfg.Pin)
		}
		if cfg.CICD != "none" && cfg.Deployment == "none" {
			return fmt.Errorf("--cicd requires a deploy target (--deploy cloudflare-pages or --deploy cloudflare-workers)")
		}
		if !cfg.Backend.IsValid() {
			return fmt.Errorf("invalid backend: %s (none, hono, elysia)", cfg.Backend)
		}
		if !cfg.ORM.IsValid() {
			return fmt.Errorf("invalid orm: %s (none, drizzle, prisma)", cfg.ORM)
		}
		if !cfg.Database.IsValid() {
			return fmt.Errorf("invalid database: %s (none, sqlite, postgres, mysql, d1)", cfg.Database)
		}
		if cfg.Database != "none" && cfg.ORM == "none" {
			return fmt.Errorf("--db requires an --orm (drizzle or prisma)")
		}
		if cfg.Database == "d1" && cfg.ORM == "prisma" {
			return fmt.Errorf("--db d1 is only supported with --orm drizzle")
		}
		if !cfg.Layout.IsValid() {
			return fmt.Errorf("invalid layout: %s (flat, monorepo)", cfg.Layout)
		}
		if cfg.Layout.IsMonorepo() && cfg.PM != "pnpm" {
			return fmt.Errorf("--layout monorepo currently requires --pm pnpm")
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
		destDir := cfg.ProjectName
		if cfg.DestDir != "" {
			destDir = cfg.DestDir
		}
		if err := pkg.ValidateDest(destDir); err != nil {
			return err
		}

		if err := pkg.Scaffold(destDir, config.Templates, cfg); err != nil {
			return fmt.Errorf("scaffold failed: %w", err)
		}

		if err := pkg.PostScaffold(destDir, cfg); err != nil {
			return err
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
	createCmd.Flags().String("cicd", "none", "CI/CD provider (none, github-actions)")
	createCmd.Flags().String("backend", "none", "Backend framework (none, hono, elysia)")
	createCmd.Flags().String("orm", "none", "ORM / database toolkit (none, drizzle, prisma)")
	createCmd.Flags().String("db", "none", "Database, requires --orm (none, sqlite, postgres, mysql, d1). d1 needs --orm drizzle")
	createCmd.Flags().String("layout", "flat", "Project layout (flat, monorepo). Defaults to monorepo when --backend is set with pnpm; monorepo splits apps/web + apps/api + packages/domain")
	createCmd.Flags().String("channel", "pinned", "Dependency version channel: pinned (vetted, >=14d old & safe) or latest")
	createCmd.Flags().String("pin", "default", "Pin strategy: default (as registry), caret, tilde, exact")
	createCmd.Flags().Bool("install", false, "Run the package manager install after scaffolding")
	createCmd.Flags().Bool("git", true, "Initialize a git repo with an initial commit")
	createCmd.Flags().String("node-engine", pkg.DefaultNodeEngine, "package.json engines.node constraint")
}
