# bungkus-cli
```
88""Yb 88   88 88b 88  dP""b8 88  dP 88   88 .dP"Y8      dP""b8 88     88
88__dP 88   88 88Yb88 dP   `" 88odP  88   88 `Ybo."     dP   `" 88     88
88""Yb Y8   8P 88 Y88 Yb  "88 88"Yb  Y8   8P o.`Y8b     Yb      88  .o 88
88oodP `YbodP' 88  Y8  YboodP 88  Yb `YbodP' 8bodP'      YboodP 88ood8 88
```

A Go CLI tool that scaffolds and configures modern frontend projects with common tooling. Single binary, no external runtime dependencies.

## Features

- **Frameworks** -- Astro, Vite
- **CSS** -- Vanilla, Tailwind CSS
- **Formatters** -- Prettier, Biome
- **Package Managers** -- bun, npm, yarn, pnpm
- **Git** -- Optional git init with Husky pre-commit hooks
- **Interactive TUI** -- Guided wizard when run without arguments

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
bungkus-cli create my-app --base astro --css tailwindcss --fmt prettier --pm bun
```

#### Flags

| Flag       | Default      | Options                      |
| :--------- | :----------- | :--------------------------- |
| `--base`   | `astro`      | `astro`, `vite`              |
| `--css`    | `vanilla`    | `vanilla`, `tailwindcss`     |
| `--fmt`    | `prettier`   | `prettier`, `biome`          |
| `--pm`     | `bun`        | `bun`, `npm`, `yarn`, `pnpm` |
| `--no-git` | `false`      | Skip git initialization      |

## Project Structure

```
main.go                         # Entrypoint
cmd/
  root.go                       # Root command (interactive wizard)
  create.go                     # Create command (flag-based)
  embed.go                      # Go embed directives
config/
  embed.go                      # Embedded config/templates FS
  templates/
    base/                       # Framework templates (astro, vite)
    css/                        # CSS templates (vanilla, tailwindcss)
    fmt/                        # Formatter configs (prettier, biome)
    shared/                     # Shared files (husky, CLAUDE.md, AGENTS.md)
internal/
  tui/
    wizard.go                   # BubbleTea interactive wizard
    loading.go                  # Spinner during scaffolding
    styles.go                   # Lip Gloss styles
pkg/
  config.go                     # Project config types & validation
  scaffold.go                   # Scaffolding logic
```

## Development

### Build

```bash
go build -o bungkus-cli .
```

### Run locally

```bash
go run . create my-app --base vite --css tailwindcss --fmt biome
```

## License

MIT
