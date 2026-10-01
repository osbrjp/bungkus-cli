package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallScriptPicksDirWithoutSudo runs install.sh in dry-run mode, which
// prints the folder it would install into (and the old copy it would leave
// behind) without touching the network.
func TestInstallScriptPicksDirWithoutSudo(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write to every folder, so the non-writable cases don't apply")
	}
	home := t.TempDir()
	userBin := filepath.Join(home, ".local", "bin")
	writable := t.TempDir()
	locked := t.TempDir()
	for _, dir := range []string{writable, locked} {
		if err := os.WriteFile(filepath.Join(dir, "bungkus-cli"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	links := t.TempDir()
	if err := os.Symlink(filepath.Join(writable, "bungkus-cli"), filepath.Join(links, "bungkus-cli")); err != nil {
		t.Fatal(err)
	}
	sys := "/usr/bin:/bin"

	cases := []struct {
		name      string
		env       []string
		wantDir   string
		wantMoved string
	}{
		{"explicit folder wins", []string{"BUNGKUS_INSTALL_DIR=/opt/custom", "PATH=" + locked + ":" + sys}, "/opt/custom", ""},
		{"fresh install goes to ~/.local/bin", []string{"PATH=" + sys}, userBin, ""},
		{"writable install is updated in place", []string{"PATH=" + writable + ":" + sys}, writable, ""},
		{"current binary from update wins over PATH", []string{"BUNGKUS_CURRENT_BIN=" + filepath.Join(writable, "bungkus-cli"), "PATH=" + sys}, writable, ""},
		{"symlinked binary updates its target", []string{"PATH=" + links + ":" + sys}, writable, ""},
		{"locked install moves to an earlier ~/.local/bin", []string{"PATH=" + userBin + ":" + locked + ":" + sys}, userBin, filepath.Join(locked, "bungkus-cli")},
		{"locked install stays when ~/.local/bin comes later", []string{"PATH=" + locked + ":" + userBin + ":" + sys}, locked, ""},
		{"locked install stays when ~/.local/bin is not on PATH", []string{"PATH=" + locked + ":" + sys}, locked, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := exec.Command("bash", "../install.sh")
			c.Env = append([]string{"HOME=" + home, "BUNGKUS_INSTALL_DRY_RUN=1"}, tc.env...)
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("install.sh: %v\n%s", err, out)
			}
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if lines[0] != tc.wantDir {
				t.Errorf("folder = %q, want %q", lines[0], tc.wantDir)
			}
			moved := ""
			if len(lines) > 1 {
				moved = strings.TrimPrefix(lines[1], "migrated-from ")
			}
			if moved != tc.wantMoved {
				t.Errorf("migrated-from = %q, want %q", moved, tc.wantMoved)
			}
		})
	}
}

func TestInstallEnvPassesTheRunningBinary(t *testing.T) {
	env := installEnv([]string{"A=1"})
	if env[0] != "A=1" || !strings.HasPrefix(env[len(env)-1], "BUNGKUS_CURRENT_BIN=/") {
		t.Errorf("installEnv = %v, want the input plus BUNGKUS_CURRENT_BIN=<absolute path>", env)
	}
}
