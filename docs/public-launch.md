# Public launch evidence

Completed 2026-10-12 after owner approval of actions 1–6, in SDK → CLI → tap order.
[bibleit-go](https://github.com/bibleit-labs/bibleit-go),
[bibleit-cli](https://github.com/bibleit-labs/bibleit-cli), and
[homebrew-tap](https://github.com/bibleit-labs/homebrew-tap) are public.
private-service remains private. Its docs PR remains draft, unmerged and undeployed.

## Release and cleanup

CLI [v0.0.2](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.0.2) was
built from merged source a5f1e605202f16446078e513e8ddce2e885efece. The guarded
[release run](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38219332962)
passed native code checks and all four installer jobs, including Windows ARM64.
All eight package hashes exactly match the approved frozen candidate; the manifest
records the final source. All 12 uploaded assets match the validated files, and
the normalized formula matches the tap. Release assets remain immutable.

The exact approved 12 v0.0.1 assets and 38 notice-deficient CI artifacts were
rechecked by ID/name/digest against verified private backups, then removed.
The v0.0.1 source/tag remain; its release explains the withdrawn downloads.
Corrected candidates and unrelated logs/artifacts were retained. The published
v0.0.2 package README remains its preparation-era snapshot; main documentation
changes belong to a future package release.

## Anonymous installation and contributor checks

| Path | Actual public validation |
| --- | --- |
| SDK | Empty module/build caches, normal Go proxy/sumdb settings, no credentials; expected SDK v0.0.1 sum, compiled README quickstart, standalone race tests/vet/build |
| CLI source and assets | Anonymous clone, normal pinned module download/verify, race tests/vet/build; all 12 asset digests and latest release verified |
| Bash installers | Latest and explicit release installation/replacement, version/help and exact notices; native Linux x86_64 and macOS ARM64, plus local macOS ARM64 |
| PowerShell installers | PowerShell 7 and 5.1, latest and explicit release installation/replacement, version/help and exact notices; native Windows AMD64 and ARM64 |
| Homebrew | Public tap/release URLs; install/test/reinstall/uninstall, exact notices and three completions; empty caches at /opt/homebrew and /home/linuxbrew/.linuxbrew; no tokens, Git helpers, local archive URLs or custom prefixes |

The [public installer matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38221644415)
and [normal public Homebrew lifecycle](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38221961063)
passed. Public native contributor CI also passed for
[SDK](https://github.com/bibleit-labs/bibleit-go/actions/runs/38220912904) and
[CLI](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38221156747).
The Windows public harness was corrected to compare canonical Git/reviewed bytes
across CRLF checkouts and to let each PowerShell engine initialize its own module
path. Downloaded installer/release bytes were unchanged. The corrected matrix
passed before tap publicity. Five successful launch validation logs were scanned
with the audit configuration; there were zero new findings. Raw logs/reports and
withdrawn packages remain in private local evidence storage.

## Protections and security

Each public repository has active main and v* tag rulesets without bypass actors.
Main requires PRs, current GitHub Actions native/style checks, resolved discussions
and linear history. Deletion/force pushes are blocked; tags cannot be updated or
deleted. Zero required approvals supports the sole maintainer. Required checks
bind to GitHub Actions application 15368. Protected merges have been verified.

Secret scanning, push protection, dependency alerts and private vulnerability
reporting are enabled and read back. SDK/CLI workflow defaults are read-only with
Actions PR creation/approval disabled; tap PR creation remains enabled for its
reviewed formula updater. No paid security features were purchased. Security
reports use security@bibleit.app. The owner confirmed MIT rights and approved
retention of audited operational records and commit email metadata.

[Executed settings](security/public-repository-settings-applied-2026-10-12.json)
and [release/install evidence](security/public-launch-2026-10-12.json) record the
actual state. Earlier readiness/audit documents are dated preparation snapshots.

## Coverage limits and separate follow-ups

Native Intel macOS and Linux ARM64 Homebrew execution are not claimed. All six
archive targets passed content/hash checks. Signing/notarization and a public
Scoop bucket are outside this release; native Windows Scoop candidate checks
passed independently. Service account/translation access remains separate from
repository access. Server docs merge/deployment requires separate approval.

Optional private repository read secrets remain stored; successful public paths
do not use them. Secret removal and underlying PAT revocation were not included
in approved actions 1–6. Assess owner control and other token uses before revocation.
