package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/spencer-osbrjp/bungkus-cli/internal/config"
	"github.com/spencer-osbrjp/bungkus-cli/internal/patcher"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init [path]",
	Short:   "Setup project boilerplate",
	Example: "bungkus init . --css tailwindcss --fmt prettier",
	Args:    cobra.ExactArgs(1),
	RunE:    runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	defaultBase := "astro"
	defaultCssFramework := "tailwindcss"
	defaultFmt := "prettier"
	defaultLinter := "eslint"

	initCmd.Flags().StringP("base", "b", defaultBase, "Project base framework")
	initCmd.Flags().String("css", defaultCssFramework, "CSS framework (e.g. tailwindcss)")
	initCmd.Flags().String("fmt", defaultFmt, "Formatter (e.g. prettier)")
	initCmd.Flags().String("linter", defaultLinter, "Linter (e.g. eslint)")
}

func runInit(cmd *cobra.Command, args []string) error {
	targetPath := args[0]
	base, _ := cmd.Flags().GetString("base")
	css, _ := cmd.Flags().GetString("css")
	formatter, _ := cmd.Flags().GetString("fmt")
	linter, _ := cmd.Flags().GetString("linter")

	// Resolve target directory
	dir, err := resolveDir(targetPath)
	if err != nil {
		return err
	}

	// Load base setup config
	setup, err := config.LoadSetup()
	if err != nil {
		return fmt.Errorf("failed to load setup config: %w", err)
	}

	// Collect all packages from base
	packages := slices.Clone(setup.Packages)

	// Collect extras to apply
	var extras []extraResult

	// Collect imports and plugins for base config patching
	var allImports []string
	var allPlugins []string

	// Process each flag
	for _, name := range []string{css, formatter, linter} {
		if name == "" {
			continue
		}
		extra, err := config.LoadExtra(name)
		if err != nil {
			return err
		}
		baseConfig, ok := extra.Base[base]
		if !ok {
			return fmt.Errorf("extra %q does not support base %q", name, base)
		}
		packages = append(packages, baseConfig.Packages...)

		// Only collect imports/plugins for base config from extras without templates
		if extra.Template == "" {
			allImports = append(allImports, baseConfig.Imports...)
			allPlugins = append(allPlugins, baseConfig.Plugins...)
		}
		extras = append(extras, extraResult{
			name:     name,
			template: extra.Template,
			base:     baseConfig,
		})
	}

	// 1. npm init -y
	fmt.Println("Initializing package.json...")
	npmInit := exec.Command("npm", "init", "-y")
	npmInit.Dir = dir
	if out, err := npmInit.CombinedOutput(); err != nil {
		return fmt.Errorf("npm init failed: %s", out)
	}

	// Patch package.json with scripts from setup config
	if err := patchPackageJSON(filepath.Join(dir, "package.json"), setup); err != nil {
		return fmt.Errorf("failed to patch package.json: %w", err)
	}
	fmt.Println("Created package.json")

	// 2. Write base config file (e.g. astro.config.mjs)
	if setup.Config.Path != "" {
		configPath := filepath.Join(dir, setup.Config.Path)
		if err := os.WriteFile(configPath, []byte(setup.Config.Template), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", setup.Config.Path, err)
		}
		fmt.Printf("Created %s\n", setup.Config.Path)
	}

	// Copy base scaffold files (e.g. src/pages/index.astro)
	if err := copyTemplateFiles(dir, setup.Files); err != nil {
		return err
	}

	// 3. Call patcher to AST-inject imports and plugins into base config
	if len(allImports) > 0 || len(allPlugins) > 0 {
		configPath := filepath.Join(dir, setup.Config.Path)
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("failed to resolve config path: %w", err)
		}
		if err := patcher.PatchBaseConfig(absConfigPath, allImports, allPlugins); err != nil {
			return fmt.Errorf("patcher failed: %w", err)
		}
	}

	// 4. Copy extra templates, patch them, and copy extra scaffold files
	for _, ec := range extras {
		// Copy and patch config template if present (e.g. .prettierrc, eslint.config.mjs)
		if ec.template != "" {
			data, filename, err := config.LoadTemplate(ec.template)
			if err != nil {
				return err
			}
			cfgPath := filepath.Join(dir, filename)
			if err := os.WriteFile(cfgPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", filename, err)
			}
			fmt.Printf("Created %s\n", filename)

			absCfgPath, err := filepath.Abs(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path for %s: %w", filename, err)
			}
			if err := patcher.PatchExtra(absCfgPath, ec.base); err != nil {
				return fmt.Errorf("patcher failed for %s: %w", filename, err)
			}
		}

		// Copy extra scaffold files (e.g. src/styles/global.css for tailwindcss)
		if err := copyTemplateFiles(dir, ec.base.Files); err != nil {
			return err
		}
	}

	// 5. npm install all packages
	fmt.Println("Installing dependencies...")
	installArgs := append([]string{"install"}, packages...)
	npmInstall := exec.Command("npm", installArgs...)
	npmInstall.Dir = dir
	npmInstall.Stdout = os.Stdout
	npmInstall.Stderr = os.Stderr
	if err := npmInstall.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	fmt.Printf("\nProject ready at %s\n", dir)
	return nil
}

type extraResult struct {
	name     string
	template string
	base     config.ExtraBase
}

func copyTemplateFiles(dir string, files []config.FileEntry) error {
	for _, f := range files {
		data, err := config.LoadTemplateFile(f.Src)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dir, f.Dest)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", f.Dest, err)
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.Dest, err)
		}
		fmt.Printf("Created %s\n", f.Dest)
	}
	return nil
}

func resolveDir(path string) (string, error) {
	if path == "." {
		return ".", nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return path, nil
}

func patchPackageJSON(pkgPath string, setup *config.Setup) error {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	// Merge scripts from setup config
	pkg["scripts"] = setup.NPM.Scripts

	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkgPath, out, 0644)
}


