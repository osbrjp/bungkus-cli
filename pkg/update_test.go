package pkg

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name            string
		current, latest string
		want            bool
	}{
		{"newer patch", "v1.4.0", "v1.4.1", true},
		{"newer minor, unprefixed current", "1.4.0", "v1.5.0", true},
		{"same version", "v1.4.0", "v1.4.0", false},
		{"older release", "v1.5.0", "v1.4.0", false},
		{"canary is older than its release", "v1.5.0-canary.2", "v1.5.0", true},
		{"release is not older than a later canary", "v1.5.0", "v1.5.0-canary.3", false},
		{"dev build is never behind", "dev", "v1.5.0", false},
		{"unparseable tag", "v1.4.0", "latest", false},
		{"empty current", "", "v1.5.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNewer(tt.current, tt.latest); got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCachedLatest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "latest-release")
	now := time.Now()

	calls := 0
	fetch := func() (string, error) { calls++; return "v1.4.0", nil }

	// Cold cache: fetches and writes through.
	if got, err := CachedLatest(path, now, time.Hour, fetch); err != nil || got != "v1.4.0" {
		t.Fatalf("cold cache = %q, %v; want v1.4.0", got, err)
	}
	if calls != 1 {
		t.Fatalf("cold cache made %d fetches, want 1", calls)
	}

	// Fresh cache: served from disk, no fetch.
	if got, err := CachedLatest(path, now, time.Hour, fetch); err != nil || got != "v1.4.0" {
		t.Fatalf("fresh cache = %q, %v; want v1.4.0", got, err)
	}
	if calls != 1 {
		t.Fatalf("fresh cache made %d fetches, want 1", calls)
	}

	// Stale cache: the TTL has passed, so it refetches.
	stale := now.Add(-2 * time.Hour)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	fetch = func() (string, error) { calls++; return "v1.5.0", nil }
	if got, err := CachedLatest(path, now, time.Hour, fetch); err != nil || got != "v1.5.0" {
		t.Fatalf("stale cache = %q, %v; want v1.5.0", got, err)
	}
	if calls != 2 {
		t.Fatalf("stale cache made %d fetches, want 2", calls)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "v1.5.0" {
		t.Errorf("cache file = %q, %v; want the refetched tag", b, err)
	}

	// A failed fetch surfaces the error rather than an empty tag.
	stale = now.Add(-2 * time.Hour)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	want := errors.New("offline")
	if _, err := CachedLatest(path, now, time.Hour, func() (string, error) { return "", want }); !errors.Is(err, want) {
		t.Errorf("failed fetch err = %v, want %v", err, want)
	}
}
