# Local release packaging and distribution

Release publication is paused. The commands below prepare files locally; no
release tag, GitHub Release, distributed formula, Scoop bucket or hosted installer
is created. Private validation branches and CI are authorized. The separate
`bibleit-labs/homebrew-tap` exists; prepared source branches may be pushed
to the private repositories as of 2026-10-11. Customer package publication remains
gated on production readiness.
Production service addresses and a published Go-client dependency remain release
gates. Local candidates still connect to `http://localhost:8080`.

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

1. Set the production HTTPS/SSH addresses, adapt the localhost-based test fixtures
   and rerun integration against the agreed server baseline. The tracked core
   protocol discrepancies are resolved.
2. Release the reviewed Go client when authorized, pin its version in go.mod,
   record dependency sums and verify builds with GOWORK=off. No replacement modules.
3. Commit reviewed sources and use a semantic version. Running the packager without
   `--local` requires Go 1.27.1, a clean checkout, a pinned client and production
   addresses. It still does not publish or tag anything.
4. The draft `release-candidate.yml` workflow validates development candidates in
   private CI, using an exact private Go-client SHA and ephemeral workspace. Its
   `BIBLEIT_CLIENT_READ_TOKEN` secret must have contents-read access only to
   `bibleit-go`. It tests native installers, Homebrew and Scoop, saving private
   validation artifacts with no Release/tag write permission. Restore standalone
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
Campaign progress is 24/25; final-image/Fly deployment verification remains open.
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
Deployment-specific TLS, image contents and rollback checks remain package 25
work. No production migration or deployed compatibility certification is claimed.

## Package 25 local deployment evidence and launch gates

The server assessment's local full suite and image driver passed (326 integration,
51 Erlang unit, 71 JavaScript, native coordinate corpus and ten SSH-bootstrap tests; 16 isolated
image/database/TLS/recovery checks). Its final local image runs on native ARM64.
These results do not replace the client's earlier verified compatibility snapshot.
Assessment progress remains 24/25 until approved deployed-boundary evidence exists.

Before customer publication:

- Choose and verify the canonical HTTPS origin. BIBLEIT_PUBLIC_URL is validated as
  an origin and requires HTTPS with secure cookies; browser Host/Origin must match.
  Align the CLI's fixed service/browser URL before testing actual approval/login,
  code exchange and revocation through the approved edge. Customers get no override.
- Verify a final server image on native amd64 and the actual Fly certificate,
  routing/proxy identity, fairness, spoofing/bypass and platform configuration.
  Passing CLI/client amd64 jobs do not certify the server image or Fly edge.
- Confirm writable server volumes for UID/GID10001. Legacy UID100 storage needs
  an approved offline ownership migration that preserves private key permissions
  and SSH host identity; do not silently replace the host key. This does not
  change the CLI's per-user credential-file permissions.
- Review the actual database attach result, role/grants and migration lineage.
  Missing sslmode uses verify-full and bad CA/hostname fails closed. Any explicit
  sslmode=disable needs approved private-mesh trust; no automatic downgrade or
  database reset is implied. Keep credential-bearing URLs out of operational logs.
- Record deployed backups/recovery and rollback, package inventory and persistent
  SSH trust across restart/replacement. Local distribution/EPMD listeners remain
  disabled; the normal Bibleit SSH protocol stays available.

No deployment, ownership migration, TLS policy change, client runtime change or
baseline rewrite is performed by this documentation update.
