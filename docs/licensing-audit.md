# Licensing and redistribution audit

Reviewed 2026-10-11. All three repositories remain private. This review does
not authorize publication, change ownership, rewrite history, or replace
published tags/assets. The owner confirmed in this chat that they have the
rights to release code, documentation and fixtures in all three repos under MIT,
including copied material and any employer/client work.

## Repository findings

| Repository | Audited hosted main | Rights and dependencies | Result |
| --- | --- | --- | --- |
| bibleit-go | `11a2858adff4545d0f57d665f4145671fd0963c8` | MIT, copyright 2022 Felipe Mamud; no external Go module dependencies; source distribution, no binary release assets | No identified licensing blocker for the audited source |
| bibleit-cli | `f121fa87ecb232f7441d93bde24189061117ca63` | Same MIT notice; pins bibleit-go v0.0.1; binaries also incorporate Go runtime, standard library and vendored modules | Future packages corrected; existing private v0.0.1 assets remain a publication blocker |
| homebrew-tap | `b7bed330459f9c3797b4230363fe7f0cd2f71ec7` | Same MIT notice; Ruby formula, Python/Bash automation; downloads CLI archives, does not vendor Homebrew itself | Formula source has no identified licensing blocker; its v0.0.1 binary target has the CLI notice gap |

Preserve the existing copyright attribution. Neither the organization name nor
a newer year is evidence of a transfer of ownership. The owner's confirmation
and current licenses establish the intended MIT grant; author names alone do
not establish ownership. Hosted authors are Felipe Mamud and, in the tap,
github-actions[bot]. No outside human contributions were identified.

