# Release maintenance

The SDK, CLI and tap are public. CLI v0.0.6 is the latest stable release and its
matching tap formula is available. All original v0.0.1 download assets were
withdrawn for missing redistribution notices; its source and tag are preserved.
[Public launch evidence](public-launch.md) records source, checksums, native public
installation results and protections. Historical preparation records remain dated
snapshots. Installation commands are in [the README](../README.md).

Published v0.0.6 pins SDK v0.0.3 and retains organization directory and application status reads. The
[completed release checklist](release-v0.0.6-readiness.md) records owner approval,
exact asset verification, the protected tap merge and anonymous native results.

## Prepare a candidate

Use Go 1.27.1, a clean committed checkout, the pinned SDK and no local replacements.
Use ordinary public Go module proxy and checksum verification.
Choose a fresh output directory; packaging refuses to overwrite an existing one.

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off python3 scripts/package_release.py --version v0.0.6 --output build/releases/v0.0.6
python3 scripts/check_release.py --assets build/releases/v0.0.6
```

The packager builds one native executable to generate completions, then six
CGO_ENABLED=0 binaries. It creates no tag, release, push or visibility change.
The manifest records the source commit, dirty/local-only flags, toolchain and
pinned SDK metadata. `--local` permits workspace/dirty development candidates;
those candidates are not eligible for publication.

| Files | Purpose |
| --- | --- |
| Darwin/Linux tar.gz, Windows ZIP, each amd64/arm64 | Executable, LICENSE, THIRD_PARTY_NOTICES, README and three shell completions |
| install.sh, install.ps1 | Per-user installers with project MIT notice |
| checksums.txt | SHA-256 for six archives and two installers |
| bibleit-cli.rb | Generated Homebrew formula with four archive hashes and pkgshare notices |
| bibleit.json | Scoop manifest for Windows amd64/arm64; no public bucket is maintained yet |
| manifest.json | Frozen input provenance and eligibility markers |

Changes to README, notices, installers, completions or build inputs require fresh
packages and updated formula hashes even if runtime code is unchanged. Published packages retain the README frozen at their release source.
Changes to the main README apply to a future release; keep existing assets immutable. Re-audit component licenses for dependency/toolchain,
crypto/build flag or target changes. Never replace binaries under an existing tag
to conceal a notice/content change.

## Offline installation checks

macOS/Linux:

```sh
BIBLEIT_INSTALL_ASSETS="$PWD/build/releases/v0.0.6" \
  BIBLEIT_INSTALL_DIR="$PWD/build/install-test" bash install.sh v0.0.6
build/install-test/bibleit version
```

Windows:

```powershell
python scripts/check_release.py --assets build/releases/v0.0.6
python scripts/check_release.py --assets build/releases/v0.0.6 --powershell powershell
./scripts/check_powershell_installer.ps1 -AssetsDir ./build/releases/v0.0.6
python scripts/check_scoop.py --assets build/releases/v0.0.6 --scoop-source ./scoop-source
```

PowerShell 7 and 5.1 require native Windows installation/replacement checks.
Linux PowerShell parser/failure tests do not establish Windows behavior. Scoop
uses a pinned/existing source checkout, a disposable root and loopback downloads.
It leaves persistent PATH and the user's real Scoop installation untouched.

Homebrew:

```sh
python3 scripts/check_homebrew.py --assets build/releases/v0.0.6
```

The fixture uses a disposable prefix, local archive URLs and normal Homebrew
sandboxing. It verifies version, completions, notice retention, reinstall and
uninstall. The tap owns production-formula comparison and its native macOS/Linux
validation; follow the tap CONTRIBUTING guide for candidate versus release inputs.
Local authenticated fixtures do not prove anonymous GitHub release access.

## Install, update and uninstall

Script installers verify the archive checksum, version and required notices
before replacing the executable. They retain LICENSE and THIRD_PARTY_NOTICES in
`bibleit-licenses` under the binary installation directory. The defaults are
`~/.local/bin` and `%LOCALAPPDATA%\Bibleit\bin`. Neither installer edits profiles,
persistent PATH or credentials, requires elevation, disables OS protections, or
removes quarantine. Packages are not platform-signed/notarized.

Updating means rerunning the installer with an approved newer version. Uninstall
by deleting the executable and its bibleit-licenses directory. Credentials remain
in the configured credential file; revoke them using `bibleit auth logout` before
removal when appropriate. Environment tokens and registered SSH keys are separate.
Homebrew installs notices in pkgshare and handles removal through brew uninstall.
Scoop retains the complete ZIP contents and manages its own uninstall lifecycle.

## Approval gates

1. Freeze and validate the final source, SDK, notices, packaged README and artifacts.
   Preserve all required notices; review compatibility and the output contract.
2. Complete native installer/package-manager checks on the supported runner
   architectures. The v0.0.3 candidate passed all six native installer targets and
   all four tap Homebrew targets; anonymous retrieval requires publication.
3. Keep the already-withdrawn v0.0.1 downloads withdrawn and preserve their private
   backups, source and tag. No additional old-release cleanup is part of v0.0.3.
4. Obtain approval before release tags/assets, tap formula merges, or visibility
   changes. The tag workflow requires a source commit on main and creates a draft
   only after validation. Candidate validation does not publish.
5. At the separately approved public cutover, validate anonymous SDK/archive and
   Homebrew retrieval, then update availability statements across all repos.

The software license does not grant service/account access, quotas or translation
redistribution rights. See [the licensing audit](licensing-audit.md) for the
reviewed rights inventory and [SECURITY.md](../SECURITY.md) for private reporting.
