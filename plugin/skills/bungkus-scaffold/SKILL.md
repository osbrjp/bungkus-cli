---
name: bungkus-scaffold
description: Scaffold a new full-stack project with bungkus-cli — a frontend (Astro/Nuxt/Vite) with a batteries-included optional backend (Hono/Elysia + Drizzle/Prisma, database, /health-check, monorepo). Translate a plain-language request ("an Astro + React site with Tailwind and zod-validated forms", "a Nuxt app with a Hono + Postgres API deployed to Cloudflare") into the correct `bungkus-cli create` command. Use when the user asks to scaffold, bootstrap, generate, or start a new frontend or full-stack project, or names bungkus / bungkus-cli.
---

# bungkus-scaffold

Drive `bungkus-cli create` from a natural-language description. bungkus-cli is
a Go binary that scaffolds Astro / Nuxt / Vite frontends with opinionated
tooling, plus an optional batteries-included backend (Hono/Elysia, an ORM +
database, a `/health-check` endpoint, and a pnpm-workspace monorepo). Your job
is to map what the user asked for onto its flags — never guess flag values,
read them from the binary.

## Prerequisites

- `bungkus-cli` on PATH. Check with `bungkus-cli --help`. If missing, tell the
  user to build it (`go build -o bungkus-cli .` in the bungkus-cli repo, then
  put it on PATH) and stop.

## Steps

### 1. Read the real options

Never rely on memory for valid flag values — they change with the registry.
Run:

```
bungkus-cli create --help
```

That prints every flag with its allowed values and default, plus the list of
`-t/--template` presets. Treat this output as the single source of truth for
this run. If a value the user wants isn't listed, say so and offer the closest
listed option — do not invent flags.

### 2. Pick a preset, then override

Presets (`-t`) set a coherent bundle of choices; flags override individual
fields. A preset sets more than its name says — typically a CSS framework plus
validation, form, query and state libraries — and `--help` does not show what
each one bundles. So:

- Use a preset when the user asked for that kind of batteries-included setup,
  and override every bundled option they did not ask for (for example
  `--css vanilla --form none --query none --validation none --state none`).
- Use `--base` with explicit flags when the user named a lean or specific
  stack; nothing is implied that way.

Never tell the user what they will get from the flag defaults or from memory;
step 3's dry run is the only source for that.

- "Astro + React" → `-t astro-react`
- "Nuxt" → `-t nuxt`
- "Vite + Vue" → `-t vite-vue`

Then add the flags the user named. Example: user wants Astro+React but
with `bun` and `biome`:

```
bungkus-cli create my-app -t astro-react --pm bun --fmt biome
```

Integration compatibility is enforced by the tool: React-only add-ons
(react-hook-form) won't apply to a Vue base and vice-versa. If the user asks
for an incompatible combo, the CLI warns and falls back to `none` — surface
that to the user rather than pretending it worked.

### 3. Confirm before scaffolding

Scaffolding writes a new directory. First run the exact command with
`--dry-run` added: it writes nothing and prints the resolved config, one line
per flag, then the packages it would install.

```
bungkus-cli create <name> [flags] --dry-run
```

Check every line against the request. If it shows a CSS framework or library
the user did not ask for, add the override and dry-run again. Then show the
user the command, the target directory and the resolved config (every option,
including the bundled libraries), and run it once they agree. If the project
name / directory already exists, stop and ask — don't overwrite.

If the binary rejects `--dry-run` as an unknown flag, tell the user to run
`bungkus-cli update`; until then avoid `-t` presets and write every option
flag out explicitly.

### 4. Run and report

```
bungkus-cli create <name> [flags]
```

Report what was created: the base, the notable add-ons, the package manager,
and the `cd <name> && <pm> install` next step. If the command printed a
warning (e.g. an add-on fell back to `none`), quote it verbatim. The CLI's
post-scaffold summary also prints the exact dev/DB steps for the chosen combo
and the local URLs — relay them rather than re-deriving.

### 5. Full-stack & database next steps

Only relevant when a `--backend` and/or `--orm` was selected. Don't hardcode
these — the post-scaffold summary and the generated `AGENTS.md` are authoritative
— but know the shape so you can guide the user:

- A `--backend` (with `pnpm`) makes the project a **monorepo**: `apps/web`
  (frontend, `http://localhost:3000`), `apps/api` (backend, `http://localhost:8000`),
  and `packages/domain` (shared types/schemas). `<pm> dev` runs web + api together.
- Every backend exposes `GET /health-check`; with an ORM it also runs a
  read-only query and returns rows — the fastest way to confirm the DB is wired.
- Database setup (run in `apps/api`, or `<pm> --filter api <script>`):
  `cp .env.example .env` → (`docker compose up -d` for postgres/mysql) →
  `db:generate` → `db:migrate` → `db:seed` (inserts dummy rows).
- **Prisma gotcha**: `db:migrate` (`prisma migrate dev`) prompts for a migration
  name and **hangs in a non-interactive shell**. Pass `db:migrate -- --name <name>`,
  or use `prisma db push` to sync without a migration file.

### 6. Hand off to the project's own docs

Every scaffolded project ships an `AGENTS.md` (and a `CLAUDE.md` pointing to it),
a `.claude/` with a permission allowlist and slash-commands (`/verify`,
`/format-fix`, `/new-component`), and — with `--test playwright` — a `.mcp.json`.
`AGENTS.md` is tailored to the exact stack and is the source of truth for that
project. After scaffolding, tell the user (or the agent taking over) to read it
before making changes, instead of re-deriving conventions here.

## Notes

- No project name / no flags at all → the user may want the interactive TUI
  wizard (`bungkus-cli` with no args). That's interactive and can't be driven
  from here; tell the user to run it in their terminal directly.
- Don't scaffold into an existing non-empty directory. Confirm first.
- The flag list and preset names in this file are illustrative — step 1's
  `--help` output overrides anything here if they ever diverge.
