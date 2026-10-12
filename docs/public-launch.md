# Public launch evidence

Completed 2026-10-12 after owner approval of actions 1–6, in SDK → CLI → tap order.
[bibleit-go](https://github.com/bibleit-labs/bibleit-go),
[bibleit-cli](https://github.com/bibleit-labs/bibleit-cli), and
[homebrew-tap](https://github.com/bibleit-labs/homebrew-tap) are public.
private service remains private. Its CLI documentation was separately approved,
merged, deployed, and verified in English, German, and Portuguese.

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
| Homebrew | Public tap/release URLs; install/test/reinstall/uninstall, exact notices and three completions; native macOS ARM64/amd64 and Linux amd64/ARM64; empty caches at /opt/homebrew, /usr/local and /home/linuxbrew/.linuxbrew; no tokens, Git helpers, local archive URLs or custom prefixes |

The [public installer matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38221644415)
and [normal public Homebrew lifecycle](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38234257265)
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

## Server documentation publication

The separately approved CLI docs deployment is complete. Its initial validation
passed 20 live checks across English, German, and Portuguese. The owner then
requested that public docs expose clients and the open command protocol while
keeping server implementation and HTTP/WebSocket references private.

The revised [Developer page](https://bibleit.app/docs/developers) is deployed.
All 45 post-deployment checks passed. Server Markdown URLs and the HTTP/WebSocket
guide return 404, while the rendered [command protocol](https://bibleit.app/docs/protocol),
[CLI guide](https://bibleit.app/docs/cli), Shell, product pages and public assets
remain available. English/German/Portuguese docs contain no server Markdown links
or implementation guides. The Developer page shows the available Go SDK and
seven planned language clients with icons; the CLI has a separate Command line
section. Health/readiness return 200 and the server repository remains private.

Detailed deployment records, private workflow links, source/image fingerprints,
and raw logs remain in ignored private storage. The public
[publication summary](security/server-docs-publication-2026-10-12.json) records
only the public-site results and remaining follow-ups.

## Completion checklist

- [x] SDK, CLI and tap launched in the approved order.
- [x] Corrected immutable release published; approved deficient assets withdrawn.
- [x] Main/release protections and available security settings applied and verified.
- [x] Anonymous SDK, installers, native contributor CI and Homebrew checks passed.
- [x] Native Intel macOS and Linux ARM64 Homebrew coverage completed.
- [x] Windows AMD64 and ARM64 PowerShell 7/5.1 checks completed.
- [x] Server CLI docs merged, deployed and verified in all three languages.
- [x] Client-focused Developer page and restricted server docs deployed and verified.
- [x] Optional private Actions read-secret copies retired; owner confirmed no token reuse.
- [x] Underlying owner PAT identities/permissions verified and both tokens revoked.

The separate [published-package fixture checks](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38234262821)
also passed on all four native targets, with installed binary hashes matching
the anonymous installations. The tap formula and release assets are unchanged.

## Coverage limits and separate follow-ups

All four native macOS/Linux Homebrew targets passed anonymous lifecycle checks,
including Intel macOS and Linux ARM64. Host and installed binary architecture
checks reject mismatches and Rosetta execution. The
[native Homebrew evidence](security/native-homebrew-2026-10-12.json) records
runner identities, standard prefixes and installed binary hashes. Intel macOS
remains [Homebrew Tier 3](https://docs.brew.sh/Support-Tiers); these passing package
checks do not change Homebrew’s support policy. All six archive targets passed
content/hash checks. Signing/notarization and a public
Scoop bucket are outside this release; native Windows Scoop candidate checks
passed independently. Service account/translation access remains separate from
repository access. The separately authorized server CLI docs deployment is complete.

The separately authorized cleanup removed `BIBLEIT_CLIENT_READ_TOKEN` from the
CLI and `BIBLEIT_CLI_READ_TOKEN` from the tap after checking all hosted branches
and tags across the five Bibleit repositories, accessible secret metadata and
local named references. The owner confirmed both tokens were created only for
these copies and were not reused elsewhere. Current CLI workflows retrieve the
public SDK without authentication; tap release/candidate reads use its generated
workflow token. Private-only historical preparation workflows, including the
CLI v0.0.1 workflows, are no longer supported for reruns after retirement; existing
tags and release assets are unchanged.

GitHub secret lists verified both copies absent. After owner account verification,
the two underlying PAT identities and selected repository permissions were checked
in GitHub settings: SDK code/metadata read and CLI Actions/code/metadata read,
with no organization permissions. Both tokens were revoked; GitHub confirmed each
deletion and their absence from the account's fine-grained token list.
See the [cleanup assessment](security/private-read-secret-cleanup-2026-10-12.json).

The approved language-client roadmap is JavaScript, Lua, Python, Ruby, Rust,
Erlang/Elixir, and PHP. The Go SDK is available; the CLI is a separate command-line
tool. Future language-client implementation and releases are separate work.

## Next CLI release: prepared, awaiting approval

The v0.0.3 UX release is prepared from source
`0e1fc6f0b1a1cb734f5b54827f7e0319bb9f3280`. All six native installer targets and
all four native tap Homebrew targets passed, including Windows ARM64 PowerShell
7/5.1 and Scoop, Intel macOS and Linux ARM64. The [readiness checklist](release-v0.0.3-readiness.md)
records the frozen inputs, exact publication actions and the [validation evidence](security/v0.0.3-release-validation.json).
Release title will be `v0.0.3`.

No new tag or release has been created. [Tap PR #11](https://github.com/bibleit-labs/homebrew-tap/pull/11)
remains held; its live formula is v0.0.2. The existing release title and all twelve
asset IDs/digests are unchanged. Anonymous v0.0.3 downloads and public package
installation remain approval-dependent post-publication gates. The private service
repository remains private and is unchanged by this release preparation.
