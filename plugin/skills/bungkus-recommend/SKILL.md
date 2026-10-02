---
name: bungkus-recommend
description: Recommend a tech stack for a new web project from what the installed bungkus-cli can actually scaffold. MUST be used before suggesting or recommending any stack, framework or tooling for a new frontend or full-stack web project when bungkus-cli is available — never answer from general knowledge first. Use when the user says "suggest a tech stack", "recommend a stack", "what stack should I use for this project", "what should I use", "which framework", is unsure which framework or tools to pick, or asks to compare the options bungkus-cli offers. Treats the user as non-technical: clears up what they are building with plain-language questions (each with an "I'm not sure" answer), never asks them to choose a technology, verifies the exact resolved config with a dry run, then recommends one stack listing every option they will get and hands over to bungkus-scaffold.
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

### 2. Understand the project first, in plain language

Treat the user as someone who is not technical. They know what they want to
build, not what it is built with. **Never ask them to choose a technology**
(a framework, a CSS approach, a library, a database). Choosing is your job.

Before any question about the stack, make sure you understand the project. If
the request is vague ("a site for my shop", "an app for my team"), ask what it
is for and what a visitor does on it, until you could describe it back in one
sentence. Do not move on with a guess.

Then ask the questions below. Rules for asking:

- Plain words only. No framework, library or tool names in a question.
- Every question offers **"I'm not sure"** as an answer, and that answer is
  always acceptable. Never push for a decision.
- Skip any question the user has already answered.
- Ask them together, in one go, as choices to pick from. Use the agent's
  question tool with selectable options when it has one; otherwise a short
  numbered list.

1. **What will people mostly do there?**
   Read information (a company site, a blog, documentation) / Do things (log
   in, fill in data, manage something) / A bit of both / I'm not sure
2. **Will visitors type information in?**
   No / A simple form or two, like "contact us" / Yes, many forms or long ones
   / I'm not sure
3. **Does it need to remember its own information,** such as accounts, orders
   or bookings?
   No / Yes / The information already lives in another system / I'm not sure
4. **Who will update the text and pictures later?**
   A developer / People who do not code, through an editing screen / I'm not
   sure
5. **How should it look?**
   We have our own design or a particular look in mind / We just want it to
   look tidy quickly / I'm not sure
6. **Is there a team that already builds things a certain way?**
   No / Yes (ask what they use, in their words) / I'm not sure
7. **Do you know where it will be put online?**
   Not yet / Yes (ask where) / I'm not sure

If an answer is unclear or two answers conflict, ask one short follow-up before
going on. Do not ask anything else.

### 3. Map the answers

Each rule picks a *kind* of option; take the actual value from step 1.

- Mostly reading → an Astro base; add the React or Vue variant only if there
  are interactive parts.
- Mostly doing things → a Vite base with React or Vue; Nuxt when the team uses
  Vue and wants routing and server rendering built in.
- React or Vue → what the team already uses. No team, or not sure → React.
- Remembers its own information → a backend with an ORM and a database. A
  backend makes the project a monorepo by default; say so in plain words ("the
  website and its server live in one project").
- Information lives in another system → no backend; a query library if pages
  load that information in the browser.
- Many or long forms → a validation library and a form library. A simple form
  or two → validation only. None → neither.
- People who do not code will edit content → a CMS, if step 1 lists one.
- Own design or a particular look → plain CSS. Tidy quickly → the CSS
  framework.
- Named a place to put it online → the deploy target, and CI/CD, only when the
  tool lists that place. Otherwise leave both out.
- A team standard for package manager, formatter or linter → use it.

**"I'm not sure" means: take the simplest option** — the tool's own default,
which for an add-on is `none` and for styling is plain CSS — and say in the
recommendation that you chose it and that it can change later. Never add an
option because it exists or because a preset includes it.

Some options only work with React or with Vue. The tool enforces this and falls
back to `none`; do not pick a mismatched pair.

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

### 5. Recommend the stack

You decide; the user does not pick. Give **one** recommendation, with no menu
of alternatives and no "which would you prefer":

- **In plain words first:** two or three sentences on what they will get and
  what it lets them do, using their own description of the project.
- **Then every resolved option, as the dry run printed it** — base, CSS,
  formatter, linter, validation, form, query, state, CMS, test, audit, desktop,
  deploy, CI/CD, backend, ORM, database, layout, package manager — including
  the ones set to `none`. One line each: the plain meaning, then the technical
  name, then why (tied to an answer they gave, or "you weren't sure, so I kept
  it simple").
- **The packages** the dry run listed beyond the framework itself, so nothing
  bundled is hidden.
- What you left out, and whether it can be added later (several tools can be
  added with `bungkus-add`; others cannot, so say which matter now).
- The exact command you dry-ran, without `--dry-run`.

End by asking only whether to go ahead and set it up. If the user objects to
something, change that one thing, dry-run again, and present the updated
recommendation.

### 6. Hand over

When the user agrees, continue with `bungkus-scaffold` using that exact
command: it confirms the target directory, runs it and reports the result. Do
not run `create` without `--dry-run` from this skill.

## Notes

- The mapping in step 3 is guidance, not a rule table. A stated team standard
  or an existing codebase convention wins over it.
- If the user is clearly technical and names specific technologies, respect
  them, but still do not turn the recommendation into a list of choices.
- Never recommend a value that step 1 did not list, and never describe a result
  that step 4 did not print.
