# Local release packaging and distribution

Publication is paused. The commands below prepare files locally; no release tag,
GitHub Release, Homebrew tap, Scoop bucket or hosted installer is created.
Production service addresses and a published Go-client dependency remain release
gates. Local candidates still connect to `http://localhost:8080`.

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
| `bibleit.rb` | Homebrew formula with actual archive checksums |
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
# Homebrew: publish the generated formula in a future bibleit-labs/homebrew-tap.
brew install bibleit-labs/tap/bibleit

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

1. Set the production HTTPS/SSH addresses and resolve outstanding server contract
   discrepancies; rerun fixed-endpoint integration against the agreed baseline.
2. Release the reviewed Go client when authorized, pin its version in go.mod,
   record dependency sums and verify builds with GOWORK=off. No replacement modules.
3. Commit reviewed sources and use a semantic version. Running the packager without
   `--local` requires Go 1.27.1, a clean checkout, a pinned client and production
   addresses. It still does not publish or tag anything.
4. Run the prepared `release-candidate.yml` workflow after publication is authorized.
   It manually validates standalone candidates and native installers on macOS,
   Linux and Windows; it saves a candidate artifact, with no Release/tag write
   permission. The workflow has not been pushed or dispatched.
5. Validate actual native Windows installation/replacement, macOS downloaded-binary
   behavior/signing policy, Homebrew install/test and Scoop install. A Ruby syntax
   check and generated checksum validation are preparation, not package-manager
   end-to-end certification.
6. Review release notes covering client version, server baseline, supported commands,
   service addresses, migration notes, OpenSSH dependency and known limitations.
   Obtain authorization to publish, then create the release and update tap/bucket.
   Consumers must never be directed to local-only candidates.

Local evidence for this increment: macOS/arm64 and disposable Linux/arm64 offline
installation/replacement and corrupt/missing/duplicate checksum checks; PowerShell
parser/failure-path checks in a disposable Linux container; all six archive layouts,
eight checksums and package manifest values; two identical archive builds. Native
Windows and actual Homebrew/Scoop installation remain pending.
