# Bibleit clients and CLI: revised implementation plan

Source branch pushes and private CI are authorized for all three prepared
repositories as of 2026-10-11. Customer release tags/assets and repository visibility
changes remain gated on production readiness. The repositories are currently
private: `bibleit-cli`, `bibleit-go` and `homebrew-tap`. This supersedes earlier
local-only source instructions. Use Conventional Commits and no codex branch prefix.
Every client targets the project-managed service, currently
`http://localhost:8080`; SSH uses `127.0.0.1:2222`. See
[endpoint-policy.md](endpoint-policy.md) and [releases.md](releases.md).

Revised 2026-10-11 against the local `../bibleit-server` checkout. Its HEAD is
`1a266cc`, with uncommitted assessment changes; this is a working-tree review,
not a compatibility claim against a published server release. This revision
supersedes the earlier sequence of extracting Go and immediately porting the
same surface into Python, Ruby, and Rust.

## Decisions retained

Use Go 1.27.1 for the CLI and Go client. Keep the CLI dependent on the
client, with browser opening, presentation, and credential persistence outside
core library operations. Use HTTPS/bearer authentication as the default finite
command transport; retain explicit SSH support. No automatic mutation retries.

Publish each language client independently under `bibleit-labs`: `bibleit-go`,
`bibleit-python`, `bibleit-ruby`, and `bibleit-rust`. Keep the Go CLI in
`bibleit-cli`, importing a pinned version of `github.com/bibleit-labs/bibleit-go`.
The server and existing reader/TUI retain separate repositories. Go client APIs
must not import the CLI. Share versioned protocol fixtures explicitly across
repositories and declare the tested server baseline with every release.

## Current completion summary

| Milestone | Completed | Remaining |
| --- | --- | --- |
| 1. Verified integration contract | 7/7 tracked items; operation matrix, fresh 34 decoder/18 encoder fixtures and full tracked HTTP/auth/SSH integration | Core protocol discrepancies resolved; browser-only API contracts and publication gates remain separate |
| 2. Core Go client | 5/5 tracked items; typed discovery, account, tokens, translations, reading and Live results; `NewClient(apiToken)` and fixed service | Future workspace/event APIs are separate |
| 3. CLI customer surface | 5/5 tracked items; single credential configuration, authentication, stable output, completion and destructive UX | Fresh private Windows/Ubuntu/macOS matrix passed; production release gates remain |
| Repository split and fixed endpoint | Complete; independent bibleit-go/bibleit-cli modules, no customer endpoint overrides, fresh fixed-endpoint HTTP/SSH integration passed | Production service addresses and standalone pinned dependency |
| 4. Workspace/event APIs | Contract requirements recorded | Server bearer API contract and implementation |
| 5. Python/Ruby/Rust clients | Separate repository naming decided | Implementation and portability checks |
| 6. Release preparation | Local six-target packaging, checksums, version injection, Homebrew/Scoop manifests, Bash/PowerShell installers; isolated native Homebrew install/test/reinstall/uninstall passed | Native Windows PowerShell 7/5.1 and Scoop lifecycle passed; production gates, tags and releases remain paused |

Completion describes tracked core scope, not all future product features. Detailed
evidence and historical increments are in [status.md](status.md).

## What changed and why it matters

| Current server evidence | Client/CLI consequence |
| --- | --- |
| Personal and organization resource accounts; scoped owner/admin/operator/viewer membership | Model authenticated identity separately from resource workspace. Do not equate a protocol actor with universal resource ownership. |
| Starter/Contributor access and organization capacity; CLI credentials have a separate cap of 10 | Read effective quotas from the server. Correct the CLI's token-quota advice; login credentials do not consume the manual-token plan allowance. |
| Token scope/expiry metadata and settings-based rotation with overlap | Distinguish login credentials from automation tokens; preserve metadata and revoked/expired/retiring error behavior. Rotation is not yet a general command-client operation. |
| SSH accepts Ed25519, RSA SHA-2, and NIST ECDSA | CLI validation now supports these algorithms while retaining OpenSSH host verification. |
| HTTP command limit: 8192 command bytes and 16384 body bytes; command and search rate limits | UTF-8 command/body limits and structured HTTP/protocol error handling are implemented and tested. |
| A separate browser-authenticated Live management WebSocket | Distinguish audience subscription from management watching. Management observers should not consume audience quota. |
| Browser-only organization settings, Live workspace movement, design, collaborations, stack reorder/removal | These features need an explicit public integration contract before CLI commands or client methods are promised. |
| Documentation reorganized into protocol, HTTP, and domain references | Use those sources and implementation tests as contract inputs; avoid treating prose alone as executable truth. |

