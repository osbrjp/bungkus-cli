package tui

import "github.com/charmbracelet/lipgloss"

var (
	focusedColor   = lipgloss.Color("205") // magenta
	unfocusedColor = lipgloss.Color("240") // gray
	selectedColor  = lipgloss.Color("42")  // green
	titleColor     = lipgloss.Color("99")  // purple

	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(focusedColor)

	unfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(unfocusedColor)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(titleColor).
			MarginBottom(1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(focusedColor).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(selectedColor).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(unfocusedColor).
			MarginTop(1)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(selectedColor).
				MarginBottom(1)

	confirmLabelStyle = lipgloss.NewStyle().
				Foreground(unfocusedColor)

	confirmValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true)
)
