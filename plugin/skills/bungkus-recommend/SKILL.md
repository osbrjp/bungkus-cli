---
name: bungkus-recommend
description: Help someone who has not decided on a stack choose what to scaffold with bungkus-cli — ask a few questions about what they are building, recommend one combination of frontend, tooling and optional backend with a reason for each choice, then hand over to bungkus-scaffold. Use when the user wants to start a new web project but is unsure which framework or tools to pick, asks "what should I use", or asks to compare the options bungkus-cli offers.
---

# bungkus-recommend

Turn "I want to build X" into one recommended `bungkus-cli create` combination.
You only recommend what the installed binary offers; the scaffolding itself is
`bungkus-scaffold`'s job.

If the user already named their stack, skip this skill and use
`bungkus-scaffold` directly.

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
answered. Do not ask about tools that have a sensible default.

1. **What is it?** Mostly content (marketing site, blog, docs) or an
   interactive app (dashboard, tool behind a login)?
2. **React, Vue, or no preference?** Follow what the team already knows.
3. **Does it need its own API and database,** or does it only read from
   existing services or a CMS?
4. **Where will it be deployed?** Only matters if a listed deploy target fits.
5. **Is there a house standard** for package manager, formatter or linter?

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
- Everything else → keep the tool's default. Do not add an option because it
  exists.

Some options only work with React or with Vue. The tool enforces this and falls
back to `none`; do not recommend a mismatched pair.

### 4. Recommend one combination

Give a single recommendation, not a menu:

- A short list of each choice with a one-line reason tied to what the user
  said.
- What you deliberately left out and when they would add it (several tools can
  be added later with `bungkus-add`; others cannot, so say which matter now).
- The exact command, built from the closest `-t` preset plus only the flags
  that differ from it.

Offer at most one alternative, and only when a real trade-off exists.

### 5. Hand over

When the user agrees, continue with `bungkus-scaffold` using that command: it
confirms the target directory, runs it and reports the result. Do not run
`create` from this skill.

## Notes

- The mapping in step 3 is guidance, not a rule table. A stated team standard
  or an existing codebase convention wins over it.
- Never recommend a value that step 1 did not list.
