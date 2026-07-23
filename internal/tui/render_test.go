package tui

import (
	"strings"
	"testing"

	"github.com/spencer-osbrjp/bungkus-cli/config"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

func TestAdvancedDropdown(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatal(err)
	}
	m := NewWizardModel()

	// Collapsed by default: the header row is present, the settings are not.
	out := m.View().Content
	if !strings.Contains(out, "Advanced options") {
		t.Fatal("advanced dropdown header missing from wizard")
	}
	if strings.Contains(out, "Channel:") {
		t.Error("settings should be hidden while the dropdown is collapsed")
	}

	// Expanded: every setting shows inline.
	m.advanced.expanded = true
	out = m.View().Content
	for _, want := range []string{"Channel:", "Pin:", "Install:", "Git init:", "Node:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expanded dropdown missing %q", want)
		}
	}
}