The HTTP command envelope still takes `{"command":"..."}` and returns protocol
v1 records. Existing command builders, decoding, and HTTP tests remain useful.
They do not establish coverage of the new organization/browser surfaces.

## Verified constraints and remaining contract gaps

1. The SECRET SET discrepancy is resolved: the reference now documents
   server-generated CREATE/ROTATE secrets, DELETE and connection authentication.
   Custom assignment remains unsupported. Client and CLI retain explicit rejection
   of `secret set`; verified fixtures prevent management becoming authentication.
2. Both WebSocket surfaces currently use browser cookies; management watching
   also checks Origin and revalidates management access. Neither is an existing
   bearer-token client event endpoint.
3. Organization appears in domain authority, but the command parser provides no
   explicit organization/workspace selection or management commands. A local
   preference cannot invent server-side workspace authorization.
4. V1 requests cannot represent embedded double quotes. This limit is now
   documented and verified: malformed quote segments fail with `bad_command`,
   while literal and trailing backslashes retain their values. Builders keep their
   restrictions. Response framing is verified through command-aware fixtures;
   any new escape or envelope grammar needs an explicit compatible version.
5. The native CLI is explicitly described as unpublished/in development in the
   server documentation. Stable CLI output is now documented in `output.md`;
   publication remains paused. The unsupported `secret set` alias has been removed.

## Delivery sequence and acceptance gates

### 1. Reconcile and verify the public integration contract

Inventory each candidate operation with its endpoint or command, authentication
mode, authority, request/response, limits, errors, and supported server baseline.
Add request fixtures beside existing response fixtures. Verify their wire
semantics with server tests, including the secret mismatch and framing cases.
The operation matrix and source-verified fixtures are complete. The assessment
resolved the secret documentation, quoting integrity and stats-field discrepancies;
all tracked core contract checks pass.

Run integration tests using a disposable server/test database: PKCE success and
failure, onboarding-required rejection, token revocation/expiry/scopes, quota
failure, HTTP 429/413, reading, Live lifecycle, and SSH identity variants.
Do not rebuild the developer database as part of compatibility verification.

**Gate:** named operations are backed by implementation tests, documentation,
and a reproducible server baseline. Local mock tests alone are insufficient.

### 2. Complete the Go client for the verified customer surface

Add discovery (`SERVER INFO`, permission-aware `HELP`), account information and
quotas, translation metadata/catalogue where authorized, and typed result models
for identity, verses, translation references, Live state/stats/stack, and token
metadata. Keep unknown fields/codes available for forward compatibility.

Extract PKCE code exchange and logout/revocation requests into library methods;
keep browser orchestration and token persistence in the CLI. Represent transport,
protocol, and HTTP JSON failures distinctly, with status and server metadata.
Enforce request limits and cancellation. Preserve the current read/search/Live
surface while reconciling draft commands with the server.

These core client items are complete and verified against the real server using
the public credential-only constructors. New methods remain subject to this gate.

**Gate:** standalone Go applications and CLI commands use the same client paths
and pass the server integration suite without personal/workspace authority leaks.

### 3. Make CLI authentication and output ready for publication

One service-bound credential configuration at a project-managed service, explicit HTTP/SSH transport selection, stable
output, shell completion and destructive-command confirmation are implemented;
see `status.md` for progress and `output.md` for the stable contract. Native
Windows, Ubuntu and macOS hosted checks passed on the earlier published code;
local Linux/arm64 and Fish checks also passed. Fixed-endpoint/API changes have
new local macOS/Linux evidence; refresh native Windows before release. Repository
placement is decided and implemented; package publication stays paused.
Support automation tokens without exposing token values in command-line arguments.
Handle username onboarding, separate CLI credential quotas, expired/revoked
credentials, and revocation failures accurately. Align explicit SSH key validation
with server-supported algorithms while retaining OpenSSH agent/config behavior.

