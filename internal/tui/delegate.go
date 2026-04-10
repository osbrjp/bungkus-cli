package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// field represents a single selectable config field.
type field struct {
	label   string
	options []option
	cursor  int
}

// option represents a single selectable value within a field.
// When isCategory is true, the option acts as a non-selectable group header.
type option struct {
	label      string
	value      string
	isCategory bool
	disabled   bool
}

// Title, Description, FilterValue implement list.DefaultItem for use with bubbles list.
func (o option) Title() string       { return o.label }
func (o option) Description() string { return "" }
func (o option) FilterValue() string { return o.label }

// wizardDelegate is a custom list delegate that renders category headers
// differently from selectable items.
type wizardDelegate struct {
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	categoryStyle lipgloss.Style
	disabledStyle lipgloss.Style
}

func (d wizardDelegate) Height() int                             { return 1 }
func (d wizardDelegate) Spacing() int                            { return 0 }
func (d wizardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d wizardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	opt, ok := item.(option)
	if !ok {
		return
	}

	if opt.isCategory {
		fmt.Fprint(w, d.categoryStyle.Render(opt.label))
		return
	}

	if opt.disabled {
		fmt.Fprint(w, d.disabledStyle.Render("   "+opt.label))
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, d.selectedStyle.Render("  "+opt.label))
	} else {
		fmt.Fprint(w, d.normalStyle.Render("   "+opt.label))
	}
}
