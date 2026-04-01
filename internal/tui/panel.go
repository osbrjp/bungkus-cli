package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Panel struct {
	Title    string
	Items    []string
	cursor   int
	selected int
}

func NewPanel(title string, items []string, defaultSelected int) Panel {
	return Panel{
		Title:    title,
		Items:    items,
		cursor:   defaultSelected,
		selected: defaultSelected,
	}
}

func (p Panel) Selected() string {
	if p.selected >= 0 && p.selected < len(p.Items) {
		return p.Items[p.selected]
	}
	return ""
}

func (p Panel) Update(msg tea.Msg) Panel {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if p.cursor > 0 {
				p.cursor--
			}
		case key.Matches(msg, keys.Down):
			if p.cursor < len(p.Items)-1 {
				p.cursor++
			}
		case key.Matches(msg, keys.Select):
			p.selected = p.cursor
		}
	}
	return p
}

func (p Panel) View(focused bool, width, height int) string {
	var b strings.Builder

	title := titleStyle.Render(p.Title)
	b.WriteString(title)
	b.WriteString("\n")

	for i, item := range p.Items {
		prefix := "  "
		style := normalStyle

		if i == p.selected {
			prefix = "* "
			style = selectedStyle
		}
		if focused && i == p.cursor {
			prefix = "> "
			if i == p.selected {
				prefix = "* "
			}
			style = cursorStyle
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, item)))
		if i < len(p.Items)-1 {
			b.WriteString("\n")
		}
	}

	// Pad to fill height
	lines := len(p.Items) + 1 // +1 for title
	for i := lines; i < height; i++ {
		b.WriteString("\n")
	}

	content := b.String()

	border := unfocusedBorder
	if focused {
		border = focusedBorder
	}

	return border.
		Width(width).
		Render(content)
}

func maxPanelHeight(panels []Panel) int {
	max := 0
	for _, p := range panels {
		h := len(p.Items) + 1 // items + title
		if h > max {
			max = h
		}
	}
	return max + 1 // padding
}

func renderPanelRow(panels []Panel, focusedIdx int, panelStartIdx int, width int) string {
	panelWidth := width/len(panels) - 2 // account for borders
	height := maxPanelHeight(panels)

	var views []string
	for i, p := range panels {
		focused := (panelStartIdx + i) == focusedIdx
		views = append(views, p.View(focused, panelWidth, height))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}
