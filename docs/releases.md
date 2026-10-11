# Local release packaging and distribution

Release publication is paused. The commands below prepare files locally; no
release tag, GitHub Release, distributed formula, Scoop bucket or hosted installer
is created. Private validation branches and CI are authorized. The separate
`bibleit-labs/homebrew-tap` exists; prepared source branches may be pushed
to the private repositories as of 2026-10-11. Customer package publication remains
gated on production readiness.
The fixed canonical addresses are configured: `https://bibleit.app`
and ordinary SSH at `bibleit.app:22`. Actual public custom-domain DNS/TLS/SSH
technical acceptance and human Google/GitHub sign-in now pass; real Turnstile
fresh-token acceptance and email delivery/verification are human-confirmed, but
real replay rejection failed and requires assessment repair/retest.
Native HTML form technical admission is qualified on repaired deployment 34d80f8.
A checksummed client snapshot and standalone candidate validation now pass.
An approved semantic Go-client version and final release-input qualification
remain launch steps.

## Public distribution versus private validation

Private source pushes and Actions artifacts are development staging, not a
customer download channel. The generated formula, Scoop manifest and installers
use ordinary GitHub Release URLs without GitHub credentials. Before public launch,
make the intended client/CLI/tap repositories and release assets publicly
accessible, or deliberately implement and test a separate authenticated/private
installation channel. Do not change repository visibility during readiness review.
A private Go client also requires Git authentication and `GOPRIVATE` for fresh
standalone builds; this is not solved by pinning a version alone.
See [Go private modules](https://go.dev/ref/mod#private-modules) and
[GitHub release asset access](https://docs.github.com/en/rest/releases/assets).

The packager prepares artifacts but does not upload, tag or publish them. Use a
reviewed manual release procedure for the first release, or add an explicit
publication workflow later. The candidate workflow has contents-read permission
and remains validation-only. See the ordered launch checklist in [plan.md](plan.md).

## Prepare and check a development candidate

Use Go 1.27.1, Python 3.11+, Git, and the ignored workspace joining this repository
with `../bibleit-go`. From the CLI repository:

```sh
python3 scripts/package_release.py --local --version v0.1.0-dev.1 --output build/releases/local-dev-1
python3 scripts/check_release.py --assets build/releases/local-dev-1
```

Choose a new output directory for each candidate. Existing directories are never
overwritten. `--local` records the candidate as development-only in `manifest.json`
and permits the unpublished workspace dependency. The version is injected into
`main.cliVersion`; no Git tag is necessary or created. Plain `go build` still
reports `dev`. The packager uses CGO_ENABLED=0, trimpath and fixed archive timestamps.
Identical inputs produce identical archive/checksum bytes; the manifest separately
records the source commit, dirty state, toolchain and client module metadata.

| Output | Contents / purpose |
| --- | --- |
| `bibleit_<version>_darwin_{amd64,arm64}.tar.gz` | macOS binary |
| `bibleit_<version>_linux_{amd64,arm64}.tar.gz` | Linux binary, no libc dependency |
| `bibleit_<version>_windows_{amd64,arm64}.zip` | Windows executable |
| `checksums.txt` | SHA-256 for six archives and both installer scripts |
| `install.sh`, `install.ps1` | Bash and PowerShell installers |
| `bibleit-cli.rb` | Homebrew formula with actual archive checksums |
| `bibleit.json` | Scoop manifest for Windows amd64/arm64 |
| `manifest.json` | Candidate build provenance and local-only marker |

Every archive includes LICENSE, README, and Bash/Zsh/Fish completion scripts.
The program itself needs no Go or Python runtime. SSH commands require a separate
OpenSSH client; HTTP commands do not. Bash installation requires curl, tar and
sha256sum or shasum. Windows installation requires PowerShell 5.1+ or PowerShell 7.

## Test installation without downloading anything

macOS/Linux:

```sh
BIBLEIT_INSTALL_ASSETS="$PWD/build/releases/local-dev-1" \
  BIBLEIT_INSTALL_DIR="$PWD/build/install-test" bash install.sh v0.1.0-dev.1
build/install-test/bibleit version
```

Windows PowerShell:

```powershell
./install.ps1 -Version v0.1.0-dev.1 -AssetsDir ./build/releases/local-dev-1 -InstallDir ./build/install-test
./build/install-test/bibleit.exe version
python scripts/check_release.py --assets build/releases/local-dev-1
./scripts/check_powershell_installer.ps1 -AssetsDir ./build/releases/local-dev-1
```

Both installers validate the selected version, require exactly one matching
checksum, extract only the expected binary, and check its version before replacing
an existing installation. They install per user, need no administrator privileges,
and leave shell profiles, persistent PATH and credential files untouched. Default
locations are `$HOME/.local/bin` and `%LOCALAPPDATA%\Bibleit\bin`; add them to PATH
when needed. Updating means rerunning the installer with a new version. Uninstall
by removing the binary; account credentials remain under the user's control.
`BIBLEIT_INSTALL_ASSETS`/`-AssetsDir` select local release files for testing, and
do not alter the client's service endpoint.

Online installs resolve latest to one concrete release tag before obtaining its
archive and checksums. Bash enforces HTTPS for downloads and redirects; PowerShell
uses the fixed GitHub HTTPS release URLs. Checksums detect damaged/mismatched files;
they are not a separate publisher signature. Neither script disables OS security
controls or removes macOS quarantine attributes.

## Distribution plan after publication resumes

These are proposed public commands, **not available installation paths today**:

```sh
# Homebrew: publish a release-matching formula in bibleit-labs/homebrew-tap.
brew install bibleit-labs/tap/bibleit-cli

# Bash: release assets host the installer; no website/domain assumption needed.
curl -fsSL https://github.com/bibleit-labs/bibleit-cli/releases/latest/download/install.sh | bash
# Specify an immutable version using: ... | bash -s -- vX.Y.Z
```

```powershell
# PowerShell: the installer is also a GitHub Release asset.
irm https://github.com/bibleit-labs/bibleit-cli/releases/latest/download/install.ps1 | iex
# Scoop: publish the generated manifest in a future bibleit-labs/scoop-bucket.
scoop bucket add bibleit https://github.com/bibleit-labs/scoop-bucket
scoop install bibleit/bibleit
```

GitHub Releases will also offer direct archives and checksums for offline/manual
installation. Homebrew installs completion files automatically. Script/manual users
can install completions using `bibleit completion bash|zsh|fish`; see the CLI help.
`go install github.com/bibleit-labs/bibleit-cli@vX.Y.Z` can be a developer path once
the Go client is released and pinned, but Go source builds report `dev` unless the
version is injected. npm adds little value to this Go binary. Defer winget,
Debian/RPM repositories, Nix and container distribution until demand justifies
their publishing and update maintenance.

This channel selection was inspired by [Resend CLI's installation paths](https://github.com/resend/resend-cli).
The formula follows [Homebrew's formula conventions](https://docs.brew.sh/Formula-Cookbook),
and the Windows manifest follows [Scoop's app manifest format](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifests).
The implementation here is written for Bibleit's Go release layout.

## Release gates and future validation

1. Canonical HTTPS/SSH constants and safe fixtures are configured. Frozen
   HTTP/SSH/import and boot validation pass at assessed 5969f2f/native f6fed37.
   Supplemental alias boot qualification passes at ef63e91. Actual public-DNS
   HTTPS/aliases/auth/SSH/rate acceptance passes on deployed c586107. The operator
   confirms Google/GitHub sign-in; real Turnstile backend success/replay remains
   an assessment/Fly gate, together with actual email delivery/verification.
   Repaired deployment 34d80f8/native f6fed37 passes real native form CSRF admission
   and updated boundary/alias checks; provider traffic was blocked in that probe;
   freeze final packaging inputs and revalidate changed inputs before release.
2. Release the reviewed Go client when authorized, pin its version in go.mod,
   record dependency sums and verify builds with GOWORK=off. No replacement modules.
3. Commit reviewed sources and use a semantic version. Running the packager without
   `--local` requires Go 1.27.1, a clean checkout, a pinned client and production
   addresses. It still does not publish or tag anything.
4. The draft `release-candidate.yml` workflow validates development candidates in
   private CI, using an exact private Go-client SHA and ephemeral workspace. Its
   `BIBLEIT_CLIENT_READ_TOKEN` secret must have contents-read access only to
   `bibleit-go`. It builds six-target archives once on Linux, then tests the same downloaded
   inputs with native installers, Homebrew and Scoop, saving private validation artifacts with no Release/tag write permission. Restore standalone
   packaging without `--local` and use GOWORK=off after the client is released.
5. Validate actual native Windows installation/replacement, macOS downloaded-binary
   behavior/signing policy, Homebrew install/test and Scoop install. A Ruby syntax
   check and generated checksum validation are preparation, not package-manager
   end-to-end certification.
6. Review release notes covering client version, server baseline, supported commands,
   service addresses, migration notes, OpenSSH dependency and known limitations.
   Obtain authorization to publish, then create the release and update tap/bucket.
   Consumers must never be directed to local-only candidates.

Local evidence for the packaging increment: macOS/arm64 and disposable Linux/arm64 offline
installation/replacement and corrupt/missing/duplicate checksum checks; PowerShell
parser/failure-path checks in a disposable Linux container; all six archive layouts,
eight checksums and package manifest values; two identical archive builds. Native
Windows PowerShell 7/5.1 installers and Scoop lifecycle now pass, together
with Linux/macOS installers and macOS Homebrew, in
[private run 38088193158](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38088193158).
Homebrew also passes local macOS/arm64 checks against the separate tap formula
in a disposable prefix; no package remains in the normal user prefix. Public
download URLs, production addresses and standalone pinned-client release builds
remain unvalidated and publication remains paused.

## Separate Homebrew repository

The formula is named `bibleit-cli`; the installed executable remains `bibleit`.
The canonical command after publication is `brew install bibleit-labs/tap/bibleit-cli`.
Homebrew maps `bibleit-labs/tap` to `bibleit-labs/homebrew-tap`; `/cli/` would
require `homebrew-cli`, while `homebrew-taps` would map to `/taps/`.
A bare `brew install bibleit-cli` requires prior tap setup (and formula trust on
Homebrew versions that require it), or future acceptance into `homebrew/core`.
The fully qualified form is the supported one-step installation command.
See [Homebrew tap naming](https://docs.brew.sh/Taps) and the
[Resend example](https://github.com/resend/homebrew-cli).

`bibleit-labs/homebrew-tap` owns `Formula/bibleit-cli.rb`, its README, license and
contribution notes. The CLI generates versioned formulas with release archive
checksums. After an authorized release, copy the matching formula into the tap,
review and validate it, then commit its version update independently. The tap provides a separate manual update-formula workflow that verifies
published artifacts and opens a PR; it never auto-merges.

```sh
python3 scripts/check_homebrew.py --assets build/releases/v0.1.0-dev.1 \
  --formula ../homebrew-tap/Formula/bibleit-cli.rb
```

The validation runner changes only archive URLs to local files. It uses the tap's
real install/test methods and checksums and writes evidence to
`build/validation/homebrew.json`. CLI archives stay local; Homebrew may fetch
its own developer gems or core metadata into the disposable prefix. This check
is not a guarantee of offline execution. On Windows, run the Scoop counterpart with an
existing/pinned source checkout:

```powershell
python scripts/check_scoop.py --assets build/releases/<candidate> --scoop-source <Scoop-source-checkout>
```

It runs real Scoop in a disposable root, checks the native executable shim and
reset/reinstallation/uninstallation, then confirms incorrect hashes are rejected.
Archive downloads use a loopback fixture; the client's fixed endpoint is unchanged.

## Tap CI setup before the first formula update

Tap setup is complete: syntax/style and seven updater regressions are green in
[run 38092320412](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092320412).
BIBLEIT_CLI_READ_TOKEN is configured and actual private CLI release-list reads
passed in [run 38092683138](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092683138).
Actions PR creation is enabled with read-only defaults; the updater job alone
requests contents-write and pull-requests-write. Keep the token scoped only to
bibleit-cli with Contents: read-only and rerun Check updater access after rotation.
Before dispatching the real update workflow, publish bibleit-cli.rb, manifest.json
and checksums.txt alongside the four macOS/Linux release archives. The
workflow consumes a reviewed version, verifies production provenance and archive
hashes, then opens a Conventional Commit release PR. It does not create CLI
releases or automatically merge the formula. The updater has fixture regression
evidence but no real-release download/PR evidence yet because no release exists.


## First-release execution checklist

Use the eight ordered gates in [plan.md](plan.md) as the readiness record.
Development candidate CI is already green; it still uses a workspace and --local,
so add a standalone production validation path after releasing and pinning the
Go client. Build from the reviewed CLI source, not an older default branch.

For an authorized manual first release:

1. Record the approved Go and CLI versions, production destinations, deployed
   server baseline and exact reviewed source commits. Complete compatibility checks.
2. Publish the approved Go client version first. Pin it in CLI, tidy module metadata
   and validate dependency retrieval with a fresh cache and GOWORK=off.
3. Commit the final CLI inputs and package without --local into a new directory.
   Inspect manifest.json (local_only=false, dirty=false, correct source and client),
   verify archive checksums/version and run final standalone/native installer checks.
4. Prepare release notes and choose customer access/signing policy. The CLI assets
   must include all six archives, checksums.txt, manifest.json, install.sh,
   install.ps1, bibleit-cli.rb and bibleit.json. Publish only the approved assets
   against the corresponding reviewed tag/source after publication authorization.
5. Dispatch the tap updater with that exact published tag. Review its generated
   release PR, run/approve tap validation and merge through normal review. Publish
   a Scoop bucket only if selected for this release; a generated manifest alone
   does not establish that channel.
6. Validate the actual customer download URLs and clean-machine installs/upgrades,
   including macOS download protections. Record support and rollback steps before
   announcing availability. Development fixtures are not a substitute for this check.

No publication command is executed as part of the readiness review. Automated
release upload, additional language clients and extra package ecosystems can be
added later without delaying a reviewed manual release of the existing core scope.


## Service recovery and credentials

The server assessment's recovery procedure quarantines restored credentials
before reopening a restored service. This is an explicit offline operator action;
it is not triggered by CLI startup, login or a client upgrade. The procedure has
only been exercised against isolated restored test databases so far. A backup can
contain credentials or grants revoked after the backup time, so recovery requires
identity and current-access reconciliation before users are re-enabled.

Package 23's final local suite passed 313 integration cases, 48 Erlang unit tests,
71 JavaScript tests and native search/fuzz tests. Its focused restore/fault cases
and 120-second database pause/restart campaign passed; restored credentials were
denied after quarantine. That result completed package 23; the subsequent package
24 retest brings current assessment progress to 24/25, with deployment verification
still open. This does not establish production
RPO/RTO or deployed backup/failover readiness; those remain release operations
checks. The client compatibility baseline is unchanged by this result.

Customer guidance after an actual approved restore/quarantine:

1. Wait for support to confirm your account and current access are ready. Old
   passwords, API tokens, browser sessions and registered user SSH keys remain
   invalid; repeated CLI requests do not restore them.
2. Complete the approved account recovery/password reset or identity-provider
   sign-in, then run `bibleit auth login` for a fresh cached HTTP credential.
   For an API-token client or automation, issue a fresh token through the approved
   account flow and replace the token in the application or secret store. Replace
   BIBLEIT_TOKEN environment values too; they override cached HTTP credentials.
3. Re-register the intended SSH public key in account settings, then use
   `bibleit --transport ssh auth login --identity <private-key-path>`. Quarantine
   revokes the server's user-key registration; it does not modify your private key
   or automatically register it again.
4. Keep SSH host verification enabled. User SSH keys and the service host key
   have different roles. A preserved service host key keeps existing trust; any
   deliberate service host-key rotation needs an independently verified support
   announcement and trust-update procedure.

Release/support documentation should describe this recovery path when the
operational procedure is approved. Existing HTTP/SSH errors and command schemas
are unchanged. No automatic credential clearing, retries, grant restoration or
runtime recovery feature is introduced by this documentation update.


## Server rollout compatibility checks

Package 24's final local suite passed 324 integration, 50 Erlang unit, 71 JavaScript
and native tests with no skips. Ordinary release/production tar checks verified
178 reviewed resources; the Linux build stage passed with OTP 29.1.1, SSH 6.0.6
and SSL 11.7.7. Independent client HTTP/SSH/import compatibility also passes the
67-source/66-test/11-native-input development snapshot. CLI schemas are unchanged.
Campaign progress was 24/25 at that checkpoint; final closure is recorded below.
Before a production release, validate the approved deployed server snapshot and
these operational requirements together:

- Migration journals must carry verified SQL filename/SHA-256 identity. Startup
  rejects missing/mismatched identity with schema_identity_mismatch. Establish
  lineage and perform an audited canonical-schema migration; do not blindly
  backfill hashes or reset an existing user/shared database. Any disposable
  development rebuild needs explicit approval.
- Provider/download HTTPS uses bounded CA/hostname verification. Reviewed imports
  must use approved HTTPS eBible sources; fetch_failed remains a safe import
  failure without a published partial installation. Keep real URL/authority
  checks in integration even when using approved local delivery fixtures.
- Use guarded server release/tar packaging and verify actual deployment assets
  exclude development private keys and assessment records. The default server
  runtime disables inbound distributed Erlang/EPMD and remote console/RPC; this
  does not disable the Bibleit SSH command transport.

These are server rollout prerequisites. The CLI does not repair database lineage,
change server administration policy or retry an unsafe import automatically.
Deployment-specific TLS/image/rollback evidence was subsequently recorded by the
assessment. These older client checks do not independently certify that newer deployment.

## Package 25 local deployment evidence and launch gates

The latest server assessment local full retest passes 331 integration, 51 Erlang
unit, 71 JavaScript, native and ten SSH-bootstrap tests. Updated ARM64 image checks
pass all 17 cases. Separate native Linux amd64 image validation also passes at
f5f1736 in [run 38101109368](https://github.com/bibleit-labs/bibleit-server/actions/runs/38101109368);
see native-validation.md for the exact image/report hashes and scope. The native
image job runs native/bootstrap/image checks, not the complete CT suite.
Independent HTTP/SSH/import integration already passed against the frozen
package-25 development source, and both client/CLI baselines were refreshed.
All those recorded source/test/native inputs match the native image report.
These results still do not certify the final deployed release.
Assessment progress was 24/25 at this local checkpoint; the later deployed closure
below supersedes its pending assessment gates.

The later production qualification is complete; the following release steps are
still required:

- Configure the fixed service/browser origin https://bibleit-server.fly.dev and
  SSH bibleit-server.fly.dev:22 together with isolated fixtures and credential
  bindings; run fresh client HTTP/SSH/import checks against a frozen newer server.
  Clients send ordinary SSH, with no PROXY preamble or customer endpoint override.
- Freeze approved release inputs, preserve qualified image/migration/host identity
  and validate changed inputs. New domain/routing changes need requalification.
- Confirm version and public/authenticated customer delivery, release/pin the Go
  dependency and validate standalone production CLI packaging and native installers.
- Publish only approved tags/assets, then validate real tap updates and customer
  installs/upgrade/rollback. Existing private workspace builds are development candidates.

No deployment, ownership migration, TLS policy or client runtime change was
performed. Baselines changed only after independent integration passed.

## Final server assessment closure — 25/25

Ledger 8923d86 closes all 25 scoped packages. Qualified deployed source is
7442010eebcfb1a2a9105db312ebaeae5f789003 with native f6fed37 and registry digest
sha256:44f2278362bd73f26a2f1c4abb78bf4d72eb02a99c177629656b72391c79fc11.
Its [native run](https://github.com/bibleit-labs/bibleit-server/actions/runs/38106224843)
and rebuilt deployment image campaigns each pass 20 checks. The final source suite
passes 334 CT, 56 Erlang unit, 71 JavaScript, native and ten bootstrap tests.

Assessment-owned deployed results pass 23 HTTPS, 11 authenticated HTTP/WSS,
four SSH authentication and 13 paired-client spoofing/isolation/recovery checks.
Seven migrations are retained; isolated on-Fly restoration and provenance-guarded
fixture cleanup pass. At this historical checkpoint these server results had not refreshed the CLI
client baseline or endpoint constants; the following increment supersedes that state.

Scope limits remain: only the Fly hostname is qualified, per-node fixed-window
admission/reset/shared NAT behavior, single Machine/volume, five-day backup retention
without continuous RPO/full failover RTO, incomplete provider form/email/OAuth
acceptance and operator deploy-token scope/expiry review. Preserve approved database
trust/grants/migrations, volume ownership and SSH host identity on future rollout.
No CLI runtime, baseline, server deployment or customer publication is changed here.

## Fixed production endpoints and fresh client compatibility — 11 October 2026

Clients now target the qualified HTTPS origin and SSH port 22. Credential-only
constructors, redirect blocking, certificate and OpenSSH host verification remain
in force. Existing schema-2 localhost credentials are rejected without overwriting
or forwarding them; back up the old config and authenticate with a fresh config.

Fresh integration uses those exact destinations through an egress-disabled
container's loopback TLS/SSH fixtures. It passes 34 decoder/18 encoder fixtures,
seven security cases, 11 Go HTTP subtests, OpenSSH subscription/pause/reconnect/
closure and JSON parity, 18 auth/key/protocol tests, and 26 import cases. Baselines
record clean server 8923d86, native f6fed37, and 68 source/70 test/11 native hashes.
TLS positive, wrong-name and untrusted-CA checks pass. This is independent client
compatibility evidence, not a new test of the public Fly routing or provider login.

Next: confirm release version and customer access, release/pin the Go dependency,
validate a standalone production candidate with GOWORK=off and no --local, then
review/publish approved assets and test the real tap updater/customer installations.

## Canonical custom-domain migration — 11 October 2026

The user confirmed bibleit.app as the main service and ordinary SSH host, with
DNS-only apex routing being prepared by the Fly chat. The legacy
live.bibleit.app/<live_id> mapping is server-owned and redirects to
bibleit.app/lives/<live_id>; API clients use the canonical origin directly and
continue to reject redirects. They expose no endpoint/profile overrides.

SSH retains the historical Fly name solely as a HostKeyAlias trust lookup, so an
existing verified key is reused and a changed key fails. The CLI does not edit
known_hosts. Old service-bound HTTP/SSH credentials fail locally and remain intact.
Back them up and sign in using a fresh config after the approved cutover.

This prepared address change passes isolated HTTP/SSH/import, including changed
saved-host-key rejection, at assessed 5969f2f. That revision’s boot change passes
canonical/provider unit and synthetic hostname configuration checks. A subsequent
pre-cutover check found a missing redirect-host environment boot parser entry.
Repair ef63e91 passes the independent direct boot regression, with committed
source/native equivalence and separate supplemental evidence. The unchanged full
client integration baseline remains 5969f2f. Actual public-DNS technical acceptance
passes on deployed c586107: 23 boundary/39 alias checks per address family, 11 auth,
10 SSH and 15 paired rate cases. The public SSH fingerprint is unchanged and
scoped synthetic fixture counts are zero. Credential-free deployment provenance
is recorded in contract/canonical-deployment-validation.json. The operator confirms
Google/GitHub sign-in; real Turnstile backend success/replay remains required
before release. New-domain client and CLI native matrices pass, as does the
revised candidate build/installer matrix at CLI bc94231. All native installer jobs
validate one shared six-target archive set. See [native-validation.md](native-validation.md)
for the exact private runs. Standalone production dependency/packaging validation
remains a separate release gate.

## Repaired native HTML form deployment qualification

Source 34d80f8/native f6fed37 is technically qualified on image
sha256:33b4ca3c13829c1c25f0745037212cc0c14e5e7768c2f78e58206c461ef8d35d,
[run 38144206075](https://github.com/bibleit-labs/bibleit-server/actions/runs/38144206075).
Real production requestSubmit carries canonical Origin/Referer and passes CSRF
before the expected missing-token denial. The probe blocks provider traffic and
creates no account/email. Updated boundary23/alias39 checks pass per address
family; full source338 CT/63 EUnit/71 JS/native/10 bootstrap passes. Credential-free
report hashes are in contract/canonical-deployment-validation.json, retaining
earlier auth/SSH/rate evidence at its tested source. Human real Turnstile backend
success/replay and actual email delivery/verification remain release gates.
The separately approved catalog rollout uses server 6ba81e2/native d6e31dc
(workflow 38144732991) and requires independent deployment/catalog qualification
before joining the release baseline. The 34d80f8/f6fed37 evidence remains
point-in-time and is not inherited by that newer image. No customer tags/assets are published.

## Focused changed-native/client qualification

The6ba81e2/d6e31dc catalogue deployment is independently qualified; the later
UTF-8 catalogue repair4243c9a passes exact-source focused client qualification:
five typed HTTP groups, Portuguese/German catalogue and named reads, five HTTP/SSH
parity cases, invalid-index publication safety and TLS3 boundaries. Recorded
69source/75test/11native hashes match committed and after-run inputs. The server
full340 CT/66 EUnit/75 JS/native/10 bootstrap suite also passes. Native d6 is
unchanged by the UTF-8 repair and the client decoder remains strict. Earlier broad
protocol integration remains separately attributed. Exact repaired deployment/public
qualification is complete at server 4243c9a/native d6e31dc, image
sha256:46670b5fba63859a6892d64c1501122ecdddfd1edc1410555eba7c165143b5c1 (workflow
38147193441): 406 committed hashes, eight successful stages, 20 actual-image startup
checks, published/running digest parity and 185 APK comparisons with zero fixed-advisory
alerts. Refreshed boundary 23/alias 39 checks pass independently on IPv4 and IPv6;
native form CSRF admission and missing-token denial pass. Production BPML has 66 strict
UTF-8 BOOK records and matching Gênesis. The five public report hashes are recorded in
canonical-deployment-validation.json. Real Turnstile backend success/replay and actual
email delivery/verification remain open.
No customer tags/assets are published.

## Main CI and draft release generation — 11 October 2026

`main` is the integration/default branch for CLI, Go client and tap. Native CI
runs build, race tests, vet and CLI runtime/shell checks on Linux, Windows and
macOS. CLI go.mod/go.sum pin client snapshot
`v0.0.0-20261011144051-464d86d7dd41`; CI and packaging use `GOWORK=off`, with
scoped private module authentication. This is an immutable source dependency,
not a first semantic Go-client release. Replace it with the approved client
release version when that release is ready.

The candidate workflow builds six archives once and validates those identical
artifacts with native installers, PowerShell, Homebrew and Scoop. It is reusable
by `release.yml`. Pushing a `v*` tag invokes native tests, verifies the tagged
commit belongs to main, builds a clean standalone candidate and runs all native
installer jobs before creating a **draft** GitHub release with archives,
checksums, manifest, Bash/PowerShell installers, formula and Scoop manifest.
Existing releases are not overwritten. No tag is created by branch pushes.
Draft generation uses GitHub's [release-create command](https://cli.github.com/manual/gh_release_create);
the workflow reuses the validation jobs through
[workflow_call](https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows).

Before approving the first tag, complete the real Turnstile/mail acceptance gates
and decide the release version and customer access. Review the generated draft
before publishing it. Private release URLs require authentication; this setup
changes neither visibility nor customer access. No customer tag/release is
created by this implementation work.

After publication, dispatch the tap's Update release formula workflow, review its
PR, and dispatch Validate released Homebrew package on that PR branch with the
same tag. It verifies the published archives and formula and performs isolated
Homebrew install/test/completion/reinstall/uninstall checks on macOS and Linux.
Merge the reviewed formula only after those checks pass. Cross-repository updater
access remains Contents-read only; PR creation uses the tap's own GITHUB_TOKEN.

Main CI verification: all 12 hosted jobs pass at CLI 13e4b72, Go 464d86d and
tap a41e453. Native CLI run 38158698440, standalone package/installers run
38158698450, Go run 38158759103, tap validation 38158709825 and tap access
38158727488 are linked in status.md. The exact candidate manifest is clean,
non-local and non-workspace with the pinned client sum; the tap updater verifies
its four formula archives. Tag-triggered draft creation and the first published
tap update remain uninvoked until a release is approved.


Real-provider follow-up: assessment ledger a40d27d confirms actual email delivery
and completed verification, plus real fresh Turnstile acceptance. Replay remains
unconfirmed after original helperfalse400/400; the assessment requests a fresh
retry with its improved classifier. Release publication stays held for that gate,
version/customer-access selection and final semantic-client candidate inputs.
Source14628f3/native d6/image28fb9a0b/run38157279015 is recorded separately from
4243's full client/boundary proof. Main is now the default and release ancestry
branch in all three repos; older master refs remain preserved.

The improved human replay test subsequently confirms duplicate token acceptance,
with CSRF/rate/continuity interference ruled out. Replay is now a failed gate,
not merely inconclusive classification. Assessment repair and deployed fresh-token/
replay retesting are required before draft publication or customer release.

The replay guard is now committed as server d160773/native d6e31dc, with full
344 CT/68 EUnit/75 JS/bootstrap/native validation and focused concurrency/restart/
expiry/database-error/HTTP regression checks green. Deployment image qualification
and genuine fresh-token/replay retest remain required before gate closure.

Replay deployment build38162034582 subsequently failed before publication: SQL008
was missing from the explicit release allowlist. A staged packaging correction
and early eight-SQL/omission coverage pass; production is unchanged and real
replay still fails there. Final corrected-image qualification awaits the Fly
owner's narrow security-release branch approval, excluding undeployed Reader
changes on server main. Only after that rollout can real-provider retesting close
the gate. Customer CLI release and tap formula publication remain held.
