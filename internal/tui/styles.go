package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary = lipgloss.Color("#88bf6e")
	ColorAccent  = lipgloss.Color("#e0af68")
	ColorMuted   = lipgloss.Color("#7a7f8a")
	ColorError   = lipgloss.Color("#f7768e")

	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	BoxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)
)
