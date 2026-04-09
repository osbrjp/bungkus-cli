package tui

import "charm.land/lipgloss/v2"

var (
	ColorPrimary = lipgloss.Color("#c6d0f5") // light text
	ColorAccent  = lipgloss.Color("#8caaee") // blue accent
	ColorMuted   = lipgloss.Color("#626880") // muted overlay
	ColorGreen   = lipgloss.Color("#a6d189") // green for selected
	ColorError   = lipgloss.Color("#e78284") // red

	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	BoxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	TitleStyle    = AccentStyle
	QuestionStyle = AccentStyle
	ActiveStyle   = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	InactiveStyle = MutedStyle
	HintStyle     = MutedStyle
)
