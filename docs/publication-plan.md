# Public repository launch plan

> Launch update, 2026-10-12: the owner approved actions 1–6 and protection at
> public cutover. The SDK, CLI and tap are now public with protected main/release
> tags. See [actual launch evidence](public-launch.md) for executed actions and
> verified installation coverage. The preparation text below is a historical
> snapshot. private service remains private; its docs PR is not merged/deployed.


Prepared on 2026-10-11 for `bibleit-labs/bibleit-go`, `bibleit-labs/bibleit-cli`,
and `bibleit-labs/homebrew-tap`. This is a preparation and approval plan;
repository visibility has not changed. `private service` remains private.

Security/history audit completed on 2026-10-11 within the scope recorded in
[security-audit.md](security-audit.md). No production credential leak was
confirmed; all scanner findings are classified. Publication still requires
the owner's disclosure decisions, remaining launch work and explicit approval.

## Current evidence and limits

| Repository | Observed state | Work needed before launch |
| --- | --- | --- |
| bibleit-go | Private; main; MIT; published private v0.0.1; no release assets | History/rights review, public module retrieval, contributor and security docs |
| bibleit-cli | Private; main; MIT; published private v0.0.1 with 12 assets; pins bibleit-go v0.0.1 | History and asset review, remove private dependency authentication, document released behavior |
| homebrew-tap | Private; main; MIT; formula points to CLI v0.0.1 | History review, anonymous package retrieval, simplify release-read authentication |

GitHub reports all three main branches as unprotected. Check organization
rulesets as well before deciding the final protection settings. At inspection,
GitHub listed 12/32/29 Actions runs and 0/47/4 artifacts for Go/CLI/tap respectively.
Those counts are exposure inventory, not evidence that their contents are safe.
Wikis and Pages are disabled in all three repositories.

Initial checkout revisions: Go `78f30bb`, CLI `e207e84`, tap `ee70274`.
The README work finished in chat `[internal identifier withheld]`:
Go `11a2858`, CLI `f121fa8`, tap `b7bed33`, pushed to private main branches.
The final READMEs were reviewed for private/public availability, current module
pin, fixed endpoints, transport/auth behavior and installation instructions;
local links resolve. They correctly mark anonymous installations as pending.
The Go quickstart compilation and released CLI version/help checks were reported
by that chat. CONTRIBUTING and archive README snapshots are covered by the separate
launch review below. Server inspection used `4c1f60e` plus an existing dirty working tree;
prepare its docs changes independently without including unrelated work.

A preliminary regex scan read all locally reachable Git blobs, commits and tag
objects: Go 33 commits/136 blobs, CLI 70 commits/520 blobs, tap 20 commits/40 blobs.
It checked common GitHub/AWS/Stripe/Slack token patterns, private-key headers,
credential URLs, email addresses, local paths and Bible asset references.
No common provider-token pattern matched. The sole private-key path matched was
CLI `contract/fixtures/tls/server.key`, documented as a disposable fixture.
Credential-URL matches include test inputs and integration database setup.
Author email addresses and older Python/client content exist in history.
CLI history includes `config/translations_books.json` and `python/LICENSE`.

This is not a complete secret audit: no dedicated scanner was installed, remote
PR refs and cached/deleted objects were not fetched, and hosted logs, artifact
contents and published binaries were not scanned. Findings below are review
tasks, not confirmed production credential leaks.

## 1. Freeze scope and inventory exposure

- [x] Finish the README chat and review its final READMEs and local links.
- [ ] Recheck README install commands and service examples against the final
  approved release and anonymous distribution paths during cutover.
- [x] Record audited main/tag SHAs, release IDs, asset hashes and the tap formula
  in the private evidence and redacted audit summary.
- [ ] Freeze final launch revisions and rescan intervening changes.
  Use a short coordinated freeze before the visibility decision.
- [x] Inventory accessible remote branches, tags, open/closed PRs and PR refs, issues and
  comments, release bodies/assets, Actions logs/artifacts/caches, LFS objects,
  attachments, packages, deployments/environments, collaborators and deploy keys.
  Check organization-inherited secrets and variables without exporting values.
  API restrictions and unavailable/cached evidence are recorded in the audit.
- [x] Distinguish local Codex checkpoint refs and the CLI `native-validation`
  branch from refs actually hosted on GitHub. Never push local snapshots or a
  mirror of the working checkout as part of this launch.

Exit: a complete exposure inventory and fixed revisions for the review.
GitHub states that Actions history and logs become public when visibility changes;
include them in the audit, not just Git files.
[GitHub visibility documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/setting-repository-visibility).

## 2. Audit secrets, private data and source disclosure

