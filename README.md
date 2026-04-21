# bungkus-cli
```
88""Yb 88   88 88b 88  dP""b8 88  dP 88   88 .dP"Y8      dP""b8 88     88
88__dP 88   88 88Yb88 dP   `" 88odP  88   88 `Ybo."     dP   `" 88     88
88""Yb Y8   8P 88 Y88 Yb  "88 88"Yb  Y8   8P o.`Y8b     Yb      88  .o 88
88oodP `YbodP' 88  Y8  YboodP 88  Yb `YbodP' 8bodP'      YboodP 88ood8 88
```

A Go CLI tool that scaffolds and configures modern frontend projects with common tooling. Single binary, no external runtime dependencies.

## Features

- **Frameworks** — Astro (vanilla, +Vue, +React), Nuxt, Vite (vanilla, +Vue, +React)
- **CSS** — Vanilla, Tailwind CSS
- **Formatters** — Biome, Prettier, OxFmt
- **Linters** — Biome, ESLint, OxLint
- **Validation** — Zod
- **Forms** — React Hook Form, TanStack Form, VeeValidate
- **Data Fetching** — TanStack Query
- **State Management** — Jotai, Zustand, Pinia, Nano Stores
- **CMS** — microCMS
- **Package Managers** — pnpm, bun, npm, yarn
- **Templates** — Named presets for common stacks (`-t astro-react`, `-t nuxt`, …)
- **Interactive TUI** — Guided wizard when run without arguments

Add-ons are integration-aware: forms, queries, and state libraries are filtered to compatible bases (e.g. VeeValidate is only offered on Vue-flavored bases; Jotai on React-flavored bases). Mismatched combinations passed via flags print a styled warning and fall back to `none`.

## Getting Started

### Install

```bash
go install github.com/spencer-osbrjp/bungkus-cli@latest
```

### Usage

Run the interactive wizard:

```bash
bungkus-cli
```

Or use the `create` command with flags:

```bash
bungkus-cli create my-app --base astro-react --css tailwindcss --fmt biome --pm pnpm
```

Start from a named template and override individual options with flags:

```bash
bungkus-cli create my-app -t nuxt --pm bun
```

#### Flags

| Flag           | Default   | Options                                                          |
| :------------- | :-------- | :--------------------------------------------------------------- |
| `--base`       | `astro`   | `astro`, `astro-vue`, `astro-react`, `nuxt`, `vite`, `vite-vue`, `vite-react` |
| `--css`        | `vanilla` | `vanilla`, `tailwindcss`                                         |
| `--fmt`        | `biome`   | `biome`, `prettier`, `oxfmt`                                     |
| `--linter`     | `biome`   | `biome`, `eslint`, `oxlint`                                      |
| `--validation` | `none`    | `none`, `zod`                                                    |
| `--form`       | `none`    | `none`, `react-hook-form`, `tanstack-form`, `veevalidate`        |
| `--query`      | `none`    | `none`, `tanstack-query`                                         |
| `--state`      | `none`    | `none`, `jotai`, `zustand`, `pinia`, `nanostores`                |
| `--cms`        | `none`    | `none`, `microcms`                                               |
| `--pm`         | `pnpm`    | `pnpm`, `bun`, `npm`, `yarn`                                     |
| `-t, --template` | —       | `astro`, `astro-react`, `astro-vue`, `nuxt`, `vite`, `vite-react`, `vite-vue` |

Flags take precedence over template presets, so `-t nuxt --pm bun` uses the Nuxt preset but overrides the package manager.

## Project Structure

```
main.go                         # Entrypoint; loads the embedded registry
cmd/
  root.go                       # Root command (launches interactive wizard)
  create.go                     # Create command (flag-based, named templates)
config/
  embed.go                      # //go:embed for registry.json and templates/
  registry.json                 # Single source of truth for all options & packages
  templates/
    base/                       # Framework templates (astro, nuxt, vite)
    css/                        # CSS templates (vanilla, tailwindcss)
    fmt/                        # Formatter configs (biome, prettier, oxfmt)
    linter/                     # Linter configs (biome, eslint, oxlint)
    form/                       # Form-library snippets
    integration/                # Per-integration snippets (react, vue)
    cms/                        # CMS integration snippets (microcms)
    pm/                         # Package manager config (npmrc, etc.)
    shared/                     # Shared files (husky, CLAUDE.md, AGENTS.md)
internal/
  tui/
    wizard.go                   # BubbleTea interactive wizard
    loading.go                  # Spinner during scaffolding
    success.go                  # Post-scaffold success box + warn helpers
    styles.go                   # Lip Gloss styles and color palette
    colors.go                   # Color tokens
pkg/
  config.go                     # ProjectConfig + typed enums + validation
  registry.go                   # Registry schema and global loader
  packagejson.go                # Data-driven package.json builder
  scaffold.go                   # Template rendering and file emission
```

## Development

### Build

```bash
go build -o bungkus-cli .
```

### Run locally

```bash
go run . create my-app --base vite-react --css tailwindcss --fmt biome
```

### Test

```bash
go test ./...
```

## License

MIT
