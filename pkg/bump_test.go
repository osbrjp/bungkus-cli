package pkg

import (
	"encoding/json"
	"reflect"
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
			return "7.0.0", true
		case "typescript":
			return "5.9.3", true // unchanged version
		}
		return "", false
	}

	res, err := BumpRegistry(content, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Changes) != 1 || res.Changes[0].To != "^7.0.0" {
		t.Fatalf("changes = %+v; want single astro -> ^7.0.0", res.Changes)
	}
	if want := `"astro": "^7.0.0"`; !contains(res.Content, want) {
		t.Errorf("content missing %q", want)
	}
	if !contains(res.Content, `"astro": "astro"`) {
		t.Error("script value was wrongly rewritten")
	}
	if !contains(res.Content, `"typescript": "~5.9.3"`) {
		t.Error("unchanged pin should keep its prefix")
	}
}

func TestNeedsBump(t *testing.T) {
	cases := []struct {
		raw, target string
		want        bool
	}{
		{"^6.1.3", "6.4.0", false}, // caret floats within the major
		{"^6.1.3", "7.0.0", true},  // major needs a human
		{"^0.45.2", "0.45.9", false},
		{"^0.45.2", "0.46.0", true}, // ^0.x floats patch only
		{"^0.0.3", "0.0.4", true},   // ^0.0.x is effectively exact
		{"~7.9.1", "7.9.4", false},
		{"~7.9.1", "7.10.0", true},   // tilde floats patch only
		{"2.5.9", "2.5.13", true},    // exact pins never float
		{"^19.3.0", "19.2.9", false}, // never downgrade
		{"2.5.13", "2.5.12", false},
	}
	for _, c := range cases {
		prefix := rangePrefix.FindString(c.raw)
		if got := needsBump(c.raw, prefix, c.target); got != c.want {
			t.Errorf("needsBump(%q, %q) = %v; want %v", c.raw, c.target, got, c.want)
		}
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

func TestBumpRegistryDesktop(t *testing.T) {
	// Proves the desktop category is scanned by BumpRegistry (it enumerates
	// option groups explicitly, so a missing group silently freezes its pins).
	content := `{
  "desktop": [
    {
      "value": "tauri",
      "packages": { "devDependencies": { "@tauri-apps/cli": "^2.11.4" } }
    }
  ]
}`
	resolve := func(name string) (string, bool) {
		if name == "@tauri-apps/cli" {
			return "3.0.0", true
		}
		return "", false
	}
	res, err := BumpRegistry(content, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Changes) != 1 || res.Changes[0].To != "^3.0.0" {
		t.Fatalf("changes = %+v; want single @tauri-apps/cli -> ^3.0.0", res.Changes)
	}
	if want := `"@tauri-apps/cli": "^3.0.0"`; !contains(res.Content, want) {
		t.Errorf("content missing %q", want)
	}
}

func TestBumpRegistryScansBackendOrmDatabase(t *testing.T) {
	// #122: these categories were missing from the pin collection, silently
	// freezing hono/drizzle/driver pins out of the freshness policy.
	content := `{
  "backend": [
    { "value": "hono", "packages": { "dependencies": { "hono": "^4.6.14" } } }
  ],
  "orm": [
    { "value": "drizzle", "packages": { "dependencies": { "drizzle-orm": "^0.38.0" } } }
  ],
  "database": [
    { "value": "postgres", "packages": { "dependencies": { "pg": "^8.13.1" } } }
  ]
}`
	bumped := map[string]string{"hono": "5.0.0", "drizzle-orm": "0.39.0", "pg": "9.0.0"}
	resolve := func(name string) (string, bool) {
		v, ok := bumped[name]
		return v, ok
	}
	res, err := BumpRegistry(content, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Changes) != len(bumped) {
		t.Fatalf("changes = %+v; want one per backend/orm/database pin", res.Changes)
	}
	for _, want := range []string{`"hono": "^5.0.0"`, `"drizzle-orm": "^0.39.0"`, `"pg": "^9.0.0"`} {
		if !contains(res.Content, want) {
			t.Errorf("content missing %q", want)
		}
	}
}

func TestOptionGroupsCoverRegistry(t *testing.T) {
	// Guards the #122 class of bug: every []OptionEntry field of Registry must
	// be returned by optionGroups(), so adding a category to the struct without
	// wiring it into the all-options scan fails here instead of silently
	// freezing its pins.
	var fields int
	rt := reflect.TypeOf(Registry{})
	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).Type == reflect.TypeOf([]OptionEntry{}) {
			fields++
		}
	}
	if groups := (&Registry{}).optionGroups(); len(groups) != fields {
		t.Errorf("optionGroups() returns %d groups but Registry has %d []OptionEntry fields — a category is missing from the all-options scan", len(groups), fields)
	}
}
