package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/osbrjp/bungkus-cli/pkg"
)

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

// successLines returns the success block shown beside the mascot: the
// "Wrapped!" line, a blank line, and the commands that start the project
// (without "cd" when it was scaffolded into the current directory).
func successLines(cfg pkg.ProjectConfig) []string {
	lines := []string{okStyle.Bold(true).Render("Wrapped! " + cfg.ProjectName + " is ready."), ""}
	if cfg.DestDir != "." {
		lines = append(lines, cmdStyle.Render("cd "+cfg.ProjectName))
	}
	return append(lines, cmdStyle.Render(cfg.PM.InstallCmd()), cmdStyle.Render(cfg.PM.RunCmd()))
}

// PrintSuccess prints the success box: the mascot beside the "Wrapped!"
// block, then the local URLs, the workspace layout, database setup and
// deploy steps that apply to cfg. Colours are dropped when stdout is not a
// terminal.
func PrintSuccess(cfg pkg.ProjectConfig) {
	_, _ = lipgloss.Println(successText(cfg))
}

// successText renders the box PrintSuccess prints.
func successText(cfg pkg.ProjectConfig) string {
	orange := lipgloss.NewStyle().Foreground(ColorAccent)

	// In a monorepo the deploy script lives in apps/web, so target it directly.
	deployRun := string(cfg.PM) + " run deploy"
	if cfg.Layout.IsMonorepo() {
		deployRun = "pnpm --filter web run deploy"
	}

	header := strings.Join(besideMascot(successLines(cfg)), "\n")

	// Local URLs the dev server serves on.
	urls := "\n\n  " + AccentStyle.Render("Local URLs:") +
		"\n    " + MutedStyle.Render("web  ") + orange.Render("http://localhost:3000")
	if cfg.Backend != "none" {
		urls += "\n    " + MutedStyle.Render("api  ") + orange.Render("http://localhost:8000")
	}
	cmds := urls

	// Monorepo layout: explain the pnpm-workspace structure.
	if cfg.Layout.IsMonorepo() {
		ws := "\n\n  " + AccentStyle.Render("Workspace (pnpm):") +
			"\n    " + MutedStyle.Render("apps/web         frontend")
		if cfg.Backend != "none" {
			ws += "\n    " + MutedStyle.Render("apps/api         backend ("+string(cfg.Backend)+")")
		}
		ws += "\n    " + MutedStyle.Render("packages/domain  shared types/schemas")
		cmds += ws
	}

	// Database setup steps (ORM selected). A server DB (postgres/mysql) ships a
	// docker-compose.yml at the root, so include `docker compose up -d`.
	if cfg.ORM != "none" {
		title := "Set up the database:"
		envSrc, envDst := ".env.example", ".env"
		gen := string(cfg.PM) + " run db:generate"
		migrate := string(cfg.PM) + " run db:migrate"
		if cfg.Layout.IsMonorepo() {
			title = "Set up the database (apps/api):"
			envSrc, envDst = "apps/api/.env.example", "apps/api/.env"
			gen, migrate = "pnpm --filter api db:generate", "pnpm --filter api db:migrate"
		}
		db := "\n\n  " + AccentStyle.Render(title) +
			"\n    " + orange.Render("cp "+envSrc+" "+envDst)
		if cfg.Database == "postgres" || cfg.Database == "mysql" {
			db += "\n    " + orange.Render("docker compose up -d")
		}
		db += "\n    " + orange.Render(gen) +
			"\n    " + orange.Render(migrate)
		cmds += db
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

	return BoxStyle.Render(header + cmds)
}