- [x] In isolated audit clones, run a maintained secret scanner such as Gitleaks
  over complete fetched history and current files, including branches/tags/PR
  refs. Scan messages/tag annotations separately where the tool does not cover
  them. Use provider rules plus Bibleit-specific token/credential patterns.
  Record scanner version, configuration, scan scope and redacted findings.
- [x] Manually inspect suspected credentials: bearer/API tokens, OAuth/browser
  codes, cookies, Live access secrets, SSH/TLS keys, DSNs, cloud/provider secrets,
  webhook URLs, test account data, screenshots and author/contact information.
  Scan binaries/archives and LFS content separately from text history.
- [x] Review fixture origins and isolation within the audited sources and assets.
  `server.key` must remain limited to the isolated
  test server and must never have authenticated production. Its README explains
  the fixture, and Python test helpers reference it. Verify packaging excludes
  it and production executables do not trust its certificate. Allowlist only
  that exact reviewed fixture; never suppress all private-key findings.
- [x] Review Go and CLI `contract/`, CLI `docs/status.md`, `docs/plan.md`,
  `docs/native-validation.md`, `docs/releases.md` and historical versions.
  They contain private-server revision hashes, Fly registry image identifiers,
  operational qualification/recovery details and coordination history. These
  are not automatically credentials, but need a deliberate disclosure decision.
  Keep public API contracts and useful compatibility summaries; move internal
  operations records to approved private storage before removing public copies.
- [x] Obtain the owner's disclosure decision for operational records and personal
  commit-email metadata. On 2026-10-12 the owner approved retaining the audited
  records and metadata in public history; no rewrite is needed.
- [x] Review older CLI Python/client code, fixtures and deleted files, including
  translation metadata and attribution. Deleting them from main does not remove
  their history. Do not publish proprietary server code or customer recordings.
  The old Python/config tree is local-only and is absent from hosted launch refs.
- [x] Review all available hosted Actions outputs, release bodies and downloadable assets.
  CLI archives include LICENSE and README from their release source; fixing the
  current README does not update those existing archive copies. Manifests and
  build metadata also need review for paths and internal identifiers.

Exit: every finding classified as synthetic/public, intentionally disclosed,
or remediated, with no unresolved secret or private-data exposure.

If a real secret is found, revoke/rotate it first. If removal is necessary,
prepare and review a coordinated history rewrite in an isolated clone, including
tags and GitHub PR/cached references; request GitHub support where needed. Obtain
explicit approval before destructive cleanup. Preserve a private recovery copy
under controlled access and prevent old clones from reintroducing the data.
Rewrites change SHAs and can affect Go version sums, releases and tap hashes.
Do not silently move v0.0.1 or replace its published bytes; decide an explicit
withdrawal/new-version strategy before proceeding.
[GitHub sensitive-data removal guidance](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository).

## 3. Confirm licensing and redistribution rights

Reviewed on 2026-10-11 in [licensing-audit.md](licensing-audit.md). The owner
confirmed the MIT grant and ownership/permissions. Source review found no
licensing blocker. CLI v0.0.1 binary assets lack Go redistribution notices;
future packaging and a verified v0.0.2 tap review branch are prepared privately.
[Release preparation](release-v0.0.2-readiness.md) records passed local/native
checks, including completed Windows/Scoop and tap native macOS/Linux validation.
Old assets remain a launch blocker until their disposition and a new release
are approved.

- [x] Confirm the owner intends MIT for code, documentation and formula in each
  repository and has rights to license all contributions. All three currently
  contain the same MIT notice, copyright 2022 Felipe Mamud; verify attribution
  and dates rather than automatically changing ownership to the organization.
- [x] Inventory dependencies, copied/adapted code, historical Python dependencies,
  fixtures, images and generated material. Identify any required third-party
  notices, including material distributed in compiled Go binaries.
- [x] Review Bible text and translation metadata separately. Software MIT does
  not grant rights to a Bible translation or third-party artwork. Record source,
  edition, attribution, redistribution terms and territorial limitations for
  any bundled excerpts/data; prefer clearly licensed synthetic test content.
  Consult the server's `docs/translation-rights-policy.md` for project policy.
- [ ] Confirm every distributed archive contains the correct license and any
  required notices. Future packages now include THIRD_PARTY_NOTICES; validate
  the final new release and resolve the deficient private v0.0.1 assets before
  public exposure. Generated Homebrew formulas retain the notices. The primary
  project/formula MIT license remains appropriate; dependencies retain their
  own terms.

Exit: a rights inventory with no unknown or incompatible material. Keep the
repositories private if required permissions cannot be established.

## 4. Prepare public documentation, CI and repository settings

- [x] Review the README chat's results and update availability, commands,
  standalone development and redistribution guidance across all three repos.
  Final public commands and anonymous retrieval remain cutover checks.
