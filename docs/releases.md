# Local release packaging and distribution

Release publication is paused. The commands below prepare files locally; no
release tag, GitHub Release, distributed formula, Scoop bucket or hosted installer
is created. Private validation branches and CI are authorized. The separate
`bibleit-labs/homebrew-tap` exists; prepared source branches may be pushed
to the private repositories as of 2026-10-11. Customer package publication remains
gated on production readiness.
The fixed canonical addresses are configured: `https://bibleit.app`
and ordinary SSH at `bibleit.app:22`. Custom-domain DNS/TLS/SSH qualification
remains pending; earlier Fly-hostname checks do not certify the new route. A released, pinned Go-client
dependency and standalone production candidate validation remain release gates.

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
   Final packaging-input freeze and actual new-domain DNS/TLS/SSH qualification remain
   assessment/Fly gates; revalidate changed inputs before release.
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
saved-host-key rejection, at assessed 5969f2f. The final boot change passes
canonical/provider unit and synthetic hostname configuration checks. Actual
DNS/TLS/SSH and authentication qualification is still required by the assessment/Fly chats before
customer release. The prior nine green native/candidate jobs cover the Fly-origin
increment, and must not be presented as new-domain validation.
