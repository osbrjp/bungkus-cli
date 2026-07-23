# ASVS 5.0 Audit — bungkus-cli (2026-07-23)

## Verdict

**Highest level fully met (app-wide): none.** Two chapters block L1, so no level passes app-wide under a strict reading.

**Level of interest (L1): not met.** The gaps are narrow and, under the tool's threat model, low-severity — most "failures" are *missing security documentation*, not missing controls. The blocking items:

- **V2 2.2.1 (L1)** — the free-form **project name / destination directory is never validated**; `create ../../foo` or `/etc/foo` writes outside the intended directory, and the name is emitted as the `package.json` `name` with no npm-name check. The single genuine code-level security gap. *(bounded: local CLI, invoking user's own privileges — no privilege escalation)*
- **V15 15.1.1 (L1)** — no documented dependency update / vulnerability-remediation policy for the tool's own Go deps; and **15.2.1** can't be verified without running `govulncheck` (`golang.org/x/mod v0.14.0` is notably old).

Everything else at L1 passes. V1, V5, V12, V16 are clean at all three levels; V13 is clean at L1. The core security concern for a scaffolding CLI — subprocess/command execution — is **solid**: every `exec.Command` uses argument slices (no shell), and the package-manager value is validated against the registry enum before execution.

## Level matrix

| Chapter | L1 | L2 | L3 |
|---|---|---|---|
| V1 Encoding & Sanitization | PASS | PASS | PASS |
| V2 Validation & Business Logic | **FAIL** | **FAIL** | **FAIL** |
| V5 File Handling | PASS | PASS | PASS |
| V12 Secure Communication | PASS | PASS | PASS |
| V13 Configuration | PASS | **FAIL** | **FAIL** |
| V15 Secure Coding & Architecture | **FAIL** | **FAIL** | **FAIL** |
| V16 Logging & Error Handling | PASS | PASS | PASS |
| **TOTAL (app-wide)** | **FAIL** | **FAIL** | **FAIL** |

L2/L3 failures in V13 and V15 are almost entirely documentation gaps plus two low-severity code items (redirect-following on the maintainer-only npm client; the hidden `bump` command shipped in the release binary). V15 L1/L3 and V13 L3 include Not-verifiable-from-repo items (`govulncheck`, Slack webhook rotation) — these make a level PARTIAL at best, never PASS.

## Scope & rationale

- **Target:** the bungkus-cli **tool itself** (`cmd/`, `pkg/`, `internal/`, `config/embed.go`, `main.go`, `.github/workflows/`). The runtime security of the *projects it generates* (template output) is explicitly out of scope — that is a separate audit of the generated app.
- **Standard:** ASVS 5.0, cached at `~/.cache/asvs/5.0`. Every verdict traces to a requirement in those files.
- **Level of interest:** L1 (all three levels graded per chapter).
- **Scanners:** none available — semgrep and checkov/tfsec not installed; `npm/pnpm audit` N/A for a Go tool. Auditors gathered evidence by reading; one dependency check (`govulncheck`) remains outstanding (see Not verifiable from repo).
- **Documented rationale for N/A-by-design (recorded verbatim, sourced from this audit's scoping + CLAUDE.md):** *bungkus-cli is a local, single-user CLI. It runs with the invoking user's own OS privileges, exposes no network-facing service (no listening sockets), performs no authentication/authorization/session management, and holds no multi-tenant or persistent user data. Subprocesses it launches (git, package managers) run as the same user, who could run them directly.* This rationale removes the authentication (V6), session (V7), and authorization (V8) chapters, and the server/service-side requirements throughout V12/V13/V15/V16.

---

## V1 — Encoding and Sanitization

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 1.2.1 | 1 | N/A-not-present | CLI emits no HTTP response / HTML / XML of its own — no `net/http` server, no HTML rendering; only files-on-disk output. Generated project HTML is out of scope. |
| 1.2.2 | 1 | N/A-not-present | No dynamic URL building from untrusted input. Only URL is fixed host `https://registry.npmjs.org/`+name (`pkg/bump.go:73`), name from trusted registry. |
| 1.2.3 | 1 | Satisfied | package.json built via `encoding/json` — `marshalPkg` uses `json.NewEncoder` (`pkg/packagejson.go:183-192`); `cfg.ProjectName`→`pkg.Name` (`:47`) contextually escaped. `SetEscapeHTML(false)` only leaves `<>&` raw (not JSON-structural). |
| 1.2.4 | 1 | N/A-not-present | No database access — no SQL/ORM driver imports in tool code. |
| 1.2.5 | 1 | Satisfied | `exec.Command(name, args...)` (`pkg/scaffold.go:309`) and `exec.Command(pm, "--version")` (`pkg/packagejson.go:264`) pass args as a slice — no shell. `pm`/git args are validated enums / static literals; `cfg.ProjectName` never reaches a subprocess arg. |
| 1.2.6 | 2 | N/A-not-present | No LDAP client. |
| 1.2.7 | 2 | N/A-not-present | No XPath/XML querying. |
| 1.2.8 | 2 | N/A-not-present | No LaTeX processing. |
| 1.2.9 | 2 | Satisfied | Dynamic regex in `pkg/bump.go:168` escapes interpolated values with `regexp.QuoteMeta`; other pattern (`:102`) is static. |
| 1.2.10 | 3 | N/A-not-present | No CSV/spreadsheet export. |
| 1.3.1 | 1 | N/A-not-present | No WYSIWYG/HTML input handling. |
| 1.3.2 | 1 | N/A-not-present | Go, no `eval`/dynamic code exec; `text/template` executes trusted embedded `.tmpl`, not user code. |
| 1.3.3 | 2 | N/A-by-design | Only untrusted-shaped input reaching a dangerous context (filesystem path) is `cfg.ProjectName`/`DestDir`→`os.MkdirAll`/`filepath.Join` (`pkg/scaffold.go:29,39,363`) with no allowlist. Rationale: user's own privileges → no trust boundary crossed. Residual noted (see V2). |
| 1.3.4 | 2 | N/A-not-present | No SVG handling. |
| 1.3.5 | 2 | N/A-not-present | No Markdown/CSS/XSL/BBCode processing in tool code. |
| 1.3.6 | 2 | N/A-by-design | Only outbound call is fixed-host `pkg/bump.go:73`; `name` from maintainer's registry.json, not runtime input. `bump` is a hidden maintainer tool. |
| 1.3.7 | 2 | Satisfied | Templates parsed from trusted embedded `.tmpl` bytes (`renderTemplate`, `pkg/scaffold.go:379-393`); untrusted `cfg` passed only as data to `t.Execute`, never to build the template string. |
| 1.3.8 | 2 | N/A-not-present | Java/JNDI — not a Go concern. |
| 1.3.9 | 2 | N/A-not-present | No memcache. |
| 1.3.10 | 2 | Satisfied | All `Printf`/`Errorf`/`Sprintf` first args are string literals — no user-controlled format strings. |
| 1.3.11 | 2 | N/A-not-present | No mail/SMTP/IMAP. |
| 1.3.12 | 3 | Satisfied | Go `regexp` uses RE2 (linear time); patterns simple/QuoteMeta-escaped. No ReDoS. |
| 1.4.1 | 2 | Satisfied | Go memory-safe; no `unsafe`, no cgo. |
| 1.4.2 | 2 | Satisfied | Only integer input is maintainer flag `--soak-days`; overflow benign/self-inflicted. |
| 1.4.3 | 2 | Satisfied | GC runtime; `defer resp.Body.Close()` (`pkg/bump.go:77`) handles the one external resource. |
| 1.1.1 | 2 | N/A-not-present | No decode/unescape of untrusted input into canonical form. |
| 1.1.2 | 2 | Satisfied | Output encoding is the final emit step (`marshalPkg`, JSON reformat `pkg/scaffold.go:398-409`) before `os.WriteFile`. |
| 1.5.1 | 1 | N/A-not-present | No XML parsing (`encoding/xml` not imported). |
| 1.5.2 | 2 | Satisfied | Deserialization into fixed typed structs (`Registry`, `Packument`); no polymorphic/gadget deserialization. |
| 1.5.3 | 3 | Satisfied | Single JSON parser everywhere; no parser divergence. |

**Roll-up: L1 PASS · L2 PASS · L3 PASS**

Overall posture is strong. Memory-safe Go CLI with no network service, database, or HTML/XML/LDAP/XPath output, so the majority of injection requirements are N/A-not-present. The two live encoding surfaces are handled correctly: package.json and all JSON output go through `encoding/json` (contextual, structurally safe), and subprocess execution uses `exec.Command` with argument slices (no shell). The one residual — not a Gap under the documented threat model — is that `cfg.ProjectName`/`DestDir` reach filesystem paths with no length limit or character allowlist (1.3.3); self-inflicted-only for a single local user (see V2 for the ticketed version).

### Tasks — V1
- [x] No action — all in-target requirements Satisfied or N/A.

---

## V2 — Validation and Business Logic

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 2.1.1 | 1 | **Gap** | Enum flags have an implicit allow-list (registry.json) documented via flag help + CLAUDE.md, but the free-form project name / dest dir (`cmd/create.go:115`) has no documented validation rule. Doc side of the 2.2.1 gap. |
| 2.1.2 | 2 | Satisfied | Combined-item consistency rules documented: flag help (`cmd/create.go:328-329`) + CLAUDE.md integration filtering. |
| 2.1.3 | 2 | N/A-by-design | Rationale: no auth / no network service / no multi-tenant data — no business-logic limits to document. |
| 2.2.1 | 1 | **Gap** | Enum inputs validated with positive allow-lists via `IsValid()` (`cmd/create.go:215-277`) — thorough. But the free-form project name / dest dir makes a security-relevant decision (where files are written) and is never validated: `cmd/create.go:115`→`:291`→`pkg/scaffold.go:29` `os.MkdirAll(destDir)`. `create ../../foo` or `/etc/foo` writes outside intended location; also rendered into package.json `name` (`pkg/packagejson.go:47`) with no npm-name check. Same gap in the TUI (`internal/tui/wizard.go:666-676`). |
| 2.2.2 | 1 | N/A-by-design | Rationale: no network service — no client/server split; all validation in the single trusted binary. |
| 2.2.3 | 2 | Satisfied | Cross-flag guards enforced: `--cicd` requires deploy (`:254`), `--db` requires `--orm` (`:266`), `d1` only with drizzle (`:269`), monorepo requires pnpm (`:275`), integration compatibility (`:279-290`). |
| 2.3.1 | 1 | N/A-not-present | Single one-shot command, no multi-step resumable flow. |
| 2.3.2 | 2 | N/A-by-design | Rationale: single-user, no multi-tenant data. |
| 2.3.3 | 2 | N/A-not-present | No transactional business operation; only mutation is file scaffolding. |
| 2.3.4 | 2 | N/A-not-present | No limited-quantity/bookable resource. |
| 2.3.5 | 3 | N/A-by-design | Rationale: no auth / multi-user context. |
| 2.4.1 | 2 | N/A-by-design | Rationale: no network service — nothing to flood. |
| 2.4.2 | 3 | N/A-by-design | Rationale: no network service. |

**Roll-up: L1 FAIL (2 gaps) · L2 FAIL (2 gaps) · L3 FAIL (2 gaps)**

For enum inputs this chapter is strong — every framework/tool/version flag is positively validated against the embedded registry allow-list, and cross-flag rules are implemented and documented. The single material weakness is the one free-form input: the project name / destination directory is passed unvalidated from CLI arg (and TUI field) straight into `os.MkdirAll`/`filepath.Join`, enforcing neither a sensible package-name structure nor preventing traversal/absolute paths outside the intended location. Severity is bounded by the "user's own privileges" rationale (no privilege escalation), but it is a genuine L1 input-validation gap (2.2.1) with a matching documentation gap (2.1.1). All other rows are legitimately N/A for a local, single-user, no-network, no-auth CLI.

### Tasks — V2

- [ ] **T-2-1: Validate the free-form project name / destination directory before scaffolding**
  - **ASVS**: 2.2.1 (L1) — resolves Gap
  - **Problem**: The project name (CLI arg / TUI field) is used unvalidated as the write destination and as the package.json `name`. Path traversal / absolute paths are accepted and passed to `os.MkdirAll`. Evidence: `cmd/create.go:115` (`cfg.ProjectName = args[0]`) → `cmd/create.go:291` → `pkg/scaffold.go:29` `os.MkdirAll(destDir, 0o755)` and `filepath.Join(destDir, ...)`; identical unvalidated path in `cmd/root.go:37` and `internal/tui/wizard.go:666-676`; rendered into `name` at `pkg/packagejson.go:47`. No `filepath.Clean`/`IsLocal`/name check exists (confirmed by repo-wide grep).
  - **Change**: Add one shared validator in `pkg` (root cause is shared — three call sites route through the same `destDir`). E.g. `func ValidateProjectName(name string) error` that: (a) allows the special `"."` case as already handled by callers; (b) for the dest path, rejects anything where `!filepath.IsLocal(destDir)` (Go 1.24 stdlib — blocks `..`, absolute, `/`-escaping paths) so output stays under cwd; (c) validates the name against npm package-name rules with a regex like `^[a-z0-9][a-z0-9._-]*$` and a length bound (npm caps at 214) since it lands in package.json `name`. Call it in `cmd/create.go` immediately after `destDir` is resolved (~`:294`, before `pkg.Scaffold`), in `cmd/root.go` before Scaffold (~`:39`), and surface the error in the TUI confirm step. No new dependency — stdlib `path/filepath` + `regexp`.
  - **Key files**: `pkg/config.go` (or a new small `pkg/validate.go`), `cmd/create.go`, `cmd/root.go`, `internal/tui/wizard.go`, and a `pkg/*_test.go` table test.
  - **Acceptance**: `bungkus-cli create ../../foo`, `create /etc/foo`, and `create "Bad Name!"` all exit with a clear validation error and write nothing; `create my-app` and `create .` behave exactly as today; a table test asserts accept/reject cases.
  - **Effort / risk**: S. Blast radius low — a guard at the entry boundary; only risk is over-tight regex rejecting a previously-working name, mitigated by the test table.

- [ ] **T-2-2: Document the project-name / dest-dir input validation rules**
  - **ASVS**: 2.1.1 (L1) — resolves Gap
  - **Problem**: No documented validation rule defines the expected structure of the free-form project-name/dest-dir input. Evidence: `--help` text in `cmd/create.go:311-334` has no entry describing name constraints; CLAUDE.md "Conventions" covers branch/commit naming, not input validation.
  - **Change**: After T-2-1 lands, document the enforced rule in two places: the positional-arg description / a note in the `create` help, and a line in CLAUDE.md "Architecture Notes" stating the project-name structure (lowercase npm-name charset, length bound) and that dest must stay local (`filepath.IsLocal`). Keep it to the rule the code enforces so doc and code match.
  - **Key files**: `cmd/create.go` (command `Short`/`Long` or a usage note), `CLAUDE.md`.
  - **Acceptance**: `bungkus-cli create --help` (or CLAUDE.md) states the accepted project-name format and the "stays within cwd" constraint, matching T-2-1.
  - **Effort / risk**: S. Docs only.
  - **Depends on**: T-2-1.

---

## V5 — File Handling

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 5.1.1 | 2 | N/A-not-present | No upload feature. Only file ops are scaffolding writes of embedded templates (`pkg/scaffold.go:16-430`) and reads of `config.Templates`/`config.RegistryJSON`. |
| 5.2.1 | 1 | N/A-not-present | No file-upload/accept path; inputs are CLI flags + a project-name string. |
| 5.2.2 | 1 | N/A-not-present | No uploaded files; only files read are the compiled-in `embed.FS` (trusted at build time). |
| 5.2.3 | 2 | N/A-not-present | No archive decompression of untrusted input; templates ship uncompressed embedded. |
| 5.2.4 | 3 | N/A-not-present | No per-user storage / upload quota — local single-user CLI. |
| 5.2.5 | 3 | N/A-not-present | No compressed-file ingestion; `copyDir` walks the trusted embedded FS only. |
| 5.2.6 | 3 | N/A-not-present | No image upload/processing. |
| 5.3.1 | 1 | N/A-not-present | No server serving files over HTTP; writes (some `0o755`, e.g. `.husky` hooks) land in the user's own project dir. |
| 5.3.2 | 1 | N/A-by-design | Rationale: user's own privileges. The one user-input-to-path flow is project name → `destDir` (`cmd/create.go:115,291-294`) → `filepath.Join`/`os.MkdirAll` (`pkg/scaffold.go:29,363`), no sanitization, but crosses no trust boundary (see observation). |
| 5.3.3 | 3 | N/A-not-present | No decompression of user archives; `copyDir` walks trusted embedded FS. |
| 5.4.1 | 2 | N/A-not-present | No file-download/serving feature; no HTTP layer. |
| 5.4.2 | 2 | N/A-not-present | No filenames in response headers — no network responses. |
| 5.4.3 | 2 | N/A-not-present | No files from untrusted sources to scan; all inputs compiled-in templates. |

**Roll-up: L1 PASS · L2 PASS · L3 PASS**

V5 targets apps that ingest, store, and serve untrusted files over a network — bungkus-cli does none of that. Its only file inputs are the build-time embedded template tree (trusted) and its only writes land in the invoking user's own project directory under their own privileges. Every requirement is N/A. The nearest thing to a finding is that the project-name argument reaches `os.MkdirAll`/`filepath.Join` with no sanitization (`cmd/create.go:115`→`pkg/scaffold.go:29,363`): a name like `../../foo` writes outside the intended directory. Under the single-user model this is a robustness/footgun nit, not a security gap — ticketed under V2 (T-2-1) rather than here.

### Tasks — V5
- [x] No action — all in-target requirements Satisfied or N/A.

---

## V12 — Secure Communication

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 12.1.1 | 1 | Satisfied | Sole outbound call uses `http.Client` with only `Timeout` set (`cmd/bump.go:33`), no TLS override anywhere. Go 1.24 default negotiates TLS 1.2/1.3, never below 1.2. |
| 12.1.2 | 2 | Satisfied | No cipher-suite override; inherits Go's secure default suites (AEAD/ECDHE-FS). |
| 12.1.3 | 2 | N/A-not-present | No mTLS client certs — no server. No `ListenAndServe`/`net.Listen`. |
| 12.1.4 | 3 | N/A-not-present | OCSP stapling is a server feature; no TLS server. |
| 12.1.5 | 3 | N/A-not-present | ECH is a server setting; no server. |
| 12.2.1 | 1 | Satisfied | One outbound request is hardcoded `https://` with no HTTP fallback — `pkg/bump.go:73` `client.Get("https://registry.npmjs.org/"+name)`; scheme is a literal. |
| 12.2.2 | 1 | Satisfied | Registry presents a publicly trusted cert; client validates via Go default (no `InsecureSkipVerify`). |
| 12.3.1 | 2 | Satisfied | Only direct outbound is HTTPS, no fallback. PM/git subprocesses delegate transport to external tools (default HTTPS registries); `pmVersion` is local. |
| 12.3.2 | 2 | Satisfied | TLS validation left at Go default — no `Transport`/`TLSClientConfig`; `InsecureSkipVerify` appears nowhere. |
| 12.3.3 | 2 | N/A-not-present | No internal HTTP service-to-service comm — single process, no listeners. |
| 12.3.4 | 2 | N/A-not-present | No internal services / self-signed CA config. |
| 12.3.5 | 3 | N/A-not-present | No microservice architecture — single binary. |

**Roll-up: L1 PASS · L2 PASS · L3 PASS**

Transport posture is clean. Exactly one direct outbound connection — an HTTPS GET to a hardcoded `registry.npmjs.org` (`pkg/bump.go:73`) using an `http.Client` that sets only a timeout, so certificate verification, TLS floor (1.2), and cipher selection stay at Go's secure defaults. No `InsecureSkipVerify`, no `TLSClientConfig`, no plaintext fallback, no listening socket. The only residual (not a V12 gap): `PostScaffold` subprocesses shell out to the user's package manager/git, whose registry transport is their own config — outside the CLI's control.

### Tasks — V12
- [x] No action — all in-target requirements Satisfied or N/A.

---

## V13 — Configuration

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 13.4.1 | 1 | N/A-by-design | Rationale: no deployed environment; ships as standalone binaries, nothing serves `.git` (not tracked). |
| 13.1.1 | 2 | **Gap** | Tool relies on external `https://registry.npmjs.org/<name>` (`pkg/bump.go:73`) but no doc enumerates the tool's outbound communication needs. |
| 13.2.1 | 2 | N/A-by-design | No backend components. Only authenticated comm is CI→GitHub via ephemeral per-run `GITHUB_TOKEN` (`release.yml:48`); npm read is anonymous. |
| 13.2.2 | 2 | Satisfied | CI tokens scoped least-privilege per workflow/job (`promote.yml:7-9`, `bump-versions.yml:11-14`, `release.yml:23-26,93`). |
| 13.2.3 | 2 | N/A-not-present | No service credentials — npm anonymous, GitHub ephemeral token; no default creds. |
| 13.2.4 | 2 | Satisfied | Outbound host is a single hardcoded constant; no user-supplied URL — effectively a one-host allowlist. |
| 13.2.5 | 2 | N/A-by-design | Rationale: no server runtime configuration. |
| 13.3.1 | 2 | Satisfied | No secrets in source/registry/workflows (all workflow secrets are `${{ secrets.* }}` refs to GitHub's encrypted store). |
| 13.3.2 | 2 | Satisfied | Secret access least-privilege: `SLACK_WEBHOOK_URL` only in `announce` job (`release.yml:130`); `GITHUB_TOKEN` scoped per job. |
| 13.4.2 | 2 | N/A-by-design | Rationale: no deployed environment / debug mode. |
| 13.4.3 | 2 | N/A-by-design | No web server — no directory listing. |
| 13.4.4 | 2 | N/A-by-design | No web server — HTTP TRACE N/A. |
| 13.4.5 | 2 | N/A-by-design | No server endpoints / monitoring surfaces. |
| 13.1.2 | 3 | N/A-not-present | No connection pooling — one `http.Client` fetching sequentially. |
| 13.1.3 | 3 | **Gap** | Mechanism present (20s timeout, no retries — `cmd/bump.go:33`, the secure default) but no doc defines the resource-management/timeout/retry strategy. |
| 13.1.4 | 3 | N/A-by-design | Tool holds no persistent secrets; CI secrets GitHub-managed. |
| 13.2.6 | 3 | **Gap** | Same as 13.1.3 — behavior is sound but not traceable to documented configuration. |
| 13.3.3 | 3 | N/A-not-present | No cryptographic operations (only sha256sum of release artifacts in CI shell). |
| 13.3.4 | 3 | Not-verifiable-from-repo | `GITHUB_TOKEN` auto-expires per run; rotation of `SLACK_WEBHOOK_URL` (`release.yml:130`) is a GitHub org secret setting not visible in the repo. |
| 13.4.6 | 3 | N/A-not-present | No backend components; CLI version string is intentional. |
| 13.4.7 | 3 | N/A-by-design | No web tier to restrict file extensions on. |

**Roll-up: L1 PASS · L2 FAIL (1 gap) · L3 FAIL (3 gaps, 1 not-verifiable)**

Supply-chain and secret hygiene are genuinely good: dependencies pinned in `go.mod` with a committed `go.sum` for integrity, `bump` only adopts stable non-deprecated releases that soaked ≥14 days (`pkg/bump.go:44-68`), all CI secrets go through GitHub's encrypted store with per-job least-privilege scoping, and a full-repo scan found zero hardcoded secrets. Every "runtime hardening" requirement is correctly N/A-by-design (no server). The only genuine misses are documentation gaps at L2/L3: the reliance on `registry.npmjs.org` and its connection resource-management strategy are undocumented even though the behavior is sound. Secure-default observation (out of scope — generated projects): ORM `.env.example` templates emit placeholder DSNs like `postgres://user:password@localhost:5432/<name>` — standard for `.example` files (real `.env` gitignored), acceptable.

### Tasks — V13

- [ ] **T-13-1: Document the tool's external communication and connection resource-management strategy**
  - **ASVS**: 13.1.1 (L2), 13.1.3 (L3), 13.2.6 (L3) — resolves three doc Gaps
  - **Problem**: The tool relies on `https://registry.npmjs.org` (`pkg/bump.go:73`) with no documentation enumerating that dependency, nor its timeout/retry/failure-handling posture. Behavior is sound (20s timeout, no retries, sequential single client — `cmd/bump.go:33`) but nothing records it.
  - **Change**: Add a short "External communication" section to `CLAUDE.md` (or `docs/dependencies.md`): (1) list the sole outbound endpoint `registry.npmjs.org`, reached only by the `bump` maintainer command, anonymous read-only; (2) state the connection policy in code — 20s timeout, no retries, sequential, failure = skip that package (`pkg/bump.go:150-152`); (3) note no other runtime egress. No code change.
  - **Key files**: `CLAUDE.md`, referencing `pkg/bump.go:71-85`, `cmd/bump.go:33`.
  - **Acceptance**: Docs name every outbound host, trigger, auth mode, timeout, retry policy; each statement maps to a cited code line.
  - **Effort / risk**: S — docs only.

- [ ] **T-13-2: Confirm SLACK_WEBHOOK_URL rotation policy**
  - **ASVS**: 13.3.4 (L3) — resolves Not-verifiable
  - **Problem**: `SLACK_WEBHOOK_URL` (`release.yml:130`) rotation is a GitHub org secret setting not visible in the repo.
  - **Change**: No code edit. Check GitHub repo/org → Settings → Secrets → Actions for `SLACK_WEBHOOK_URL`, confirm owner + rotation cadence (Slack webhook URLs don't auto-expire, so document a manual interval); record alongside T-13-1's doc. If none exists, establish one.
  - **Key files**: `.github/workflows/release.yml:130`; GitHub Actions secrets console.
  - **Acceptance**: A documented rotation owner + interval exists, or the secret is confirmed removed if the announce step is dropped.
  - **Effort / risk**: S — console check + one doc line.
  - **Depends on**: runtime access to GitHub org secrets settings.

---

## V15 — Secure Coding and Architecture

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 15.1.1 | 1 | **Gap** | No doc defines remediation timeframes for the CLI's own Go deps. No SECURITY.md, no dependabot/renovate/govulncheck for Go modules. (`bump-versions.yml` targets the *scaffolded projects'* npm pins, not `go.mod`.) |
| 15.1.2 | 2 | Satisfied | `go.mod`/`go.sum` are a maintained inventory with integrity hashes verified against the Go checksum DB; all module paths fully-qualified public repos. |
| 15.1.3 | 2 | N/A-by-design | Rationale: no network service — subprocess installs are user-initiated. |
| 15.1.4 | 3 | **Gap** | No documentation flags any dependency as a risky component. |
| 15.1.5 | 3 | **Gap** | "Dangerous functionality" (subprocess exec `pkg/scaffold.go:309`, `pkg/packagejson.go:264`; outbound HTTP `pkg/bump.go:73`) not documented as a security-sensitive area with its safeguards. |
| 15.2.1 | 1 | Not-verifiable-from-repo | No scanner output; cannot confirm deps free of known vulns from source. `golang.org/x/mod v0.14.0` is notably old. Settle with `govulncheck ./...`. |
| 15.2.2 | 2 | N/A-by-design | Rationale: no network service — no availability surface. |
| 15.2.3 | 2 | **Gap** | Maintainer-only `bump` command compiled into and invokable from the distributed binary — `cmd/bump.go:20` `Hidden: true` hides it from help but it remains functional shipped tooling. |
| 15.2.4 | 3 | Satisfied | Go modules pin versions + checksums (sum.golang.org); all paths fully qualified — no dependency-confusion substitution surface. |
| 15.2.5 | 3 | N/A-by-design | Rationale: no privilege boundary crossed; subprocesses run as the user. Sandboxing not meaningfully applicable. |
| 15.3.1 | 1 | N/A-not-present | No API/HTTP server returns objects across a trust boundary; package.json written to local disk. |
| 15.3.2 | 2 | **Gap** | Outbound client (`cmd/bump.go:33`) sets no `CheckRedirect`, so Go default follows up to 10 redirects. Low severity (hardcoded trusted host, maintainer-only path). |
| 15.3.3 | 2 | N/A-not-present | No mass-assignment; config built field-by-field from typed enum-validated flags. |
| 15.3.4 | 2 | N/A-not-present | No proxy/middleware, client IP, request logging, or rate limiting. |
| 15.3.5 | 2 | Satisfied | Go statically typed, no implicit coercion; typed enums further constrain values — type confusion not expressible. |
| 15.3.6 | 2 | N/A-not-present | Tool codebase is Go; no JS (generated output out of scope). |
| 15.3.7 | 2 | N/A-not-present | No HTTP server accepting request parameters. |
| 15.4.1 | 3 | Satisfied | Only shared global `globalRegistry` (`pkg/registry.go:77`) assigned once at startup (`main.go:16`), read-only after; TUI serializes state through single-threaded `Update`. |
| 15.4.2 | 3 | N/A-by-design | Only check-then-use pairs (`fs.Stat`→`fs.Sub`) are against the immutable embedded FS; disk writes cross no privilege boundary. |
| 15.4.3 | 3 | N/A-not-present | No locking primitives in app code. |
| 15.4.4 | 3 | N/A-not-present | No thread pool / resource scheduler. |

**Roll-up: L1 FAIL (1 gap, 1 not-verifiable) · L2 FAIL (3 gaps) · L3 FAIL (5 gaps)**

The chapter's headline concern for this tool — subprocess/command-execution safety — is genuinely solid: every `exec.Command` is called with separate argument slices and no shell (`pkg/scaffold.go:309`, `pkg/packagejson.go:264`), so there is no shell-metacharacter injection surface; the package-manager value is validated against the registry enum (`cmd/create.go:227`; TUI constrained to `registry.PackageManagers`) before execution; and install commands originate from the trusted embedded `registry.json`, not free user input. Go's memory safety and static typing eliminate whole defensive-coding categories. The failing rows are almost entirely **missing security documentation** (15.1.1/15.1.4/15.1.5), plus two low-severity code items (default redirect-following in the maintainer-only npm fetch, and the hidden `bump` command shipped in the production binary) and one unverifiable dependency-vuln check. The single most important gap is **15.1.1 (L1)**: no documented remediation/update policy for the CLI's own Go dependencies, and consequently no way to assert 15.2.1 — run `govulncheck` and adopt a policy to close both.

### Tasks — V15

- [ ] **T-15-1: Confirm the CLI's Go dependencies are free of known vulnerabilities**
  - **ASVS**: 15.2.1 (L1) — verify Not-verifiable-from-repo
  - **Problem**: No scanner evidence in-repo; `golang.org/x/mod v0.14.0` (go.mod) is well behind current releases and a candidate for a known advisory.
  - **Change**: Run `govulncheck ./...` at repo root; record output. Add a `govulncheck` step to `.github/workflows/run-tests.yml` so it runs on every push. Remediate any advisory by bumping the affected module (`go get <module>@<fixed>` then `go mod tidy`).
  - **Key files**: `go.mod`, `go.sum`, `.github/workflows/run-tests.yml`
  - **Acceptance**: `govulncheck ./...` exits clean; CI fails on any future advisory.
  - **Effort / risk**: S; low blast radius (dep bumps, gated by existing `go test ./...`).

- [ ] **T-15-2: Document a dependency update & vulnerability-remediation policy**
  - **ASVS**: 15.1.1 (L1) — Gap
  - **Problem**: No document defines risk-based remediation timeframes for the tool's own Go components. `bump-versions.yml` only refreshes scaffolded-project npm pins, not `go.mod`.
  - **Change**: Add `SECURITY.md` (or README section): remediation SLAs by severity (e.g. critical ≤7 days, high ≤30, moderate ≤90), that `go.mod`/`go.sum` is the component inventory, and the enforcing mechanism (`govulncheck` in CI from T-15-1, plus Dependabot for the Go ecosystem via `.github/dependabot.yml`).
  - **Key files**: `SECURITY.md` (new), `.github/dependabot.yml` (new), `README.md`
  - **Acceptance**: A reader finds the update cadence + remediation window for Go deps; Dependabot opens PRs for `gomod`.
  - **Effort / risk**: S; docs + config.
  - **Depends on**: T-15-1.

- [ ] **T-15-3: Disable redirect-following on the npm-registry HTTP client**
  - **ASVS**: 15.3.2 (L2) — Gap
  - **Problem**: `cmd/bump.go:33` builds `&http.Client{Timeout: 20 * time.Second}` (used by `pkg/bump.go:71`) with no `CheckRedirect`, so Go's default follows up to 10 redirects.
  - **Change**: Set `client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }`. If npm genuinely redirects and must be followed, add a bounded same-host-only `CheckRedirect` with a comment documenting intent.
  - **Key files**: `cmd/bump.go`
  - **Acceptance**: A 3xx from `registry.npmjs.org` is not transparently followed to an arbitrary host; non-redirecting behaviour unchanged.
  - **Effort / risk**: S; maintainer-only path.

- [ ] **T-15-4: Exclude the maintainer `bump` command from the released binary**
  - **ASVS**: 15.2.3 (L2) — Gap
  - **Problem**: `cmd/bump.go` registers the `bump` maintainer/CI tool into the production binary (`cmd/bump.go:71` `rootCmd.AddCommand(bumpCmd)`); `Hidden: true` (`:20`) only hides it from help.
  - **Change**: Gate `bump` behind a build tag — move `cmd/bump.go` to `//go:build bump` with a no-op registration for default builds; CI runs `go run -tags bump . bump …`. Update `.github/workflows/bump-versions.yml` to pass `-tags bump`.
  - **Key files**: `cmd/bump.go`, `.github/workflows/bump-versions.yml`
  - **Acceptance**: Default `go build` produces a binary where `bungkus-cli bump` is unknown; the scheduled workflow still runs it with `-tags bump`.
  - **Effort / risk**: S; isolated. (Alternatively accept as documented low-severity if maintainers prefer a single build target.)

- [ ] **T-15-5: Document the tool's "dangerous functionality"**
  - **ASVS**: 15.1.5 (L3) — Gap
  - **Problem**: No doc identifies subprocess execution (`pkg/scaffold.go:309`, `pkg/packagejson.go:264`) and outbound HTTP (`pkg/bump.go:73`) as security-sensitive.
  - **Change**: In `SECURITY.md`, add a "Dangerous functionality" section listing these sites and their existing safeguards (args-array `exec.Command` no shell; PM validated against registry enum before exec; install commands only from the embedded trusted registry; hardcoded HTTPS host).
  - **Key files**: `SECURITY.md` (shared with T-15-2), referencing `pkg/scaffold.go`, `pkg/packagejson.go`, `pkg/bump.go`
  - **Acceptance**: A reviewer can locate every `os/exec` and network call and its stated control from the doc.
  - **Effort / risk**: S; docs only.
  - **Depends on**: T-15-2.

- [ ] **T-15-6: Document third-party "risky components"**
  - **ASVS**: 15.1.4 (L3) — Gap
  - **Problem**: No doc flags any dependency as a risky component (unmaintained/EOL/vuln-history).
  - **Change**: In `SECURITY.md`, add a "Risky components" note — after reviewing `go.mod`, either list any lib meeting the risky criteria with rationale/mitigation, or state that a review found none and give the review cadence.
  - **Key files**: `SECURITY.md` (shared), `go.mod`
  - **Acceptance**: The doc contains a dated risky-component assessment covering the `go.mod` set.
  - **Effort / risk**: S; docs only.
  - **Depends on**: T-15-2.

---

## V16 — Security Logging and Error Handling

| Req | Level | Verdict | Evidence / reason |
|---|---|---|---|
| 16.1.1 | 2 | N/A-by-design | Rationale: no centralized log store. No logging framework; only `fmt.Printf`/`Fprintf` to the user's terminal. |
| 16.2.1 | 2 | N/A-by-design | Output goes to the user's own terminal; no security-event log entries produced. |
| 16.2.2 | 2 | N/A-by-design | No security event logs/timestamps emitted. |
| 16.2.3 | 2 | N/A-by-design | No log files/services written (only `os.WriteFile` for scaffolded project files). |
| 16.2.4 | 2 | N/A-by-design | No machine-processed logs; terminal output human-facing. |
| 16.2.5 | 2 | N/A-by-design | No security logging exists; tool handles no credentials/tokens. |
| 16.3.1 | 2 | N/A-by-design | Rationale: no authentication events to audit-log. |
| 16.3.2 | 2 | N/A-by-design | Rationale: no authorization events to audit-log. |
| 16.3.3 | 2 | N/A-by-design | No security controls to bypass; flag validation is not an attacker-facing boundary. |
| 16.3.4 | 2 | N/A-by-design | No security-log store (errors incl. the npm path are surfaced to stderr + up the RunE chain, but no security log to write to). |
| 16.4.1 | 2 | N/A-not-present | No log store to encode into; user input echoed to terminal is not persisted. |
| 16.4.2 | 2 | N/A-not-present | No persisted logs to protect from tamper. |
| 16.4.3 | 2 | N/A-not-present | No log transmission / network log sink. |
| 16.5.1 | 2 | Satisfied | Errors wrapped with `%w`, descriptive-but-non-sensitive; no stack traces (no `panic`/`recover`), no secrets. cobra prints the error (no `SilenceErrors`), main exits 1 (`cmd/root.go:61-66`). |
| 16.5.2 | 2 | Satisfied | Only external resource is the npm registry in `bump`: `http.Client{Timeout: 20s}` (`cmd/bump.go:33`) + graceful degradation — failed fetch logged and package skipped (`cmd/bump.go:37-42`). |
| 16.5.3 | 2 | Satisfied | No fail-open: all config validation runs before any file write (`cmd/create.go:215-290`) and aborts on error; every Scaffold sub-step returns `err` immediately. (Caveat: partial output not rolled back on mid-scaffold failure — robustness/UX, not fail-open.) |
| 16.5.4 | 3 | N/A-by-design | Go has no exceptions (chapter note excuses Go). Errors returned up the cobra `RunE` chain, printed, `os.Exit(1)`. Single-shot local CLI — a crash affects only that invocation. |

**Roll-up: L1 PASS (no L1 requirements in chapter) · L2 PASS · L3 PASS**

For a local single-user scaffolding CLI this chapter is largely about a threat model the tool doesn't have — no auth, authorization, session state, persisted log store, or attacker-facing boundary — so every V16.1–V16.4 security-logging requirement is legitimately N/A (verified by search: no logging framework, no log sinks, no credential handling). The error-handling requirements that apply are met: errors consistently wrapped with `%w` and surfaced up the cobra `RunE` chain to stderr, non-zero exit, no `panic`/`recover` leaking stack traces, error text carries no secrets, and the single external-resource path has a timeout and degrades by skipping. Non-security observation: a mid-scaffold failure leaves partially-written project files with no rollback — a robustness rough edge, not a security fail-open (the run still aborts).

### Tasks — V16
- [x] No action — all in-target requirements Satisfied or N/A.

---

## N/A chapters

- **V3 Web Frontend Security** — N/A-not-present. No web frontend; the CLI has no browser-facing surface (no `net/http` server, no HTML rendering of its own).
- **V4 API and Web Service** — N/A-not-present. No API/web service; no listening sockets.
- **V6 Authentication** — N/A-by-design. Local single-user CLI with no authentication (rationale in Scope).
- **V7 Session Management** — N/A-by-design / not-present. No sessions.
- **V8 Authorization** — N/A-by-design. No access-control model; runs with the invoking user's own OS privileges.
- **V9 Self-contained Tokens** — N/A-not-present. No token issuance/verification.
- **V10 OAuth and OIDC** — N/A-not-present. No OAuth/OIDC flows.
- **V11 Cryptography** — N/A-not-present. The tool performs no cryptographic operations (release-artifact `sha256sum` is a CI shell step, not tool code).
- **V14 Data Protection** — N/A-not-present. Handles/stores no sensitive or personal data of its own.
- **V17 WebRTC** — N/A-not-present. No WebRTC.

## Not verifiable from repo

- **15.2.1 (L1) — dependency vulnerabilities.** No scanner output in-repo. Settle by running `govulncheck ./...` at repo root (T-15-1); `golang.org/x/mod v0.14.0` is the leading candidate for an advisory.
- **13.3.4 (L3) — `SLACK_WEBHOOK_URL` rotation.** The secret is consumed at `release.yml:130` but its rotation cadence is a GitHub org secret setting not visible in the repo. Settle in the GitHub Actions secrets console (T-13-2).
