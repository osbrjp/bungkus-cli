package tui

import (
	"fmt"
	"strings"
)

func confirmView(path string, panels [4]Panel, width int) string {
	var b strings.Builder

	title := confirmTitleStyle.Render("Confirm your selections")
	b.WriteString(title)
	b.WriteString("\n\n")

	rows := []struct{ label, value string }{
		{"Path", path},
		{"Base", panels[0].Selected()},
		{"CSS", panels[1].Selected()},
		{"Formatter", panels[2].Selected()},
		{"Linter", panels[3].Selected()},
	}

	for _, r := range rows {
		label := confirmLabelStyle.Render(fmt.Sprintf("  %-12s", r.label))
		value := confirmValueStyle.Render(r.value)
		b.WriteString(label + value + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Enter: confirm  |  Esc: go back"))

	return focusedBorder.Width(width - 2).Render(b.String())
}
