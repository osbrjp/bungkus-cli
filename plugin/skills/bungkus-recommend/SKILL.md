---
name: bungkus-recommend
description: Recommend a tech stack for a new web project from what the installed bungkus-cli can actually scaffold. MUST be used before suggesting or recommending any stack, framework or tooling for a new frontend or full-stack web project when bungkus-cli is available — never answer from general knowledge first. Use when the user says "suggest a tech stack", "recommend a stack", "what stack should I use for this project", "what should I use", "which framework", is unsure which framework or tools to pick, or asks to compare the options bungkus-cli offers. Asks a few questions, verifies the exact resolved config with a dry run, recommends one combination listing every option the user will get, then hands over to bungkus-scaffold.
---

# bungkus-recommend

Turn "I want to build X" into one recommended `bungkus-cli create` command
whose result you have verified. The scaffolding itself is `bungkus-scaffold`'s
job.

If the user already named their full stack, skip this skill and use
`bungkus-scaffold` directly.

## The rule: check before you recommend

**Never state what the user will get from memory, from this file, or from the
flag defaults in `--help`.** A `-t` preset sets several options at once
(typically a CSS framework plus validation, form, query and state libraries),
and `--help` shows only each flag's own default, not what a preset changes.
Recommending "the defaults" on top of a preset is how a user ends up with a
CSS framework and four libraries they did not ask for.

So, every time, before any recommendation is shown to the user:

1. Read the options the installed binary offers (step 1).
2. Run the exact command you intend to recommend with `--dry-run` and read the
   resolved config it prints (step 4).

If either cannot be done, say so and stop. Do not fall back to a guess.

## Steps

### 1. Read the real options

```
bungkus-cli create --help
```

This lists every flag with its allowed values and default, and the `-t`
presets. Recommend only values that appear there. If `bungkus-cli` is not on
PATH, tell the user to install it and stop.

### 2. Ask only what changes the answer

Ask the questions below in one message, skipping any the user already
answered.

1. **What is it?** Mostly content (marketing site, blog, docs) or an
   interactive app (dashboard, tool behind a login)?
2. **React, Vue, or no preference?** Follow what the team already knows.
3. **How should it be styled?** A utility CSS framework, or their own
   stylesheet. Ask this explicitly: every preset turns a CSS framework on.
4. **Does it need its own API and database,** or does it only read from
   existing services or a CMS?
5. **Where will it be deployed?** Only matters if a listed deploy target fits.
6. **Is there a house standard** for package manager, formatter or linter?

### 3. Map the answers

Each rule picks a *kind* of option; take the actual value from step 1.

- Content-first → an Astro base; add the React or Vue variant only if they need
  interactive components.
- Interactive app → a Vite base in their framework; Nuxt when they chose Vue
  and want routing and server rendering built in.
- Needs an API → a backend, plus an ORM and a database only if it stores data.
  A backend makes the project a monorepo by default; say so.
- Forms with user input → a validation library, and a form library for more
  than a couple of fields.
- Data fetched from an API on the client → a query library. Shared client
  state across pages → a state library that matches the framework. Otherwise
  leave both out.
- Deploy target and CI/CD → only when they named a target the tool supports.
- Anything the user did not ask for and the answers do not call for → `none`
  (or the plain option). Do not add an option because it exists or because a
  preset includes it.

Some options only work with React or with Vue. The tool enforces this and falls
back to `none`; do not recommend a mismatched pair.

### 4. Build the command and verify it with a dry run

Build the command one of two ways:

- **`--base` plus explicit flags** — nothing is implied. Prefer this when the
  user wants a lean setup or their own choices.
- **`-t <preset>` plus overrides** — only when the preset's bundle is mostly
  what the user wants. Every bundled option they do not want needs an explicit
  override, for example `--css vanilla --form none --query none
  --validation none --state none`.

Then run it with `--dry-run` added. It writes nothing and prints the resolved
config, one line per flag, followed by the packages it would install:

```
bungkus-cli create <name> <flags> --dry-run
```

Compare every line against what the user asked for. If a line shows something
they did not ask for, add the override and run the dry run again. Repeat until
the resolved config matches. Only then recommend.

If the binary rejects `--dry-run` as an unknown flag, it is too old to show a
resolved config: tell the user to run `bungkus-cli update`. Until they do, use
only the `--base` form with **every** option flag written out explicitly, never
a `-t` preset, and say that the result could not be verified.

### 5. Recommend one combination

Give a single recommendation, not a menu:

- **Every resolved option, as the dry run printed it** — base, CSS, formatter,
  linter, validation, form, query, state, CMS, test, audit, desktop, deploy,
  CI/CD, backend, ORM, database, layout, package manager — including the ones
  set to `none`. One line each, with a reason tied to what the user said where
  you made a choice.
- **The packages** the dry run listed for anything beyond the framework itself,
  so libraries a preset bundles are visible by name.
- What you deliberately left out and when they would add it (several tools can
  be added later with `bungkus-add`; others cannot, so say which matter now).
- The exact command you dry-ran, without `--dry-run`.

Offer at most one alternative, and only when a real trade-off exists.

### 6. Hand over

When the user agrees, continue with `bungkus-scaffold` using that exact
command: it confirms the target directory, runs it and reports the result. Do
not run `create` without `--dry-run` from this skill.

## Notes

- The mapping in step 3 is guidance, not a rule table. A stated team standard
  or an existing codebase convention wins over it.
- Never recommend a value that step 1 did not list, and never describe a result
  that step 4 did not print.
