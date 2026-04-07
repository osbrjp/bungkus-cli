package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Scaffold creates the project directory and renders templates into it.
func Scaffold(destDir string, templates fs.FS, cfg ProjectConfig) error {
	baseDir := "templates/base/" + string(cfg.Base)

	baseFS, err := fs.Sub(templates, baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base templates: %w", err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Copy base templates
	if err := copyDir(baseFS, destDir, cfg); err != nil {
		return err
	}

	// Copy CSS templates into src/styles/
	cssDir := "templates/css/" + string(cfg.CSS)
	cssFS, err := fs.Sub(templates, cssDir)
	if err != nil {
		return fmt.Errorf("failed to read css templates: %w", err)
	}

	stylesDir := filepath.Join(destDir, "src", "styles")
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
			return os.MkdirAll(destPath, 0755)
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

	perm := os.FileMode(0644)
	if strings.Contains(destPath, ".husky") {
		perm = 0755
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", destPath, err)
	}
	defer f.Close()

	if err := t.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to render template %s: %w", destPath, err)
	}

	return nil
}

func copyFile(srcFS fs.FS, srcPath string, destPath string) error {
	data, err := fs.ReadFile(srcFS, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	return os.WriteFile(destPath, data, 0644)
}
