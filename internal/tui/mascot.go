package tui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// mascotPixels is the owner's packet mascot, one rune per pixel: G body,
// E eyes, T and U the left and right halves of the paper wrap, '.' empty.
// The row count must stay even: two pixel rows make one cell.
var mascotPixels = []string{
	"....GG....",
	"...GGGG...",
	"...GGGG...",
	"..GEGGEG..",
	".TTTTUUUU.",
	"..........",
}

// mascotLegs is the row drawn under the half-block body: heavy box-drawing
// legs at the columns where the wrap's feet sit.
const mascotLegs = "   ┛  ┛"

// Mascot colours. The legs use the painted muted green-grey because black
// legs would vanish on the Daun Teduh background.
var (
	tokMascotBody  = token{2, 71, "#34ab52", ""}
	tokMascotEyes  = token{0, 233, "#0b120d", ""}
	tokMascotWrapL = token{3, 180, "#d9c574", ""}
	tokMascotWrapR = token{11, 222, "#e7d075", ""}
	tokMascotLegs  = token{7, 109, "#9aab9c", ""}
)

var mascotTokens = map[byte]token{
	'G': tokMascotBody,
	'E': tokMascotEyes,
	'T': tokMascotWrapL,
	'U': tokMascotWrapR,
}

// Mascot poses: the body shifted down by 0, 1 or 2 pixels in a fixed
// 4-row box. The duck loop dips from standing to crouched and back.
const (
	poseStand = 0
	poseHalf  = 1
	poseFull  = 2
)

var duckLoop = []int{poseStand, poseHalf, poseFull, poseFull, poseHalf, poseStand}

// duckInterval is the time each frame of duckLoop is shown.
const duckInterval = 120 * time.Millisecond

// animate is false without a terminal on stdout, where frames would only be
// written to a pipe or a file.
var animate = profile != colorprofile.NoTTY

// mascotRows is the height of every mascot frame in cells.
const mascotRows = 4

// mascot renders the mascot in the standing pose.
func mascot() []string { return mascotFrame(poseStand) }

// mascotFrame renders the mascot with its body shifted down by shift pixel
// rows (0–2) as mascotRows rows of 10 columns. The body is drawn in half
// blocks, two pixel rows per cell: both pixels in the same colour give '█';
// only the top or bottom pixel set gives '▀' or '▄' in its colour; two
// different colours give '▀' with the top as foreground and the bottom as
// background; neither gives a space on the terminal's own background. When
// the bottom row is left empty by the shift, the legs fill it; at a full
// shift the wrap covers them.
func mascotFrame(shift int) []string {
	blank := strings.Repeat(".", len(mascotPixels[0]))
	px := make([]string, 0, mascotRows*2)
	for range shift {
		px = append(px, blank)
	}
	px = append(px, mascotPixels...)
	for len(px) < mascotRows*2 {
		px = append(px, blank)
	}

	rows := make([]string, 0, mascotRows)
	for y := 0; y < mascotRows*2; y += 2 {
		var b strings.Builder
		top, bottom := px[y], px[y+1]
		empty := true
		for x := range len(top) {
			t, tok := mascotTokens[top[x]]
			bt, btok := mascotTokens[bottom[x]]
			st := lipgloss.NewStyle()
			switch {
			case !tok && !btok:
				b.WriteByte(' ')
				continue
			case tok && btok && top[x] == bottom[x]:
				b.WriteString(st.Foreground(t.color()).Render("█"))
			case tok && btok:
				b.WriteString(st.Foreground(t.color()).Background(bt.color()).Render("▀"))
			case tok:
				b.WriteString(st.Foreground(t.color()).Render("▀"))
			default:
				b.WriteString(st.Foreground(bt.color()).Render("▄"))
			}
			empty = false
		}
		if y == mascotRows*2-2 && empty {
			rows = append(rows, mascotLegRow())
			continue
		}
		rows = append(rows, b.String())
	}
	return rows
}

// mascotLegRow renders mascotLegs in the leg colour, padded to the sprite
// width.
func mascotLegRow() string {
	legs := strings.TrimLeft(mascotLegs, " ")
	pad := len(mascotLegs) - len(legs)
	return strings.Repeat(" ", pad) +
		lipgloss.NewStyle().Foreground(tokMascotLegs.color()).Render(legs) +
		strings.Repeat(" ", len(mascotPixels[0])-lipgloss.Width(mascotLegs))
}

// besideMascot places lines to the right of the standing mascot, both
// top-aligned with a 4-column gap; lines past the mascot's height are
// indented to the same column.
func besideMascot(lines []string) []string {
	sprite := mascot()
	gap := strings.Repeat(" ", 4)
	out := make([]string, 0, max(len(lines), len(sprite)))
	for i := range max(len(lines), len(sprite)) {
		left := strings.Repeat(" ", len(mascotPixels[0]))
		if i < len(sprite) {
			left = sprite[i]
		}
		right := ""
		if i < len(lines) {
			right = lines[i]
		}
		out = append(out, left+gap+right)
	}
	return out
}
