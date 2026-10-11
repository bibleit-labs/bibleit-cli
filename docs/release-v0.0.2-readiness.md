# Corrected CLI v0.0.2 release preparation

Prepared 2026-10-11. Repositories remain private. No release/tag was created;
v0.0.1 assets and both default branches remain unchanged.

## Resumed validation

On owner instruction, Windows validation resumed. Documentation candidate
b802ea3fa74ef1d1b34d4173b3ef1ee51724f9d6 at
[run 38176529757](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38176529757)
built all six targets and passed native Linux and Windows PowerShell 7/5.1,
parser/failure and Scoop lifecycle checks. It includes the updated README.
Native macOS Homebrew also passed; the full candidate run succeeded.
Final CI-hardened source validation is being completed.
The current README updates require a final fresh candidate; no release/tag,
old-asset changes or visibility change has occurred.

## Earlier candidate and tap evidence

Corrected candidate source: `9fb7bfc4565ffc3adacbcecf8343042e056ffc8c` on
`licensing-redistribution-audit`, using the pinned SDK v0.0.1 and Go 1.27.1.
[Candidate CI](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38174450362)
builds six archives once and passes those exact inputs to native installer jobs.
It produces a clean, standalone production candidate, not a local-only build.
The earlier run exposed Windows checkout line-ending conversion; .gitattributes
now preserves LF license bytes rather than weakening exact notice checks.

Tap review branch: `release/v0.0.2`. Its formula is generated from verified CI
archives; the updater rejects missing/inconsistent notices and missing pkgshare
installation. Candidate validation downloads the exact successful CLI run,
checks workflow/event/repository/source SHA/version and verifies hashes and
formula equality before installation. Published-release validation remains
available when no candidate run ID is supplied.

The build passed and all eight checksums match the first candidate; the line-ending
fix did not change the shipped bytes. Native code tests passed on all three hosted
platforms at [run 38174449692](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38174449692).
The candidate's three installation jobs did not start: GitHub reports failed
account payments or a spending limit. Overall candidate CI is therefore failing,
and the tap workflow deliberately cannot accept it as a successful candidate yet.
This remains a release gate, not a waived check.

Local checks passed on the exact fresh candidate:

- Darwin arm64: all six archive layouts/eight checksums, Bash installation and
  replacement, notice retention, latest/download fixture and rejection paths.
- Linux arm64: actual archive execution and the same installer checks in an offline,
  native-architecture Docker Linux container.
- PowerShell 7.4 in offline Linux: parser and failure paths, including a valid
  checksum with a missing notice. Native Windows behavior remains unverified.
- Tap-owned Homebrew on native Darwin arm64: exact generated formula verification,
  install/test/reinstall/completions, byte-for-byte notices after install/reinstall,
  and uninstall, entirely in a disposable prefix.
- Tap syntax/style CI passed at
  [run 38174382048](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38174382048);
  all 13 updater regression tests passed locally.

See [the validation summary](security/v0.0.2-validation-summary-2026-10-11.json)
for exact archive hashes, GitHub artifact digest and compact local evidence.
The earlier owner hold was lifted; the resumed Windows checks above supersede
its pending status.
Tap-owned Linux Homebrew validation also remains outstanding. The public-docs
update changes README, which is embedded in the archives: the recorded candidate
and proposed formula remain evidence for source 9fb7bfc only. Build a fresh clean
candidate from the final approved documentation/source, regenerate the tap hashes,
and validate those exact artifacts before release. Once Actions capacity is
available, use that new successful candidate run for tap validation. Do not treat
the old source's checks as approval for new package bytes.

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
