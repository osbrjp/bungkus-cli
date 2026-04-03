# CLAUDE.md

## Project Overview

**bungkus-cli** is a Go CLI tool that scaffolds and configures modern frontend projects (Astro, Vite) with common tooling (Tailwind, Prettier, ESLint, Biome). Single binary distribution — no external runtime dependencies.

## Tech Stack

- **Go 1.24+** — CLI, orchestration, template rendering
- **Cobra** — CLI command framework
- **BubbleTea / Lip Gloss** — Interactive TUI
- **Sprig** — Template functions for `text/template`
- **`text/template` + `.tmpl` files** — Config file generation

## Project Structure

```
main.go                     # Entrypoint
cmd/
  root.go                   # Root cobra command
  init.go                   # Init command (main orchestration)
internal/
  config/config.go          # Config schema & loaders
  tui/                      # BubbleTea TUI components
  runner/                   # Command execution utilities
config/
  bases/*.json              # Base framework configs (astro, vite)
  extras/*.json             # Extra tool configs (tailwindcss, prettier, eslint, biome)
  templates/                # .tmpl template files per tool/framework
```

## Build & Run

```bash
go build -o bungkus-cli .
go install .
```

## Key Commands

```bash
bungkus init [path] [flags]
bungkus init . --css tailwindcss --fmt prettier --linter eslint
bungkus init my-app -b astro --css tailwindcss
```

## Architecture Notes

- Templates and configs are embedded in the binary via `//go:embed`
- Config files are generated using Go `text/template` with Sprig functions and `.tmpl` files
- Config JSONs define packages, template data, and tool metadata per base/extra combination
- The TUI launches when no path argument is provided

## Conventions

- Commit messages use conventional commits (feat, fix, chore)
- Branch naming: `i{issue#}-{date}-{seq}` (e.g., `i13-20260327-1705`)
- GitHub repo: `spencer-osbrjp/bungkus-cli`
