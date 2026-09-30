package tui

import "testing"

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

// Palette tokens must match the Daun Pisang spec: the dark terminal-fallback
// set in bungkus-mc docs/DESIGN.md §2.1. bungkus-cli never paints a
// background, so it implements that set only. Update both repos together.
func TestPaletteMatchesSpec(t *testing.T) {
	spec := []struct {
		name string
		got  token
		want token
	}{
		{"fg-muted", tokMuted, token{7, 245, "#8a99a8"}},
		{"fg-dim", tokDim, token{8, 240, "#555555"}},
		{"border", tokBorder, token{8, 235, "#2a2a2a"}},
		{"accent", tokAccent, token{3, 216, "#ffaa88"}},
		{"ok", tokOK, token{2, 107, "#7fb069"}},
		{"warn", tokWarn, token{11, 185, "#e8c547"}},
		{"err", tokErr, token{9, 209, "#ff7361"}},
		{"info", tokInfo, token{12, 68, "#6f8fc7"}},
	}
	for _, tc := range spec {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s = %+v, spec says %+v", tc.name, tc.got, tc.want)
			}
		})
	}
}
