# The OSBR Standard Repository

A template repository for creating standardized repositories over organization.

## Getting Started

### Install

Download the latest release binary for your platform (`darwin`/`linux` × `arm64`/`amd64`) and verify its SHA256:

```bash
curl -fsSL https://raw.githubusercontent.com/spencer-osbrjp/bungkus-cli/main/install.sh | bash
```

Defaults to `/usr/local/bin/bungkus-cli` (uses `sudo` if needed). Override with `BUNGKUS_INSTALL_DIR`:

```bash
BUNGKUS_INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/spencer-osbrjp/bungkus-cli/main/install.sh | bash
```

Update to the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/spencer-osbrjp/bungkus-cli/main/update.sh | bash
```

Confirm:

```bash
bungkus-cli --version
```

If you have Go installed and prefer to build from source:

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
