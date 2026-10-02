---
name: go-best-practices
description: Rules every agent must follow when writing, editing or reviewing Go in this repo — gofmt-clean, idiomatic Go with errors returned and wrapped, registry-driven options, and table-driven tests. Use for any .go file, go.mod change, or Go code review in bungkus-cli.
---

# Go best practices for bungkus-cli

These record what the code already does. They sit on top of the OSBR handbook's
[Golang Style Guide](https://handbook.osbrjp.com/style-guide-golang) (source:
`osbrjp/handbook`, `doc/style-guide-golang.md`), `CLAUDE.md` and `SECURITY.md`.
If a rule here conflicts with those, they win.

Before you say a Go change is done, all four must pass with no output from the
first two:

```sh
gofmt -l .
go vet ./...
go build -o /dev/null .
go test ./...
```

CI (`.github/workflows/run-tests.yml`) also runs `govulncheck` and scaffolds
real projects; a template or `registry.json` change can break those without
breaking `go test`.

## 1. Formatting and layout

- **`gofmt` is the formatter.** No other formatter or linter config exists; do
  not add one as a side effect of another change.
- **Follow** [Effective Go](https://go.dev/doc/effective_go) and
  [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- **Imports** are two groups: standard library, then everything else.
- **Packages:** `cmd/` holds cobra commands and flag handling only, `pkg/` holds
  the logic (config, registry, package.json, scaffolding), `internal/tui/` holds
  the BubbleTea wizard, `config/` holds only the embedded assets. Logic that a
  command and the wizard both need goes in `pkg/`.
- **Keep exports narrow.** Lowercase by default; export only what another
  package calls.

## 2. Registry first

- **`config/registry.json` is the single source of truth** for bases, add-ons
  and package versions. Adding an option means a JSON entry and usually a
  template directory. Do not hardcode an option value, a package name or a
  version in Go.
- **Option kinds are typed strings** (`type FormLib string`) with an
  `IsValid()` method that asks the registry, and `IsValidIntegration(base)`
  where the option is tied to React or Vue. A new kind follows that shape.
- **Template presets** in `cmd/create.go` start from `pkg.NewProjectConfig()`
  and override fields, so a new `ProjectConfig` field inherits its default.
- **Flags override only when typed:** check `cmd.Flags().Changed(name)` before
  applying a flag value over a preset.

## 3. Errors

- **Return `error`; handle it at the call site.** Wrap with `%w` and say what
  was being done: `fmt.Errorf("failed to write package.json: %w", err)`.
  A few older call sites return a bare `err`; wrap in new code.
- **Messages** are lowercase, have no trailing punctuation, and quote user
  input with `%q`.
- **No `panic`** for anything a user can cause. The only panics are
  `regexp.MustCompile` on constant patterns at package init.
- **`os.Exit` only in `main.go` and `cmd.Execute`.** Everything else returns
  the error up to cobra's `RunE`.
- **Sentinel errors** (`ErrUnknownAddOption`) are compared with `errors.Is`.
- **Nothing ignored silently.** The accepted `_` discards are cobra flag
  getters for flags the command itself registered
  (`v, _ := cmd.Flags().GetString("pm")`) and best-effort writes whose failure
  must not fail the command (the update-check cache). Anything else is handled.
- **Incompatible add-ons warn and fall back to `none`**; they are not errors.

## 4. State and types

- **No new mutable globals.** The existing package-level variables are the
  cobra commands, the embedded assets, lookup tables and styles that are set
  once, and the registry (`InitRegistry` in `main`, read through
  `GetRegistry()`). Pass anything else as an argument.
- **Copy slices and maps** before handing them across a package boundary if the
  caller could change them.
- **No interfaces yet.** Add one only where the consumer needs two
  implementations; define it in the consuming package and keep it small.
- **Limits are named constants** with a comment naming the source
  (`maxProjectNameLen` mirrors npm's cap).

## 5. Files, paths and processes

- **Validate before writing.** A project name goes through
  `ValidateProjectName` and every destination through `ValidateDest`, so a
  scaffold can never leave the current directory.
- **Permissions:** `0o755` for directories, `0o644` for files.
- **Processes** are started with `exec.Command` / `exec.CommandContext` and an
  argument slice, never by building a shell string from user input. The one
  `bash -c` call (the updater's fixed install pipeline) takes no user input;
  do not add another.
- **Templates** are `text/template` files under `config/templates/`, read from
  the embedded FS. Never read templates or the registry from disk at runtime.

## 6. Comments

- **Exported identifiers and non-obvious unexported ones** get a doc comment
  that starts with the identifier's name and says what it guarantees and why
  (see `pkg/validate.go`).
- **Inline comments explain why,** not what: a quirk, a limit with a reason, a
  workaround with its cause. Reference the issue when one exists (`(#122)`).
- **No edit narration** ("added", "now uses", "fixed per review"), no restating
  the next line, no crediting a tool. History goes in commit messages.

## 7. Tests

- **Tests live next to the code** (`pkg/config_test.go`), in the same package.
- **Table-driven** with `t.Run` per case; a case struct has a `name` and the
  inputs and expectations.
- **Standard library only:** `t.Errorf("Fn(%q) = %v, want %v", in, got, want)`.
  No assertion library.
- **Exercise the real embedded registry,** not a fixture copy, so a
  `registry.json` change is covered automatically.
- **Filesystem tests** use `t.TempDir()`; helpers call `t.Helper()`.

## 8. Dependencies

- **Check the standard library and existing modules first** (cobra, BubbleTea,
  Lip Gloss, `golang.org/x/mod`). The binary ships with no runtime dependencies;
  keep it that way.
- **A new module** must come from a known host and be justified in the PR;
  `SECURITY.md` sets the remediation policy and `govulncheck` runs in CI.
- **`go.mod` and `go.sum` are committed together.**

## 9. Review checklist

- [ ] The four commands at the top pass.
- [ ] No option value, package name or version hardcoded in Go that belongs in `registry.json`.
- [ ] Every error is returned and wrapped with `%w`; no new `panic`, no `os.Exit` outside `main.go` / `cmd.Execute`.
- [ ] No new mutable global, no new shell string, no write outside the validated destination.
- [ ] New logic has a table-driven test next to it.
- [ ] A `create` flag or preset change also updates `plugin/skills/bungkus-scaffold/SKILL.md` and bumps the plugin version.
- [ ] Comments explain why; no edit narration.
