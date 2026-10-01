package cmd

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	goInstalled := &debug.BuildInfo{Main: debug.Module{Path: goModule, Version: "v1.9.1"}}
	localBuild := &debug.BuildInfo{Main: debug.Module{Path: goModule, Version: "(devel)"}}
	cases := []struct {
		name        string
		stamped     string
		bi          *debug.BuildInfo
		ok          bool
		wantVersion string
		wantMethod  installMethod
	}{
		{"release stamp wins", "v1.9.1", goInstalled, true, "v1.9.1", methodRelease},
		{"go install records the tag", "dev", goInstalled, true, "v1.9.1", methodGo},
		{"local go build stays dev", "dev", localBuild, true, "dev", methodRelease},
		{"no build info stays dev", "dev", nil, false, "dev", methodRelease},
		{"empty stamp is treated as dev", "", goInstalled, true, "v1.9.1", methodGo},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, m := resolveVersion(tc.stamped, tc.bi, tc.ok)
			if v != tc.wantVersion || m != tc.wantMethod {
				t.Errorf("resolveVersion = (%q, %d), want (%q, %d)", v, m, tc.wantVersion, tc.wantMethod)
			}
		})
	}
}

func TestUpdateCommand(t *testing.T) {
	if got := updateCommand(methodRelease); got != "bungkus-cli update" {
		t.Errorf("release: %q", got)
	}
	if got := updateCommand(methodGo); got != "go install github.com/osbrjp/bungkus-cli@latest" {
		t.Errorf("go: %q", got)
	}
}

// writeBin creates an executable bungkus-cli in a new temp folder and
// returns the folder.
func writeBin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bungkus-cli"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCopiesOnPath(t *testing.T) {
	local, gobin := writeBin(t), writeBin(t)
	links := t.TempDir()
	if err := os.Symlink(filepath.Join(local, "bungkus-cli"), filepath.Join(links, "bungkus-cli")); err != nil {
		t.Fatal(err)
	}
	notExec := t.TempDir()
	if err := os.WriteFile(filepath.Join(notExec, "bungkus-cli"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()

	path := strings.Join([]string{empty, local, notExec, links, gobin, local}, string(os.PathListSeparator))
	got := copiesOnPath(path)
	want := []string{filepath.Join(local, "bungkus-cli"), filepath.Join(gobin, "bungkus-cli")}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("copiesOnPath = %v, want %v (symlinks, duplicates and non-executables skipped)", got, want)
	}
}

func TestDuplicateNote(t *testing.T) {
	local, gobin := writeBin(t), writeBin(t)
	sep := string(os.PathListSeparator)
	self, _ := filepath.EvalSymlinks(filepath.Join(gobin, "bungkus-cli"))

	if note := duplicateNote(self, gobin); note != "" {
		t.Errorf("a single copy needs no note, got %q", note)
	}

	note := duplicateNote(self, local+sep+gobin)
	for _, want := range []string{
		filepath.Join(local, "bungkus-cli") + " (runs)",
		filepath.Join(gobin, "bungkus-cli") + " (this one)",
		"rm " + filepath.Join(gobin, "bungkus-cli"),
	} {
		if !strings.Contains(note, want) {
			t.Errorf("note lacks %q:\n%s", want, note)
		}
	}
}
