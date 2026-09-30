package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// State colours must stay tellable apart when the terminal downgrades to 256 or 16 colours.
func TestPaletteDistinctAtLowColor(t *testing.T) {
	toks := map[string]token{"accent": tokAccent, "ok": tokOK, "warn": tokWarn, "err": tokErr}
	for a, ta := range toks {
		for b, tb := range toks {
			if a >= b {
				continue
			}
			if ta.ansi256 == tb.ansi256 {
				t.Errorf("%s and %s share 256-colour index %d", a, b, ta.ansi256)
			}
			if ta.ansi == tb.ansi {
				t.Errorf("%s and %s share 16-colour index %d", a, b, ta.ansi)
			}
		}
	}
	if tokMuted.ansi256 == tokInfo.ansi256 || tokMuted.ansi == tokInfo.ansi {
		t.Error("muted and info collide at low colour")
	}
}

// Palette tokens must match the Daun Pisang spec in bungkus-mc
// docs/DESIGN.md §2.1: the dark terminal-fallback set (indices and hex) and
// the painted Daun Teduh hexes. Update both repos together.
func TestPaletteMatchesSpec(t *testing.T) {
	spec := []struct {
		name string
		got  token
		want token
	}{
		{"fg-muted", tokMuted, token{7, 245, "#8a99a8", "#9aab9c"}},
		{"fg-dim", tokDim, token{8, 240, "#555555", "#5c7062"}},
		{"border", tokBorder, token{8, 235, "#2a2a2a", "#3b4d40"}},
		{"accent", tokAccent, token{3, 216, "#ffaa88", ""}},
		{"ok", tokOK, token{2, 107, "#7fb069", ""}},
		{"warn", tokWarn, token{11, 185, "#e8c547", ""}},
		{"err", tokErr, token{9, 209, "#ff7361", ""}},
		{"info", tokInfo, token{12, 68, "#6f8fc7", ""}},
	}
	for _, tc := range spec {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %+v, spec says %+v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestPaintedBackgroundMatchesSpec(t *testing.T) {
	if paintBG != "#1c2a21" || paintFG != "#d6e2d3" {
		t.Errorf("painted bg/fg = %s/%s, spec says #1c2a21/#d6e2d3", paintBG, paintFG)
	}
}

func TestUTF8LocaleSelectsBoxDrawing(t *testing.T) {
	cases := []struct {
		name string
		vars []string
		want bool
	}{
		{"LANG utf-8", []string{"", "", "en_US.UTF-8"}, true},
		{"LC_ALL wins", []string{"C", "", "ja_JP.UTF-8"}, false},
		{"lowercase utf8", []string{"", "C.utf8", ""}, true},
		{"nothing set", []string{"", "", ""}, false},
		{"POSIX", []string{"", "", "POSIX"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUTF8Locale(tc.vars...); got != tc.want {
				t.Errorf("isUTF8Locale(%q) = %v, want %v", tc.vars, got, tc.want)
			}
		})
	}
	if in, focus := borders(false); in != lipgloss.ASCIIBorder() || focus != lipgloss.ASCIIBorder() {
		t.Error("non-UTF-8 terminals must get ASCII borders")
	}
	if _, focus := borders(true); focus != lipgloss.ThickBorder() {
		t.Error("focused sections must get the heavy border on UTF-8 terminals")
	}
}
