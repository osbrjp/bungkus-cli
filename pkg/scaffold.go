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
	baseDir := "templates/base/astro"

	baseFS, err := fs.Sub(templates, baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base templates: %w", err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	err = fs.WalkDir(baseFS, ".", func(path string, d fs.DirEntry, err error) error {
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

		data, err := fs.ReadFile(baseFS, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Render .tmpl files, copy everything else as-is
		if strings.HasSuffix(path, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
			return renderTemplate(destPath, string(data), cfg)
		}

		return os.WriteFile(destPath, data, 0644)
	})

	return err
}

func renderTemplate(destPath string, tmplContent string, cfg ProjectConfig) error {
	t, err := template.New(filepath.Base(destPath)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", destPath, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", destPath, err)
	}
	defer f.Close()

	if err := t.Execute(f, cfg.CSS); err != nil {
		return fmt.Errorf("failed to render template %s: %w", destPath, err)
	}

	return nil
}
