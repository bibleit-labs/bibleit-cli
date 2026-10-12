# Release maintenance

The repositories remain private. CLI v0.0.1 exists privately, but its binary
packages lack required notices. Corrected v0.0.2 and a matching tap formula are
staged for review. Windows installer/Scoop checks passed on the documentation
candidate; current evidence and remaining gates are in [release readiness](release-v0.0.2-readiness.md).
Dated operational evidence is in [historical status](status.md), rather than this
current procedure. Public installation commands are in [the README](../README.md).

## Prepare a candidate

Use Go 1.27.1, a clean committed checkout, the pinned SDK and no local replacements.
While the SDK is private, configure authorized Git access and GOPRIVATE first.
Choose a fresh output directory; packaging refuses to overwrite an existing one.

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off python3 scripts/package_release.py --version v0.0.2 --output build/releases/v0.0.2
python3 scripts/check_release.py --assets build/releases/v0.0.2
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
packages and updated formula hashes even if runtime code is unchanged. The
public-docs update changes the packaged README; regenerate the final candidate
before release approval. Re-audit component licenses for dependency/toolchain,
crypto/build flag or target changes. Never replace binaries under an existing tag
to conceal a notice/content change.

## Offline installation checks

macOS/Linux:

```sh
BIBLEIT_INSTALL_ASSETS="$PWD/build/releases/v0.0.2" \
  BIBLEIT_INSTALL_DIR="$PWD/build/install-test" bash install.sh v0.0.2
build/install-test/bibleit version
```

Windows:

```powershell
python scripts/check_release.py --assets build/releases/v0.0.2
python scripts/check_release.py --assets build/releases/v0.0.2 --powershell powershell
./scripts/check_powershell_installer.ps1 -AssetsDir ./build/releases/v0.0.2
python scripts/check_scoop.py --assets build/releases/v0.0.2 --scoop-source ./scoop-source
```

PowerShell 7 and 5.1 require native Windows installation/replacement checks.
Linux PowerShell parser/failure tests do not establish Windows behavior. Scoop
uses a pinned/existing source checkout, a disposable root and loopback downloads.
It leaves persistent PATH and the user's real Scoop installation untouched.

Homebrew:

```sh
python3 scripts/check_homebrew.py --assets build/releases/v0.0.2
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
   architectures. Windows checks resumed and passed on the documentation candidate;
   final artifact validation and anonymous retrieval remain separate gates.
3. Review the old v0.0.1 asset/artifact withdrawal proposal. Keep private evidence
   and preserve its tag/source history; perform cleanup only after approval.
4. Obtain approval before release tags/assets, the staged tap merge, or visibility
   changes. The tag workflow requires a source commit on main and creates a draft
   only after validation. Candidate validation does not publish.
5. At the separately approved public cutover, validate anonymous SDK/archive and
   Homebrew retrieval, then update availability statements across all repos.

The software license does not grant service/account access, quotas or translation
redistribution rights. See [the licensing audit](licensing-audit.md) for the
reviewed rights inventory and [SECURITY.md](../SECURITY.md) for private reporting.
