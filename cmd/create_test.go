package cmd

import (
	"testing"

	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/pkg"
)

// TestTemplatePresetsAreValid runs every preset through the same enum checks
// `create` applies to flags. A typo in a preset (e.g. vite-vue once set
// Linter "prettier") otherwise only surfaces when a user runs -t <name>.
func TestTemplatePresetsAreValid(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatalf("InitRegistry: %v", err)
	}
	for name, factory := range templates {
		c := factory()
		checks := map[string]bool{
			"base":       c.Base.IsValid(),
			"css":        c.CSS.IsValid(),
			"fmt":        c.Fmt.IsValid(),
			"linter":     c.Linter.IsValid(),
			"validation": c.Validation.IsValid(),
			"form":       c.Form.IsValid(),
			"query":      c.Query.IsValid(),
			"state":      c.State.IsValid(),
			"cms":        c.CMS.IsValid(),
			"pm":         c.PM.IsValid(),
		}
		for field, ok := range checks {
			if !ok {
				t.Errorf("template %q: invalid %s", name, field)
			}
		}
	}
}
