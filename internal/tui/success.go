package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/spencer-osbrjp/bungkus-cli/pkg"
)

// PrintCICDSkipped prints a styled warning that CI/CD was skipped because no
// deploy target was selected.
func PrintCICDSkipped() {
	tag := WarnStyle.Render(" WARN ")
	msg := fmt.Sprintf(
		"%s %s requires a deploy target — skipping CI/CD workflow",
		tag,
		AccentStyle.Render("CI/CD"),
	)
	fmt.Println(msg)
}

// PrintSkippedIntegration prints a styled warning that a library is skipped
// because it's not compatible with the chosen base framework.
func PrintSkippedIntegration(lib, base string) {
	tag := WarnStyle.Render(" WARN ")
	msg := fmt.Sprintf(
		"%s %s is not supported on %s — skipping %s",
		tag,
		AccentStyle.Render(lib),
		AccentStyle.Render(base),
		SkipStyle.Render(lib),
	)
	fmt.Println(msg)
}

// PrintSuccess prints a styled success box with get-started instructions.
func PrintSuccess(cfg pkg.ProjectConfig) {
	header := PrimaryStyle.Render("✔ ") + "Project scaffolded at " + AccentStyle.Render(cfg.ProjectName)

	// Only show "cd <name>" when scaffolded into a new subfolder, not when using ".".
	var cdLine string
	if cfg.DestDir != "." {
		cdLine = "\n    " + lipgloss.NewStyle().Foreground(ColorOrange).Render("cd "+cfg.ProjectName)
	}

	orange := lipgloss.NewStyle().Foreground(ColorOrange)

	// In a monorepo the deploy script lives in apps/web, so target it directly.
	deployRun := string(cfg.PM) + " run deploy"
	if cfg.Layout.IsMonorepo() {
		deployRun = "pnpm --filter web run deploy"
	}

	runLine := orange.Render(cfg.PM.RunCmd())
	if cfg.Layout.IsMonorepo() && cfg.Backend != "none" {
		runLine += MutedStyle.Render("   # runs apps/web + apps/api")
	}

	cmds := fmt.Sprintf(
		"\n\n  %s%s\n    %s\n    %s",
		AccentStyle.Render("Get started:"),
		cdLine,
		orange.Render(cfg.PM.InstallCmd()),
		runLine,
	)

	// Monorepo layout: explain the workspace and the API's DB setup so the
	// pnpm-workspace structure isn't a surprise after scaffolding.
	if cfg.Layout.IsMonorepo() {
		ws := "\n\n  " + AccentStyle.Render("Workspace (pnpm):") +
			"\n    " + MutedStyle.Render("apps/web         frontend")
		if cfg.Backend != "none" {
			ws += "\n    " + MutedStyle.Render("apps/api         backend ("+string(cfg.Backend)+")")
		}
		ws += "\n    " + MutedStyle.Render("packages/domain  shared types/schemas")
		cmds += ws

		if cfg.ORM != "none" {
			cmds += "\n\n  " + AccentStyle.Render("Set up the database (apps/api):") +
				"\n    " + orange.Render("cp apps/api/.env.example apps/api/.env") +
				"\n    " + orange.Render("pnpm --filter api db:generate") +
				"\n    " + orange.Render("pnpm --filter api db:migrate")
		}
	}

	cicdSecrets := fmt.Sprintf("%s\n    %s\n    %s",
		MutedStyle.Render("1. Set GitHub secrets (once):"),
		MutedStyle.Render("   gh secret set CLOUDFLARE_API_TOKEN"),
		MutedStyle.Render("   gh secret set CLOUDFLARE_ACCOUNT_ID  ← wrangler whoami"),
	)

	if cfg.Deployment == "cloudflare-pages" {
		if cfg.CICD == "github-actions" {
			cmds += fmt.Sprintf(
				"\n\n  %s\n\n    %s\n    %s",
				AccentStyle.Render("Deploy to Cloudflare Pages (CI/CD):"),
				cicdSecrets,
				MutedStyle.Render("2. git push"),
			)
		} else {
			cmds += fmt.Sprintf(
				"\n\n  %s\n\n    %s\n    %s\n    %s",
				AccentStyle.Render("Deploy to Cloudflare Pages:"),
				MutedStyle.Render("1. wrangler login (once)"),
				MutedStyle.Render("2. wrangler pages project create "+cfg.ProjectName+" (once)"),
				MutedStyle.Render("3. "+deployRun),
			)
		}
	} else if cfg.Deployment == "cloudflare-workers" {
		if cfg.CICD == "github-actions" {
			cmds += fmt.Sprintf(
				"\n\n  %s\n\n    %s\n    %s",
				AccentStyle.Render("Deploy to Cloudflare Workers (CI/CD):"),
				cicdSecrets,
				MutedStyle.Render("2. git push"),
			)
		} else {
			cmds += fmt.Sprintf(
				"\n\n  %s\n\n    %s\n    %s",
				AccentStyle.Render("Deploy to Cloudflare Workers:"),
				MutedStyle.Render("1. wrangler login (once)"),
				MutedStyle.Render("2. "+deployRun),
			)
		}
	}

	fmt.Println(BoxStyle.Render(header + cmds))
}
