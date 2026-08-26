package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <option> [dir]",
	Short: "Add a tool to an existing project.",
	Long: `Add a tool from the bungkus registry to an existing project without
touching your code.

Renders the option's config templates into the project (existing files are
NEVER overwritten — they are skipped and reported) and additively merges the
option's dependencies and scripts into package.json (existing versions and
scripts are never changed; other package.json fields, key order, and the
file's own indent style are preserved).

GitHub Actions workflow files (.github/**) are always written to the git
repository root, even when adding inside a subdirectory such as apps/web.

The package manager is detected from package.json's packageManager field or
a lockfile (searched upward to the git root), and the base framework from
package.json dependencies; use --pm / --base to override. Run "bungkus-cli
add" with no option to list what can be added.`,
	Args: cobra.MaximumNArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("pm", "", "Package manager override (pnpm, bun, npm, yarn); detected otherwise")
	addCmd.Flags().String("base", "", "Base framework override; detected from package.json otherwise")
	addCmd.Flags().String("deploy", "", "Deploy target for CI/CD options (cloudflare-pages, cloudflare-workers)")
}

func printAddable() {
	fmt.Println("Addable options:")
	for _, c := range pkg.AddableOptions() {
		fmt.Printf("  %-8s %s\n", c.Name+":", strings.Join(c.Options, ", "))
	}
	fmt.Println("\ncicd options require --deploy.")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		printAddable()
		return nil
	}
	opt := args[0]
	dir := "."
	if len(args) == 2 {
		dir = args[1]
	}

	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return fmt.Errorf("no package.json found in %s — 'add' works on an existing project (use 'create' to scaffold a new one)", dir)
	}
	cfg, err := pkg.DetectProject(dir, raw)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("pm") {
		v, _ := cmd.Flags().GetString("pm")
		cfg.PM = pkg.PackageManager(v)
	}
	if cmd.Flags().Changed("base") {
		v, _ := cmd.Flags().GetString("base")
		cfg.Base = pkg.BaseFramework(v)
	}
	if cmd.Flags().Changed("deploy") {
		v, _ := cmd.Flags().GetString("deploy")
		cfg.Deployment = pkg.DeployTarget(v)
	}
	if cfg.PM == "" {
		return fmt.Errorf("could not detect the package manager (no or ambiguous lockfile) — pass --pm (pnpm, bun, npm, yarn)")
	}
	if !cfg.PM.IsValid() {
		return fmt.Errorf("invalid package manager: %s (pnpm, bun, npm, yarn)", cfg.PM)
	}
	if cfg.Base == "" {
		return fmt.Errorf("could not detect the base framework from package.json — pass --base (astro, astro-react, astro-vue, nuxt, vite, vite-react, vite-vue)")
	}
	if !cfg.Base.IsValid() {
		return fmt.Errorf("invalid base framework: %s", cfg.Base)
	}
	if cfg.Deployment != "none" && !cfg.Deployment.IsValid() {
		return fmt.Errorf("invalid deploy target: %s (cloudflare-pages, cloudflare-workers)", cfg.Deployment)
	}

	rep, err := pkg.Add(dir, config.Templates, cfg, opt)
	if errors.Is(err, pkg.ErrUnknownAddOption) {
		fmt.Printf("unknown option %q\n\n", opt)
		printAddable()
		return errors.New("nothing added")
	}
	if err != nil {
		return err
	}
	printAddReport(rep, dir, opt, cfg)
	return nil
}

func printAddReport(rep *pkg.AddReport, dir, opt string, cfg pkg.ProjectConfig) {
	fmt.Printf("Added %s (%s) to %s\n\n", opt, strings.Join(rep.Categories, ", "), dir)
	for _, f := range rep.CreatedFiles {
		if strings.HasPrefix(f, "..") {
			fmt.Printf("  created   %s   (repo root)\n", f)
		} else {
			fmt.Printf("  created   %s\n", f)
		}
	}
	for _, f := range rep.SkippedFiles {
		fmt.Printf("  skipped   %s (already exists — kept yours)\n", f)
	}
	if len(rep.DepsAdded)+len(rep.DepsSkipped)+len(rep.ScriptsAdded)+len(rep.ScriptsSkipped) > 0 {
		fmt.Println("\n  package.json:")
		for _, d := range rep.DepsAdded {
			fmt.Printf("    + %s\n", d)
		}
		for _, s := range rep.ScriptsAdded {
			fmt.Printf("    + scripts  %s\n", s)
		}
		for _, d := range rep.DepsSkipped {
			fmt.Printf("    = %s (already present — kept yours)\n", d)
		}
		for _, s := range rep.ScriptsSkipped {
			fmt.Printf("    = scripts  %s (already defined — kept yours)\n", s)
		}
	}

	if rep.WorkflowRelocated {
		fmt.Printf("\n  note: workflow(s) written to %s/.github — review job steps if this is a monorepo.\n", rep.GitRoot)
	}
	if rep.NoGitWarning {
		for _, f := range rep.CreatedFiles {
			if strings.HasPrefix(filepath.ToSlash(f), ".github/") {
				fmt.Printf("\n  warning: no git repository found — .github/ written under %s;\n", dir)
				fmt.Println("  GitHub Actions only reads .github/ at the repo root.")
				break
			}
		}
	}

	if rep.PkgJSONChanged {
		fmt.Printf("\nRun `%s` to install the new dependencies.\n", cfg.PM.InstallCmd())
	} else if len(rep.CreatedFiles) == 0 {
		fmt.Printf("\nNothing to do — everything %s provides already exists.\n", opt)
	}
}
