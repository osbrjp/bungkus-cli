# Security

bungkus-cli is a local, single-user CLI that scaffolds projects. It runs with
the invoking user's own OS privileges, exposes no network service, and stores
no credentials or user data. This document records the security-relevant
aspects of the **tool itself** (not the projects it generates).

## Reporting a vulnerability

Open a private security advisory on the GitHub repository, or email the
maintainer. Please do not file public issues for undisclosed vulnerabilities.

## Dependency management & remediation policy

- **Inventory.** `go.mod` + `go.sum` are the authoritative list of all direct
  and transitive third-party components. `go.sum` carries integrity hashes
  verified against the Go checksum database; every module path is a
  fully-qualified public repository (no dependency-confusion surface).
- **Scanning.** `govulncheck ./...` runs in CI on every push
  (`.github/workflows/run-tests.yml`) and fails the build on any advisory that
  affects called code.
- **Updates.** Dependabot opens PRs for the `gomod` and `github-actions`
  ecosystems weekly (`.github/dependabot.yml`).
- **Remediation windows** (by advisory severity affecting called code):
  - Critical: ≤ 7 days
  - High: ≤ 30 days
  - Moderate / Low: ≤ 90 days

  Advisories in required-but-uncalled modules are bumped opportunistically
  (e.g. on the next Dependabot PR) rather than under an SLA.

## External communication

The tool is offline at scaffold time. Its only outbound network call is:

- **Host:** `https://registry.npmjs.org/<package>` (hardcoded, HTTPS only)
- **Trigger:** the maintainer-only `bump` command
  (`cmd/bump.go`, gated behind the `bump` build tag — not in released binaries)
- **Auth:** none (anonymous, read-only version metadata)
- **Client policy** (`cmd/bump.go`): 20s request timeout, no retries, sequential
  requests, redirects are **not** followed (`CheckRedirect` returns
  `ErrUseLastResponse`); a failed fetch skips that one package rather than
  aborting. TLS is left at Go's secure defaults (no `InsecureSkipVerify`).

The package-manager and git subprocesses launched by `PostScaffold` perform
their own network access using the user's own tool configuration; that is
outside this tool's control.

## Dangerous functionality

Security-sensitive operations and their safeguards:

- **Subprocess execution** — `pkg/scaffold.go` (`PostScaffold`: `git …`, the
  package-manager install command) and `pkg/packagejson.go` (`pm --version`).
  All calls use `exec.Command` with an argument slice (no shell), so there is
  no shell-metacharacter injection surface. The package-manager value is
  validated against the registry enum before execution, and install commands
  originate only from the embedded, trusted `config/registry.json`.
- **Filesystem writes** — scaffolding writes under the destination directory.
  The project name / destination is validated (`pkg/validate.go`,
  `ValidateProjectName` + `ValidateDest`) to a single local path segment, so
  absolute paths and `..` traversal are rejected before any write.
- **Outbound HTTP** — see "External communication" above.

## Risky components

Reviewed 2026-07-23 against `go.mod`: no unmaintained, end-of-life, or
known-vulnerable third-party component identified (`govulncheck` clean).
Reassess on each Dependabot PR and at least quarterly.
