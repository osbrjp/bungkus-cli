package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/osbrjp/bungkus-cli/pkg"
)

func TestKnownUpdateReadsOnlyTheCache(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("BUNGKUS_NO_UPDATE_CHECK", "")
	path, err := pkg.UpdateCachePath()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name, cached, current, want string
	}{
		{"no cache", "", "1.8.0", ""},
		{"newer release", "v1.9.0", "1.8.0", "v1.9.0"},
		{"untagged cache", "1.9.0", "v1.8.0", "v1.9.0"},
		{"up to date", "v1.8.0", "1.8.0", ""},
		{"dev build", "v1.9.0", "dev", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove(path)
			if tc.cached != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tc.cached+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := knownUpdate(tc.current); got != tc.want {
				t.Errorf("knownUpdate(%q) with cache %q = %q, want %q", tc.current, tc.cached, got, tc.want)
			}
		})
	}

	t.Setenv("BUNGKUS_NO_UPDATE_CHECK", "1")
	if got := knownUpdate("1.8.0"); got != "" {
		t.Errorf("update checks disabled, got %q", got)
	}
}
