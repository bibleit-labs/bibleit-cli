# Security and Git history audit

Audit completed on 2026-10-11 for the three private client/distribution repositories.
**No production credential leak was confirmed in the scanned material.** All
scanner findings were classified; repository visibility remains private and
publication is still blocked on the disclosure decisions and launch checks below.
This report is an audit snapshot, not permission to publish.

## 1. Scope and evidence

Fresh, non-shallow audit mirrors fetched all advertised heads/tags and available
PR head/merge refs. `git fsck --full` passed in all three mirrors. GitHub metadata,
issues/PRs/reviews, release bodies/assets, available Actions logs/artifacts,
repository and inherited secret names, variables, keys, environments, deployments,
collaborators, forks, caches and package inventory were collected read-only.

| Repository | Audited main | Commits | Distinct blobs | Hosted PRs | Actions runs / downloaded logs | Artifacts |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| bibleit-go | 11a2858adff4545d0f57d665f4145671fd0963c8 | 34 | 137 | 0 | 13 / 13 | 0 |
| bibleit-cli | f121fa87ecb232f7441d93bde24189061117ca63 | 64 | 401 | 0 | 34 / 34 | 50 |
| homebrew-tap | b7bed330459f9c3797b4230363fe7f0cd2f71ec7 | 21 | 41 | 1 | 30 / 29 | 4 |

Each repository has one hosted branch, main. Go and CLI each have the v0.0.1 tag.
Tap PR 1's head, `61dcf0c`, is included even though its release branch was deleted.
None of the runs had earlier rerun attempts to collect. Tap run `38166277869`
has no jobs; its log download returns 404. Its metadata was reviewed, but there
are no available execution logs to scan for that run.

No LFS pointers, forks, caches, environments, deployments, deploy keys,
repository variables or inherited Actions secrets/variables were returned.
No linked GitHub attachments were found in the collected discussion metadata.
Visible organization package listings for container, npm, RubyGems, Maven,
Docker and NuGet returned no packages. No tracked server Erlang/BEAM files,
database dumps or `.bt`/`.bidx`/`.bsearch` assets were found in hosted history.
This does not replace the separate translation-rights review of text fixtures.

GitHub's final metadata check confirmed all three repositories private.
All raw evidence stays in a mode-0700 temporary directory, outside Git; files
are created with a restrictive umask. Only this report, a redacted summary and
the scanner configuration are intended for the audit commit. The collector
removes API-generated temporary clone tokens from repository metadata copies.

## 2. Scanner and coverage

Gitleaks **8.30.1**, downloaded from its official release and checksum-verified:

```text
darwin/arm64 archive SHA256
b40ab0ae55c505963e365f271a8d3846efbc170aa17f2607f13df610a9aeb6a5
```

The [audit configuration](security/gitleaks-audit.toml) extends all default rules
with Bibleit bearer tokens (`bt_` plus 43 Base62 characters), browser authorization
codes/flow IDs (`bcc_`/`bcf_` plus 43 URL-safe characters), credential URLs and
long sensitive query values. These formats were checked against server source;
a synthetic self-test detected the bearer/code/flow examples and produced fully
redacted results. No custom allowlist or baseline suppressed findings. Inline
`gitleaks:allow` directives and repository ignore fingerprints were disabled.

Separate scans covered full Git diffs including merges, every reachable blob,
commit and annotated-tag object, current source, hosted metadata and extracted
logs/artifacts/assets. Objects were exported separately so old snapshots and
commit/tag messages were covered beyond diff scanning.

All 136 downloadable outer archives were safely expanded with path/type checks.
CLI artifact archives contained 228 nested package archives; 102 were distinct
by hash. These were expanded too. The resulting 108 executable instances contain
72 distinct binary hashes. Gitleaks scanned printable strings from every distinct
binary because ordinary directory scanning skips binary files. Direct binary-byte
checks also found zero Bibleit bearer/browser codes, private-key headers, or
`BIBLEIT_RUNTIME_FIXTURE`/`CA`/`ADDRESS` controls.

