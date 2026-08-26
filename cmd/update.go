package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bungkus-cli to the latest release.",
	Long: `Replace this binary with the latest published release.

Resolves the newest release tag from GitHub and, when it is newer than the
running version, re-runs the official install script — the same one the README
documents — so downloads are checksum-verified exactly as on a first install.

Use --check to report what is available without installing anything.`,
	Args: cobra.NoArgs,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().Bool("check", false, "Report whether an update is available, without installing")
}

func runUpdate(cmd *cobra.Command, _ []string) error {
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
		fmt.Printf("update available: %s → %s\nrun: bungkus-cli update\n", pkg.NormalizeVersion(current), latest)
		return nil
	}

	fmt.Printf("updating %s → %s\n", pkg.NormalizeVersion(current), latest)
	// pipefail so a failed download is not swallowed by the pipe into bash.
	install := exec.CommandContext(cmd.Context(), "bash", "-c",
		"set -o pipefail; curl -fsSL "+pkg.InstallScriptURL+" | bash")
	install.Stdin, install.Stdout, install.Stderr = os.Stdin, os.Stdout, os.Stderr
	return install.Run()
}
