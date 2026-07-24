# Bungkus-cli

A scaffolding CLI for modern frontend projects — with an optional backend
(Hono / Elysia), ORM + database (Drizzle / Prisma), and a pnpm-workspace
monorepo layout when you build full-stack.

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

Scaffold full-stack — a backend turns the project into a pnpm-workspace monorepo:

```bash
bungkus-cli create my-app --base astro-react --backend hono --orm drizzle --db postgres
```

#### Flags

| Flag             | Default      | Options                                                                       |
| :--------------- | :----------- | :---------------------------------------------------------------------------- |
| `--base`         | `astro`      | `astro`, `astro-vue`, `astro-react`, `nuxt`, `vite`, `vite-vue`, `vite-react` |
| `--css`          | `vanilla`    | `vanilla`, `tailwindcss`                                                      |
| `--fmt`          | `biome`      | `biome`, `prettier`, `oxfmt`                                                  |
| `--linter`       | `biome`      | `biome`, `eslint`, `oxlint`                                                   |
| `--validation`   | `none`       | `none`, `zod`                                                                 |
| `--form`         | `none`       | `none`, `react-hook-form`, `tanstack-form`, `veevalidate`                     |
| `--query`        | `none`       | `none`, `tanstack-query`                                                      |
| `--state`        | `none`       | `none`, `jotai`, `zustand`, `pinia`, `nanostores`                            |
| `--test`         | `none`       | `none`, `playwright`                                                          |
| `--audit`        | `none`       | `none`, `lhci`                                                                |
| `--cms`          | `none`       | `none`, `microcms`                                                            |
| `--backend`      | `none`       | `none`, `hono`, `elysia`                                                      |
| `--orm`          | `none`       | `none`, `drizzle`, `prisma`                                                   |
| `--db`           | `none`       | `none`, `sqlite`, `postgres`, `mysql`, `d1`                                   |
| `--layout`       | `flat`       | `flat`, `monorepo`                                                            |
| `--deploy`       | `none`       | `none`, `cloudflare-pages`, `cloudflare-workers`                              |
| `--cicd`         | `none`       | `none`, `github-actions`                                                      |
| `--pm`           | `pnpm`       | `pnpm`, `bun`, `npm`, `yarn`                                                  |
| `--channel`      | `pinned`     | `pinned` (vetted, ≥14d old & safe), `latest`                                 |
| `--pin`          | `default`    | `default`, `caret`, `tilde`, `exact`                                          |
| `--install`      | `false`      | run the package manager install after scaffolding                            |
| `--git`          | `true`       | initialize a git repo with an initial commit                                 |
| `--node-engine`  | `>=22.12.0`  | `package.json` `engines.node` constraint                                     |
| `-t, --template` | —            | `astro`, `astro-react`, `astro-vue`, `nuxt`, `vite`, `vite-react`, `vite-vue` |

Flags take precedence over template presets, so `-t nuxt --pm bun` uses the Nuxt preset but overrides the package manager.

Combination rules the CLI enforces:

- `--cicd` requires `--deploy` (using `--cicd github-actions` without a deploy target is an error).
- `--db` requires `--orm`; `--db d1` is only supported with `--orm drizzle`.
- `--layout monorepo` currently requires `--pm pnpm`.
- Selecting a `--backend` (with pnpm) defaults `--layout` to `monorepo`; pass `--layout flat` to keep everything in one package.
- `project-name` must be a valid npm package name (lowercase letters, digits, `.`, `-`, `_`, starting with a letter or digit); use `.` to scaffold into the current directory.

### Backend & full-stack (monorepo)

When you select a `--backend`, the project is scaffolded as a **pnpm-workspace monorepo**:

```
my-app/
  package.json          # private workspace root (pnpm -r dev/build, husky)
  pnpm-workspace.yaml
  apps/
    web/                 # the frontend (astro/nuxt/vite) + its tooling
    api/                 # the backend (hono/elysia) + orm/db
  packages/
    domain/              # shared contract: zod schemas (with --validation zod) or plain types
```

- **`--backend hono`** runs on Node via `tsx`; **`--backend elysia`** runs on Bun. `pnpm dev` runs `apps/web` (`http://localhost:3000`) and `apps/api` (`http://localhost:8000`) together. Every backend exposes `GET /health-check`; with an ORM selected it also runs a read-only query against the database and returns the rows, so you can confirm the DB wiring end-to-end.
- **`--orm drizzle` / `--orm prisma`** add the config, a `db/` client, `.env.example`, and `db:generate` / `db:migrate` / `db:seed` scripts under `apps/api`. `db:seed` inserts a couple of dummy rows so a fresh DB (and `/health-check`) returns real data. `web` and `api` both depend on `packages/domain` via `workspace:*`.
- **`--db postgres` / `--db mysql`** also generate a root `docker-compose.yml` whose credentials match `.env.example`, so `docker compose up -d` gives you a working database. `sqlite` needs nothing extra; `d1` targets Cloudflare Workers (pair with `--deploy cloudflare-workers`).

The post-scaffold summary prints the get-started steps for your exact combo (install, dev, and — when a server database is selected — `docker compose up -d` plus the `db:generate` / `db:migrate` commands).

### AI-agent-ready

Every scaffolded project ships with files that make it work well with Claude Code and other AI agents out of the box:

- **`AGENTS.md`** (and a `CLAUDE.md` pointing to it) — describes your *exact* stack: real commands, both dev URLs, the monorepo map, the ORM/DB workflow, and the `/health-check` probe.
- **`.claude/settings.json`** — a permission allowlist for routine dev commands (package manager, `docker compose` when a server DB is selected, `wrangler` for Cloudflare, read-only git) so agents don't stall on prompts.
- **`.claude/commands/`** — project slash-commands: `/verify` (typecheck + build + test, and curl the health-check), `/format-fix`, and `/new-component` (follows the repo's naming + JSDoc conventions).
- **`.mcp.json`** — with `--test playwright`, a Playwright MCP server so an agent can drive the running app in a browser.

## Project Structure

```
main.go                         # Entrypoint; loads the embedded registry
cmd/
  root.go                       # Root command (launches interactive wizard)
  create.go                     # Create command (flag-based, named templates)
  bump.go                       # Maintainer-only version bump (//go:build bump)
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
    backend/                    # Backend servers (hono, elysia)
    orm/                        # ORM config + db client (drizzle, prisma)
    database/                   # docker-compose for server DBs (postgres, mysql)
    monorepo/                   # Workspace pieces (root, api tsconfig, domain src)
    deploy/                     # Deploy configs (wrangler.jsonc per target)
    cicd/                       # CI/CD workflows (github-actions/<target>/)
    pm/                         # Package manager config (pnpm-workspace.yaml, .npmrc)
    shared/                     # Shared files (husky, CLAUDE.md, AGENTS.md)
internal/
  tui/
    wizard.go                   # BubbleTea interactive wizard (Frontend / Backend tabs)
    loading.go                  # Spinner during scaffolding
    success.go                  # Post-scaffold success box + warn helpers
    styles.go                   # Lip Gloss styles and color palette
    colors.go                   # Color tokens
pkg/
  config.go                     # ProjectConfig + typed enums
  registry.go                   # Registry schema and global loader
  packagejson.go                # Data-driven package.json builder (web/api/domain/root)
  scaffold.go                   # Template rendering and file emission
  validate.go                   # Project-name / destination validation
  bump.go                       # Version-bump resolution (used by cmd/bump.go)
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