| Scan | Go findings | CLI findings | Tap findings | Classification |
| --- | ---: | ---: | ---: | --- |
| Hosted Git history | 60 | 63 | 0 | Source fingerprints, test URLs/database credentials, disposable TLS fixture |
| Current source | 50 | 51 | 0 | Same categories |
| All exported objects | 203 | 215 | 0 | Same categories across historical snapshots |
| Sanitized hosted metadata | 0 | 0 | 0 | No findings |
| Extracted outer logs/artifacts/assets | 0 | 0 | 0 | No findings |
| CLI nested package contents | — | 0 | — | No findings |
| CLI binary strings | — | 0 | — | No findings |

Counts overlap between scans and historical snapshots; they are not distinct
credentials. There are **zero unresolved scanner findings**.

Reproduce the primary scans in isolated audit directories, using the pinned
scanner and this configuration; store reports outside the source repository:

```sh
gitleaks git "$AUDIT_MIRROR" --config "$AUDIT_CONFIG" \
  --log-opts='--all --full-history -m' --ignore-gitleaks-allow \
  --gitleaks-ignore-path "$EMPTY_IGNORE_PATH" --redact=100 \
  --report-format json --report-path "$PRIVATE_REPORT"

gitleaks dir "$AUDIT_EXTRACTED_FILES" --config "$AUDIT_CONFIG" \
  --ignore-gitleaks-allow --gitleaks-ignore-path "$EMPTY_IGNORE_PATH" \
  --redact=100 --report-format json --report-path "$PRIVATE_REPORT"
```

Use different report paths for each scope. Gitleaks exit 1 means findings require
classification, not a scan failure. The summary records configuration and release
asset hashes. Preserve raw evidence privately if needed beyond temporary-directory
retention; do not upload audit clones, scanner output or logs to a public repo.

## 3. Finding classifications

1. **Server fingerprints — false-positive secrets, disclosure still undecided.**
   Generic-key rules matched SHA-256 values in `files_sha256`/related source,
   test and native fingerprint maps in contract JSON. The original values and
   JSON context were checked; these are file hashes, not credentials. Go and
   CLI operational narratives additionally expose private-server revision/image
   identifiers, migration/recovery details and coordination history. Their
   publication remains an owner disclosure decision.
2. **Credential URLs — synthetic fixtures.** Go validation tests and historical
   CLI profile tests deliberately use `user:pass`, `user:secret` or
   `user:password` on example hosts and assert rejection. Integration scripts
   use `bibleit_test`/`bibleit-test-only` only for disposable test PostgreSQL
   containers/loopback. No production DSN was identified.
3. **Private key — disposable TLS fixture.** Findings are confined to
   `contract/fixtures/tls/server.key`, including historical versions. The paired
   certificates identify themselves as disposable. The fixture README explains
   its isolation; Python helpers use it for fake servers and the Go trust/dial
   override is in `_test.go`. The release packager copies only the executable,
   license, README and completions, verified against actual release archives.
   Shipped executables contain none of the fixture controls or private-key
   headers. No production use was found in this scope; this is not an audit of
   production secret stores. Keep exact fixture provenance rather than suppressing
   all private-key detections.
4. **Personal identity metadata — publication decision.** Commit/tag metadata
   contains personal author email addresses; tap history also contains a bot
   address outside GitHub's noreply domain. Those values were not copied into
   this report. Email-pattern hits in tap logs are Homebrew bottle filenames,
   not contact addresses. Path hits there are standard runner/Linuxbrew paths.
   Public commits disclose author/committer metadata, even after file cleanup.
5. **Local history — keep separate from hosted publication refs.** The older
   Python/config tree identified in the preliminary audit exists in local-only
   refs, not the freshly fetched hosted history. A supplemental Gitleaks scan
   of all local CLI refs produced 85 findings, all classified as fingerprints,
   synthetic test URLs/database credentials or the same fixture key. Local
   `native-validation` and Codex checkpoint refs must not be pushed as a mirror.

