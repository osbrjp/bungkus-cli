package tui

import "charm.land/lipgloss/v2"

var (
	// Lackluster palette
	ColorLuster = lipgloss.Color("#deeeed")
	ColorLack   = lipgloss.Color("#708090")
	ColorOrange = lipgloss.Color("#ffaa88")
	ColorYellow = lipgloss.Color("#abab77")
	ColorGreen  = lipgloss.Color("#789978")
	ColorBlue   = lipgloss.Color("#7788AA")
	ColorError  = lipgloss.Color("#D70000")
	ColorDim    = lipgloss.Color("#555555")
	ColorDark   = lipgloss.Color("#2a2a2a")
	ColorGray1  = lipgloss.Color("#080808")
	ColorGray3  = lipgloss.Color("#2a2a2a")

	PrimaryStyle = lipgloss.NewStyle().Foreground(ColorLuster)
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorOrange).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorDim)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	BoldStyle    = lipgloss.NewStyle().Bold(true).Foreground(ColorLuster)
	BoxStyle     = lipgloss.NewStyle().
			Border(lipgloss.ASCIIBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	TitleStyle    = AccentStyle
	QuestionStyle = AccentStyle
	ActiveStyle   = lipgloss.NewStyle().Foreground(ColorLuster).Bold(true)
	InactiveStyle = MutedStyle
	HintStyle     = MutedStyle

	// Focus-aware border styles
	ActiveBorder   = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).BorderForeground(ColorLack).Padding(0, 1)
	InactiveBorder = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).BorderForeground(ColorDark).Padding(0, 1)

	// Panel title style — distinct from group labels inside panels
	PanelTitleStyle = lipgloss.NewStyle().Foreground(ColorLuster).Bold(true).Underline(true)

	// Footer keybinding styles
	FooterBarStyle  = lipgloss.NewStyle().MarginTop(1).Foreground(ColorDim)
	FooterKeyStyle  = lipgloss.NewStyle().Foreground(ColorLuster).Bold(true)
	FooterDescStyle = lipgloss.NewStyle().Foreground(ColorLack)
	FooterSepStyle  = lipgloss.NewStyle().Foreground(ColorDim)
)