- [x] Correct all three CONTRIBUTING guides; ordinary unit tests work without
  private-server access and the CLI pins SDK v0.0.1 without a required workspace.
- [x] Replace the CLI release/native-validation/endpoint guides with current
  procedures and mark retained rollout ledgers as historical evidence.
- [x] Add SECURITY.md, bug-report and PR templates in all three repositories.
  The owner approved security@bibleit.app for private reports. Community conduct
  and CODEOWNERS policies can be added if contributor governance needs them.
- [x] Prepare CLI native/candidate/release workflows for public dependencies:
  retain conditional authenticated private downloads until cutover, use ordinary
  public module downloads afterwards, and make reusable read secrets optional.
- [x] Prepare tap reads to stop using the private cross-repository token in public
  mode while retaining scoped formula-PR writes. Revoke obsolete tokens only
  after replacement paths pass at launch.
- [x] Pin external Actions to upstream full SHAs and prepare weekly Dependabot
  update review in all three repos. Preserve read-only PR execution, manual
  release/formula triggers and the main ancestry gate. Actionlint passes.
  Native workflow execution and anonymous public-fork retrieval are separate checks.
- [ ] Configure main protection/rulesets, passing checks, force-push/deletion
  restrictions and release-tag protections. Choose review requirements that
  work for the actual maintainer count and formula bot, without bypassing checks.
  Enable available secret scanning/push protection, dependency alerts and private
  vulnerability reporting; verify settings after visibility changes.

