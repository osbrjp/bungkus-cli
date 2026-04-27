# CLAUDE.md

## Project Overview

**bungkus-cli** is a Go CLI tool that scaffolds and configures modern frontend projects (Astro, Nuxt, Vite) with common tooling (Tailwind, Biome, Prettier, ESLint, OxLint, OxFmt) and opinionated add-ons (Zod, React Hook Form, TanStack Form/Query, VeeValidate, Pinia, Jotai, Zustand, Nano Stores, microCMS). Single-binary distribution — no external runtime dependencies.

## Tech Stack

- **Go 1.24+** — CLI, orchestration, template rendering
- **Cobra** — CLI command framework
- **BubbleTea / Lip Gloss** — interactive TUI
- **Sprig** — template functions for `text/template`
- **`text/template` + `.tmpl` files** — config file generation

## Project Structure

```
main.go                     # Entrypoint; loads the embedded registry
cmd/
  root.go                   # Root command (interactive TUI wizard)
  create.go                 # Create command; named templates + flag overrides
config/
  embed.go                  # //go:embed for registry.json and templates/
  registry.json             # Single source of truth for bases, add-ons, packages
  templates/                # .tmpl files per base/tool/integration
internal/
  tui/                      # BubbleTea components + lipgloss styles
pkg/
  config.go                 # ProjectConfig + typed enums (FormLib, StateLib, ...)
  registry.go               # Registry schema + global loader
  packagejson.go            # Data-driven package.json builder
  scaffold.go               # Template rendering and file emission
  *_test.go                 # Table-driven tests
```

## Build, Run, Test

```bash
go build -o bungkus-cli .
go run . create my-app -t astro-react --pm bun
go test ./...
```

## Key Commands

```bash
bungkus-cli                                      # interactive TUI wizard
bungkus-cli create my-app --base astro-react    # flag-driven
bungkus-cli create my-app -t nuxt --pm bun       # template preset + flag override
```

Flags (see `--help` for full list): `--base`, `--css`, `--fmt`, `--linter`, `--validation`, `--form`, `--query`, `--state`, `--cms`, `--pm`, `-t/--template`.

## Architecture Notes

- **Registry-driven**: `config/registry.json` is the single source of truth. Every base framework, add-on, and package version lives there. Adding an option means editing the JSON and (usually) adding a template directory — no Go changes required for pure package additions.
- **Embedded assets**: `registry.json` and every `.tmpl` are compiled into the binary via `//go:embed`.
- **Config generation**: `pkg/packagejson.BuildPackageJSON` merges base + css + formatter + linter + validation + form + query + state + cms packages in a fixed order, then applies cross-cutting rules (e.g. prettier+tailwindcss adds `prettier-plugin-tailwindcss`, react-hook-form+zod adds `@hookform/resolvers`).
- **Integration filtering**: options declare `requiresIntegration: ["react"]` / `["vue"]` / `["react", "vue"]` in the registry. `FormLib.IsValidIntegration`, `QueryLib.IsValidIntegration`, and `StateLib.IsValidIntegration` check compatibility against the base's effective integration (Nuxt is remapped to Vue). The TUI hides incompatible options; the CLI warns and falls back to `none`.
- **Template presets**: `cmd/create.go` defines named factory functions (e.g. `astro-react`, `nuxt`, `vite-vue`). Each starts from `pkg.NewProjectConfig()` and overrides fields, so new ProjectConfig fields inherit defaults automatically.
- **Flag/template interaction**: flags only override template/default values when the user explicitly types them (`cmd.Flags().Changed(name)`), so cobra defaults can't silently clobber a preset.
- **TUI launches** when no subcommand/path is provided; `wm.Cfg` is handed to `pkg.Scaffold` on confirm.

## Conventions

- Commit messages: conventional commits (`feat:`, `fix:`, `test:`, `chore:`)
- Branch naming: `i{issue#}-{date}-{seq}` (e.g. `i37-20260414-1741`)
- GitHub repo: `spencer-osbrjp/bungkus-cli`
- Tests live next to the code (`pkg/config_test.go`, `pkg/packagejson_test.go`). Table-driven; they exercise the real embedded registry so changes to `registry.json` are covered automatically.
