# Public launch approval package

Prepared 2026-10-12. This document specifies requested actions; it does not grant
approval. All repositories remain private. No PR has been merged, release/tag
created, or old asset/artifact removed. Server deployment is excluded.

## Reviewed changes and evidence

- [SDK PR 1](https://github.com/bibleit-labs/bibleit-go/pull/1): documentation,
  security contact, pinned Actions and Dependabot; review source 5bd8418.
- [CLI PR 1](https://github.com/bibleit-labs/bibleit-cli/pull/1): corrected packaging,
  version-free README, public CI preparation and dated audit/release evidence.
- [Tap PR 2](https://github.com/bibleit-labs/homebrew-tap/pull/2): exact corrected
  formula, notice validation, native lifecycle checks and public access preparation.
- [Server docs PR 1]([private service reference withheld]): implemented
  CLI documentation in Markdown and all three rendered locales; source edbc361.
  Keep the server private; merge/deploy its docs separately after public installs.

Frozen package candidate f0c080c passed CLI run 38177668163. Six archives and both
installers have eight approved hashes in
[final validation](security/v0.0.2-final-validation-2026-10-11.json).
Tap run 38178224204 passed exact formula/native Darwin arm64/Linux x86_64 lifecycle.
Windows AMD64 PowerShell 7/5.1 and Scoop passed; Windows arm64 runtime is a recorded
coverage gap, not a release check claimed complete. Linux/macOS arm64 installer
checks passed locally. Signing/notarization is outside this release scope.

[Gate review](security/launch-gate-review-2026-10-12.json) supplements the full
[security audit](security-audit.md) and [licensing audit](licensing-audit.md).
Owner MIT ownership/permissions and security@bibleit.app are confirmed. New commit
history, new completed logs/artifacts and current discussion/release metadata
have zero new scanner findings. All 39 new downloads were available; 24 unique
nested packages retained notices and binary byte/string checks passed. Raw reports
remain private. The fresh merge-head checks started after this inventory snapshot;
review their results/logs and recheck any later changes before visibility changes.

## Disclosure decisions needed

Approve retaining audited operational records and personal author/committer email
metadata in hosted history, or choose removal preparation before launch. The
operational records include docs/status.md, docs/plan.md and contract deployment
validation narratives in CLI/SDK: private server revision/image identifiers,
source fingerprints and recovery/coordination history. The audit found no
production credential leak. Current historical banners do not erase old history.

A removal choice requires a separate reviewed plan for affected Git history,
Go tags/checksums and hosted references/assets. No history rewrite is proposed
or authorized by this approval package.

## Requested release, cleanup and public-launch actions

1. Approve the three SDK/CLI/tap PRs for squash merge with Conventional Commit
   titles and no codex prefix. Require current head checks and no unresolved
   changes/findings. Record exact heads immediately before each merge.
2. Keep repos private while preparing the corrected CLI v0.0.2 draft. Build/test
   the final merged source through the guarded release workflow, then compare
   all eight artifact hashes, notices and formula with the frozen candidate.
   Manifest source commit must reflect the final source; do not copy old provenance.
   If package hashes change, stop publication and refresh formula/native evidence.
3. Preserve private copies, then remove only the 12 exact old release assets in
   [withdrawal proposal](security/v0.0.1-asset-withdrawal-proposal-2026-10-11.json),
   and only the 38 exact deficient CI artifacts in
   [cleanup proposal](security/deficient-artifact-cleanup-proposal-2026-10-11.json).
   All IDs/digests were rechecked present/unchanged on 2026-10-12. Recheck again at
   execution; preserve corrected candidates and unrelated logs/report artifacts.
   Keep the v0.0.1 source/tag; add an explanatory withdrawn-downloads release note
   pointing to the corrected release. Deletions are irreversible.
4. Publish the verified corrected release while CLI remains private, then make
   SDK public first and prove empty-cache normal Go proxy/sumdb retrieval without
   credentials. Validate SDK exported docs and public contributor execution.
5. Make CLI public, configure protections/security features and validate public
   dependency CI, all released assets, latest and explicit-version Bash/PowerShell
   installers on native runners. Stop at any unexpected checksum/access result.
6. Make tap public with the matching approved formula; configure protections and
   prove normal public brew install/test/reinstall/uninstall on macOS/Linux without
   Git credentials, cached assets, local archive URLs or custom prefixes.
   Tap publicity alone cannot make private CLI assets downloadable.
7. After advertised paths pass, prepare availability-copy PRs without inserting
   README version pins. Coordinate separate server-doc merge/deployment approval.
   Retire private read secrets only after public replacement paths pass; revocation
   of the underlying tokens needs confirmed owner control/other-use assessment.

## Repository protections and security features

SDK/CLI read defaults and disabled Actions PR creation/approval are applied and
verified. Tap read defaults/PR creation remain unchanged. Ruleset inspection still
returns the private-plan HTTP 403; no protections have been applied.

[Settings proposal](security/public-repository-settings-proposal-2026-10-11.json)
requires PRs, current GitHub Actions checks, resolved threads and linear history;
blocks main deletion/force pushes and release-tag update/deletion. Zero approval
count supports the sole maintainer; no bypass is proposed. Use squash/rebase
merges. Configure available dependency alerts, secret scanning/push protection and
private vulnerability reporting once eligible, then read back actual state.
If protection cannot be applied, stop further launch actions and report the gap.
Official schemas: [rulesets](https://docs.github.com/en/rest/repos/rules),
[private vulnerability reporting](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/configure-vulnerability-reporting/configure-for-a-repository).

Public anonymous installation cannot be proved while required repositories/assets
are private. Approval for the sequence above includes the visibility changes needed
for those checks. Failure pauses the sequence before advancing to another repository
or announcing availability; publicity already applied cannot retract downloaded copies.
