package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
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

	// Copy base templates
	if err := copyDir(baseFS, destDir, cfg); err != nil {
		return err
	}

	// Copy integration templates (e.g. astro-react → integration/astro/react)
	if entry.Integration != "" {
		integrationDir := "templates/integration/" + entry.Group + "/" + entry.Integration

		integrationFS, err := fs.Sub(templates, integrationDir)
		if err != nil {
			return fmt.Errorf("failed to read %v integration templates: %w", entry.Integration, err)
		}
		if err := copyDir(integrationFS, destDir, cfg); err != nil {
			return err
		}
	}

	// Copy CSS templates into the framework's styles directory
	cssDir := "templates/css/" + string(cfg.CSS)
	cssFS, err := fs.Sub(templates, cssDir)
	if err != nil {
		return fmt.Errorf("failed to read css templates: %w", err)
	}

	stylesDir := filepath.Join(destDir, entry.StylesDir)
	if err := copyDir(cssFS, stylesDir, cfg); err != nil {
		return err
	}

	// Copy formatter templates
	fmtDir := "templates/fmt/" + string(cfg.Fmt)
	fmtFS, err := fs.Sub(templates, fmtDir)
	if err != nil {
		return fmt.Errorf("failed to read formatter templates: %w", err)
	}

	if err := copyDir(fmtFS, destDir, cfg); err != nil {
		return err
	}

	// Copy linter templates (skip if same tool as formatter, e.g. biome)
	if string(cfg.Linter) != string(cfg.Fmt) {
		linterDir := "templates/linter/" + string(cfg.Linter)
		linterFS, err := fs.Sub(templates, linterDir)
		if err != nil {
			return fmt.Errorf("failed to read linter templates: %w", err)
		}
		if err := copyDir(linterFS, destDir, cfg); err != nil {
			return err
		}
	}

	// Copy Package Manager templates (some PMs like bun/npm have no templates)
	pmDir := "templates/pm/" + string(cfg.PM)
	if _, err := fs.Stat(templates, pmDir); err == nil {
		pmFS, err := fs.Sub(templates, pmDir)
		if err != nil {
			return fmt.Errorf("failed to read package manager templates: %w", err)
		}
		if err := copyDir(pmFS, destDir, cfg); err != nil {
			return err
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
		if err := copyDir(cmsFS, destDir, cfg); err != nil {
			return err
		}
	}

	// Copy shared templates (husky, etc.)
	sharedFS, err := fs.Sub(templates, "templates/shared")
	if err != nil {
		return fmt.Errorf("failed to read shared templates: %w", err)
	}

	return copyDir(sharedFS, destDir, cfg)
}

func copyDir(srcFS fs.FS, destDir string, cfg ProjectConfig) error {
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the embedded .git directory
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
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

	// Format JSON files after rendering
	if strings.HasSuffix(destPath, ".json") {
		var parsed json.RawMessage
		if err := json.Unmarshal(output, &parsed); err != nil {
			return fmt.Errorf("failed to parse rendered JSON %s: %w", destPath, err)
		}
		formatted, err := json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON %s: %w", destPath, err)
		}
		output = append(formatted, '\n')
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
