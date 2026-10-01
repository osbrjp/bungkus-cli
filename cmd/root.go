/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/osbrjp/bungkus-cli/config"
	"github.com/osbrjp/bungkus-cli/internal/tui"
	"github.com/osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bungkus-cli",
	Short: "A frontend scaffolding cli tool.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tui.UpdateAvailable = knownUpdate(cmd.Root().Version)
		tui.UpdateCommand = updateCommand(installedBy)
		wizardResult, err := tea.NewProgram(tui.NewWizardModel(config.Templates)).Run()
		if err != nil {
			return err
		}

		wm, ok := wizardResult.(tui.WizardModel)
		if !ok || wm.Canceled {
			return nil
		}
		if wm.Err != nil {
			return fmt.Errorf("scaffold failed: %w", wm.Err)
		}
		if !wm.Created {
			return nil
		}

		// The wizard scaffolded the files; install and git init stream their
		// output, so they run here on the normal screen.
		cfg := wm.Cfg
		destDir := cfg.ProjectName
		if cfg.DestDir != "" {
			destDir = cfg.DestDir
		}
		if err := pkg.PostScaffold(destDir, cfg); err != nil {
			return err
		}

		tui.PrintSuccess(cfg)
		return nil
	},
}

// knownUpdate returns the newer release tag the last update check cached
// (e.g. "v1.9.0"), or "" when nothing newer than current is cached or update
// checks are disabled. It only reads the cache file, never the network, so
// the wizard can announce the update on its first frame.
func knownUpdate(current string) string {
	if os.Getenv("BUNGKUS_NO_UPDATE_CHECK") != "" {
		return ""
	}
	path, err := pkg.UpdateCachePath()
	if err != nil {
		return ""
	}
	tag, err := os.ReadFile(path)
	if err != nil || !pkg.IsNewer(current, strings.TrimSpace(string(tag))) {
		return ""
	}
	return pkg.NormalizeVersion(string(tag))
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// SetVersion sets the version `--version` prints and the wizard header shows.
// v is the build-time stamp; without one, the version Go recorded for a
// `go install` build is used (see resolveVersion), which also decides how
// `update` works.
func SetVersion(v string) {
	bi, ok := debug.ReadBuildInfo()
	v, installedBy = resolveVersion(v, bi, ok)
	rootCmd.Version = v
	tui.Version = v
}

func Execute() {
	notify := startUpdateCheck()
	err := rootCmd.Execute()
	notify()
	if err != nil {
		os.Exit(1)
	}
}

// startUpdateCheck looks for a newer release alongside the real command and
// returns a func that prints a one-line hint once that command is done. The
// check never delays or fails the command it rides along with: it is skipped
// unless stderr is a terminal, and a network error stays silent.
func startUpdateCheck() func() {
	silent := func() {}
	if os.Getenv("BUNGKUS_NO_UPDATE_CHECK") != "" || !isTerminal(os.Stderr) {
		return silent
	}
	if len(os.Args) > 1 && os.Args[1] == "update" {
		return silent
	}

	current := rootCmd.Version
	latest := make(chan string, 1)
	go func() { latest <- pkg.AvailableUpdate(current) }()

	return func() {
		select {
		case tag := <-latest:
			if tag != "" {
				fmt.Fprintf(os.Stderr, "\na newer version is available (%s → %s) — run: %s\n",
					pkg.NormalizeVersion(current), tag, updateCommand(installedBy))
			}
		case <-time.After(500 * time.Millisecond):
			// Still in flight; not worth holding the shell for.
		}
	}
}
