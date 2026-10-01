package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bungkus-cli to the latest release.",
	Long: `Replace this binary with the latest published release.

Resolves the newest release tag from GitHub and, when it is newer than the
running version, re-runs the official install script — the same one the README
documents — so downloads are checksum-verified exactly as on a first install.

A binary installed with "go install" is updated with go install instead, so
the copy in GOBIN stays the one that is updated. If bungkus-cli is installed
more than once on PATH, update says which copy runs.

Use --check to report what is available without installing anything.`,
	Args: cobra.NoArgs,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().Bool("check", false, "Report whether an update is available, without installing")
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	if note := duplicateNote(selfPath(), os.Getenv("PATH")); note != "" {
		fmt.Fprint(os.Stderr, note)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	latest, err := pkg.LatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("could not resolve the latest release: %w", err)
	}

	current := rootCmd.Version
	if current == "" || current == pkg.DevVersion {
		fmt.Printf("this is a %s build — latest release is %s\n", pkg.DevVersion, latest)
		fmt.Printf("install it with: curl -fsSL %s | bash\n", pkg.InstallScriptURL)
		return nil
	}

	if !pkg.IsNewer(current, latest) {
		fmt.Printf("already up to date (%s)\n", pkg.NormalizeVersion(current))
		return nil
	}

	if check, _ := cmd.Flags().GetBool("check"); check {
		fmt.Printf("update available: %s → %s\nrun: %s\n", pkg.NormalizeVersion(current), latest, updateCommand(installedBy))
		return nil
	}

	fmt.Printf("updating %s → %s\n", pkg.NormalizeVersion(current), latest)
	if installedBy == methodGo {
		return goInstall(cmd, latest)
	}
	// pipefail so a failed download is not swallowed by the pipe into bash.
	install := exec.CommandContext(cmd.Context(), "bash", "-c",
		"set -o pipefail; curl -fsSL "+pkg.InstallScriptURL+" | bash")
	install.Stdin, install.Stdout, install.Stderr = os.Stdin, os.Stdout, os.Stderr
	install.Env = installEnv(os.Environ())
	return install.Run()
}

// installEnv returns env plus BUNGKUS_CURRENT_BIN, the resolved path of the
// running binary, so install.sh updates this copy in place (or, when its
// folder isn't writable, moves it to ~/.local/bin) instead of guessing from
// PATH. When the path can't be determined, install.sh falls back to PATH.
func installEnv(env []string) []string {
	exe, err := os.Executable()
	if err != nil {
		return env
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return append(env, "BUNGKUS_CURRENT_BIN="+exe)
}

// goInstall updates a binary installed with go install by running
// `go install <module>@<tag>`, which writes to the same GOBIN without sudo.
// When go isn't on PATH it prints the command instead of failing.
func goInstall(cmd *cobra.Command, tag string) error {
	target := goModule + "@" + tag
	goBin, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("installed with go install; update with: go install %s\n", target)
		return nil
	}
	install := exec.CommandContext(cmd.Context(), goBin, "install", target)
	install.Stdin, install.Stdout, install.Stderr = os.Stdin, os.Stdout, os.Stderr
	return install.Run()
}
