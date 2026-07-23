package pkg

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPickVersion(t *testing.T) {
	now := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	iso := func(d time.Time) string { return d.Format(time.RFC3339) }
	old := iso(now.Add(-100 * 24 * time.Hour)) // ~14 weeks old, soaked
	fresh := iso(now.Add(-3 * 24 * time.Hour)) // 3 days old, still soaking

	doc := Packument{
		Versions: map[string]struct {
			Deprecated json.RawMessage `json:"deprecated"`
		}{
			"1.0.0":      {},
			"1.2.0":      {},                                            // newest soaked & safe -> want this
			"1.3.0":      {},                                            // too fresh
			"2.0.0-rc.1": {},                                            // prerelease
			"1.1.0":      {Deprecated: json.RawMessage(`"do not use"`)}, // deprecated
		},
		Time: map[string]string{
			"1.0.0":      old,
			"1.1.0":      old,
			"1.2.0":      old,
			"1.3.0":      fresh,
			"2.0.0-rc.1": old,
		},
	}

	got, ok := PickVersion(doc, now, DefaultSoak)
	if !ok || got != "1.2.0" {
		t.Fatalf("PickVersion = %q, %v; want 1.2.0, true", got, ok)
	}

	// Nothing qualifies -> ok=false.
	none := Packument{
		Versions: map[string]struct {
			Deprecated json.RawMessage `json:"deprecated"`
		}{"9.9.9-beta": {}},
		Time: map[string]string{"9.9.9-beta": old},
	}
	if _, ok := PickVersion(none, now, DefaultSoak); ok {
		t.Fatal("PickVersion should reject prerelease-only package")
	}
}

func TestBumpRegistry(t *testing.T) {
	// Minimal registry with a scripts/name collision ("astro" appears as both a
	// script value and a dependency version) to prove scripts stay untouched.
	content := `{
  "bases": [
    {
      "value": "astro",
      "packages": {
        "scripts": { "astro": "astro" },
        "dependencies": { "astro": "^6.1.3" }
      }
    }
  ],
  "commonPackages": { "devDependencies": { "typescript": "~5.9.3" } }
}`

	resolve := func(name string) (string, bool) {
		switch name {
		case "astro":
			return "6.4.0", true
		case "typescript":
			return "5.9.3", true // unchanged version
		}
		return "", false
	}

	res, err := BumpRegistry(content, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Changes) != 1 || res.Changes[0].To != "^6.4.0" {
		t.Fatalf("changes = %+v; want single astro -> ^6.4.0", res.Changes)
	}
	if want := `"astro": "^6.4.0"`; !contains(res.Content, want) {
		t.Errorf("content missing %q", want)
	}
	if !contains(res.Content, `"astro": "astro"`) {
		t.Error("script value was wrongly rewritten")
	}
	if !contains(res.Content, `"typescript": "~5.9.3"`) {
		t.Error("unchanged pin should keep its prefix")
	}
}

func TestApplyPinStrategy(t *testing.T) {
	cases := []struct {
		in   string
		pin  PinStrategy
		want string
	}{
		{"^6.1.3", PinExact, "6.1.3"},
		{"~5.9.3", PinCaret, "^5.9.3"},
		{"4.4.6", PinTilde, "~4.4.6"},
		{"^6.1.3", PinDefault, "^6.1.3"}, // default falls through unchanged
		{"latest", PinExact, "latest"},   // non-semver left alone
		{"workspace:*", PinCaret, "workspace:*"},
	}
	for _, c := range cases {
		if got := applyPinStrategy(c.in, c.pin); got != c.want {
			t.Errorf("applyPinStrategy(%q, %q) = %q; want %q", c.in, c.pin, got, c.want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
