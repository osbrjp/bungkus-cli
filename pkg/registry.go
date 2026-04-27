package pkg

import (
	"encoding/json"
	"slices"
)

// Registry holds all option metadata loaded from config/registry.json.
// This is the single source of truth — adding a new framework only
// requires a JSON entry and template files, no Go code changes.

type Packages struct {
	Scripts         map[string]string `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

type BaseEntry struct {
	Value       string   `json:"value"`
	Label       string   `json:"label"`
	TemplateDir string   `json:"templateDir"`
	StylesDir   string   `json:"stylesDir"`
	Group       string   `json:"group"`
	Integration string   `json:"integration,omitempty"`
	EntryPoint  string   `json:"entryPoint,omitempty"`
	Private     bool     `json:"private,omitempty"`
	Packages    Packages `json:"packages"`
}

type OptionEntry struct {
	Value               string              `json:"value"`
	Label               string              `json:"label"`
	ExcludeGroups       []string            `json:"excludeGroups,omitempty"`
	RequiresIntegration []string            `json:"requiresIntegration,omitempty"`
	Packages            Packages            `json:"packages"`
	IntegrationPackages map[string]Packages `json:"integrationPackages,omitempty"`
}

func (o *OptionEntry) ExcludesGroup(group string) bool {
	return slices.Contains(o.ExcludeGroups, group)
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
	Linters         []OptionEntry `json:"linters"`
	Validation      []OptionEntry `json:"validation"`
	Form            []OptionEntry `json:"form"`
	Query           []OptionEntry `json:"query"`
	State           []OptionEntry `json:"state"`
	CMS             []OptionEntry `json:"cms"`
	PackageManagers []PMEntry     `json:"packageManagers"`
	CommonPackages  Packages      `json:"commonPackages"`
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

func (r *Registry) HasLinter(value string) bool {
	for _, e := range r.Linters {
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

func (r *Registry) GetCSS(value string) *OptionEntry {
	for i := range r.CSS {
		if r.CSS[i].Value == value {
			return &r.CSS[i]
		}
	}
	return nil
}

func (r *Registry) GetFormatter(value string) *OptionEntry {
	for i := range r.Formatters {
		if r.Formatters[i].Value == value {
			return &r.Formatters[i]
		}
	}
	return nil
}

func (r *Registry) GetLinter(value string) *OptionEntry {
	for i := range r.Linters {
		if r.Linters[i].Value == value {
			return &r.Linters[i]
		}
	}
	return nil
}

func (r *Registry) GetCMS(value string) *OptionEntry {
	for i := range r.CMS {
		if r.CMS[i].Value == value {
			return &r.CMS[i]
		}
	}
	return nil
}

func (r *Registry) HasCMS(value string) bool {
	return r.GetCMS(value) != nil
}

func (r *Registry) GetValidation(value string) *OptionEntry {
	for i := range r.Validation {
		if r.Validation[i].Value == value {
			return &r.Validation[i]
		}
	}
	return nil
}

func (r *Registry) HasValidation(value string) bool {
	return r.GetValidation(value) != nil
}

func (r *Registry) GetForm(value string) *OptionEntry {
	for i := range r.Form {
		if r.Form[i].Value == value {
			return &r.Form[i]
		}
	}
	return nil
}

func (r *Registry) HasForm(value string) bool {
	return r.GetForm(value) != nil
}

func (r *Registry) GetQuery(value string) *OptionEntry {
	for i := range r.Query {
		if r.Query[i].Value == value {
			return &r.Query[i]
		}
	}
	return nil
}

func (r *Registry) HasQuery(value string) bool {
	return r.GetQuery(value) != nil
}

func (r *Registry) GetState(value string) *OptionEntry {
	for i := range r.State {
		if r.State[i].Value == value {
			return &r.State[i]
		}
	}
	return nil
}

func (r *Registry) HasState(value string) bool {
	return r.GetState(value) != nil
}
