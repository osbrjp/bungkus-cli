package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Scaffold creates the project directory and renders templates into it.
func Scaffold(destDir string, templates fs.FS, cfg ProjectConfig) error {
	entry := globalRegistry.GetBase(string(cfg.Base))
	if entry == nil {
		return fmt.Errorf("unknown base framework: %s", cfg.Base)
	}

	baseDir := "templates/base/" + entry.TemplateDir

	baseFS, err := fs.Sub(templates, baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base templates: %w", err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Layout-aware roots. Flat keeps everything at destDir. Monorepo splits the
	// frontend into apps/web and the backend into apps/api, leaving shared
	// tooling and the workspace manifest at destDir (the root).
	webDir := destDir
	apiDir := destDir
	if cfg.Layout.IsMonorepo() {
		webDir = filepath.Join(destDir, "apps", "web")
		apiDir = filepath.Join(destDir, "apps", "api")
		if err := os.MkdirAll(webDir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", webDir, err)
		}
	}

	// Frontend package.json (in monorepo mode this omits backend/orm deps)
	pkgJSON, err := BuildPackageJSON(cfg)
	if err != nil {
		return fmt.Errorf("failed to build package.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(webDir, "package.json"), pkgJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	// Copy base templates, skipping entry point when integration provides its own
	var skip string
	if entry.Group == "vite" && entry.EntryPoint != "" {
		skip = "main.ts"
	}
	if err := copyDir(baseFS, webDir, cfg, skip); err != nil {
		return err
	}

	// Copy integration templates (e.g. astro-react → integration/astro/react)
	if entry.Integration != "" {
		integrationDir := "templates/integration/" + entry.Group + "/" + entry.Integration

		integrationFS, err := fs.Sub(templates, integrationDir)
		if err != nil {
			return fmt.Errorf("failed to read %v integration templates: %w", entry.Integration, err)
		}
		if err := copyDir(integrationFS, webDir, cfg, ""); err != nil {
			return err
		}
	}

	// Copy CSS templates into the framework's styles directory
	cssDir := "templates/css/" + string(cfg.CSS)
	cssFS, err := fs.Sub(templates, cssDir)
	if err != nil {
		return fmt.Errorf("failed to read css templates: %w", err)
	}

	stylesDir := filepath.Join(webDir, entry.StylesDir)
	if err := copyDir(cssFS, stylesDir, cfg, ""); err != nil {
		return err
	}

	// Copy formatter templates
	fmtDir := "templates/fmt/" + string(cfg.Fmt)
	fmtFS, err := fs.Sub(templates, fmtDir)
	if err != nil {
		return fmt.Errorf("failed to read formatter templates: %w", err)
	}

	if err := copyDir(fmtFS, webDir, cfg, ""); err != nil {
		return err
	}

	// Copy linter templates (skip if same tool as formatter, e.g. biome)
	if string(cfg.Linter) != string(cfg.Fmt) {
		linterDir := "templates/linter/" + string(cfg.Linter)
		linterFS, err := fs.Sub(templates, linterDir)
		if err != nil {
			return fmt.Errorf("failed to read linter templates: %w", err)
		}
		if err := copyDir(linterFS, webDir, cfg, ""); err != nil {
			return err
		}
	}

	// Copy Package Manager templates to the root (pnpm-workspace.yaml, .npmrc)
	pmDir := "templates/pm/" + string(cfg.PM)
	if _, err := fs.Stat(templates, pmDir); err == nil {
		pmFS, err := fs.Sub(templates, pmDir)
		if err != nil {
			return fmt.Errorf("failed to read package manager templates: %w", err)
		}
		if err := copyDir(pmFS, destDir, cfg, ""); err != nil {
			return err
		}
	}

	if cfg.Test != "none" {
		testDir := "templates/test/" + string(cfg.Test)
		if _, err := fs.Stat(templates, testDir); err == nil {
			testFS, err := fs.Sub(templates, testDir)
			if err != nil {
				return fmt.Errorf("failed to read test templates: %w", err)
			}
			if err := copyDir(testFS, webDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	if cfg.Audit != "none" {
		auditDir := "templates/audit/" + string(cfg.Audit)
		if _, err := fs.Stat(templates, auditDir); err == nil {
			auditFS, err := fs.Sub(templates, auditDir)
			if err != nil {
				return fmt.Errorf("failed to read audit templates: %w", err)
			}
			if err := copyDir(auditFS, webDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	if cfg.CMS != "none" {
		if cfg.Base.IsVite() {
			return fmt.Errorf("cms integration is currently not supported in vite project")
		}

		group, err := cfg.Base.GetGroup(string(cfg.Base))
		if err != nil {
			return err
		}

		cmsDir := "templates/cms/" + string(group)
		cmsFS, err := fs.Sub(templates, cmsDir)
		if err != nil {
			return fmt.Errorf("failed to read CMS templates: %w", err)
		}
		if err := copyDir(cmsFS, webDir, cfg, ""); err != nil {
			return err
		}
	}

	if cfg.Deployment != "none" {
		deployDir := "templates/deploy/" + string(cfg.Deployment)
		if _, err := fs.Stat(templates, deployDir); err == nil {
			deployFS, err := fs.Sub(templates, deployDir)
			if err != nil {
				return fmt.Errorf("failed to read deployment templates: %w", err)
			}
			if err := copyDir(deployFS, webDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	if cfg.CICD != "none" && cfg.Deployment != "none" {
		cicdDir := "templates/cicd/" + string(cfg.CICD) + "/" + string(cfg.Deployment)
		if _, err := fs.Stat(templates, cicdDir); err == nil {
			cicdFS, err := fs.Sub(templates, cicdDir)
			if err != nil {
				return fmt.Errorf("failed to read cicd templates: %w", err)
			}
			if err := copyDir(cicdFS, webDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	if cfg.Backend != "none" {
		backendDir := "templates/backend/" + string(cfg.Backend)
		if _, err := fs.Stat(templates, backendDir); err == nil {
			backendFS, err := fs.Sub(templates, backendDir)
			if err != nil {
				return fmt.Errorf("failed to read backend templates: %w", err)
			}
			if err := copyDir(backendFS, apiDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	if cfg.ORM != "none" {
		ormDir := "templates/orm/" + string(cfg.ORM)
		if _, err := fs.Stat(templates, ormDir); err == nil {
			ormFS, err := fs.Sub(templates, ormDir)
			if err != nil {
				return fmt.Errorf("failed to read orm templates: %w", err)
			}
			if err := copyDir(ormFS, apiDir, cfg, ""); err != nil {
				return err
			}
		}
	}

	// Copy shared templates (husky, etc.) to the root
	sharedFS, err := fs.Sub(templates, "templates/shared")
	if err != nil {
		return fmt.Errorf("failed to read shared templates: %w", err)
	}
	if err := copyDir(sharedFS, destDir, cfg, ""); err != nil {
		return err
	}

	if cfg.Layout.IsMonorepo() {
		return scaffoldMonorepo(destDir, apiDir, templates, cfg)
	}
	return nil
}

// scaffoldMonorepo writes the workspace-only pieces: the private root
// package.json, the shared packages/domain contract, and (when a backend or orm
// is selected) apps/api's package.json + tsconfig. The web and api source have
// already been rendered by Scaffold into their app directories.
func scaffoldMonorepo(destDir, apiDir string, templates fs.FS, cfg ProjectConfig) error {
	rootPkg, err := BuildRootPackageJSON(cfg)
	if err != nil {
		return fmt.Errorf("failed to build root package.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "package.json"), rootPkg, 0o644); err != nil {
		return fmt.Errorf("failed to write root package.json: %w", err)
	}

	// Root-level files (workspace .gitignore for the hoisted node_modules)
	rootFS, err := fs.Sub(templates, "templates/monorepo/root")
	if err != nil {
		return fmt.Errorf("failed to read monorepo root templates: %w", err)
	}
	if err := copyDir(rootFS, destDir, cfg, ""); err != nil {
		return err
	}

	// Shared domain contract
	domainDir := filepath.Join(destDir, "packages", "domain")
	if err := os.MkdirAll(domainDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", domainDir, err)
	}
	domainPkg, err := BuildDomainPackageJSON(cfg)
	if err != nil {
		return fmt.Errorf("failed to build domain package.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(domainDir, "package.json"), domainPkg, 0o644); err != nil {
		return fmt.Errorf("failed to write domain package.json: %w", err)
	}
	domainFS, err := fs.Sub(templates, "templates/monorepo/domain")
	if err != nil {
		return fmt.Errorf("failed to read domain templates: %w", err)
	}
	if err := copyDir(domainFS, domainDir, cfg, ""); err != nil {
		return err
	}

	// apps/api scaffolding (package.json + tsconfig) when a backend/orm exists
	if cfg.Backend != "none" || cfg.ORM != "none" {
		if err := os.MkdirAll(apiDir, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", apiDir, err)
		}
		apiPkg, err := BuildAPIPackageJSON(cfg)
		if err != nil {
			return fmt.Errorf("failed to build api package.json: %w", err)
		}
		if err := os.WriteFile(filepath.Join(apiDir, "package.json"), apiPkg, 0o644); err != nil {
			return fmt.Errorf("failed to write api package.json: %w", err)
		}
		apiFS, err := fs.Sub(templates, "templates/monorepo/api")
		if err != nil {
			return fmt.Errorf("failed to read api templates: %w", err)
		}
		if err := copyDir(apiFS, apiDir, cfg, ""); err != nil {
			return err
		}
	}
	return nil
}

// PostScaffold runs optional post-generation steps: installing dependencies
// and initializing a git repo with an initial commit. Subprocess output is
// streamed to the terminal. Safe to call unconditionally — each step is gated
// by its config flag. git init is skipped when scaffolding into an existing
// directory (DestDir ".").
func PostScaffold(destDir string, cfg ProjectConfig) error {
	run := func(name string, args ...string) error {
		c := exec.Command(name, args...)
		c.Dir = destDir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if cfg.Install {
		fields := strings.Fields(cfg.PM.InstallCmd())
		if len(fields) > 0 {
			if err := run(fields[0], fields[1:]...); err != nil {
				return fmt.Errorf("dependency install failed: %w", err)
			}
		}
	}

	if cfg.GitInit && cfg.DestDir != "." {
		for _, args := range [][]string{
			{"init"},
			{"add", "."},
			{"commit", "--no-verify", "-m", "initial commit"},
		} {
			if err := run("git", args...); err != nil {
				return fmt.Errorf("git init failed: %w", err)
			}
		}
	}
	return nil
}

func copyDir(srcFS fs.FS, destDir string, cfg ProjectConfig, skip string) error {
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the embedded .git directory
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}

		// Skip package.json.tmpl — package.json is generated programmatically
		if !d.IsDir() && d.Name() == "package.json.tmpl" {
			return nil
		}

		// Skip entry point file when integration provides its own
		if skip != "" && !d.IsDir() {
			resolved := strings.TrimSuffix(path, ".tmpl")
			if filepath.Base(resolved) == skip {
				return nil
			}
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Render .tmpl files, copy everything else as-is
		if strings.HasSuffix(path, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
			return renderTemplate(srcFS, path, destPath, cfg)
		}

		return copyFile(srcFS, path, destPath)
	})
}

func renderTemplate(srcFS fs.FS, srcPath string, destPath string, cfg ProjectConfig) error {
	data, err := fs.ReadFile(srcFS, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	t, err := template.New(filepath.Base(destPath)).Parse(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", destPath, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("failed to render template %s: %w", destPath, err)
	}

	output := buf.Bytes()

	// Format JSON files after rendering (skip if JSONC with comments)
	if strings.HasSuffix(destPath, ".json") {
		var parsed json.RawMessage
		if err := json.Unmarshal(output, &parsed); err == nil {
			var fmtBuf bytes.Buffer
			enc := json.NewEncoder(&fmtBuf)
			enc.SetIndent("", "  ")
			enc.SetEscapeHTML(false)
			if err := enc.Encode(parsed); err == nil {
				output = fmtBuf.Bytes()
			}
		}
	}

	perm := os.FileMode(0o644)
	if strings.Contains(destPath, ".husky") {
		perm = 0o755
	}

	return os.WriteFile(destPath, output, perm)
}

func copyFile(srcFS fs.FS, srcPath string, destPath string) error {
	data, err := fs.ReadFile(srcFS, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	return os.WriteFile(destPath, data, 0o644)
}
