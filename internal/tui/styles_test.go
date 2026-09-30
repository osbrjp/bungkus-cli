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