Exit: reviewed public-ready changes and settings, normal CI independent of
private server access, and explicit release/security ownership.
[GitHub Actions secure-use reference](https://docs.github.com/en/actions/reference/security/secure-use).

## 5. Prepare private service documentation updates

Prepared locally on 2026-10-11: current CLI reference, existing Markdown mentions,
all three locale catalogs, the rendered CLI/developer links and the public-doc
allowlist. Availability stays private. Deployment and full server integration
checks remain separate; unrelated server implementation work is excluded.

Documentation validation: 77 local Markdown link destinations resolved; offline
v0.0.2 help/version and auth/translation/live/completion usage checked; eight
Erlang docs-rendering tests and nine docs JavaScript tests passed. Erlang checks
used the existing compiled renderer with current templates/catalogs, not a new
server build. No live browser, database-backed server suite or deployment was run.
The final version-free packaged README and matching formula passed fresh candidate
and tap native validation; exact evidence is in the release readiness record.


The rendered docs are owned by `src/http/bibleit_http_docs.erl`,
`priv/templates/docs-page.html` and `priv/i18n/docs-{en,pt,de}.json`.
Editing only files under `docs/` will leave the live CLI page stale.

| Location in private service | Prepared change |
| --- | --- |
| priv/i18n/docs-en.json, docs-pt.json, docs-de.json | Replace docs_cli_preview/intro/body coming-soon text in every locale; describe implemented capabilities and private availability |
| priv/templates/docs-page.html | Remove the CLI planned badge; show auth, translation and Live examples; link the detailed CLI reference and Go/CLI/tap repositories |
| docs/reference/http.md | Replace forthcoming-CLI wording with implemented PKCE login behavior and link to CLI documentation |
| docs/reference/protocol.md | Update implemented CLI transport/storage description; keep raw local-server SSH examples distinct from the fixed production CLI |
| docs/developers/architecture.md | Update browser/CLI flow status and links; preserve the existing browser/session flow |
| docs/developers/development.md | Clarify that raw OpenSSH can target the development server but the distributed CLI targets bibleit.app; do not promise an endpoint flag |
| docs/index.html | Remove invented personal output and add the private-availability CLI reference |
| docs/maintenance.md | Update the local CLI status and release-check procedure; keep distribution validation as the publication gate |
| priv/config/public-docs.txt | Add only reference/cli.md; preserve existing routes and the operations-record exclusion |

Keep the website quickstart concise and link detailed CLI output/exit-code
contracts to the public CLI docs. Explain HTTP as the default, explicit
`--transport ssh`, OpenSSH prerequisites/registered public keys, `BIBLEIT_TOKEN`
for automation, and credential storage without real account values. Link Go
installation from the developers page and the tap from CLI installation.
Do not imply public repositories make the server or hosted service open source.

Prepare the changes while private; publish the availability copy after anonymous
installation passes. Follow `docs/maintenance.md`: locale key/placeholder parity,
unchanged command text across locales, routes/anchors, signed-in/out navigation,
keyboard and narrow layout review. Run relevant Erlang docs/i18n tests and docs
JavaScript tests, then the required server suite with a disposable database.
Retain the home-page separation checked by `bibleit_http_docs_tests.erl`; add
released-install assertions on the CLI page. Server deployment is a separate
reviewed action, coordinated with its existing work.

See [public CI and settings preparation](public-ci-readiness.md) for the exact
reviewable settings proposal and the private-plan ruleset limitation. No settings
payload has been applied.

## 6. Approval and staged cutover

The [concrete approval package](launch-approval.md) records the current PRs,
release inputs, disclosure choices, exact cleanup and launch sequence. SDK/CLI
Actions PR creation/approval is now disabled; rulesets remain feature-blocked.

Homebrew needs public access to both the tap and the CLI release assets; opening
the tap alone only enables discovery/tapping. The Go library is not a runtime
dependency of the downloaded executable. If Homebrew distribution is prioritized,
coordinate CLI and tap access in the same approved launch window. The order below
also makes Go public first so CLI source builds and contributor CI work normally;
choose the approved launch grouping before cutover.

Present one concrete approval package: exact reviewed refs and release hashes,
redacted audit results, rights decisions, hosted cleanup decisions, README and
server docs diffs, CI/settings changes, release strategy, and validation matrix.
The user has required approval before public visibility changes and fresh public
installation checks in the README chat. Request that approval only after the
preparation package is complete. This plan does not grant it.

1. Make **bibleit-go** public first, then prove v0.0.1 or an approved replacement
   retrieves anonymously using the normal Go proxy/checksum database and an empty
   module cache. Check public package documentation and exported API visibility.
2. Activate the prepared CLI public-dependency CI changes. Run clean standalone
   tests/builds and a fork-equivalent PR path without secrets, sibling modules,
   local replace directives or GOPRIVATE. Finish reviewing all CLI assets.
3. Make **bibleit-cli** public. Verify approved release assets, checksums, manifests,
   installer scripts and latest/version-pinned URLs anonymously. A documentation
   cleanup does not require retagging; use a new release for any changed binaries
   or archive contents, and verify its generated formula before launch.
4. Make **homebrew-tap** public once its formula targets the approved accessible
   CLI release. Update through the existing verified formula PR flow if needed,
   and run syntax/style plus published-package checks before merging.
5. Validate every advertised distribution path below, then publish the reviewed
   server documentation, remove pending-availability README copy, and announce.

Public Go modules use proxy/checksum infrastructure; preserve released version
content and verify the existing CLI go.sum against public retrieval.
[Go module authentication reference](https://go.dev/ref/mod#authenticating).

## 7. Anonymous validation and completion criteria

Use clean machines/runners with no gh login, Git credential helper, cached module
or package, GH_TOKEN/GITHUB_TOKEN, private tokens or local asset overrides.
Use a disposable install directory/prefix. HTTP service commands may use a
dedicated test account; distribution downloads and builds must remain anonymous.

| Path | Required evidence |
| --- | --- |
| Go library | Empty-cache go get of approved version, go mod verify, compiled quickstart; default proxy/sumdb, no GOPRIVATE or Git credentials |
| CLI source | GOWORK=off download/verify/test/vet/build; no sibling checkout; fork PR CI succeeds without secrets |
| Bash/curl | Public script download and latest/version-pinned install on macOS/Linux; version/help, PATH guidance, reinstall and failed-checksum preservation |
| PowerShell | Public script download and latest/version-pinned install on Windows PowerShell 5.1 and PowerShell 7; version/help, replacement, PATH guidance and failure preservation |
| Homebrew | brew install bibleit-labs/tap/bibleit-cli from a fresh tap/cache on macOS/Linux, brew test, completions, reinstall/uninstall; real GitHub archive URLs |
| Manual archives | All six approved archives accessible and hash-correct; execute version/help on available native runners; record arm64 coverage gaps instead of claiming six native runs |
| Service quickstart | Browser auth/info/read/search/logout; registered-key SSH and subscription; output/exit codes match docs using synthetic Live data and cleanup |
| Public docs | Three locales render correctly; all repo/install links and anchors work anonymously; no coming-soon or localhost-CLI claims remain |

Scoop has a release manifest but no public bucket in the inspected release
records. Do not advertise a bucket-based command unless a reviewed public bucket
is explicitly added and tested; document manual manifest installation only if
chosen as part of launch scope. Record the macOS signing/notarization and Windows
code-signing status accurately; evaluate them separately from checksum integrity.

Complete when audits and rights review are closed, the three repos are public,
approved release versions and formulas agree, public CI and advertised installs
pass, public docs are deployed, and obsolete cross-repo read tokens are revoked.
Record results with exact refs, hashes, platforms and timestamps.

If an exposure appears during cutover, stop announcement and revoke affected
credentials immediately; coordinate withdrawal and cleanup. Returning a repo to
private cannot retract copies already downloaded or forked. Keep the repositories
private whenever a blocking finding remains unresolved.
