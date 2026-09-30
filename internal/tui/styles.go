package tui

import (
	"image/color"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Daun Pisang palette (dark); the token spec lives in bungkus-mc docs/DESIGN.md §2.
//
// On truecolor terminals the wizard paints the "Daun Teduh" leaf-green
// background and text colour, as bungkus-mc does, and uses the painted hexes.
// Elsewhere (256/16 colours, no TTY, or BUNGKUS_BACKGROUND=terminal) it keeps
// the terminal's own background and uses the fallback set, whose explicit
// 16/256 indices keep accent, warn and err apart after downsampling.
//
// Detect reads env/isatty only; it never queries the terminal.
var (
	profile = colorprofile.Detect(os.Stdout, os.Environ())
	pick    = lipgloss.Complete(profile)
	painted = profile == colorprofile.TrueColor && os.Getenv("BUNGKUS_BACKGROUND") != "terminal"
	unicode = isUTF8Locale(os.Getenv("LC_ALL"), os.Getenv("LC_CTYPE"), os.Getenv("LANG"))
)

// token is one palette entry: the fallback set's 16-colour index, 256-colour
// index and hex, plus the painted-set hex when it differs (empty = same hex).
type token struct {
	ansi, ansi256 int
	hex           string
	paint         string
}

func (t token) color() color.Color {
	if painted && t.paint != "" {
		return lipgloss.Color(t.paint)
	}
	return pick(lipgloss.Color(strconv.Itoa(t.ansi)), lipgloss.Color(strconv.Itoa(t.ansi256)), lipgloss.Color(t.hex))
}

var (
	tokMuted  = token{7, 245, "#8a99a8", "#9aab9c"} // Surat khabar — hints, secondary text
	tokDim    = token{8, 240, "#555555", "#5c7062"} // decorative only
	tokBorder = token{8, 235, "#2a2a2a", "#3b4d40"} // inactive border
	tokAccent = token{3, 216, "#ffaa88", ""}        // Minyak — titles, commands to copy
	tokOK     = token{2, 107, "#7fb069", ""}        // Daun — selection, success, focus
	tokWarn   = token{11, 185, "#e8c547", ""}       // Kuning — warnings
	tokErr    = token{9, 209, "#ff7361", ""}        // Sambal — errors
	tokInfo   = token{12, 68, "#6f8fc7", ""}        // Biru — ids, links

	// Painted background and text (Daun Teduh); used only when painting.
	paintBG = "#1c2a21"
	paintFG = "#d6e2d3"
)

// ThemeBG and ThemeFG are the terminal colours the views set while running.
// Both are nil when not painting, which leaves the terminal's own colours.
var ThemeBG, ThemeFG color.Color

func init() {
	if painted {
		ThemeBG, ThemeFG = lipgloss.Color(paintBG), lipgloss.Color(paintFG)
	}
}

// paint applies the theme's terminal colours to a view. Bubble Tea restores
// the user's own colours when the program exits.
func paint(v tea.View) tea.View {
	v.BackgroundColor, v.ForegroundColor = ThemeBG, ThemeFG
	return v
}

// isUTF8Locale reports whether the first set locale variable (in POSIX
// precedence order) names a UTF-8 encoding, so box-drawing borders render.
func isUTF8Locale(vars ...string) bool {
	for _, v := range vars {
		if v != "" {
			v = strings.ToLower(v)
			return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
		}
	}
	return false
}

// borders returns the inactive and focused border shapes: light and heavy
// box-drawing on UTF-8 terminals, plain ASCII otherwise.
func borders(utf8 bool) (inactive, focused lipgloss.Border) {
	if utf8 {
		return lipgloss.NormalBorder(), lipgloss.ThickBorder()
	}
	return lipgloss.ASCIIBorder(), lipgloss.ASCIIBorder()
}

var inactiveShape, focusedShape = borders(unicode)

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
			Border(inactiveShape).
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

	// Focus-aware border styles: focused sections get the heavy Daun-green border.
	ActiveBorder   = lipgloss.NewStyle().Border(focusedShape).BorderForeground(ColorOK).Padding(0, 1)
	InactiveBorder = lipgloss.NewStyle().Border(inactiveShape).BorderForeground(ColorDim).Padding(0, 1)

	// Panel title style — distinct from group labels inside panels
	PanelTitleStyle = lipgloss.NewStyle().Bold(true).Underline(true)

	// Footer status bar: a mode badge followed by the key hints, as in bungkus-mc.
	StatusModeStyle = lipgloss.NewStyle().Foreground(ColorOK).Reverse(true).Bold(true)
	FooterBarStyle  = lipgloss.NewStyle().MarginTop(1).Foreground(ColorDim)
	FooterKeyStyle  = lipgloss.NewStyle().Bold(true)
	FooterDescStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	FooterSepStyle  = lipgloss.NewStyle().Foreground(ColorDim)
)
