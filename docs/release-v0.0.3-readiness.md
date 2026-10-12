# v0.0.3 release preparation

Prepared 2026-10-12. Release title: **v0.0.3**. Publication requires owner approval.
The current public release and tap remain v0.0.2. Existing releases are immutable.
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
is held for release approval; its evidence-only follow-up preserves the validated
formula and helper bytes. Intel macOS remains [Homebrew Tier 3](https://docs.brew.sh/Support-Tiers).

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

## Exact actions awaiting approval

1. Create and push immutable tag `v0.0.3` at source commit
   `0e1fc6f0b1a1cb734f5b54827f7e0319bb9f3280`. Let the protected release workflow
   rerun native validation and create its draft.
2. Download all twelve draft assets and compare every SHA-256 to the approved
   record, including the source manifest. Normalize its formula and compare it
   byte-for-byte with the held tap PR. Stop on any difference.
3. Set the draft title to exactly `v0.0.3`, replace its notes with the reviewed
   release notes, and publish it as the latest stable release. The workflow
   already uses the tag name as the draft title; verify it before publication.
4. Verify anonymous retrieval of the twelve assets and their recorded hashes,
   then squash-merge the verified head of tap PR #11 through branch protection
   with `chore(formula): update bibleit-cli to v0.0.3`.
5. Run the public installer workflow with `version=v0.0.3` on all six native
   targets and the public Homebrew workflow on all four macOS/Linux targets.
   Require empty caches, ordinary public URLs, no credentials, exact executable
   and notice bytes, auth help with a preserved legacy config, and completion files.
6. Record publication URLs, final verification results and the remaining limits.
   Keep older releases unchanged and the private service repository private.

No tag, draft or published release is created during preparation. The staged tap
formula is not merged. Anonymous v0.0.3 retrieval and a real `brew upgrade`
remain post-publication checks. Signing/notarization, a public Scoop bucket,
future language clients and server changes are separate work.
