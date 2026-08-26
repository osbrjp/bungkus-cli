package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/pkg"
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

func TestTabSwitch(t *testing.T) {
	if err := pkg.InitRegistry(config.RegistryJSON); err != nil {
		t.Fatal(err)
	}
	m := NewWizardModel() // focus starts on the project name field

	press := func(m tea.Model, k string) tea.Model {
		next, _ := m.Update(tea.KeyPressMsg{Code: rune(k[0])})
		return next
	}

	fe := press(m, "]").(WizardModel)
	if fe.screen != screenBackend {
		t.Fatalf("] from the name field should open the backend tab, got screen %d", fe.screen)
	}
	if got := press(fe, "[").(WizardModel).screen; got != screenWizard {
		t.Fatalf("[ should return to the frontend tab, got screen %d", got)
	}
}
