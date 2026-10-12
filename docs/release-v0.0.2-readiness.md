# Corrected CLI v0.0.2 release preparation

Prepared 2026-10-11. Repositories remain private. No release/tag was created;
v0.0.1 assets and both default branches remain unchanged.

## Final candidate validation

Follow-up 2026-10-12: source `cbe588d81aa300d59f5f5fc76f705e5e4476b73b`
passed [candidate run 38179740828](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38179740828),
including native Windows ARM64 PowerShell 7/5.1 and Scoop lifecycle/rejection checks.
All eight package hashes equal the frozen f0c080c inputs below, and the tap verifier
generates the exact reviewed formula. [Final ARM64 evidence](security/v0.0.2-final-validation-2026-10-12.json)
records source, jobs and artifact digests; the ARM64 runtime gap is now closed.

Frozen package source: `f0c080c1505346aee78dde9e0e24f80bb9d76a74`.
[CLI candidate run 38177668163](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38177668163)
succeeded: six clean archives, eight checksums, native Linux/macOS installers,
Windows AMD64 PowerShell 7 and 5.1, parser/rejection checks and Scoop lifecycle.
All archives contain exact LICENSE, THIRD_PARTY_NOTICES and the version-free README.
Additional exact-candidate checks passed locally on Darwin arm64 and offline,
unprivileged native Linux arm64, including the Bash latest-download fixture.
The original candidate validated Windows arm64 archives/hashes only; the follow-up
above adds native runtime/installer evidence on the same package bytes.

[Tap native run 38178224204](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38178224204)
succeeded on Darwin arm64 and Linux x86_64 at tap source
`ab1fc27f1ef7b3722ac6a5648df1949d30072a96`. Both jobs verified candidate provenance,
all package hashes and exact formula equality, then install/test/reinstall,
completions, byte-for-byte notices and uninstall in isolated Homebrew prefixes.
The updated read token's Actions/Contents access worked. Downloads occur after
Homebrew setup, which replaces the checkout; a permission preflight runs before it.
Tap style/syntax run 38178072711 and all 13 updater regressions passed.

Hardened native CLI code tests passed on Linux/macOS/Windows at run 38176879632;
SDK native tests passed on all three at run 38176883572. Subsequent source changes
are documentation only. Actions are pinned and private/public authentication is
conditional; public anonymous dependency retrieval awaits approved cutover.

[Final validation evidence](security/v0.0.2-final-validation-2026-10-11.json)
records exact source, all package hashes, artifact digests and native scope.
Evidence-only documentation commits after the frozen source do not change packaged
inputs. Any final merge/tag build must be validated and compared with these inputs.
Authenticated artifact access and local URL fixtures do not prove anonymous public
installation. No release, tag, merge, visibility or old-asset change was made.

## Earlier evidence

[Earlier validation summary](security/v0.0.2-validation-summary-2026-10-11.json)
records source `9fb7bfc4565ffc3adacbcecf8343042e056ffc8c` and the initial billing
block. That candidate's installer jobs did not start. Its package bytes differ
from the final README candidate and are retained solely as historical evidence.
The owner lifted the Windows hold; the successful final checks above supersede
all earlier pending Windows and tap Linux statuses.

## Recommended disposition of v0.0.1 — approval required

Recommended: release the corrected v0.0.2 packages through the private review
process, then withdraw all 12 downloadable v0.0.1 assets before any public launch.
Preserve the existing Git tag/source history and retain a release note explaining
that binary downloads were withdrawn because redistribution notices were missing;
point it to the corrected version. Verified copies of all 12 originals and their digests have been preserved
in the ignored private `.local-backups/private-release-evidence/v0.0.1` directory. Do not rebuild or overwrite binaries under v0.0.1 or move its tag: that
would invalidate existing checksums and hide the distinction between old and
corrected packages. Removing only the six archives leaves stale installers,
checksums and package-manager metadata, so withdraw the coherent set of 12.
The release API currently reports `immutable=false`.

The concrete proposed asset IDs/digests are in
[the withdrawal inventory](security/v0.0.1-asset-withdrawal-proposal-2026-10-11.json).
No withdrawal was performed. A separate notice file added beside the old archives
would not travel with already downloaded/extracted packages or fix the old
installers' notice retention, so it is a weaker option than a new corrected version.

Also remove or let expire the 38 historical CI artifacts identified by the audit
as containing 228 embedded CLI archives without THIRD_PARTY_NOTICES. Preserve
private audit evidence first. The precise proposed artifact IDs and archive hashes
are in [the artifact cleanup inventory](security/deficient-artifact-cleanup-proposal-2026-10-11.json).
This list reflects the audited inventory; recheck it at launch and exclude all
corrected candidate artifacts. Do not delete logs/runs indiscriminately: report-only
validation artifacts do not need binary licensing cleanup.

Signed-in users with repository read access can download workflow artifacts;
therefore a public repo would provide another path to old deficient binaries.
[GitHub artifact access](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/download-workflow-artifacts).
Artifact deletion is irreversible and is explicitly held for approval.
[GitHub artifact removal](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/remove-workflow-artifacts).

## Approval and final release sequence

1. Review the CLI/tap branches and successful evidence. Freeze the approved release
   source SHA; if merging changes that SHA/inputs, validate the final commit again.
2. Create the v0.0.2 tag/draft using the approved workflow; confirm its final assets
   match approved inputs and native validation. Release publication is a separate
   action from making repositories public.
3. Execute the approved v0.0.1 withdrawal/artifact cleanup and update its explanatory
   note; keep the tag/source history and private evidence intact.
4. Publish the corrected release and merge its verified tap formula in the agreed
   launch sequence. Recheck final privacy/anonymous module and asset access during
   the separately approved visibility change.

v0.0.2 runtime behavior is unchanged. New script installers reject older packages
missing notices, including v0.0.1, before replacing an existing executable.
macOS/Windows platform signing and notarization remain outside this change;
normal OS protections remain enabled. The six targets include arm64 packages,
but native CI runs only on architectures provided by the selected hosted runners.
