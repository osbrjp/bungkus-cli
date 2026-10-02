# Bungkus-cli

A scaffolding CLI for modern frontend projects — with an optional backend
(Hono / Elysia), ORM + database (Drizzle / Prisma), and a workspace
monorepo layout (pnpm, bun, npm or yarn) when you build full-stack.

## Getting Started

### Install

Download the latest release binary for your platform (`darwin`/`linux` × `arm64`/`amd64`) and verify its SHA256:

```bash
curl -fsSL https://raw.githubusercontent.com/osbrjp/bungkus-cli/main/install.sh | bash
```

Installs to `~/.local/bin/bungkus-cli`, so no `sudo` is needed; if that folder
isn't on your `PATH`, the installer prints the line to add. When bungkus-cli is
already installed, it updates that copy in place. Choose another folder with
`BUNGKUS_INSTALL_DIR` (set it on `bash`, which runs the script):

```bash
curl -fsSL https://raw.githubusercontent.com/osbrjp/bungkus-cli/main/install.sh | BUNGKUS_INSTALL_DIR=/opt/bin bash
```

Installed in `/usr/local/bin` by an older installer? The next update goes to
`~/.local/bin` instead of asking for `sudo`, as long as `~/.local/bin` comes
before `/usr/local/bin` on your `PATH`, and tells you how to remove the old copy.
Otherwise it updates `/usr/local/bin` with `sudo`, as before.

Update to the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/osbrjp/bungkus-cli/main/update.sh | bash
```

Confirm:

```bash
bungkus-cli --version
```

If you have Go installed and prefer to build from source:

```bash
go install github.com/osbrjp/bungkus-cli@latest
```

This puts the binary in `$GOBIN` (default `~/go/bin`). Keep one install
method: if bungkus-cli is also installed with the script above, only the copy
that comes first on your `PATH` runs (check with `type -a bungkus-cli` in fish,
`which -a bungkus-cli` in zsh/bash). `bungkus-cli update` lists the copies when
there is more than one. For local development, build without installing so the
dev build never shadows the real one:

```bash
go build -o /tmp/bk . && /tmp/bk
```

### Updating

```bash
bungkus-cli update          # replace this binary with the latest release
bungkus-cli update --check  # report what is available, install nothing
```

`update` updates the way bungkus-cli was installed:

- **Install script:** re-runs it, so downloads stay checksum-verified and the
  copy you run is replaced in place (no `sudo`; see Install).
- **`go install`:** runs `go install github.com/osbrjp/bungkus-cli@<latest>`,
  which updates the copy in `$GOBIN`. Without `go` on `PATH` it prints that
  command instead.

`bungkus-cli --version` reports the release version for both; a local
`go build` reports `dev`.

Once a day, other commands check for a newer release in the background and
print a one-line hint when one exists. The check never blocks or fails a
command; set `BUNGKUS_NO_UPDATE_CHECK=1` to turn it off.

### Colours

On truecolor terminals the wizard paints the Daun Pisang leaf-green
background while it runs and restores your terminal's colours on exit.
Set `BUNGKUS_BACKGROUND=terminal` to keep your own background. 256- and
16-colour terminals always keep their own background.

### Usage

Run the interactive wizard:

```bash
bungkus-cli
```

The wizard lists every choice in a steps pane, edits the selected one in an
options pane (incompatible options stay visible, greyed, with the reason), and
previews the equivalent `bungkus-cli create …` command, layout and
dependencies. Keys: `j/k` or arrows move, `h/l` or `tab` switch pane, `space`
picks, `enter` picks and moves on, typing a step's number (`1`–`21`, shown
beside each step) jumps to it, `0` or `r` jumps to review, `?` lists every key.
Under 100 columns it shows one pane at a time.

Or use the `create` command with flags:

```bash
bungkus-cli create my-app --base astro-react --css tailwindcss --fmt biome --pm pnpm
```

Start from a named template and override individual options with flags:

```bash
bungkus-cli create my-app -t nuxt --pm bun
```

Scaffold full-stack — a backend turns the project into a workspace monorepo:

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
| `--dry-run`      | `false`      | print the resolved config and its packages, then exit without writing        |
| `-t, --template` | —            | `astro`, `astro-react`, `astro-vue`, `nuxt`, `vite`, `vite-react`, `vite-vue` |

Flags take precedence over template presets, so `-t nuxt --pm bun` uses the Nuxt preset but overrides the package manager. A preset sets several options at once (for example a CSS framework plus validation, form, query and state libraries); add `--dry-run` to see exactly what a preset plus your flags resolves to before anything is written.

Combination rules the CLI enforces:

- `--cicd` requires `--deploy` (using `--cicd github-actions` without a deploy target is an error).
- `--db` requires `--orm`; `--db d1` is only supported with `--orm drizzle`.
- Selecting a `--backend` defaults `--layout` to `monorepo` (with any `--pm`); pass `--layout flat` to keep everything in one package.
- `project-name` must be a valid npm package name (lowercase letters, digits, `.`, `-`, `_`, starting with a letter or digit); use `.` to scaffold into the current directory.

### Backend & full-stack (monorepo)

When you select a `--backend`, the project is scaffolded as a **workspace monorepo** for the chosen package manager:

```
my-app/
  package.json          # private workspace root (dev/build for every app, husky)
  pnpm-workspace.yaml   # pnpm only; bun/npm/yarn list workspaces in package.json
  apps/
    web/                 # the frontend (astro/nuxt/vite) + its tooling
    api/                 # the backend (hono/elysia) + orm/db
  packages/
    domain/              # shared contract: zod schemas (with --validation zod) or plain types
