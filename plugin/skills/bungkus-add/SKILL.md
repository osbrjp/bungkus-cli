---
name: bungkus-add
description: Add a tool to an existing frontend project with bungkus-cli — a formatter, linter, Playwright tests, Lighthouse CI, a Cloudflare deploy config, or a GitHub Actions workflow — without touching the project's code. Translate a plain-language request ("add Biome to this project", "set up GitHub Actions deploy to Cloudflare Pages") into the correct `bungkus-cli add` command. Use when the user asks to add, install, or set up one of these tools in a project that already exists, or names `bungkus-cli add`.
---

# bungkus-add

Drive `bungkus-cli add` from a natural-language description. `add` renders an
option's config files into an existing project and merges its dependencies and
scripts into `package.json`. It never overwrites: an existing file, dependency
version or script is kept and reported as skipped. Your job is to map the
request onto one option and the right directory — never guess option names,
read them from the binary.

For a project that does not exist yet, use `bungkus-scaffold` instead.

## Prerequisites

- `bungkus-cli` on PATH. Check with `bungkus-cli --help`. If missing, tell the
  user to install it (see the bungkus-cli README) and stop.
- A `package.json` in the target directory. `add` refuses without one.

## Steps

### 1. Read the real options

```
bungkus-cli add
```

With no option this lists everything that can be added, grouped by category,
generated from the tool's registry. Treat it as the single source of truth for
this run. If what the user wants isn't listed, say so and offer the closest
listed option — do not invent one. `bungkus-cli add --help` lists the flags.

### 2. Pick the directory

`add` works on the directory that holds the `package.json` to change: the
current directory by default, or the optional second argument.

- Flat project → the project root.
- Monorepo (`apps/web`, `apps/api`) → the app the tool belongs to, usually
  `apps/web`. Ask if it is not obvious.

GitHub Actions workflows are the exception the tool handles itself: they are
always written to the git repository root, wherever you run it.

### 3. Let it detect; override only on request

The package manager is detected from `package.json`'s `packageManager` field or
a lockfile, and the base framework from the dependencies. Pass a flag only when
the tool says it could not detect something:

- `--pm` — the error names it when there is no lockfile, or several.
- `--base` — the error names it when the framework can't be told from
  `package.json`.
- `--deploy` — CI/CD options need a deploy target. It is detected from an
  existing wrangler config; otherwise ask the user which target they deploy to
  rather than choosing for them.

### 4. Confirm, then run

Show the exact command and the directory it will change, then run it:

```
bungkus-cli add <option> [dir] [flags]
```

Add one option per command. For several tools, run them one after another, the
deploy target before the CI/CD workflow so the second needs no `--deploy`.

### 5. Report what happened

Relay the tool's report rather than re-deriving it:

- `created` files, and any marked `(repo root)`.
- `skipped` files and `=` lines in the `package.json` block: these already
  existed and were kept. Say so plainly — the tool did not merge them, so the
  user may need to reconcile their existing config by hand.
- Any `note:` or `warning:` line, quoted verbatim (a workflow relocated to the
  repo root in a monorepo; no git repository found).
- The install command it prints when `package.json` changed. Run it only if
  the user asks.

"Nothing to do" means everything the option provides was already there.

## Notes

- `add` only covers standalone tools: config files plus dependencies. It does
  not add things that need edits inside the app's source (CSS framework, form,
  state or query libraries, a backend). For those, say it is not supported by
  `add` and point to the library's own setup guide.
- The option names in this file's description are illustrative — step 1's
  listing overrides anything here if they ever diverge.