Finalize human output, stable typed JSON, NDJSON for supported streams, stderr
errors, exit codes, help, completion, cancellation, and destructive-command UX.
Do not silently fall back from bearer credentials to SSH. Secret-authorized SSH
subscription requires authentication and subscription on the same connection;
a separate one-shot secret command cannot prepare a later connection.

**Gate:** documented customer examples work against the verified server baseline;
Go 1.27.1 builds and representative native tests pass on Linux/macOS/Windows.

### 4. Extend the server integration surface for workspaces and events

Agree a bearer-authenticated contract for workspace discovery/selection, workspace
Live creation/movement, member-scoped authority, and richer Live control. Decide
whether to extend v1 commands or add versioned JSON resources; keep the choice
behind typed client methods and reuse server authorization/domain logic.

Define audience and management event contracts separately: initial snapshot,
event types, revisions, reconnect/resynchronization, revocation, workspace moves,
access loss, and quota accounting. A snapshot on reconnect is not event replay.
Do not scrape dashboard HTML or reuse CSRF-protected browser forms as client APIs.
Personal token scopes must not grant organization powers without active membership;
the server contract must explicitly define how scopes and workspace authority combine.

**Gate:** server-side bearer authorization and contract tests exist before these
features appear as supported client methods. This milestone requires server work;
this plan revision does not modify the server checkout.

### 5. Validate portability with Python, then Ruby and Rust

Implement Python against the verified core contract first, using idiomatic
exceptions, cancellation, and typed results. Reuse request/response fixtures and
the integration suite. Port to Ruby and Rust after Python reveals cross-language
ambiguities. Add workspace/events only when milestone 4's contract is ready.

**Gate:** each package states the operations, transports, language/runtime minimum,
and tested server baseline it supports. Releases need not have equal feature coverage.

### 6. Publish independently

Set stable module/package paths before registry publication. Release CLI and
clients independently, run change-scoped CI, and maintain a shared feature matrix.
Publish CLI binaries/checksums for Linux/macOS/Windows on amd64/arm64, test native
TLS/login/process behavior, and document the optional external OpenSSH dependency.
Coordinate server documentation with the first CLI release. Add installer and
package-manager distribution after release artifacts and upgrade paths work.
Local preparation now provides six versioned archives, checksums, license and
completion files, Bash/PowerShell installers, generated Homebrew/Scoop manifests,
and a manual candidate-validation workflow. See `releases.md` for offline testing
and proposed distribution commands. No hosted installer, customer-ready formula, bucket or release
has been published; tap source and CI are pushed privately. The separate tap repository exists and local Homebrew lifecycle
checks pass. Native Windows CLI, PowerShell 7/5.1 installer and Scoop lifecycle checks now pass
in authorized private CI. See `native-validation.md` for exact revisions/runs.

## Latest integration evidence

The latest 2026-10-11 local refresh is recorded in CLI validation commit
`fe719db` and Go-client commit `17e146b`, against server HEAD
`1a266cc8bf6aa726c4890b78ddf95a3aadda0ba0` plus assessment changes. The server
snapshot includes fail-closed request quoting, the corrected generated-secret
contract and immutable translation publication. `contract/server-baseline.json`
records 63 source and 59 test fingerprints. It freezes the tested assessment working
tree, including later plan-support changes; subsequent edits require a fresh
snapshot.
Its review date uses UTC, while this progress record uses the local date.

Passing coverage: 34 decoder/18 encoder fixtures, 7 targeted security cases,
11 HTTP subtests, actual OpenSSH subscription/credential/output checks, 17
auth/key/protocol EUnit tests, 10 real-HTTP import-access cases and 15 publication
cases. The optional `--imports` runner executes the assessment suite cases and
callbacks with local fixtures in the disposable namespace. All checks passed;
this is focused coverage, not the full server security campaign. Later assessment
edits need a fresh snapshot. Browser-only workspace/event APIs and production
release gates remain open.

Homebrew preparation uses `bibleit-labs/homebrew-tap`, formula `bibleit-cli`,
and executable `bibleit`. The future one-step installation command is
`brew install bibleit-labs/tap/bibleit-cli`. The formula at tap commit `7834751` passed isolated
local install/test/reinstall/completions/uninstall using development archives.
The generator emits `bibleit-cli.rb`; no customer-ready formula or release was published. The prepared tap source branch
may now be pushed privately.
Homebrew itself may fetch developer gems/core metadata into the disposable prefix.