```

- **`--backend hono`** runs on Node via `tsx`; **`--backend elysia`** runs on Bun. `<pm> run dev` runs `apps/web` (`http://localhost:3000`) and `apps/api` (`http://localhost:8000`) together. Every backend exposes `GET /health-check`; with an ORM selected it also runs a read-only query against the database and returns the rows, so you can confirm the DB wiring end-to-end.
- **`--orm drizzle` / `--orm prisma`** add the config, a `db/` client, `.env.example`, and `db:generate` / `db:migrate` / `db:seed` scripts under `apps/api`. `db:seed` inserts a couple of dummy rows so a fresh DB (and `/health-check`) returns real data. `web` and `api` both depend on `packages/domain` (`@repo/domain`).
- **Per package manager**: the root scripts and the `@repo/domain` link follow `--pm`.

  | `--pm` | Workspace list        | `@repo/domain` | Root `dev`                                     | Run one app's script          |
  | :----- | :-------------------- | :------------- | :--------------------------------------------- | :---------------------------- |
  | pnpm   | `pnpm-workspace.yaml` | `workspace:*`  | `pnpm --recursive --parallel run dev`          | `pnpm --filter api run <s>`   |
  | bun    | `workspaces` field    | `workspace:*`  | `bun run --filter '*' dev`                     | `bun run --filter api <s>`    |
  | npm    | `workspaces` field    | `*`            | `concurrently` over `npm run dev -w <app>`     | `npm run <s> -w api`          |
  | yarn   | `workspaces` field    | `*`            | `concurrently` over `yarn workspace <app> run dev` | `yarn workspace api run <s>` |

  npm has no `workspace:` protocol and yarn 1 rejects it; `*` links the local workspace on npm and on yarn classic and Berry alike. npm and yarn run workspace scripts one at a time, so the root `dev` uses `concurrently` to keep both dev servers up.
- **`--db postgres` / `--db mysql`** also generate a root `docker-compose.yml` whose credentials match `.env.example`, so `docker compose up -d` gives you a working database. `sqlite` needs nothing extra; `d1` targets Cloudflare Workers (pair with `--deploy cloudflare-workers`).

The post-scaffold summary prints the get-started steps for your exact combo (install, dev, and — when a server database is selected — `docker compose up -d` plus the `db:generate` / `db:migrate` commands).

### AI-agent-ready

Every scaffolded project ships with files that make it work well with Claude Code and other AI agents out of the box:

