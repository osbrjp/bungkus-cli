package pkg

import "encoding/json"

// Registry holds all option metadata loaded from config/registry.json.
// This is the single source of truth — adding a new framework only
// requires a JSON entry and template files, no Go code changes.

type BaseEntry struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Category    string `json:"category"`
	TemplateDir string `json:"templateDir"`
	StylesDir   string `json:"stylesDir"`
	Group       string `json:"group"`
	Integration string `json:"integration,omitempty"`
}

type OptionEntry struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type PMEntry struct {
	Value      string `json:"value"`
	Label      string `json:"label"`
	InstallCmd string `json:"installCmd"`
	ExecCmd    string `json:"execCmd"`
	RunCmd     string `json:"runCmd"`
}

type Registry struct {
	Bases           []BaseEntry   `json:"bases"`
	CSS             []OptionEntry `json:"css"`
	Formatters      []OptionEntry `json:"formatters"`
	PackageManagers []PMEntry     `json:"packageManagers"`
}

var globalRegistry *Registry

func InitRegistry(data []byte) error {
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	globalRegistry = &r
	return nil
}

func GetRegistry() *Registry {
	return globalRegistry
}

func (r *Registry) GetBase(value string) *BaseEntry {
	for i := range r.Bases {
		if r.Bases[i].Value == value {
			return &r.Bases[i]
		}
	}
	return nil
}

func (r *Registry) HasBase(value string) bool {
	return r.GetBase(value) != nil
}

func (r *Registry) HasCSS(value string) bool {
	for _, e := range r.CSS {
		if e.Value == value {
			return true
		}
	}
	return false
}

func (r *Registry) HasFormatter(value string) bool {
	for _, e := range r.Formatters {
		if e.Value == value {
			return true
		}
	}
	return false
}

func (r *Registry) GetPM(value string) *PMEntry {
	for i := range r.PackageManagers {
		if r.PackageManagers[i].Value == value {
			return &r.PackageManagers[i]
		}
	}
	return nil
}

func (r *Registry) HasPM(value string) bool {
	return r.GetPM(value) != nil
}