## Simplified authentication and tap CI

Named profiles are removed before the first release. The CLI stores one cached
HTTP token and optional SSH identity path in a version-2 service-bound file. HTTP
is the default; --transport ssh is explicit. SSH login preserves the HTTP token;
auth info reports the selected transport's actor, which may differ between HTTP
and SSH credentials. BIBLEIT_TOKEN overrides HTTP data requests only and is never
saved. Logout revokes the cached HTTP token and clears cached selections after
successful revocation; failure preserves them. Old profile/legacy files remain
untouched and require backup plus fresh login. Secure Unix modes and Windows ACLs
remain mandatory. Profiles are not part of future workspace selection.

The separate tap now has push/PR/manual tap syntax/style validation and release
updater regressions. Its manual versioned updater downloads the CLI-generated
formula and release archives, verifies checksums/provenance and opens a reviewed
release/vX.Y.Z PR with a Conventional Commit; it never merges or creates a CLI
release. Private CLI downloads need a contents-read BIBLEIT_CLI_READ_TOKEN secret
in the tap. Actions PR creation must be permitted before invoking publication.
The generated formula, manifest and checksums must accompany the CLI release.
Unpublished development URLs are not fetched by tap syntax CI; actual installation
continues in CLI candidate CI and must be rechecked against published release URLs.

## First release readiness review and ordered next steps

Reviewed 2026-10-11 against local sources and authenticated GitHub state. This
review is complete; the first customer release is not ready. Core scope and
private development validation are complete. There are no Go, CLI or tap releases
published. Python/Ruby/Rust clients, workspace/events, additional package channels
and automated release uploading are outside the initial release requirement.

### Completed preparation

