package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spencer-osbrjp/bungkus-cli/pkg"
	"github.com/spf13/cobra"
)

// bumpCmd is a maintainer tool (not part of the scaffolding flow): it refreshes
// the version pins in config/registry.json to the newest releases that have
// soaked for at least --soak-weeks and are not deprecated. Intended to run in
// CI on a schedule and open a PR with the result.
var bumpCmd = &cobra.Command{
	Use:    "bump",
	Short:  "Refresh registry.json version pins to the newest safe, soaked releases",
	Hidden: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("registry")
		write, _ := cmd.Flags().GetBool("write")
		days, _ := cmd.Flags().GetInt("soak-days")

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		now := time.Now()
		minAge := time.Duration(days) * 24 * time.Hour
		client := &http.Client{Timeout: 20 * time.Second}

		resolve := func(name string) (string, bool) {
			doc, err := pkg.FetchPackument(client, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ! %s: %v\n", name, err)
				return "", false
			}
			return pkg.PickVersion(doc, now, minAge)
		}

		res, err := pkg.BumpRegistry(string(raw), resolve)
		if err != nil {
			return err
		}

		for _, c := range res.Changes {
			fmt.Printf("  %-32s %s -> %s\n", c.Name, c.From, c.To)
		}
		fmt.Printf("\n%d pin(s) updated, %d package(s) skipped (soak >= %dd)\n",
			len(res.Changes), len(res.Skipped), days)

		if len(res.Changes) == 0 {
			return nil
		}
		if !write {
			fmt.Println("dry run — pass --write to apply")
			return nil
		}
		if err := os.WriteFile(path, []byte(res.Content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Printf("wrote %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bumpCmd)
	bumpCmd.Flags().String("registry", "config/registry.json", "Path to registry.json")
	bumpCmd.Flags().Bool("write", false, "Apply changes (default: dry run)")
	bumpCmd.Flags().Int("soak-days", 14, "Minimum age in days before adopting a release")
}