The MIT grant permits redistribution subject to retaining its copyright and
permission notice. [Official MIT license](https://opensource.org/license/mit).
The grant covers these projects' software/docs/owned fixtures, not every Bible
translation obtainable through the API, service access, or third-party trademarks.

## Scope and provenance

Reused the complete fresh hosted mirrors and downloaded release evidence from
[the security audit](security-audit.md): all 34/64/21 reachable hosted commits,
137/401/41 distinct blobs, and the tap PR refs, for Go/CLI/tap respectively.
Reviewed historical file paths and license/copy/adaptation headers, dependency
manifests, current sources/tests/docs/fixtures, release layouts, and build scripts.
Historical source paths total 35/65/15 respectively. Their file types are source,
configuration, documentation and the CLI's disposable TLS PEM/key fixture;
no bundled images, fonts, artwork or Bible corpus were identified.

The local CLI native-validation/checkpoint refs contain older Python files and
translation configuration, but the fresh hosted CLI mirror does not. Those local
refs remain excluded from publication; publishing them requires a separate
rights review. Python scripts in hosted history use the standard library and
are project automation; Python itself is not shipped in the six CLI archives.
Go, Homebrew, Scoop, CI actions, containers and the private server are execution
or development tools/dependencies, not automatically redistributed as part of
the source repositories or CLI archives. Their downloaded packages retain their
own licenses; a future bundled tool/container distribution needs another review.

No copied source requiring an incompatible license was identified in project
files. This conclusion combines inspected evidence with the owner's ownership
confirmation; searching copyright strings alone cannot establish provenance.

## Translation content and names

Contract fixtures use synthetic text such as `Fixture verse`, `John 3:16`,
and an invented string exercising Portuguese characters, quotes and backslashes.
The integration harness creates its own short shepherd-themed fixture sentences;
it does not copy Psalms from a publisher. Go integration assertions search for
`shepherd` in service/test responses. Catalog fixtures contain translation names,
book names/counts and synthetic records; no actual translation corpus is shipped.
Generated shell completions come from project code. TLS material is generated
for the isolated fixture, as established in the security audit.

`rights_status=public_domain` and `Publisher credit` in protocol tests are test
fields, not clearance evidence for a real edition. Keep future data reviews
separate from software licensing. The private server's
`docs/translation-rights-policy.md` requires exact edition/source evidence,
attribution/trademark review and worldwide clearance; the client MIT license
cannot substitute for that evidence.

The World English Bible publisher describes its text as public domain and its
name as a trademark. References in tests/examples identify the translation;
they do not grant a right to brand modified text as that edition or imply
publisher endorsement. [Publisher rights statement](https://ebible.org/eng-web/copr.htm).
No blanket public-domain claim is made for KJV or other names in fixtures.
No additional Bible-text permission was needed for the synthetic material found.

## Binary component inventory

Reviewed `go list -deps -json` for all six CGO_ENABLED=0 Go 1.27.1 release targets,
using the pinned standalone SDK: Darwin amd64/arm64, Linux amd64/arm64,
Windows amd64/arm64. Package counts are 201/200/201/200/204/203; these are compiler
input packages, not a claim that every symbol survives linker elimination.
Inspected selected Go and assembly files for additional embedded notices.

| Component | Terms/notices retained |
| --- | --- |
| bibleit-go v0.0.1 | Full SDK MIT notice, from the pinned module |
| Go runtime and standard library | Full BSD-3-Clause notice, from installed Go 1.27.1 |
| Go-vendored golang.org/x/crypto, x/net, x/sys, x/text | Each full BSD-3-Clause license from the exact installed toolchain |
| fiat-crypto-derived scalar and P-256 arithmetic | Both upstream notice/disclaimer blocks embedded in Go source |
| Sun Microsystems-derived math routines | 1993 and 2004 copyright/permission notices, including SunPro attribution |
| Cephes-derived math routines | Moshier copyright and accompanying free-use/no-guarantee notice carried by Go |
| Rijndael/AES reference implementation | Original author, public-domain declaration and disclaimer block carried by Go |

These permissive notices do not change the project's primary MIT license.
The Go BSD license requires its notice, conditions and disclaimer in materials
provided with binary distributions. [Official Go license](https://github.com/golang/go/blob/master/LICENSE).
BoringCrypto/OpenSSL source is present in the toolchain but excluded from the
reviewed default CGO_ENABLED=0 package graphs. Compiler/tool dependencies under
`src/cmd/vendor`, Go test data and test-only fixture keys are not shipped.
Custom crypto experiments, different toolchains/targets or new modules need a
fresh inventory. The notice bundle conservatively retains notices for selected
packages even where the linker may eliminate unused routines.

Homebrew/Scoop `license MIT` continues to describe the upstream project's primary
license. The CLI's third-party code retains its own terms and is documented in
THIRD_PARTY_NOTICES. Formula license metadata alone is not a substitute for
redistribution notices. [Homebrew license metadata](https://docs.brew.sh/Formula-Cookbook#fill-in-the-license).

## Corrections prepared privately

- Added root THIRD_PARTY_NOTICES with the exact component license/notice texts.
- Package builder includes LICENSE and THIRD_PARTY_NOTICES in all six archives.
- Downloadable Bash/PowerShell scripts carry the full project MIT notice.
- Both installers validate required notices before replacing a binary and retain
  them in `bibleit-licenses` beneath the installation directory. New installers
  reject archives without notices, including old v0.0.1 packages.
- Generated Homebrew formulas install both notices in pkgshare. Scoop extracts
  the complete ZIP, and its validator now requires THIRD_PARTY_NOTICES too.
- Archive/native installation checks verify exact notice contents and rejection
  of a correctly checksummed archive with a missing notice.
- README distinguishes software licensing from translation rights.

The currently hosted tap formula remains pinned to v0.0.1. It cannot install a
notice absent from those archives. Update it using the generated formula from
a new, verified release before public installation. No current formula hashes,
release assets or tags were changed.

## Remaining release decision

The six actual private v0.0.1 release archives include only project LICENSE,
README, completions and executable. Their MIT notice covers both project and
SDK attribution, but the additional Go/embedded notices are missing. The two
separately downloadable v0.0.1 installer scripts also lack an embedded MIT
copyright/permission notice. Adding a
file to current Git does not fix already published archives or their embedded
installer scripts/formula. Before making CLI assets public, prepare a new
version with these fixes, validate it on native platforms, update the tap,
and explicitly decide how to withdraw or remediate the old deficient assets.
Making the whole CLI repo public would expose old release assets too; merely
pointing the tap at a newer version is insufficient. Asset replacement/removal,
retagging and publication require the owner's separate approval.

Source licensing review is complete for the audited hosted revisions; binary
redistribution remediation remains pending a new release and old-asset decision.
Recheck changed code/dependencies and final release evidence before approval.

## Validation

Local-only candidate `v0.0.2-licensing.1` built successfully for all six targets,
using the pinned standalone SDK v0.0.1 and Go 1.27.1. It is intentionally marked
local-only/dirty and is not suitable for publishing as-is. No release or tag was
created. Checks passed:

- Six archive layouts, exact license/notice bytes and all eight checksums.
- Native Darwin arm64 installation, replacement, notice retention and the Bash
  latest/download path using offline fixtures.
- Missing-notice rejection with a valid checksum; invalid-version,
  missing/duplicate-checksum and corruption rejection preserve the old binary.
- PowerShell 7.4 syntax and failure paths in a cached Linux container with no
  network, including valid-checksum/missing-notice rejection. This does not
  establish native Windows installation/replacement or PowerShell 5.1 behavior.
- Generated Homebrew Ruby syntax and its explicit pkgshare notice installation;
  exact SDK/Go/four vendored-module license texts are in the notice bundle.
- Standalone installers carry the complete MIT notice; Python compilation,
  Bash syntax and Git whitespace checks pass.

Native Windows/Scoop and native Homebrew notice-retention checks remain final
release gates. The unchanged SDK/tap source required no licensing code changes.
Private visibility was rechecked after this work; all three remain private.