There is no evidence from this audit that a credential-driven history rewrite
or production credential rotation is necessary. If the owner chooses to remove
operational narratives or personal metadata, prepare an explicit removal plan
covering history, tags, PR/cached refs and relevant hosted assets; removing files
from main alone is insufficient. Obtain approval before destructive cleanup.

## 4. Released artifacts and source controls

Go v0.0.1 resolves to `78f30bba8bcf0026a85af826b393c4c65baab379`;
CLI v0.0.1 resolves to `40a826008facf76caccd1452140a5071a7c589e3`.
All 12 CLI assets were collected and inspected. All eight entries in
`checksums.txt` match their downloads. The six archives contain only the executable,
LICENSE, README and completion files. Manifests, installers, formula and release
bodies were scanned too. Existing archives retain their original README snapshots;
the newer README commits do not replace them.

Manual source review confirmed fixed HTTPS endpoints, blocked HTTP redirects,
context/timeouts and response bounds, loopback browser callbacks with random
state/PKCE, private config-file permissions/Windows DACL application before writes,
and OpenSSH public-key authentication with host-key verification. SSH uses
`accept-new` trust on first use; it is not a pre-pinned host key. Applications
supplying a custom HTTP client remain responsible for its transport trust.

Bearer credentials/authorization results and Live secrets are intentionally
available to their owners; callers must not log `AuthorizationResult.Fields`
or secret-bearing `Command.String()` values. Those caveats are documented.
This review is focused on publication exposure and secret handling, not a full
penetration test or a dependency-vulnerability assessment.

Existing `GOWORK=off go test -race -count=1 ./...` suites passed on darwin/arm64
for both Go and CLI. The CLI run used a separate module/build cache and authenticated
private dependency retrieval. No production login, mutation or public installation
was performed, and no new Windows/Linux execution coverage is claimed.

## 5. Repository security gaps

- GitHub reports secret scanning, code scanning and Dependabot alerts disabled.
  Main protection/ruleset endpoints return the private-plan upgrade/public-access
  restriction. Configure available controls once public access is approved, or
  separately choose a private-plan change; do not change visibility to bypass it.
- Default Actions permissions are read-only, but all three repositories allow
  Actions to approve PR reviews. Only the tap requires automated PR creation for
  its updater; review this combined creation/approval setting per repository.
- Most third-party Actions still use mutable version tags. Pin reviewed commit
  SHAs and establish updates before running public contributor workflows.
- CLI uses `BIBLEIT_CLIENT_READ_TOKEN`; tap uses `BIBLEIT_CLI_READ_TOKEN`.
  No values were exposed in logs/scanned files. Secret APIs expose names, not
  plaintext or actual token scope; least-privilege/expiry claims cannot be proven
  from names alone. Retain private-repo access until replacement public paths work,
  then revoke obsolete read tokens. Fork PR CI must eventually work without them.
- A maintained private vulnerability-reporting route and SECURITY.md are still
  needed as part of public-launch preparation.

## 6. Decisions and next work

The security/history audit is complete within the stated scope. Before a public
launch, decide whether to retain or remove the operational records and personal
commit metadata, complete licensing/rights review, prepare contributor CI and
repository controls, and validate the advertised installation paths after approval.

Making the **tap alone** public permits discovery/tapping but cannot enable
installation from its private CLI release URLs. Homebrew distribution needs both
the tap and CLI assets public; the Go library's visibility is not a runtime
requirement for the downloaded binary. For a Homebrew-first launch, coordinate
CLI and tap access in the same approved window. Publishing Go first remains the
simpler path for public CLI source builds and contributor CI.

Do not retag or replace immutable v0.0.1 content as incidental cleanup. Any required
rewrite or asset withdrawal needs an approved version/checksum/formula strategy.
Recheck exact refs and rescan changes before the visibility decision. GitHub cached
objects/deleted refs cannot be exhaustively enumerated with ordinary repository
APIs; where genuine sensitive data requires removal, follow GitHub's support-led
cleanup procedure. Server source/production credentials remain outside this audit.

Sources: [Gitleaks](https://github.com/gitleaks/gitleaks),
[GitHub sensitive-data removal guidance](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository).