- Go client native matrix: Windows/Ubuntu/macOS passed at 17e146b in
  [run 38091438398](https://github.com/bibleit-labs/bibleit-go/actions/runs/38091438398).
- CLI native and development candidate matrices: passed at fe719db with exact
  client 17e146b in [native run 38091481226](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481226)
  and [candidate run 38091481159](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481159).
  These cover single credential configuration, output/completion, native installers,
  PowerShell 7/5.1, Homebrew and Scoop against development fixtures. Six targets
  cross-build; Windows arm64 has not been validated on native arm64 hardware.
- Tap Homebrew syntax/style and seven updater regressions passed at 04e8b16 in
  [run 38092320412](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092320412).
  Private CLI release-list access using BIBLEIT_CLI_READ_TOKEN passed at bf6a46b in
  [run 38092683138](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092683138).
  Actions PR creation is enabled, default workflow permissions stay read-only,
  and only the updater job requests write permissions. The user confirmed the
  secret selects only bibleit-cli with Contents: read-only.
- Prepared sources and documentation are pushed to private branches. Subsequent
  documentation commits do not change the tested runtime code. No tags, releases,
  merges or visibility changes were performed during this review.

### Remaining release gates, in order

| Order | Item | Current status / required input | Done when |
| --- | --- | --- | --- |
| 1 | Confirm version, scope and customer access | Project decision; propose Go/CLI v0.1.0 for the existing core command surface, Homebrew/Bash/PowerShell/manual archives | Version and supported transports/architectures recorded; choose public customer artifacts or implement authenticated delivery; choose whether Scoop bucket ships now or later |
| 2 | Configure production service and safe test fixtures | Blocked on actual HTTPS URL, SSH host/port and deployed service | Change project-owned constants without exposing endpoint overrides; isolate disposable fixtures from production; verify stored credential bindings safely reject the old service |
| 3 | Freeze and validate server compatibility | Pending agreed committed/deployed server revision | Refresh synchronized client/CLI baseline only after full HTTP/SSH/auth/import integration passes; run explicit controlled production TLS, browser login/token revocation and SSH host verification checks |
| 4 | Release Go client and pin standalone CLI dependency | After gates 1–3; go.mod still requires v0.0.0 | Review/promote the client source, publish approved semantic version, pin it in CLI and record sums; clean-cache GOWORK=off test/vet/build succeeds without sibling checkout or replace directives; private access uses appropriate Git credentials/GOPRIVATE |
| 5 | Validate exact production CLI candidate | After gate 4; current candidate workflow uses --local and go work | Add a standalone production candidate validation path using the released client, GOWORK=off and no --local; clean Go 1.27.1 source produces six archives, installer scripts, formula, Scoop manifest, checksums and provenance; rerun relevant native/platform/installer checks on those exact inputs |
| 6 | Finalize release sources, notes and delivery policy | Before publishing CLI; release/default branches still need review | Promote reviewed source through normal review; notes name client version/server baseline, protocol limits, old-config migration/revocation, OpenSSH dependency and supported architectures; choose signing policy and test downloaded macOS binaries with normal OS protections; record manual release procedure or implement upload automation |
| 7 | Publish CLI assets and update package channels | After gate 6 and publication authorization | Approved tag matches source/manifest; customer-accessible assets are attached and hashes verified; dispatch tap updater for real release, review its PR and run/approve validation before merging; publish Scoop bucket only if selected |
| 8 | Verify fresh customer installs and maintenance | After publication, before announcing availability | Published Homebrew, Bash, PowerShell and manual installs/upgrade paths pass on clean native hosts; selected Scoop channel passes; document uninstall, support, credential preservation, token rotation and package rollback procedure |

The publication dependency order is Go client → pinned/validated CLI → CLI release
assets → reviewed Homebrew formula (and Scoop bucket if selected). A reviewed
manual first release is sufficient; an automated tag/upload workflow is optional.
Native installer validation already passed for development candidates, so repeat
it on final changed release inputs rather than treating all tooling as unbuilt.

### Concrete evidence and limits

- The service constants remain http://localhost:8080 in the Go client and
  127.0.0.1:2222 in the CLI. The production packager intentionally rejects these
  placeholders. Native process tests and real-server harness use loopback fixtures;
  adapt them with the endpoint change while keeping the public API credential-only.
- Both contract/server-baseline.json files are identical and record server HEAD
  1a266cc8bf6aa726c4890b78ddf95a3aadda0ba0 plus an assessment working-tree snapshot.
  At this review, 1/63 recorded source fingerprints differs:
  src/http/bibleit_http_listener.erl. All 59 recorded test fingerprints match.
  This is drift evidence, not a regression finding; compatibility with that later
  edit has not been tested. Do not replace baseline hashes merely to hide drift.
- CLI go.mod still requires bibleit-go v0.0.0. Draft CI pins exact client 17e146b
  in an ephemeral workspace and packages with --local. Green draft CI does not
  certify standalone dependency retrieval or production release provenance.
- All three repositories are private. Generated ordinary GitHub release URLs do
  not authenticate customers. The intended open source launch requires approved
  public client/CLI/tap access, or a deliberately tested authenticated channel.
  Server visibility remains private. Source push authorization does not authorize
  visibility changes or customer release publication.
- The tap formula is still v0.1.0-dev.1 with unpublished URLs. Read-secret and PR
  policy setup are complete; actual release archive verification, bot PR creation
  and customer formula installation await the first production CLI release.
- Current default branches are CLI package-manager-validation, Go
  private-platform-validation and tap formula-preparation. Review/promote the
  approved source before release; do not accidentally tag stale default-branch
  code. Continue on integration-refresh (CLI/client) and formula-preparation (tap),
  using Conventional Commits and no codex prefix.

### Next actionable increment

Obtain the production HTTPS URL, SSH host/port and intended deployed server
revision, plus the first-release version/access decision. Then change service
constants and endpoint-dependent fixtures together and run the frozen integration
gate. Until deployment inputs are ready, keep validation as private development
staging. The seven core contract items and tap updater setup remain complete.

## Sources reviewed

- `../bibleit-server/docs/reference/protocol.md`
- `../bibleit-server/docs/reference/http.md`
- `../bibleit-server/docs/domain-model.md`
- `../bibleit-server/docs/live-design.md`
- `../bibleit-server/src/protocol/bibleit_protocol.erl`
- `../bibleit-server/src/http/bibleit_http_view.erl`
- `../bibleit-server/src/http/bibleit_http_listener.erl`
- `../bibleit-server/src/http/bibleit_http_management_ws.erl`
- `../bibleit-server/src/auth/bibleit_authorization.erl`
- `../bibleit-server/src/http/bibleit_cli_auth.erl`
- `../bibleit-server/src/ssh/bibleit_ssh_listener.erl`
