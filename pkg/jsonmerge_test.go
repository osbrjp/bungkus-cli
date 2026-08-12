package pkg

import (
	"strings"
	"testing"
)

func TestMergeAddPackages(t *testing.T) {
	base := `{
  "name": "my-app",
  "author": {
    "name": "Spencer",
    "url": "https://example.com"
  },
  "license": "MIT",
  "scripts": {
    "dev": "astro dev",
    "format": "prettier --write ."
  },
  "dependencies": {
    "astro": "^5.0.0"
  },
  "devDependencies": {
    "prettier": "^3.0.0"
  }
}
`

	cases := []struct {
		name        string
		orig        string
		add         Packages
		wantChanged bool
		contains    []string
		notContains []string
		wantErr     bool
	}{
		{
			name: "adds dep and script, preserves unknown fields",
			orig: base,
			add: Packages{
				Scripts:         map[string]string{"lhci": "lhci autorun"},
				DevDependencies: DevDependencies{"@lhci/cli": "^0.14.0"},
			},
			wantChanged: true,
			contains: []string{
				"\"lhci\": \"lhci autorun\"",
				"\"@lhci/cli\": \"^0.14.0\"",
				// unknown field kept verbatim, interior formatting intact
				"\"author\": {\n    \"name\": \"Spencer\",\n    \"url\": \"https://example.com\"\n  }",
				"\"license\": \"MIT\"",
			},
		},
		{
			name: "never changes an existing script or version",
			orig: base,
			add: Packages{
				Scripts:         map[string]string{"format": "biome format ."},
				DevDependencies: DevDependencies{"prettier": "^9.9.9"},
			},
			wantChanged: false,
			contains:    []string{"prettier --write .", "\"prettier\": \"^3.0.0\""},
			notContains: []string{"biome format .", "^9.9.9"},
		},
		{
			name:        "dep in the other section is skipped, not moved or duplicated",
			orig:        base,
			add:         Packages{Dependencies: Dependencies{"prettier": "^9.9.9"}},
			wantChanged: false,
		},
		{
			name:        "dep in peerDependencies is skipped",
			orig:        `{"peerDependencies": {"react": "^19.0.0"}}`,
			add:         Packages{Dependencies: Dependencies{"react": "^18.0.0"}},
			wantChanged: false,
			notContains: []string{"^18.0.0"},
		},
		{
			name:        "creates a missing section at the end",
			orig:        `{"name": "x"}`,
			add:         Packages{Scripts: map[string]string{"lhci": "lhci autorun"}},
			wantChanged: true,
			contains:    []string{"\"scripts\": {\n    \"lhci\": \"lhci autorun\"\n  }"},
		},
		{
			name:    "malformed json errors",
			orig:    `{"name": `,
			add:     Packages{Scripts: map[string]string{"a": "b"}},
			wantErr: true,
		},
		{
			name:        "duplicate top-level keys are deduped on write",
			orig:        `{"license": "MIT", "license": "ISC", "scripts": {}}`,
			add:         Packages{Scripts: map[string]string{"lhci": "lhci autorun"}},
			wantChanged: true,
			contains:    []string{"\"license\": \"ISC\""},
			notContains: []string{"\"MIT\""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := &AddReport{}
			out, err := MergeAddPackages([]byte(tc.orig), tc.add, rep)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("MergeAddPackages: %v", err)
			}
			if rep.PkgJSONChanged != tc.wantChanged {
				t.Errorf("PkgJSONChanged = %v, want %v", rep.PkgJSONChanged, tc.wantChanged)
			}
			if !tc.wantChanged && string(out) != tc.orig {
				t.Errorf("no-op merge must return the input byte-identical\ngot:  %q\nwant: %q", out, tc.orig)
			}
			for _, s := range tc.contains {
				if !strings.Contains(string(out), s) {
					t.Errorf("output missing %q\n%s", s, out)
				}
			}
			for _, s := range tc.notContains {
				if strings.Contains(string(out), s) {
					t.Errorf("output should not contain %q\n%s", s, out)
				}
			}
		})
	}
}

func TestMergeAddPackagesPreservesKeyOrder(t *testing.T) {
	orig := `{"zebra": 1, "apple": 2, "scripts": {"z": "1", "a": "2"}}`
	rep := &AddReport{}
	out, err := MergeAddPackages([]byte(orig), Packages{Scripts: map[string]string{"m": "3"}}, rep)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !(strings.Index(s, "zebra") < strings.Index(s, "apple") && strings.Index(s, "apple") < strings.Index(s, "scripts")) {
		t.Errorf("top-level key order not preserved:\n%s", s)
	}
	// existing script order kept, new key appended after
	if !(strings.Index(s, "\"z\"") < strings.Index(s, "\"a\"") && strings.Index(s, "\"a\"") < strings.Index(s, "\"m\"")) {
		t.Errorf("script key order not preserved:\n%s", s)
	}
}