- **`AGENTS.md`** (and a `CLAUDE.md` pointing to it) — describes your *exact* stack: real commands, both dev URLs, the monorepo map, the ORM/DB workflow, and the `/health-check` probe.
- **`.claude/settings.json`** — a permission allowlist for routine dev commands (package manager, `docker compose` when a server DB is selected, `wrangler` for Cloudflare, read-only git) so agents don't stall on prompts.
- **`.claude/commands/`** — project slash-commands: `/verify` (typecheck + build + test, and curl the health-check), `/format-fix`, and `/new-component` (follows the repo's naming + JSDoc conventions).
- **`.mcp.json`** — with `--test playwright`, a Playwright MCP server so an agent can drive the running app in a browser.

### Agent plugin

The `bungkus` plugin lets Claude Code or Codex turn a plain request ("an Astro + React site with Tailwind") into the right `bungkus-cli create` command, add a tool to an existing project with `bungkus-cli add`, and recommend a stack when you haven't picked one. It also sets up [bungkus-mc](https://github.com/osbrjp/bungkus-mc), the terminal mission control for agent sessions. This repo is its marketplace:

```bash
# Claude Code
claude plugin marketplace add osbrjp/bungkus-cli@release
claude plugin install bungkus@bungkus-cli

# Codex
codex plugin marketplace add osbrjp/bungkus-cli@release
codex plugin add bungkus@bungkus-cli
```

`@release` pins the marketplace to the stable branch, so the plugin changes only with a stable release. Without it the marketplace follows `main`, the canary line, which is what you want only when testing the plugin before a release.

One machine holds one copy of the marketplace. To switch between stable and canary, remove it and add it again:

```bash
claude plugin marketplace remove bungkus-cli   # Codex: codex plugin marketplace remove bungkus-cli
claude plugin marketplace add osbrjp/bungkus-cli@release
claude plugin install bungkus@bungkus-cli
```

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
    pm/                         # Package manager config (pnpm-workspace.yaml, .npmrc, .yarnrc.yml)
    shared/                     # Shared files (husky, CLAUDE.md, AGENTS.md)
internal/
  tui/
    wizard.go                   # BubbleTea wizard model: keys, picks, in-pane scaffolding
    steps.go                    # Step list, registry-derived options, equivalent create command
    view.go                     # Bento layout: header, steps/options/preview panes, status bar
    mascot.go                   # Half-block mascot and its ducking animation
    success.go                  # Post-scaffold success box + warn helpers
    styles.go                   # Lip Gloss styles and color palette
pkg/
  config.go                     # ProjectConfig + typed enums
  registry.go                   # Registry schema and global loader
  packagejson.go                # Data-driven package.json builder (web/api/domain/root)
  scaffold.go                   # Template rendering and file emission
  validate.go                   # Project-name / destination validation
  bump.go                       # Version-bump resolution (used by cmd/bump.go)
plugin/                         # `bungkus` Claude Code / Codex plugin (scaffold, add, recommend, mc-setup skills)
.claude-plugin/marketplace.json # Claude Code marketplace listing the plugin
.agents/plugins/marketplace.json # Codex marketplace listing the plugin
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

### Plugin

`plugin/` holds the `bungkus` plugin, and this repo is its marketplace (`.claude-plugin/marketplace.json` for Claude Code, `.agents/plugins/marketplace.json` for Codex). It has four skills: `bungkus-scaffold` turns a plain request into a `bungkus-cli create` command, `bungkus-add` does the same for `bungkus-cli add` on an existing project, `bungkus-recommend` helps someone who hasn't chosen a stack pick one, and `bungkus-mc-setup` installs and configures bungkus-mc. The content of `bungkus-mc-setup` is owned by the bungkus-mc repo (`skills/bungkus-mc-setup/SKILL.md` there): change it there first, then copy the file here and bump the plugin version. When a `create` or `add` flag, option or preset changes, update the matching skill under `plugin/skills/` in the same PR and bump `version` in both `plugin/.claude-plugin/plugin.json` and `plugin/.codex-plugin/plugin.json`.

## License

MIT
