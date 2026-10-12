# v0.0.3 release evidence

Published 2026-10-12 after owner approval of actions 1–6. Release title: **v0.0.3**.
[The latest stable release](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.0.3) and live tap formula are v0.0.3. Existing releases remain immutable.
The private service repository is unchanged and remains private.

## Frozen inputs

- CLI source: `0e1fc6f0b1a1cb734f5b54827f7e0319bb9f3280`, a protected merge on main.
- Go: `1.27.1`; SDK: `github.com/bibleit-labs/bibleit-go v0.0.1`, checksummed, no workspace/replacement.
- Twelve assets: six native archives, two installers, checksums, formula, Scoop manifest and provenance manifest.
- Each archive includes the exact current README, LICENSE, THIRD_PARTY_NOTICES and Bash/Zsh/Fish completions.
- ELF, Mach-O and PE executable headers match all six advertised target architectures.
- No literal private repository name occurs in the shipped executables.
- The checkout-only README LF fix changes provenance but preserves every archive/installer hash.

The [reviewed release notes](release-v0.0.3-notes.md) describe auth help, completion
setup, legacy configuration recovery and direct SSH host trust. The README keeps
installation guidance independent of release versions. The SSH change does not
reuse trust recorded only under an older hostname; changed known keys remain
rejected. Neither native validation nor package installation modifies real user
credentials or shell startup files.

## Native validation

All gates passed on the frozen inputs. The [validation record](security/v0.0.3-release-validation.json)
contains all twelve asset digests, six executable hashes/headers, native reports,
source commits, artifact identity/digest and the exact tap formula/helper hashes.

| Gate | Native targets | Result |
| --- | --- | --- |
| Bash installer, local help and legacy config preservation | Linux AMD64/ARM64; macOS Intel/ARM64 | Passed |
| PowerShell 7 and 5.1 install/replacement and failure paths | Windows AMD64/ARM64 | Passed |
| Scoop install/shim/reset/reinstall/uninstall and wrong hash rejection | Windows AMD64/ARM64; correct architecture archives downloaded | Passed |
| Homebrew install/test/reinstall/uninstall, notices, completions and auth help | macOS Intel/ARM64; Linux AMD64/ARM64 | Passed |

[CLI candidate 38244704911](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38244704911)
and [tap validation 38245499054](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38245499054)
both completed successfully. Tap validation used commit
`2f472a026ffe875f16b128d8d47f641056adbab7`. All installed executable hashes match
the reviewed archives; host/header checks verify the native Homebrew targets and
reject Rosetta. The tap's production formula exactly matches the normalized
CLI-generated formula. [Tap PR #11](https://github.com/bibleit-labs/homebrew-tap/pull/11)
was squash-merged through protection at `65bff4e45fa8075619ec7afc47b1f4f70c1eeb31`;
its evidence-only follow-up preserved the validated formula and helper bytes. Intel macOS remains [Homebrew Tier 3](https://docs.brew.sh/Support-Tiers).

Local standalone packaging, race tests/vet, auth help with an unchanged legacy
config, workflow lint, and all 16 tap verifier tests passed. Documentation and
workflow review passed protected native CI. Completion initialization requires
the user's shell setup described in the packaged README; files are installed,
but installers do not edit shell startup files.

The first run stopped on Windows because its worktree README used CRLF while the
packages contained canonical LF. The corrected source pins README checkout
endings. A later Intel macOS uninstall timed out fetching full Homebrew core
history after installation/test/reinstall succeeded. The fixture now seeds core
metadata from the runner cache or a shallow public checkout before the lifecycle.
Superseded runs are retained as diagnostic history, not release evidence.

## Approved publication actions — completed

1. Created immutable annotated tag `v0.0.3` at the approved frozen source.
   [The tag workflow](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38247813670) passed native race/process/shell checks and all six
   native installer targets, including Windows ARM64 PowerShell 7/5.1 and Scoop.
2. Downloaded all twelve draft assets and compared every SHA-256 and GitHub digest
   with the approved record. The provenance manifest and normalized production
   formula matched exactly.
3. Published [v0.0.3](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.0.3) as latest stable, with title exactly `v0.0.3`
   and the [reviewed release notes](release-v0.0.3-notes.md).
4. Downloaded all twelve published assets anonymously and verified their hashes,
   then merged [tap PR #11](https://github.com/bibleit-labs/homebrew-tap/pull/11)
   at the approved head through required checks with a conventional commit.
5. [Anonymous installers](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38249509803) passed on all six native targets,
   using latest and explicit release URLs. Windows ran PowerShell 7 and 5.1.
   [Anonymous Homebrew](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38248930792) passed install/test/reinstall/uninstall
   on macOS ARM64/Intel and Linux AMD64/ARM64 at standard prefixes with empty
   download caches. Reports verify native architectures, exact executable and
   notice bytes, updated help, preserved legacy configuration and Homebrew completions.
   The real local Homebrew upgrade from v0.0.2 to v0.0.3 also passed; user credential
   configuration and shell startup files were unchanged.
6. Recorded publication metadata, anonymous reports and the actual upgrade in
   [the validation record](security/v0.0.3-release-validation.json). The previous
   v0.0.2 release title/notes and all twelve asset IDs/sizes/digests remain unchanged.
   The private service repository remains private.

The first public installer run passed macOS/Linux but Windows stopped before
installation: the default PowerShell shell did not expand Bash-style `$VERSION`.
[PR #16](https://github.com/bibleit-labs/bibleit-cli/pull/16) selected Bash for the
workflow invocation. The final six-target run above passed from protected main;
actual Windows installation still uses PowerShell 7 and 5.1. This harness-only
change did not modify the published tag or assets. Two jobs in the final run
initially encountered GitHub HTTP 504 downloads; those failed jobs were rerun
without changing source or assets. The record includes the final run attempt.

## Remaining limits

No publication or native installation gate remains pending. Intel macOS remains
[Homebrew Tier 3](https://docs.brew.sh/Support-Tiers). Signing/notarization,
a public Scoop bucket, future language clients and private service changes are
separate work. Shell initialization still requires the setup in the packaged README.
