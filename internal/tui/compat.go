package tui

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// updateCompat marks formatter and linter options as disabled/enabled based on:
// 1. Base framework group exclusions (e.g. oxfmt/oxlint excluded for astro)
// 2. Cross-field: biome formatter → only biome linter, and vice versa
func (m wizardModel) updateCompat() wizardModel {
	reg := pkg.GetRegistry()
	base := m.selectedValue(0)
	baseEntry := reg.GetBase(base.value)
	if baseEntry == nil {
		return m
	}

	selectedFmt := m.selectedValue(int(focusFmt) - 1)
	selectedLinter := m.selectedValue(int(focusLinter) - 1)

	// Update formatter list
	fmtIdx := int(focusFmt) - 1
	var fmtItems []list.Item
	for _, f := range reg.Formatters {
		disabled := f.ExcludesGroup(baseEntry.Group)
		if selectedLinter.value == "biome" && f.Value != "biome" {
			disabled = true
		}
		fmtItems = append(fmtItems, option{label: f.Label, value: f.Value, disabled: disabled})
	}
	m.lists[fmtIdx].SetItems(fmtItems)
	m = m.fixSelection(fmtIdx, fmtItems)

	// Update linter list
	linterIdx := int(focusLinter) - 1
	var linterItems []list.Item
	for _, l := range reg.Linters {
		disabled := l.ExcludesGroup(baseEntry.Group)
		if selectedFmt.value == "biome" && l.Value != "biome" {
			disabled = true
		}
		linterItems = append(linterItems, option{label: l.Label, value: l.Value, disabled: disabled})
	}
	m.lists[linterIdx].SetItems(linterItems)
	m = m.fixSelection(linterIdx, linterItems)

	return m
}

// fixSelection moves the cursor to the first enabled item if the current selection is disabled.
func (m wizardModel) fixSelection(listIdx int, items []list.Item) wizardModel {
	if sel, ok := m.lists[listIdx].SelectedItem().(option); ok && sel.disabled {
		for i, item := range items {
			if opt, ok := item.(option); ok && !opt.disabled {
				m.lists[listIdx].Select(i)
				break
			}
		}
	}
	return m
}

// buildConfig assembles a ProjectConfig from the current wizard selections.
func (m wizardModel) buildConfig() pkg.ProjectConfig {
	name := m.projectName()
	return pkg.ProjectConfig{
		ProjectName: name,
		Base:        pkg.BaseFramework(m.selectedValue(0).value),
		CSS:         pkg.CSSFramework(m.selectedValue(1).value),
		Fmt:         pkg.Formatter(m.selectedValue(2).value),
		Linter:      pkg.Linter(m.selectedValue(3).value),
		PM:          pkg.PackageManager(m.selectedValue(4).value),
		NoGit:       m.selectedValue(5).value == "no",
	}
}

// projectName returns the entered name or the placeholder default.
func (m wizardModel) projectName() string {
	name := strings.TrimSpace(m.textInput.Value())
	if name == "" {
		name = "my-app"
	}
	return name
}

// selectedValue returns the selected option's value for a given list index.
// Skips category headers and returns the first selectable option as fallback.
func (m wizardModel) selectedValue(listIdx int) option {
	item := m.lists[listIdx].SelectedItem()
	if opt, ok := item.(option); ok && !opt.isCategory {
		return opt
	}
	for _, o := range m.fields[listIdx].options {
		if !o.isCategory {
			return o
		}
	}
	return m.fields[listIdx].options[0]
}

// summaryView renders a confirmation screen with all selections and install instructions.
func (m wizardModel) summaryView() string {
	var b strings.Builder

	name := m.projectName()
	base := m.selectedValue(0)
	css := m.selectedValue(1)
	fmtOpt := m.selectedValue(2)
	linter := m.selectedValue(3)
	pm := m.selectedValue(4)
	git := m.selectedValue(5)

	label := MutedStyle.Width(20)
	value := ActiveStyle

	b.WriteString(TitleStyle.Render("bungkus-cli") + "\n\n")
	b.WriteString(AccentStyle.Render("  Summary") + "\n\n")

	rows := []struct{ l, v string }{
		{"Project Name", name},
		{"Base", base.label},
		{"CSS", css.label},
		{"Formatter", fmtOpt.label},
		{"Linter", linter.label},
		{"Package Manager", pm.label},
		{"Git", git.label},
	}
	for _, r := range rows {
		b.WriteString("  " + label.Render(r.l) + value.Render(r.v) + "\n")
	}

	b.WriteString("\n" + HintStyle.Render("  enter scaffold • esc go back"))

	return b.String()
}

func (m wizardModel) headerView() string {
	art := `88""Yb 88   88 88b 88  dP""b8 88  dP 88   88 .dP"Y8      dP""b8 88     88
88__dP 88   88 88Yb88 dP   ` + "`" + `" 88odP  88   88 ` + "`" + `Ybo."     dP   ` + "`" + `" 88     88
88""Yb Y8   8P 88 Y88 Yb  "88 88"Yb  Y8   8P o.` + "`" + `Y8b     Yb      88  .o 88
88oodP ` + "`" + `YbodP' 88  Y8  YboodP 88  Yb ` + "`" + `YbodP' 8bodP'      YboodP 88ood8 88`
	return AccentStyle.Margin(1, 0).Render(art) + "\n"
}

func (m wizardModel) footerView() string {
	return "\n" + MutedStyle.Render("  tab/shift+tab navigate  ↑/↓ select  enter confirm  esc quit")
}

// setUpView renders the setup fields as a 3-column grid of bordered lists.
func (m wizardModel) setUpView() string {
	colWidth := max(m.width/3, 20)

	var boxes []string
	for i := range m.fields {
		focused := m.focus == uint(i+1)

		borderColor := ColorMuted
		if focused {
			borderColor = ColorAccent
		}

		style := lipgloss.NewStyle().
			Width(colWidth - 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

		boxes = append(boxes, style.Render(m.lists[i].View()))
	}

	var rows []string
	for i := 0; i < len(boxes); i += 3 {
		end := min(i+3, len(boxes))
		row := lipgloss.JoinHorizontal(lipgloss.Top, boxes[i:end]...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
