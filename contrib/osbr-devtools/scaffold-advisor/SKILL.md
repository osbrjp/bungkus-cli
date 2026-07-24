---
name: scaffold-advisor
description: Turn a plain-English project brief into the right `bungkus-cli create` command and run it. Use when someone wants to start a new frontend/full-stack project, asks "which bungkus-cli flags should I use", or describes an app they want scaffolded.
---

# bungkus-cli scaffold advisor

Map a plain-English brief to a single `bungkus-cli create` command, confirm it, then run it.

## Method

1. Read the brief for signals: framework, interactivity, styling, backend/DB needs, deploy target, testing.
2. Choose flags using the tables below. When a signal is missing, take the **default** — don't interrogate the user for every field.
3. Show the assembled command and a one-line rationale. Run it once confirmed.

## Flags

| Flag | Values | Default | Pick when |
|------|--------|---------|-----------|
| `--base` | `astro`, `astro-react`, `astro-vue`, `nuxt`, `vite`, `vite-react`, `vite-vue` | `astro` | Content/marketing → `astro`; interactive islands → `astro-react`/`astro-vue`; Vue app → `nuxt`; SPA → `vite-react`/`vite-vue` |
| `--css` | `vanilla`, `tailwindcss` | `vanilla` | Utility-first styling → `tailwindcss` |
| `--fmt` | `prettier`, `biome`, `oxfmt` | `biome` | Team standard; `biome` doubles as linter |
| `--linter` | `biome`, `eslint`, `oxlint` | `biome` | Match `--fmt` unless a specific linter is required |
| `--validation` | `none`, `zod` | `none` | Runtime schema validation / typed forms → `zod` |
| `--form` | `none`, `tanstack-form` | `none` | Complex forms (react/vue base only) |
| `--query` | `none`, `tanstack-query` | `none` | Client data-fetching/caching (react/vue base only) |
| `--state` | `none`, `jotai`, `zustand`, `pinia`, `nanostores` | `none` | react → `jotai`/`zustand`; vue/nuxt → `pinia`; framework-agnostic → `nanostores` |
| `--cms` | `none`, `microcms` | `none` | microCMS content (not supported on Vite bases) |
| `--backend` | `none`, `hono`, `elysia` | `none` | Needs an API. `hono` (Node) is the safe default; `elysia` runs on Bun |
| `--orm` | `none`, `drizzle`, `prisma` | `none` | Needs a database |
| `--db` | `none`, `sqlite`, `postgres`, `mysql`, `d1` | `none` | Requires `--orm`. Local/simple → `sqlite`; production → `postgres`; Cloudflare → `d1` (drizzle only) |
| `--deploy` | `none`, `cloudflare-pages`, `cloudflare-workers` | `none` | Cloudflare Pages for static/SSR, Workers for edge/API |
| `--test` | `none`, `playwright` | `none` | Wants e2e tests (also adds a Playwright MCP for agents) |
| `--pm` | `bun`, `npm`, `yarn`, `pnpm` | `pnpm` | Team default `pnpm`; monorepo requires `pnpm` |
| `--layout` | `flat`, `monorepo` | auto | Auto-upgrades to `monorepo` when a backend is set with `pnpm` |
| `-t/--template` | `astro`, `astro-react`, `astro-vue`, `nuxt`, `vite`, `vite-react`, `vite-vue` | — | Preset shortcut; individual flags still override |

## Compatibility rules (don't emit invalid combos)

- `--db` requires `--orm`; `--db d1` only with `--orm drizzle`.
- `--form`/`--query` need a React or Vue base (Nuxt counts as Vue). On a plain `astro`/`vite` base they fall back to `none`.
- `--state`: `jotai`/`zustand` → React bases; `pinia` → Vue/Nuxt; `nanostores` → any.
- `--cms microcms` is not supported on Vite bases.
- A backend (with `pnpm`) makes it a monorepo: `apps/web` + `apps/api` + `packages/domain`.

## Examples

- "Marketing site, Tailwind" → `bungkus-cli create site --base astro --css tailwindcss`
- "React dashboard with forms and data fetching" → `bungkus-cli create app --base vite-react --css tailwindcss --validation zod --form tanstack-form --query tanstack-query --state zustand`
- "Full-stack app, Postgres, deploy to Cloudflare" → `bungkus-cli create app --base astro-react --backend hono --orm drizzle --db postgres --deploy cloudflare-pages --test playwright`
- "Vue app on Cloudflare Workers with D1" → `bungkus-cli create app --base nuxt --backend hono --orm drizzle --db d1 --deploy cloudflare-workers`

After scaffolding, the project's own `AGENTS.md` documents the exact dev/DB/deploy workflow — read it before making changes.
