package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	configdata "github.com/spencer-osbrjp/bungkus-cli/config"
)

type Setup struct {
	Base     string      `json:"base"`
	Packages []string    `json:"packages"`
	NPM      NPMConfig   `json:"npm"`
	Config   ConfigFile  `json:"config"`
	Files    []FileEntry `json:"files"`
}

type NPMConfig struct {
	Scripts   map[string]string `json:"scripts"`
	Overrides map[string]string `json:"overrides,omitempty"`
}

type ConfigFile struct {
	Path     string `json:"path"`
	Template string `json:"template"`
}

type FileEntry struct {
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	Executable bool   `json:"executable,omitempty"`
}

type Extra struct {
	Extra    string               `json:"extra"`
	Template string               `json:"template,omitempty"`
	Base     map[string]ExtraBase `json:"base"`
}

type ExtraBase struct {
	Packages  []string          `json:"packages"`
	Imports   []string          `json:"imports"`
	Plugins   []string          `json:"plugins"`
	Spreads   []string          `json:"spreads"`
	JsonMerge map[string]any    `json:"jsonMerge,omitempty"`
	Scripts   map[string]string `json:"scripts,omitempty"`
	Files     []FileEntry       `json:"files"`
}

func LoadSetup(base string) (*Setup, error) {
	data, err := configdata.FS.ReadFile(filepath.Join("bases", base+".json"))
	if err != nil {
		return nil, fmt.Errorf("unknown base %q: %w", base, err)
	}
	var s Setup
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadTemplateFile reads a file from templates/{src}.
func LoadTemplateFile(src string) ([]byte, error) {
	fullPath := filepath.Join("templates", src)
	data, err := configdata.FS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", src, err)
	}
	return data, nil
}

// LoadTemplate reads a template file and returns content + filename.
func LoadTemplate(templatePath string) ([]byte, string, error) {
	data, err := LoadTemplateFile(templatePath)
	if err != nil {
		return nil, "", err
	}
	filename := filepath.Base(templatePath)
	return data, filename, nil
}

func LoadExtra(name string) (*Extra, error) {
	data, err := configdata.FS.ReadFile(filepath.Join("extras", name+".json"))
	if err != nil {
		return nil, fmt.Errorf("unknown extra %q: %w", name, err)
	}
	var e Extra
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
