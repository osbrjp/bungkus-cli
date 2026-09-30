package tui

import (
	"image/color"
	"os"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Daun Pisang palette (dark): the token spec lives in bungkus-mc docs/DESIGN.md §2.
// Rules: never paint a background; primary text is the terminal's default
// foreground; every colour carries explicit 16/256 values because automatic
// downsampling collapses accent, warn and err into the same red.
//
// Detect reads env/isatty only; it never queries the terminal.
var pick = lipgloss.Complete(colorprofile.Detect(os.Stdout, os.Environ()))

// token is one palette entry: 16-colour index, 256-colour index, truecolor hex.
type token struct {
	ansi, ansi256 int
	hex           string
}

func (t token) color() color.Color {
	return pick(lipgloss.Color(strconv.Itoa(t.ansi)), lipgloss.Color(strconv.Itoa(t.ansi256)), lipgloss.Color(t.hex))
}

var (
	tokMuted  = token{7, 245, "#8a99a8"}  // Surat khabar — hints, secondary text
	tokDim    = token{8, 240, "#555555"}  // decorative only
	tokBorder = token{8, 235, "#2a2a2a"}  // inactive border
	tokAccent = token{3, 216, "#ffaa88"}  // Minyak — titles, commands to copy
	tokOK     = token{2, 107, "#7fb069"}  // Daun — selection, success
	tokWarn   = token{11, 185, "#e8c547"} // Kuning — warnings
	tokErr    = token{9, 209, "#ff7361"}  // Sambal — errors
	tokInfo   = token{12, 68, "#6f8fc7"}  // Biru — ids, links
)

var (
	ColorMuted  = tokMuted.color()
	ColorDim    = tokDim.color()
	ColorBorder = tokBorder.color()
	ColorAccent = tokAccent.color()
	ColorOK     = tokOK.color()
	ColorWarn   = tokWarn.color()
	ColorErr    = tokErr.color()
	ColorInfo   = tokInfo.color()

	PrimaryStyle = lipgloss.NewStyle()
	AccentStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorDim)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorErr).Bold(true)
	WarnStyle    = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)
	SkipStyle    = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	BoxStyle     = lipgloss.NewStyle().
			Border(lipgloss.ASCIIBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	TitleStyle    = AccentStyle
	QuestionStyle = AccentStyle
	ActiveStyle   = lipgloss.NewStyle().Bold(true)
	InactiveStyle = MutedStyle
	HintStyle     = MutedStyle

	// Selection uses reverse video instead of a painted background.
	CursorStyle   = lipgloss.NewStyle().Reverse(true).Bold(true)
	SelectedStyle = lipgloss.NewStyle().Foreground(ColorOK).Reverse(true)

	// Focus-aware border styles
	ActiveBorder   = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).BorderForeground(ColorMuted).Padding(0, 1)
	InactiveBorder = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder()).BorderForeground(ColorBorder).Padding(0, 1)

	// Panel title style — distinct from group labels inside panels
	PanelTitleStyle = lipgloss.NewStyle().Bold(true).Underline(true)

	// Footer keybinding styles
	FooterBarStyle  = lipgloss.NewStyle().MarginTop(1).Foreground(ColorDim)
	FooterKeyStyle  = lipgloss.NewStyle().Bold(true)
	FooterDescStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	FooterSepStyle  = lipgloss.NewStyle().Foreground(ColorDim)
)
