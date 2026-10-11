# Implementation status

Updated 2026-10-11. Counts describe explicit acceptance items, not estimated
percentages of the entire project. The full plan is in `plan.md`.

## Current increment: main and release CI

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Promote reviewed source to main | Complete | All three repos now default to main; previous default history preserved; no release created |
| Standalone CLI dependency | Complete locally | Exact Go snapshot and checksums; GOWORK=off tests pass |
| Build/test/draft release workflows | Complete; hosted validation passed | CLI/Go native matrices and all installer jobs pass; six standalone archives built; tag-triggered draft generation configured |
| Homebrew tap continuation | CI/configuration complete | Tap CI/access pass, eight updater regressions pass and actual standalone candidate formula verification passes; published-package validation added on macOS/Linux; first actual release/install still pending |

Hosted evidence: CLI native [38158698440](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38158698440)
and candidate [38158698450](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38158698450)
use source 13e4b72; Go native [38158759103](https://github.com/bibleit-labs/bibleit-go/actions/runs/38158759103)
uses 464d86d; tap validation [38158709825](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38158709825)
and access [38158727488](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38158727488)
use a41e453. Candidate manifest confirms clean, GOWORK=off, non-local assets and
the exact pinned Go-client checksum. Documentation-only updates retain these
explicitly tested revisions.

Progress: **4/4 infrastructure items complete**, with all 12 hosted jobs green.
Draft creation itself runs on the first approved tag; no release was created
during these validation runs.

No tags, published releases or visibility changes are part of this increment.
Real Turnstile replay rejection remains the provider release gate; email delivery/verification and fresh-token acceptance are human-confirmed.

## Current increment: canonical custom-domain migration

The user confirmed https://bibleit.app and ordinary SSH bibleit.app:22 through
the assessment chat during the fixed-endpoint work. The earlier Fly-hostname
increment below is complete; the new route needs its own evidence.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Configure canonical endpoints and trust migration | Complete | Go HTTP/browser origin and CLI SSH use bibleit.app; HostKeyAlias preserves the verified Fly-host known_hosts entry solely as a trust lookup; no user endpoint overrides |
| Update isolated fixtures and credential tests | Complete | Verified TLS fixture names bibleit.app; old localhost and Fly configurations reject without overwrite; real changed-host-key regression added |
| Refresh HTTP/SSH/import compatibility | Complete | Full run at 5969f2f (69 source/72 test/11 native); supplemental direct alias boot regression at ef63e91 (69 source/73 test/11 native) passes; sole runtime difference is parser registration; evidence retained separately |
| Publish source and record fresh private CI | Complete | Go 58505aa native matrix and CLI bc94231 native/candidate matrices pass; CLI pins client 59dfaa4 (documentation-only since 58505aa); one Linux build supplies identical archives to native Windows/Linux/macOS installer checks |

Progress: **4/4 custom-domain client items complete**. Local Go/CLI race tests and
vet pass; Darwin TLS/PTY/Bash/Zsh checks pass. The Turnstile/canonical-host boot
change at 5969f2f is qualified; all Erlang inputs used in that run match the commit.
The subsequent missing
BIBLEIT_HTTP_REDIRECT_HOSTS parser entry is repaired at ef63e91 and independently
passes direct boot qualification. Actual public-route technical acceptance now
passes, and the operator confirms Google/GitHub sign-in. Real Turnstile backend
success/replay remains pending.
The client uses the canonical origin directly and does not follow aliases. Dirty
metadata preserves unrelated frontend working-tree edits; final packaging inputs
still need freezing.
Assessment-owned actual public checks confirm the DNS-only cutover, HTTPS/aliases
on both address families, authentication, SSH fingerprint and paired rate behavior.
This public evidence is separate from local fixtures. No customer release
tag/assets are published; client work makes no production data changes.

Fresh private CI passes all ten jobs: three [Go native jobs](https://github.com/bibleit-labs/bibleit-go/actions/runs/38126756591),
three [CLI native jobs](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38128631673),
and the build plus three [native candidate installer jobs](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38128631681).
The repaired candidate workflow eliminates duplicate cross-compilation; Windows
PowerShell 7/5.1 and Scoop, macOS Homebrew and Linux installers pass on the same
six-target archive set. CLI runtime source is unchanged since 9243489; the
client's Go source is unchanged since its tested 58505aa revision.
The earlier Windows candidate timeout was resolved by this build-sharing change.
Detailed evidence is in [native-validation.md](native-validation.md). The full
isolated integration log is build/validation/final-canonical-integration-2026-10-11.log;
supplemental direct boot evidence is contract/server-boot-validation.json and
build/validation/alias-boot-validation-2026-10-11.log. No broad client rerun was
needed for the sole parser-registration change; the assessment’s repaired source
suite passes 337 CT/63 EUnit/71 JavaScript/native 65536/10 bootstrap checks.

### Actual canonical-domain technical acceptance

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Public DNS/TLS/alias routing | Complete | Two authorities plus Cloudflare/Google resolve the new DNS-only routes; 23 boundary and 39 alias checks pass on each address family |
| Authentication, SSH and paired rate behavior | Complete | 11 auth, 10 actual canonical-DNS SSH and 15 paired rate checks pass; retained public SSH fingerprint confirmed; scoped fixture cleanup counts are zero |
| Real provider sign-in | Complete | Operator confirms Google and GitHub OAuth sign-in at bibleit.app |
| Native HTML form/browser approval repair | Complete for technical admission | Deployed 34d80f8 real requestSubmit carries canonical Origin/Referer, passes CSRF and reaches missing-token denial; same-origin HTML policy retains sensitive no-referrer; provider traffic blocked, no account/email |
| Real Turnstile fresh-token acceptance | Complete, human confirmed | Original helper freshAcceptedBeforeInputValidation=true; real signup also completed; source14628f3, assessment ledger a40d27d |
| Actual email delivery and verification | Complete, human confirmed | Check-email page, delivered message and verification into the new account confirmed; identifiers and message contents omitted |
| Real Turnstile replay rejection | Failed; assessment repair in progress | Improved same-body human test reached invalid-email validation twice; human denial false; CSRF/rate/continuity interference ruled out; repair and genuine fresh-token/replay retest required |

Progress: **6/7 public acceptance items complete**; the client increment remains
4/4. The preceding technical probes qualified source c586107, runtime repair ef63e91/native f6fed37,
image sha256:1140436c3d9e7de1b1f80e88aa208a0a6c06cff4b7ba371e0f5ddcbb5b6ace22,
from [deployment 38129282912](https://github.com/bibleit-labs/bibleit-server/actions/runs/38129282912).
403 source hashes and 20 actual-image checks matched. Credential-free provenance
is in contract/canonical-deployment-validation.json, synchronized with the Go client.
The operator confirms Google/GitHub sign-in. The old app’s live.bibleit.app
certificate was removed; the new server retains all three Issued certificates.
Legacy app/data/volumes and Pages publication are preserved. Customer tags/assets
remain held; no client protocol/runtime change is required. A subsequent native
HTML form bug affects email signup, login, CLI approval and settings. Earlier
fetch/explicit-Origin probes missed actual browser Origin:null serialization.
Repair 34d80f811b3ac9179d6afe0966185e2b3517181e is now deployed and technically
qualified with native f6fed37, 403 source hashes and 20 image checks, from
[run 38144206075](https://github.com/bibleit-labs/bibleit-server/actions/runs/38144206075).
Image sha256:33b4ca3c13829c1c25f0745037212cc0c14e5e7768c2f78e58206c461ef8d35d
passes 23 boundary/39 alias checks per address family and real native form CSRF
admission. Full source suite passes 338 CT/63 EUnit/71 JS/native/10 bootstrap.
The native probe blocks providers and creates no account or email; real
Turnstile backend success/replay and actual email delivery/verification remain
pending. Qualification freeze/customer release stays held for that manual gate.
The separate catalog rollout uses server 6ba81e2/native d6e31dc in
[workflow 38144732991](https://github.com/bibleit-labs/bibleit-server/actions/runs/38144732991).
It is outside this point-in-time 34d80f8/f6fed37 qualification and requires its own
deployment/catalog and changed-input acceptance before inclusion.

### Catalog native rollout and focused client compatibility

The separately approved catalog release is operationally qualified at server
6ba81e2d2ec7d4bfbe5bbf819e77438a0f205665/native
d6e31dc8eb7c478e2f915e7bd156b047a2cb279e, image
sha256:7456bfec37b1d2285041387ef6135685dc707748ea82adbff60417e96c5fec47,
[run 38144732991](https://github.com/bibleit-labs/bibleit-server/actions/runs/38144732991).
405 source hashes, eight successful build stages, 20 actual-image checks,
native amd64, zero cleanup counts, matching published/running digest and 185 APK
comparisons with zero fixed-advisory alerts were independently verified. Refreshed
23 boundary/39 alias checks pass per explicit address family; native browser CSRF
admission passes with provider traffic blocked and no account/email. Existing
BPML has 66 books, omitted Luke17:36 is notfound and17:37 reads exactly.

Focused client compatibility found a separate localized UTF-8 server bug:
BPML/LUT1912 book-name literals emit Latin-1 bytes. Strict Go decoding correctly
rejects the finite catalogue; sparse present/omitted/chapter reads and invalid
index rejection preserving the prior generation pass. Repair 4243c9a is committed
and passes the full340 CT/66 EUnit/75 JS/native65536/10 bootstrap server suite;
all198 names and dashboard/protocol66-book catalogues agree across3languages.
Independent --catalog-only client qualification now passes five typed HTTP groups,
Portuguese/German catalogues and accented named reads, five HTTP/SSH parity cases,
one invalid-index publication case and three TLS checks at exact4243c9a/native
d6e31dc. All69source/75test/11native hashes match that commit and remain unchanged
after validation. The SDK/CLI decoder remains strict; no runtime decoding change.
server-catalog-validation.json retains this focused evidence separately. Exact repaired
deployment/public qualification is complete at server 4243c9a/native d6e31dc, image
sha256:46670b5fba63859a6892d64c1501122ecdddfd1edc1410555eba7c165143b5c1 (workflow
38147193441): 406 committed hashes, eight successful stages, 20 actual-image startup
checks, published/running digest parity and 185 APK comparisons with zero fixed-advisory
alerts. Refreshed boundary 23/alias 39 checks pass independently on IPv4 and IPv6;
native form CSRF admission and missing-token denial pass. Production BPML has 66 strict
UTF-8 BOOK records and matching Gênesis. The five public report hashes are recorded in
canonical-deployment-validation.json. Real Turnstile backend success/replay and actual
email delivery/verification remain open.
Earlier broad
client protocol integration stays attributed to its tested source. No customer
release; real Turnstile/email acceptance remains open separately.

## Previous increment: fixed Fly endpoints and compatibility refresh

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Configure service addresses | Complete | Go/client browser origin https://bibleit-server.fly.dev; ordinary CLI SSH bibleit-server.fly.dev:22; credential-only constructors, redirects blocked, TLS/host verification retained |
| Isolate tests and preserve credentials | Complete | Egress-disabled namespace and verified TLS/hostname fixtures; direct host integration rejected; native test-only loopback dialer; old HTTP/SSH bindings rejected without overwrite |
| Rerun HTTP/SSH/import checks | Complete | Clean server 8923d86/native f6fed37; 34 decoder/18 encoder, seven security, 11 HTTP, real OpenSSH lifecycle/JSON parity, 18 auth/key/protocol and 26 import cases; synchronized 68 source/70 test/11 native hashes |
| Update release plan and private CI | Complete | Plan marks endpoint/client compatibility work complete; exact client b081081/CLI 26bbfaa native and candidate matrices pass on Windows/Ubuntu/macOS |

Progress: **4/4 increment acceptance items complete**; assessment remains **25/25**.
Local Go/client and CLI race tests/vet pass; Darwin process TLS/PTY and Bash/Zsh
checks pass. Linux CI supplies Fish; Windows CI verifies the test-only TLS path.
Integration uses synthetic local principals and a unique disposable database,
which was removed. No production principal, deployment or shared database changed.
Full ignored log: build/validation/production-endpoint-integration-2026-10-11.log.

Fresh private CI passes all nine jobs: [Go native](https://github.com/bibleit-labs/bibleit-go/actions/runs/38125983246),
[CLI native](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38126046169),
and [candidate packaging/installers](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38126046149).
This includes Linux Fish and native Homebrew/Scoop/Bash/PowerShell 7/5.1 lifecycle checks.

Release work still requires version/customer-access decisions, a released pinned
Go dependency, standalone production-candidate validation and approved customer
assets followed by real tap/installation checks. No release tags/assets are published.

## Previous increment: final assessment closure and release-plan reconciliation

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Verify assessment closure | Complete | Ledger 8923d86 records 25/25 scoped packages complete; source 7442010 deployed, native f6fed37 unchanged; 334 CT/56 Erlang/71 JavaScript/native/ten bootstrap tests pass |
| Record production qualification | Complete | HTTPS bibleit-server.fly.dev, ordinary SSH port 22; 23 HTTPS/11 authenticated HTTP-WSS/4 SSH/13 paired checks pass; native/deployed image campaigns pass 20 checks |
| Reconcile release plan | Complete | Removes obsolete deployment/unknown-address blockers; records fresh client integration, fixed-address configuration, version/access and standalone packaging/install gates |

Progress: **3/3 closure-review items complete; assessment 25/25**. The current
CLI/client runtime still targets localhost and the verified contract baseline is
an older development snapshot. No endpoint, runtime, baseline, release tag or
production mutation is performed in this documentation follow-up. Final deployed
admission code needs a fresh independent client integration refresh before release.
SSH clients continue normal SSH and must not inject a PROXY preamble.

Evidence: [native source 7442010 run](https://github.com/bibleit-labs/bibleit-server/actions/runs/38106224843),
assessment ledger 8923d86 and final qualification section. See plan.md for the exact
production image digest, scoped results and operational limits. The qualified
Fly hostname does not qualify a future custom domain or establish capacity/failover.

## Previous increment: native amd64 server validation coordination

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Coordinate with assessment | Complete | Assessment owns server/CI/ledger work; isolated private branch security-native-validation/package-25 preserves shared main |
| Review native amd64 server results | Complete | Run 38101109368 at f5f1736 passes on Linux x86_64/native amd64 Docker; native corpus, ten SSH-bootstrap and 17 final-image checks pass; downloaded hashes/image/cleanup verified |
| Update and publish release plan | Complete | Exact native evidence recorded; local architecture gate closed for this snapshot; approved deployed Fly and release-freeze gates remain |

Progress: **3/3 tracked items complete**. [Native server run](https://github.com/bibleit-labs/bibleit-server/actions/runs/38101109368)
uses clean native f6fed37. Validation image:
sha256:ab28d63404454a94ac1d891e3d0ae1e042a8ab559c314403c0de37f6adf42706.
All eight stages exit zero; both owned image-tag cleanup codes are zero.
Node24 action pins and artifact upload pass; zero job annotations and no Node20
deprecation warning in the completed log.
The 389-file report matches every existing client source/test/native baseline hash.
The two newer boundary test files do not rewrite the independent client snapshot.

The native image job does not run the full CT suite. The separate local full retest
passes 331 integration/51 Erlang/71 JavaScript/native/ten SSH-bootstrap tests;
ARM64 image checks also pass 17 cases. No application/native source fix was needed.
No client runtime, verified baseline, tag or deployment changes here. Assessment
was **24/25** at this native checkpoint, with Fly boundary evidence then pending.
The final closure above supersedes that assessment blocker. This validation image
is not a registry-published release.

## Previous increment: package 25 final local results and compatibility refresh

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Review package 25 final results | Complete | 326 integration, 51 Erlang unit, 71 JavaScript, native, ten SSH-bootstrap and 16 final-image checks pass; assessment stays 24/25 pending approved Fly boundary evidence |
| Rerun HTTP/SSH/import integration | Complete | 34 decoder/18 encoder fixtures, seven security cases, 11 HTTP subtests, SSH lifecycle/output parity, 17 auth/key/protocol tests and 26 import cases pass |
| Synchronize baselines and update release plan | Complete | Identical 67 source/66 test/11 native fingerprints; includes boot/public-origin fixes; client 3bec5bc and CLI 438c0ca pushed; all three hosted matrices pass |

Progress: **3/3 requested items complete**. Fresh hosted evidence:

- [Go client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38099677783): Windows/macOS/Ubuntu passed at 3bec5bc.
- [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38099709263): all three platforms passed at 438c0ca with exact client 3bec5bc.
- [CLI candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38099709267): all three platforms passed at the same inputs; six-target packaging, Homebrew, PowerShell 7/5.1 and Scoop checks passed.

Documentation-only completion commits follow the tested CLI revision.
Server c2beef5 plus working-tree changes and clean
native f6fed37 identify the frozen development snapshot. All input sets matched
immediately after the run, and its disposable database was removed. Ignored log:
build/validation/package-25-integration-2026-10-11.log.

No client runtime or server source change was required. Production origin/SSH
configuration, an agreed committed/deployed release, native amd64 server-image
execution, Fly proxy fairness, explicit database trust and volume ownership remain
release gates. The CLI/client CI matrices do not certify the server deployment.

## Previous increment: package 24 final results and compatibility refresh

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Verify final assessment result | Complete | 324 integration, 50 Erlang unit, 71 JavaScript and native tests pass, no skips; release/tar resource checks and Linux build stage pass; campaign 24/25, Fly verification pending |
| Rerun HTTP/SSH/import integration | Complete | 34 decoder/18 encoder fixtures, seven security cases, 11 HTTP subtests, SSH lifecycle/output parity, 17 auth/key/protocol tests and 26 import cases pass; no client runtime change |
| Synchronize baselines and update release plan | Complete | Server c2beef5 plus working-tree changes, native f6fed37; 67 source/66 test/11 native hashes; baseline bytes identical; draft CI pins client 3c43133 |

Progress: **3/3 requested items complete**. Fresh hosted evidence:

- [Go client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38098572121): Windows/macOS/Ubuntu passed at 3c43133.
- [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38098645524): all three platforms passed at 990215d with exact client 3c43133.
- [CLI candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38098645722): all three platforms passed at the same inputs; six-target packaging, PowerShell 7/5.1, Scoop and Homebrew checks passed.

Documentation-only completion commits follow the tested CLI revision.
Recorded inputs matched immediately after the isolated run; its database was
removed. Package 25 subsequently edited server boot/public-origin handling and
two security tests outside the verified snapshot; native inputs still match. Full ignored log: build/validation/package-24-integration-2026-10-11.log.
This validates a development snapshot. Production endpoints, committed/deployed
baseline, migration lineage, standalone released dependency and customer installs
remain release gates. No shared database, server runtime source or release tag is
changed by this increment.

Package-25 local follow-up: the assessment's 326 integration/51 Erlang/71
JavaScript/native/ten SSH-bootstrap tests and 16 isolated image/database checks
pass. Overall progress stays 24/25 because approved Fly deployed-boundary evidence
is missing. Canonical public-origin login, server UID10001 volume ownership with
preserved SSH identity, explicit database trust and native amd64 server-image
verification are now recorded as launch gates. These are separate from the green
CLI native jobs. No client runtime or verified baseline update is performed here.

## Previous increment: package 23 final-results review

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Verify final assessment result | Complete | Assessment confirms 313 integration cases, 48 Erlang unit tests, 71 JavaScript tests and native search/fuzz tests pass without skips; test container/volume removed; package 23 complete for bounded local scope |
| Update release gates and recovery evidence | Complete | Campaign now 23/25; packages 24–25 remain; local fault/restore success distinguished from pending production recovery/deployment verification |
| Publish documentation | Complete | Plan, release guide and status updated on integration-refresh with a Conventional Commit; runtime and verified server/native baseline unchanged |

Progress: **3/3 requested review items complete**. Package 23's local fixtures
validate restore quarantine and controlled degradation, not production RPO/RTO,
backup automation, whole-volume crash consistency, OOM or physical disk exhaustion.
Customer fresh-credential/SSH-key guidance remains applicable after an actual
approved restore. No client integration or production restore is claimed here.

Package-24 follow-up: the final full suite and local release/build checks now
pass; see the current compatibility refresh above. Release documentation now records migration
identity/lineage requirements, bounded verified HTTPS import policy and guarded
server packaging/runtime defaults. CLI commands/error names and native revision
are unchanged. Campaign progress is 24/25. No runtime edits or database reset
were needed; the verified baseline was refreshed only after independent integration.

## Previous increment: assessment compatibility refresh

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Check latest assessment results | Complete review | Package 22 final suite passes: 310 integration, 48 Erlang unit, 71 JavaScript and native C programs, no skips; overall campaign 22/25, packages 23–25 pending |
| Rerun isolated HTTP/SSH/import integration | Complete | 34 decoder/18 encoder fixtures, seven security cases, 11 HTTP subtests, real SSH lifecycle/output parity, 17 auth/key/protocol tests and 25 import cases pass; disposable database removed |
| Synchronize baselines and update release plan | Complete | Server c2beef5 working-tree snapshot, native f6fed37; 65 source/62 test/11 native fingerprints; client and CLI baseline bytes identical; private CI pins updated to client 8de21e9 |

Progress: **3/3 requested items complete**. Current compatibility passes with no
client regression. Fresh hosted verification:

- [Go client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38096465359): Windows/Ubuntu/macOS passed at 8de21e9.
- [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38096515700): all three platforms passed at e737565 with client 8de21e9.
- [CLI candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38096515593): all three platforms passed at the same inputs; six-target packaging, PowerShell 7/5.1, Scoop and Homebrew checks passed.

Documentation-only completion commits follow the tested CLI source revision.
This is development integration evidence: the server working tree has assessment changes and the
campaign remains 22/25 complete. The committed/deployed baseline, production
addresses, released Go dependency and final standalone/customer-install gates
remain open. The assessment has subsequently added two package-23 test files
outside this frozen snapshot; recorded source/test/native inputs still match.
No server edits or release tags/assets are created here.

Package-23 follow-up: the final local assessment suite now passes and campaign
progress is 23/25; see the current review above. Customer guidance covers fresh
credentials and SSH key re-registration after an approved restore/quarantine and
identity/grant reconciliation. CLI runtime and the verified baseline are unchanged;
no production restore is claimed.

## Previous increment: first-release readiness review

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Review current sources and hosted evidence | Complete | All three repos private and clean; no published releases; client/CLI native and development candidate CI green; tap syntax and actual secret access green |
| Identify and order remaining release gates | Complete | Eight ordered gates in plan.md: version/access, production addresses/fixtures, server baseline, released Go dependency, standalone production candidate, release notes/source promotion, publication/tap PR, customer installs |
| Update and push release documentation | Complete | plan.md, releases.md and this status revised; Conventional Commits on integration-refresh; no tags, releases, merges or visibility changes |

Progress: **3/3 review items complete**. This measures the review, not launch
readiness. Production HTTP/SSH constants are still loopback placeholders and
CLI go.mod still requires bibleit-go v0.0.0. One of 63 recorded server source
fingerprints differs (bibleit_http_listener.erl); all 59 recorded test fingerprints
match. No integration rerun is claimed for the later server edit.

The next work needs production addresses and an agreed deployed server revision,
plus the release version/customer access decision. Final standalone validation,
real tap release-download/PR flow and customer install checks remain outstanding.

Assessment follow-up: native libbibleit index/cache hardening and management
WebSocket keepalive changes were reported after the review. Code inspection found
no CLI/Go management WebSocket transport or ping cadence assumption, so no matching
client code change is needed. Full server suite results were pending at notification;
freeze server plus native submodule revisions and rerun compatibility before release.

## Previous increment: Homebrew updater access and permissions

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Private CLI access preflight | Complete | BIBLEIT_CLI_READ_TOKEN configured; Actions verified private CLI metadata and release-list access at tap bf6a46b in [run 38092683138](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092683138); no releases, tags or PRs created |
| Actions PR permissions | Complete | Final API check confirms can_approve_pull_request_reviews=true with default_workflow_permissions=read; organization restriction was lifted during setup; update job alone requests write permissions |
| Validate and record setup | Complete | Seven updater regressions, shell/YAML checks and hosted Homebrew syntax/style pass at tap 04e8b16 in [run 38092320412](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38092320412); actual secret access check also passed |

Progress: **3/3 tracked setup items complete**. The user confirmed the token is
configured for only bibleit-cli with Contents: read-only; the Actions access check
verifies it can read the private CLI, without exposing the credential. This does
not independently inspect the token's full scope. Formula downloads and PR
creation remain an end-to-end gate for the first real production CLI release.

## Previous increment: single credential configuration and tap CI

Branches remain integration-refresh (CLI/client) and formula-preparation (tap).
Source pushes to the three private repositories are authorized. No tags, customer
releases, automatic merges or repository visibility changes are performed.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Remove named profiles | Complete | One schema-v2 service-bound config; HTTP default, explicit --transport ssh; cached token and optional identity retained; no profile commands/flag/environment selection/output/completion; old files rejected and untouched |
| Add Homebrew tap CI | Complete implementation | Tap syntax/style plus seven updater regressions; manually dispatched release updater verifies published CLI-generated formula/archives and opens a Conventional Commit PR without auto-merge |
| Validate, update plan and push | Complete | Local race/vet/process/PTY/Bash/Zsh and full frozen-server integration pass; synchronized client baseline; client/CLI native, candidate and tap matrices all green; prepared source and documentation pushed |

Progress: **3/3 tracked items complete**. The local server integration log is
build/validation/single-config-integration-2026-10-11.log (ignored): 34 decoder/18
encoder fixtures, 7 security cases, 11 HTTP subtests, real SSH lifecycle and
single-credential/output parity, 17 auth/key/protocol tests, 10 import-access and
15 publication cases. The snapshot includes 63 source and 59 test fingerprints;
new plan-support assessment files are included in the snapshot, not additional
client test coverage. The harness is a focused integration gate, not the full
server security campaign.

Tap update publication is prepared but not invoked: no CLI release exists. Private
release reads need BIBLEIT_CLI_READ_TOKEN (Contents: read only to bibleit-cli) in
the tap; Actions PR-creation policy must permit the write job. These setup steps
are documented in the tap. Tap syntax CI does not fetch unpublished candidate
URLs; real lifecycle checks run through the CLI candidate workflow.

Fresh hosted evidence for this increment:

- [Client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38091438398): all three platforms passed at 17e146b.
- [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481226): Windows/Ubuntu/macOS race/vet/build/process, config ACL/permissions, and Linux Bash/Zsh/Fish passed at fe719db with exact client 17e146b.
- [Candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481159): all three platforms passed at fe719db; six-target packaging, native installers, PowerShell 7/5.1, Scoop and Homebrew.
- [Tap syntax/style and updater regressions](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38091389888): passed at d08561b. The updater publication workflow was not invoked; no actual release exists.

Documentation-only commits follow the tested runtime commit. The assessment
continued changing server sources/tests after the frozen run; those later edits
are outside this baseline. Rerun against the agreed deployed baseline before
release. The current increment introduces no server edits and found no client
integration regression. Production addresses, released/pinned dependency,
standalone production packaging and customer-accessible publication remain open.

## Previous increment: release readiness and private source publication

Updated authorization: all prepared source work may be pushed to the private
CLI, Go-client and Homebrew tap repositories. This supersedes earlier local-only
source restrictions; customer release tags/assets and visibility changes remain
separate readiness gates. Branches remain `integration-refresh` (CLI/client) and
`formula-preparation` (tap), with Conventional Commits and no codex prefix.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Review remaining release gates | Complete | Production addresses/deployment, endpoint fixture adaptation, released/pinned dependency, GOWORK=off build, final candidate/platform checks, customer-accessible distribution and release sequence recorded |
| Update next steps and validation configuration | Complete | Ordered nine-step checklist in plan.md; current client SHA pinned in both CLI workflows; integration-refresh now triggers candidate checks; stale local-only policy corrected |
| Push prepared work and refresh hosted verification | Complete | Client 310259e, CLI validation source 0ce1c47 and tap dc2c4a5 pushed; all client/CLI native and package-candidate jobs passed |

Progress: **3/3 tracked readiness-review items complete**. This is a review/source
publication increment, not a customer release. Repository privacy was verified
through authenticated interactive zsh; no visibility changes, tags or releases.

Fresh hosted evidence (2026-10-11):

- [Client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38090370470): Windows/Ubuntu/macOS race, vet and examples passed at 310259e.
- [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38090456759): Windows/Ubuntu/macOS race/vet/build/process and Linux Bash/Zsh/Fish passed at 0ce1c47 with exact client 310259e.
- [Candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38090456683): six-target development packaging, native installers, Windows PowerShell 7/5.1 and Scoop, and macOS Homebrew passed at the same CLI/client inputs. Private validation artifacts remain development-only.

These runs use the development endpoint and workspace. The ordered production
release checklist is in plan.md; green draft CI does not close those remaining
gates. Documentation-only completion commits follow the tested source revision.
All prepared commits are pushed, with no default-branch merges, visibility
changes, release tags or customer releases.

## Previous increment: Homebrew naming and assessment protocol fixes

Branches: CLI/client `integration-refresh`, tap `formula-preparation`. Local only:
no pushes, tags, releases or CI runs. Validation is captured by CLI `6d0ebf8`,
Go client `2b68ccc` and tap `7834751`; the server remains assessment-owned.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Prepare the separate Homebrew package | Complete | User selected homebrew-tap with formula bibleit-cli; generator and tap agree; actual install/test/reinstall/completions/uninstall pass in a disposable prefix |
| Check quoting/SECRET SET and update the client | Complete | Server rejects malformed quoting, preserves literal/trailing backslashes and documents generated secrets; Go restrictions retained, README/contract clarified, seven additional request fixtures and builder regressions added |
| Fresh integration, baselines and plan | Complete | 34 decoder/18 encoder fixtures, 7 security cases, 11 HTTP subtests, real OpenSSH checks, 17 auth/key/protocol EUnit tests, 10 import-access and 15 publication cases; Go/CLI race and vet pass |

Progress: **3/3 requested items complete**. Milestone 1 is now **7/7 tracked core
contract items complete**. The quote integrity and SECRET SET discrepancies are
closed by verified rejection/documentation, not by introducing new request escapes
or custom secret assignment. All other Go README behavior/limits remain valid.

The future customer command is `brew install bibleit-labs/tap/bibleit-cli`; the
installed executable is `bibleit`. The tap's development formula URLs remain
unpublished. The checker uses local CLI archives and a disposable prefix; Homebrew
still fetched developer gems and core metadata, so the report no longer claims
fully offline execution. The user's normal Homebrew prefix was untouched.

The synchronized baseline includes 63 source and 58 test fingerprints and matched
the server working tree at completion. The baseline review date is UTC; this
record uses the local date. Evidence is ignored locally under:

- `build/validation/homebrew-cli-naming.json`
- `build/validation/protocol-integration-2026-10-11.log`

The assessment subsequently confirmed a green full `make test` on 2026-10-11:
248 Common Test cases, 48 Erlang unit tests, 71 JavaScript tests and the native
C test, with no skips. This is server-owned evidence recorded in
`../bibleit-server/docs/developers/security-test-results.md`, separate from our
client integration run; its protocol contract matches the client safeguards.
Production HTTPS/SSH addresses, a published/pinned Go client, standalone GOWORK=off
validation and publication authorization remain release gates. Browser-only
workspace/event contracts and the other language clients remain future work.

## Previous increment: latest assessment translation validation

Branch: `integration-refresh`. Local only: no pushes, tags, releases or CI runs.
Inputs: CLI harness revision `26de771`, Go-client revision `5138077`, server HEAD
`1a266cc8bf6aa726c4890b78ddf95a3aadda0ba0` plus assessment working-tree changes.
The new immutable publication module and translation/registry changes are now
included in `contract/server-baseline.json`, with 63 source and 57 test fingerprints.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Inspect latest assessment changes | Complete | Immutable generation publication, registry generation-aware handles, local import-access and publication suites; existing HTTP/SSH checks extended with optional `--imports` |
| Validate frozen source and integration | Complete | 27 decoder/18 encoder fixtures; 7 targeted security cases; all 11 Go HTTP subtests; real OpenSSH lifecycle/profile/output checks; 8 auth/key EUnit tests; 10 import-access and 14 publication cases |
| Update baselines, plan and local commits | Complete | CLI/client baselines synchronized; source and test fingerprints matched the server at completion; disposable database removed; remaining contract/publication gates retained |

Progress: **3/3 tracked assessment-validation items complete**. All 24 import
cases passed, including concurrent installs/archives, retained native handles,
idempotent provenance/retries, registry outage, legacy upgrade, install/delete
races, path containment, staged write/index/manifest failures and process death
before/after publication. Import access cases use real HTTP authorization and
loopback archive fixtures. Publication faults use the assessment's production-code
fixture clone at filesystem/NIF boundaries. Suite cases execute their actual
setup/teardown callbacks; this is focused execution, not a full Common Test campaign.
No client/server regression or runtime fix was needed. Later assessment edits
require a new snapshot; the matching fingerprints are a completion-time observation.

Reproduce:

```sh
python3 scripts/check_server_container.py --imports --baseline contract/server-baseline.json
```

Output: `build/validation/assessment-integration-2026-10-10.log` (ignored local
evidence). This closes the previous translation/registry snapshot gap; it does
not close production deployment, browser-only API contracts or the full server
security assessment.

## Previous increment: fresh full HTTP/SSH integration

Branch: `integration-refresh`. Everything in this increment stays local; no pushes,
tags, releases or hosted CI runs. The test used CLI source revision `045fd20` and
Go-client revision `9332d35` (the canonical LIVE STATS fixture refresh), against
server HEAD `1a266cc8bf6aa726c4890b78ddf95a3aadda0ba0` plus its uncommitted assessment changes. Exact server
source fingerprints are recorded in `contract/server-baseline.json`.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Baseline inspection and isolation | Complete | Go 1.27.1; clean CLI/client inputs; frozen server working tree; Linux fixture NIFs rebuilt; fixed localhost:8080 in a disposable network namespace |
| Full tracked HTTP/SSH integration | Complete | 27 decoder/18 encoder fixtures; 7 targeted security cases; all 11 Go HTTP subtests; real CLI/OpenSSH subscription, pause/disconnect/reconnect/deletion closure and profile/output parity; 8 auth/key EUnit tests |
| Evidence, plan and local commits | Complete | CLI/client server fingerprints refreshed; exact frozen snapshot recorded; disposable database removed; plan retains production and browser-only contract gates |

Progress: **3/3 tracked integration-refresh items complete**. Log:
`build/validation/full-integration-refresh-2026-10-10.log` (ignored local evidence).
The first run also passed (`full-integration-2026-10-10.log`); a second run
included a fetcher edit detected after its snapshot. The assessment continued
editing `bibleit_translation.erl` and `bibleit_translation_registry.erl` after
the second snapshot; those subsequent edits are outside this baseline.
The repaired LIVE STATS encoder passed canonical fixtures and real subscriber
statistics. There were no integration regressions requiring client or server fixes.
The SSH algorithm-negative tests intentionally log rejected handshakes; all eight
EUnit tests pass. This is the full existing client/CLI integration harness, not the
entire server security assessment or coverage of browser-only organization APIs.

Reproduce:

```sh
python3 scripts/check_server_container.py --baseline contract/server-baseline.json
```

## Previous increment: package-manager and native Windows validation

The user authorized private CI validation after recreating `bibleit-cli` and
`bibleit-go` privately. Validation branches have no codex prefix; no release tags
or distributed binaries/formulas are published. `homebrew-tap` is separate and
its development formula preparation remains local.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Homebrew install lifecycle | Complete on local macOS/arm64 | Real generated formula in disposable prefix: install, formula tests, completions, reinstall and uninstall passed; normal user prefix untouched |
| Go-client native matrix | Complete | `private-platform-validation`, commit `3d15ec6`; Windows/Ubuntu/macOS race tests, vet and builds passed in private run 38086747099 |
| CLI native and Scoop execution | Complete | CLI matrix passed at `1e58c54` in run 38088193147; release-candidate matrix passed in run 38088193158, including native PowerShell 7/5.1 installation/replacement/failure checks and real Scoop shim/reset/reinstall/uninstall/hash rejection |
| Separate tap and verification records | Complete | Local tap commit `61ad753` on `formula-preparation`; actual tap formula passed isolated Homebrew checks; native reports downloaded under ignored `build/validation/ci-38088193158`; no tap push |

Progress: **4/4 tracked validation items complete**. GitHub disables deploy keys
on `bibleit-go`; user selected `BIBLEIT_CLIENT_READ_TOKEN` with contents-read only
on that repository and configured it. The current production endpoint and standalone client-release
gates are unchanged. See `docs/releases.md` for repeatable local checks.

[Go-client native CI passed](https://github.com/bibleit-labs/bibleit-go/actions/runs/38086747099).
[CLI native CI passed](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38088193147).
[Installer/Homebrew/Scoop CI passed](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38088193158).

Native Windows exposed a PowerShell null-to-empty conversion during binary
replacement; `[NullString]::Value` fixes the backup-path argument. The harness
now resets inherited module paths when launching another PowerShell runtime.
Scoop checksum rejection is verified by its explicit error and absence of an
installed executable/current link/shim, since this engine can return zero after
a pipeline install failure. Validation uses real archive hashes and isolated roots.

Server follow-up (2026-10-10): rechecked the Go client's duplicate-field comment.
The previous `LIVE STATS` encoder prepended `connections` while retaining the same
field from the statistics map. At the user's request, the security assessment
chat was asked to emit one authoritative count and cover empty/nonempty subscriber
lists and stale conflicting counts. The client already handles canonical output,
accepts identical duplicates for compatibility and rejects conflicting duplicates.
The assessment fixed the encoder and reports three protocol regressions plus a
green full suite. Local canonical CLI/client fixtures now use one field; a legacy
client fixture preserves duplicate compatibility. Fresh read-only server checks
passed 27 decoder and 18 encoder fixtures, and Go-client race tests passed.
Client fixture/comment refresh: local commit `9332d35`; no client push.
Earlier full-integration fingerprints remain historical; this refresh did not
rerun full HTTP/SSH integration or publish a new client revision.

## Previous increment: local release packaging and installation paths

Branch: `local-release-packaging`. Everything stays local: no pushes, tags,
GitHub Releases, hosted workflow runs, Homebrew tap or Scoop bucket changes.

| Item | Status | Evidence / limits |
| --- | --- | --- |
| Distribution design | Complete | Homebrew, Bash/cURL, PowerShell, direct archives and Scoop selected; future channels and maintenance scope documented |
| Archives and build metadata | Complete | Go 1.27.1; six amd64/arm64 macOS/Linux/Windows archives; version injection, license/completions, eight SHA-256 entries; repeat builds have identical archive/checksum bytes |
| Installer and package-manager preparation | Complete locally | Checksum-required per-user installers, version validation and safe replacement; generated Homebrew formula and Scoop manifest use real archive hashes |
| Local validation and future native gate | Complete for local scope | macOS/arm64 and disposable Linux/arm64 offline install/reinstall and failure cases pass; PowerShell parser/failure cases pass in a Linux container; Ruby syntax and manifest checks pass; native Windows and actual Homebrew/Scoop install/test remain pending |
| Reconcile plan and local Conventional Commit | Complete | Plan marks core/client split/fixed-endpoint work complete; release docs distinguish local preparation from production/publication; commit stays local |

Progress: **5/5 tracked local-preparation items complete**. Candidate files are
ignored under `build/releases/`. The prepared manual workflow validates future
standalone candidates and native installers with read-only repository permissions;
it has not been pushed or run. Production packaging currently rejects the
unpublished `v0.0.0` client dependency unless `--local` is explicit. Publication
remains paused. See [releases.md](releases.md) for reproduction and release gates.

## Previous increment: simple API-token clients and real fixed-service integration

All work remains local; publication and release tags stay paused. Both branches
are `simple-client-integration`, with no codex prefix. Server files are read into
a frozen snapshot and are not edited. The complete fixed-endpoint run passed on
2026-10-10; its baseline is `contract/server-baseline.json`.

| Item | Status | Evidence |
| --- | --- | --- |
| Choose credential-only API | Complete | `NewClient(apiToken)`; API token terminology; separate custom HTTP transport constructor; validation error return retained |
| Simplify Go client and update CLI | Complete | Config/AuthConfig removed; NewAuthClient() takes no options; examples and callers updated; no endpoint override |
| Local verification | Complete | Go client and CLI race tests, vet and build pass with Go 1.27.1; invalid credentials rejected before requests |
| Real fixed-endpoint integration | Complete | Public constructors at localhost:8080 in a disposable Linux container namespace; 27 decoder/18 encoder fixtures, 7 security cases, Go HTTP/domain/auth checks, actual CLI/OpenSSH lifecycle, 4 PKCE and 4 SSH EUnit cases pass |
| Document baseline and commit locally | Complete | Updated endpoint policy, client examples, container harness instructions and matching baselines; local Conventional Commits only |

Progress: **5/5 tracked items complete**. Host port 8080 was occupied; the isolated
runner used its own namespace and temporary PostgreSQL instance with no published
ports. The developer's existing HTTP process remained running with the same PID.
Native Windows was not rerun for this increment, and these checks are not a full
server security assessment or a production-endpoint release certification.

Run locally:

```sh
python3 scripts/check_server_container.py --baseline contract/server-baseline.json
```

The container runner requires Docker and build-image/package downloads on its
first use. It uses a cached Go 1.27.1 image, Erlang 29, and rebuilds fixture NIFs
from the frozen source. No developer .env, host keys, live database, Docker socket
mount, or private server source is baked into the tools image. Source mounts are
read-only; compilation and test data live in disposable directories/containers.

## Previous increment: fixed service and local-only work

The user paused all GitHub pushes, release tags and binary publication, and is
handling deletion/recreation of the public repositories. Previous publication
and hosted evidence below are historical. The server stays separate/private.

| Item | Status | Evidence |
| --- | --- | --- |
| Fixed Go client/authentication service | Complete | No public Endpoint field; one `ServerEndpoint` constant at `http://localhost:8080`; request, exchange and logout routing verified; endpoint environment ignored |
| Fixed CLI profiles and authentication | Complete | No endpoint flag/completion; account credentials and transport only; foreign bindings and unbound legacy bearer credentials rejected before requests or config writes |
| Preserve local work and unpublished dependency | Complete | Both pre-change Git bundles verified in ignored `.local-backups/`; sibling workspace uses both modules; unreleased client `v0.0.0` placeholder |
| Validate local draft | Complete | Both modules race/vet/build on macOS and Linux; native process/PTY; Linux Fish/Bash/Zsh; CLI six-target cross-build; both integration binaries compile |
| Document and commit locally | Complete | Endpoint policy, examples, contribution notes and release hold updated; Conventional Commits on `fixed-service-endpoint`; no pushes |

Progress: **5/5 fixed-service/local-only items complete**. Native Windows and
real-server integration were not rerun for this draft. GitHub cleanup belongs
to the user and is not included in this progress count. Before release, replace
the placeholder with production HTTPS/SSH addresses, publish a reviewed Go
client version, pin it in the CLI, and validate with `GOWORK=off` and hosted CI.

## Milestone 1: verify the integration contract

| Item | Status | Evidence / remaining work |
| --- | --- | --- |
| Operation and transport inventory | Complete | `contract/features.md` separates server availability, authentication, and client support |
| Shared request/response fixtures checked against source | Complete | 27 decoder cases and 18 encoder cases verified by compiling current Erlang source |
| Client safeguards for known incompatibilities | Complete | SECRET SET no longer becomes authentication; identical stats fields accepted, conflicts rejected; request size and JSON errors covered; command-aware framing handles single-record translation replies with count |
| Isolated server authentication/security checks | Complete for targeted coverage | 7 security cases (including SSH revocation), 4 PKCE cases, and 4 SSH cases; isolated source/dependency snapshot and disposable container/database; no developer .env |
| Go client against a real HTTP listener | Complete for core smoke coverage | Ping, identity, translation add/remove/list, named verse/numeric chapter reads, search, Live create/list/info/stats/start/stop/pause/delete/push, quota errors, expired credentials, and not-found handling |
| Complete integration coverage required by milestone gate | Complete for named core operations | HTTP translation/Live/expiry/quota checks plus actual CLI OpenSSH verse/pause/disconnect/reconnect/closure; assessment fixed normal deletion during stream authorization changes |
| Resolve server documentation/parser and encoder discrepancies | Waiting on server changes | SECRET SET mismatch and repeated connections in stats; assessment fixed scoped-token encoding and SSH deletion closure |

Progress: **6 of 7 tracked milestone-1 items complete**. Source fixtures,
targeted server security checks, all Go HTTP checks and the CLI OpenSSH
subscription check pass against the updated baseline. The assessment fixed
both scoped-token encoding and the intermediate normal-deletion regression.

The milestone gate remains open until the server contract discrepancies are
resolved. These checks cover the named core operations, not the full security
assessment or future workspace/event APIs.

## Remaining milestones

| Milestone | Status | Progress |
| --- | --- | --- |
| 2. Complete verified Go client surface | Complete for tracked core | 5/5 acceptance items and complete HTTP/SSH integration pass; future API expansion remains separate |
| 3. Customer CLI profiles/auth/output | Complete for tracked scope | 5/5: profiles, credentials, output, completion/destructive UX and representative native platform validation |
| 4. Workspace and event integration APIs | Pending server contract | Existing browser endpoints are not bearer-token client APIs |
| 5. Python, Ruby, Rust clients | Not started | Wait for verified core contract |
| 6. Independent publication | Local preparation complete; publishing paused | Six-target archives/checksums, installers and Homebrew/Scoop manifests prepared; production addresses, real client version, standalone release builds and publication pending; native Windows/package-manager validation complete |

## Reproduce verification

```sh
go test -race ./...
go vet ./...
python3 scripts/check_server.py
python3 scripts/check_server.py --integration --baseline contract/server-baseline.json
```

The source check needs Erlang/OTP with its JSON module. Integration additionally
needs Go 1.27.1, Docker with the local `postgres:17-alpine` image, built server
dependencies/NIF, and the sibling server checkout. It creates a uniquely named
loopback-only temporary container and random database, compiles source into a
temporary source/dependency snapshot, skips asset symlinks, detects concurrent Erlang source edits during copying, and removes its container on success or failure. It does
not build or modify the server checkout or load the development .env.

The regular Go test run skips live-server integration; the isolated harness runs
it explicitly. Server tests and client HTTP tests are separate evidence, both
reported by the harness. Shared malformed/future-response fixtures are decoder
tests, not claims that the server encoder emits malformed or hypothetical data.

## Previous increment and assessment coordination

The server security assessment chat is `01a12273-bf8b-7e30-826c-6354f7307511`.
Its latest completed turn reports fixes for unverified OAuth emails and disabled
user/account authority. This client work reads that checkout and uses a copied
snapshot; it does not edit the server or its assessment files.

Completed in this increment:

- Fixed finite reply framing for account translation list/add/remove. These
  responses contain `count` but have no `END`. HTTP and CLI SSH exec now pass
  the validated command into the decoder; truncated multi-record replies still fail.
- Added synthetic translation data to the disposable fixture, with real native
  indexes. No external translation download is needed.
- Verified expired bearer rejection and Live quota errors through the Go client;
  manual token quota rejection is a separate server fixture assertion.
- Added SSH revocation and key tests. Server tests cover Ed25519, RSA SHA-2,
  ECDSA P-256/P-384/P-521, and rejection of weak/legacy algorithms. They do not
  yet prove the CLI accepts every server-supported key type.
- Copied source, dependencies, native binaries, and selected public runtime assets
  into the temporary harness directory; local certificates, SSH keys and
  development data are excluded.
- Passed Go race tests and vet on Go 1.27.1.

## Previous increment: CLI SSH identity and subscriptions

- CLI explicit identities now accept Ed25519, RSA >=2048 bits, and ECDSA
  P-256/P-384/P-521. OpenSSH validates the public-key contents. Generated-key
  tests exercise all accepted types and rejection of weak RSA.
- Subscription decoding requires an initial acknowledgement, preserves server
  error metadata, rejects incomplete records and unexpected envelopes, and
  accepts verse records larger than the previous scanner's 64 KiB limit.
- Unexpected disconnects fail explicitly. Live closure succeeds; revoked access
  fails. Ctrl-C cancels the OpenSSH child, and all exit paths reap it.
- A new root-package integration test uses a disposable RSA identity and
  temporary known_hosts file. It receives a pushed verse, pauses the Live,
  disconnects, reconnects to a fresh paused snapshot, and receives closure
  after deletion. This proves explicit reconnect, not automatic replay.
- Go race tests, vet, build, 17 decoder fixtures, 7 encoder fixtures, 7 security
  cases, 8 PKCE/SSH cases, Go HTTP checks, and the CLI SSH test pass.

The assessment chat is active again and is changing server resilience behavior.
One snapshot failed to compile while those edits were underway; a fresh snapshot
then compiled and all integration checks passed. No server edits were made here.

Milestone 1 still awaits the server SECRET SET documentation/parser discrepancy
and duplicate stats field resolution.

## Previous increment: discovery and account results

| Milestone 2 acceptance item | Status | Evidence / remaining work |
| --- | --- | --- |
| Server info and permission-aware discovery | Complete | Typed GetServerInfo/Help; builders; CLI server help; real HTTP checks |
| Personal account and effective quotas | Complete | Typed GetAccountInfo/ListQuotas; CLI account info/quotas; numeric/count validation, zero/unlimited distinction and raw fields |
| Translation metadata and reference catalogue | Complete | Typed Go methods and CLI; real metadata/catalogue checks, source ID preservation and strict hierarchy validation |
| Typed identity, reading, Live and token metadata | In progress | Identity, readings and Live results verified; token metadata remains |
| Library PKCE exchange and credential revocation | Pending | Browser interaction stays in CLI; extract the transport methods |

Milestone 2 progress: **3 of 5 tracked acceptance items complete**. Milestone 1
remains **6 of 7**, awaiting server discrepancies.

Account notifications revealed that the server emits opaque Erlang maps with
spaces and, for larger maps, physical line breaks. The decoder now preserves
balanced compound values without evaluating them. Regression checks reject
malformed maps and newlines outside compounds. Typed results retain Raw for
future fields and original logical wire records.

The disposable PostgreSQL readiness probe now waits for TCP, avoiding the
initialization-only Unix socket which can briefly report ready before restart.
No server checkout changes were made.

## Previous increment: translation metadata and reference catalogue

| Item | Status | Evidence |
| --- | --- | --- |
| Validated INFO/CATALOG command builders | Complete | Reject empty/injected arguments; shared server decoder fixtures |
| Typed translation metadata | Complete | Names, edition, rights, attribution/trademark and unknown fields retained; real HTTP checks pass after assessment fix |
| Typed reference catalogue | Complete | Ordered source IDs, localized names, sparse chapter numbers and verse counts; strict hierarchy/count/duplicate validation |
| CLI translation info/catalog | Complete | Same builders and existing HTTP/SSH adapters; local CLI tests pass |
| Live catalogue and existing core integration | Passing | Installed synthetic data, 19 books, Psalms 1 and 23, missing translation; existing HTTP/SSH/security checks pass |
| Live metadata integration | Passing | Assessment session fixed binary encoder keys; fresh snapshot verification passes; no server files edited here |
| Documentation and shared fixtures | Complete | 21 decoder fixtures and 12 server encoder fixtures, including binary metadata keys |

Milestone 2 is now **3/5 complete**. Milestone 1 remains **6/7** pending its
server protocol discrepancies. Both metadata and catalogue now pass against a
fresh server snapshot after the assessment's encoder fix.

The original metadata failure was left visible, and the assessment was messaged
at the user's request. The harness now finishes independent HTTP/SSH/EUnit
checks before reporting client failures; failing runs do not overwrite the
passing baseline. Go race tests, vet, build, 21 request decoder fixtures,
12 response encoder fixtures, 7 security cases, 8 PKCE/SSH cases, and HTTP/CLI SSH
integration pass.



## Previous increment: typed identity, reading and Live results

| Item | Status | Evidence |
| --- | --- | --- |
| Identity model | Complete | GetIdentity preserves provider, display identity, roles/permissions and unknown fields; strict auth boolean; real HTTP check |
| Reading/search models | Complete | Single verse, chapter, whole book, search and empty search; coordinates remain optional where absent; real indexed data |
| Live state models | Complete | Typed create/get/list/settings; omitted manager fields distinct from false/zero; real HTTP checks |
| Live statistics and stack | Complete | Connection counters/details verified with a real SSH subscriber, durations/revision, stack positions; malformed counts/positions rejected |
| Live mutation results | Complete | Push/pop payloads and start/stop/pause/resume/clear/delete acknowledgements; real lifecycle checks; no retries |
| Raw API compatibility and example | Complete | Existing raw methods retained; standalone read example uses typed values |
| Token metadata | Pending | Still required by the broader milestone-2 typed-results acceptance item |

Requested slice: **3/3 domains complete** (identity, reading, Live). Milestone 2
remains **3/5 acceptance items complete**: token metadata is the remaining part
of item 4; library PKCE exchange/revocation is item 5. Milestone 1 remains 6/7.

Go race tests, vet, build, 26 request fixtures, 15 encoder fixtures and isolated
HTTP/security/PKCE/SSH integration pass. No server files were edited here.
Next at that point: typed personal token metadata, then library authentication transport.


## Previous increment: typed token metadata

| Item | Status | Evidence |
| --- | --- | --- |
| Token command builder and CLI listing | Complete | AccountTokensCommand and account tokens; existing HTTP/SSH adapters and output modes |
| Typed token metadata | Complete locally | ListTokens/TokenList/TokenMetadata; nil versus zero/false/empty, timestamps, scopes, issuer/source and raw fields |
| Decoder compatibility | Complete | Server control-character escaping; Unicode labels; count/type/ID/numeric/boolean validation and error preservation |
| Compiled source fixtures | Passing | 27 request cases and 17 response encoder cases, including empty token list and control escapes |
| Real scoped-token listing | Blocked by server | HTTP 500: permission tuples passed as character-list text; active scoped and restricted credentials expose regression |
| Existing integration | Passing | Other HTTP operations, CLI SSH subscription, 7 security cases and 8 PKCE/SSH cases |

Progress: **3/4 requested implementation/verification items complete** (command,
model, decoder; real token integration remains blocked). Milestone 2 remains
**3/5** until its typed-results acceptance item is fully verified. Milestone 1
remains **6/7** and now also tracks the token scope encoder discrepancy.

Local Go race tests, vet and build pass on Go 1.27.1. The failing scoped-list integration leaves the
last passing baseline untouched. No server files were changed here. Server
fixes stay with the assessment. On 2026-10-10, the user authorized messaging
the assessment; the confirmed scoped-token encoder failure and suggested CSV
fix were sent to chat 01a12273-bf8b-7e30-826c-6354f7307511. Next: verify the assessment fix, then extract library PKCE
exchange and revocation.


## Previous increment: authorization transport and token verification

| Item | Status | Evidence |
| --- | --- | --- |
| Token metadata server compatibility | Complete | Assessment's scope CSV fix; real manual/scoped listings, permission denial, escaped labels and expired/revoked filtering |
| Library code exchange | Complete | NewAuthClient/ExchangeCode, typed result and retained JSON fields; bounded form/response, cancellation and redirect rejection |
| Library revocation | Complete | RevokeCredential; HTTP 204 acknowledgement, status/code/metadata errors and no retries |
| CLI delegation and failed logout | Complete | Browser exchange and logout use library methods; failed revocation preserves memory/disk profile; 401 clears unusable profile |
| Real authentication lifecycle | Passing | Exchange/use, code reuse, bad verifier consumption, revoke/use-after-revoke and repeat-revoke denial |
| Full integration gate | Passing | Fresh snapshot: all HTTP, CLI SSH, 7 security cases, 8 PKCE/SSH cases; baseline updated after assessment fixed normal closure |

Requested increment: **5/5 implementation/verification items complete**.
Milestone 2 has **5/5 tracked acceptance items complete** and its full
integration checks pass. Milestone 1 is back to **6/7**, awaiting the existing
SECRET SET documentation/parser mismatch and duplicate stats field resolution.

The user authorized ongoing coordination of integration regressions with the
assessment. The assessment fixed both scope encoding and normal SSH closure;
a fresh full run passes and updates server-baseline.json. Go 1.27.1 race tests,
vet and build pass; 27 request and 18 response encodings verify against compiled
source. No server files were changed here. Next: endpoint-bound named CLI
profiles and stable output; retain the unresolved protocol safeguards.


## Previous increment: endpoint-bound profiles and stable output

| Requested item | Status | Evidence |
| --- | --- | --- |
| Profile storage and migration | Complete | Version-1 named profiles, explicit transport/endpoint, active selection, atomic 0600 writes; legacy credentials bind to binary defaults and migrate on write |
| Selection, authentication and automation | Complete | Profile add/use/list/show/remove; --profile/BIBLEIT_PROFILE precedence; HTTP-only BIBLEIT_TOKEN is never saved; logout retains binding and other profiles; no HTTP-to-SSH fallback |
| Stable finite JSON | Complete | Schema version 1, typed record fields, absent versus empty distinction, unknown string retention, structured stderr errors and exit codes; profile listings omit tokens |
| Subscription NDJSON | Complete | Per-record envelopes, event names, native base64-JSON verse payload, retained raw record, unexpected EOF/revocation handling and cancellation on output failure |
| Verification and documentation | Complete | Race/vet/build; full isolated HTTP/SSH/security integration; matching HTTP/SSH JSON; six cross-compilation targets; output contract and profile examples |

Requested increment: **5/5 complete**. Milestone 2 remains **5/5** and milestone 1
remains **6/7**. Milestone 3 is **3/5 tracked acceptance items complete**:

| Milestone 3 item | Status | Remaining work |
| --- | --- | --- |
| Named endpoint-bound profiles and explicit transport | Complete | Implemented and real-server routing verified |
| Browser/SSH authentication and automation credentials | Complete for current server surface | Library transport, cached revocation semantics, scopes/expiry/quota and SSH validation checked |
| Stable table/raw/JSON/NDJSON output and errors | Complete for documented version 1 | See output.md; JSON is typed protocol records, not Go domain struct serialization |
| Release UX, shell completion and destructive-command behavior | In progress | Add completion and finalize destructive-command UX; subscriptions already support Ctrl-C and explicit reconnect |
| Native supported-platform validation | Complete for representative gate | Hosted Windows/Ubuntu/macOS pass; local macOS/arm64 and Linux/arm64 pass; Fish runtime passes; native Windows/arm64 is outside the matrix |

The expanded harness tests CLI JSON on real HTTP account/discovery/translation/
reading/Live replies, verifies HTTP/SSH Ping envelopes match, and renders actual
SSH verse/pause/reconnect/closure events through NDJSON. All existing 7 security
and 8 PKCE/SSH cases pass; the successful baseline is updated. No server files
were edited here. The existing SECRET SET/stats safeguards remain.

Next: shell completion and destructive-command UX, then native platform CI and
publication/package-path decisions. Workspace APIs still await a public server
contract; no organization/browser endpoint wrappers were introduced.


## Previous increment: shell completion and destructive-command UX

| Item | Status | Evidence |
| --- | --- | --- |
| Command/flag/profile completion | Complete | Shared local candidate helper; values, nested Live verbs, local names, terminator handling; no server requests or profile writes |
| Bash, Zsh and Fish script generation | Complete with runtime caveat | Bash/Zsh syntax and actual executable candidate forwarding pass; Zsh source/autoload supported; Fish is unavailable locally and its runtime check remains pending |
| Scoped destructive-command confirmation | Complete | Live delete/all, stack clear/pop, secret rotate/delete; endpoint/profile/target prompt; bulk deletion requires `delete all`; cancellation/EOF sends no request |
| Automation and stable errors | Complete | Strict boolean --yes; JSON/noninteractive confirmation_required exit 2; cancellation exit 1; dispatch exactly once, invalid commands rejected, credentials never displayed |
| Verification and documentation | Complete for available local tools | Go 1.27.1 race/vet/build, native macOS PTY cancellation, six CGO-free cross-builds; README/help/output contract updated |

Requested implementation increment: **5/5 complete**, with **Fish runtime
verification pending** because Fish is not installed locally. Milestone 3 is now
**4/5 acceptance items complete**; native Linux/Windows validation remains its
platform gate. Cross-compilation is not native platform validation. Milestones
1 and 2 remain **6/7** and **5/5**, respectively.

The destructive guard runs after command validation and before the shared
HTTP/SSH request path. Request-boundary tests prove missing confirmation sends
zero requests and --yes dispatches once. Existing client/server transport code
and server files are unchanged; this increment uses local tests and shell/PTY
checks and does not claim a new full server-integration run.

Next: native platform CI (including Fish runtime validation), then independent
publication and package-path decisions.


## Previous increment: native validation and Fish runtime

| Item | Status | Evidence / remaining work |
| --- | --- | --- |
| Native CI and reproducible process harness | Complete locally | Ubuntu/Windows/macOS matrix; exact Go 1.27.1; race/vet/build; profile replacement, spaced paths, loopback HTTP, JSON and destructive guards; workflow YAML parsed |
| Native Linux validation | Passing on arm64 | Disposable official Go Linux container on Docker's Linux/arm64 VM; race/vet/build and actual CLI/PTY checks pass (PTY, not emulation of Windows) |
| Fish runtime validation | Passing | Fish 3.6.0 syntax and nine actual complete -C cases; no HTTP fixture requests or profile writes; Bash/Zsh same cases pass |
| Windows-specific correctness and tests | Implemented, cross-compiled | User-only config ACL before writing credentials; actual saved DACL checks before/after replacement; real console-handle/CRLF test; Windows test binary compiles |
| Native Windows execution and hosted CI evidence | Pending environment | No Windows host; GitHub CLI unauthenticated; local checkout uncommitted. Ready workflow must be published/run before this item can pass |

Progress: **4/5 requested tracked items delivered; native Windows execution is
pending**. Milestone 3 stays **4/5 acceptance items complete** until the native
platform gate passes. Milestones 1/2 stay **6/7** and **5/5**. The user was asked
which Windows execution environment to use; publication has not been assumed.

The successful Linux log is `/tmp/bibleit-linux-native.log`. macOS native process,
HTTP-fixture, PTY and shell checks also pass. Native process smoke fixtures are
separate from real-server integration; no fresh full server suite is claimed by
this increment. See `native-validation.md` for exact coverage and reproduction.
No server files were edited. The disposable Linux container is removed after
execution, and its read-only source snapshot excludes Git/config/developer data.


## Publication preparation

The user selected GitHub Actions after publishing this checkout. The initial
commit is prepared on `native-validation`; the native workflow will
run on push. GitHub SSH returned `Repository not found` for the configured
`mittel-labs/bibleit-cli` remote, and GitHub CLI is unauthenticated. This does not
establish whether the repository is absent or inaccessible to the current SSH
identity. The user selected a different repository; its URL and authenticated access are
required before publication and native Windows execution.


The replacement destination is `mittel-labs/bibleit`. Its existing main branch
contains the Python TUI, website and native-library submodule. The publication
branch preserves those files and merges their history with the prepared CLI;
only README/gitignore overlap, and both projects' contents are retained. The Go
module/import path is now `github.com/mittel-labs/bibleit`. Publishing a feature
branch triggers native CI and does not merge into main. SSH remote access works;
GitHub CLI authentication remains required for inspecting private Actions runs.


The branch is published as `native-validation`; the user requested no codex
prefix, and the earlier prefixed remote reference was removed after the new
branch push succeeded. Hosted Ubuntu (including Fish) and macOS checks passed.
The first Windows run exposed a test assumption about icacls ACL-file encoding;
the reader now accepts UTF-16LE without a BOM and UTF-8. Actual user-only DACL
verification remains strict. A native Windows rerun is required before closing
the platform gate. Public status/annotations are readable without CLI login;
failed Go test output is now included in safely escaped CI annotations.


The Windows rerun revealed explicit SYSTEM/Administrators grants surviving
inheritance removal. Config saving now replaces the whole file DACL through
Windows security APIs with a protected current-user-only grant. The ACL test
also recognizes equivalent well-known SID aliases emitted by icacls; it still
requires exactly one grant. Native rerun remains required.


## Latest verification: published native platform gate

| Item | Status | Evidence |
| --- | --- | --- |
| Publication staging | Complete | `mittel-labs/bibleit` branch `native-validation`; existing Python/TUI/native-library/website content and main preserved; earlier codex-prefixed branch removed at user's request |
| Hosted native Linux/Ubuntu | Pass | Go 1.27.1 race/vet/build, process/loopback fixture, PTY confirmation; Bash/Zsh/Fish completion |
| Hosted native Windows | Pass | Go 1.27.1 race/vet/build and native process fixture; real config DACL before/after replacement; console handles and CRLF |
| Hosted native macOS | Pass | Go 1.27.1 race/vet/build, process fixture and PTY confirmation |
| Verification records | Complete | Native validation documentation and plan updated; exact verified code SHA and run linked below |

Requested native validation: **5/5 complete**. Milestone 3: **5/5 tracked
acceptance items complete**. Milestones 1/2 remain **6/7** and **5/5**.
[All native jobs passed](https://github.com/mittel-labs/bibleit/actions/runs/38081432573)
at code commit `08e4a31d0ff468b07e01408788f8b94a42995dc0`.

Windows validation first exposed ACL fixture encoding assumptions, then actual
explicit SYSTEM/Administrators grants left by inheritance removal. The final
implementation replaces the entire config DACL through Windows security APIs
before writing credentials, and protects it from inheritance. Its test requires
exactly one current-user grant, including equivalent SDDL SID aliases. The
corrected native race suite and subsequent process checks pass.

The staging branch is not merged. After reviewing the existing Bibleit repository,
the recommendation is a separate repository for server clients and CLI, with
shared contracts and independent package releases. That migration/repository
creation has not been authorized or performed; the current feature branch keeps
validated work available for either destination. Publication of packages remains
milestone 6, distinct from publishing this validation branch.


## Repository separation (2026-10-10)

The user's selected layout supersedes earlier monorepo recommendations:
`bibleit-labs/bibleit-cli` contains the CLI; `bibleit-labs/bibleit-go` contains
the independent Go client. Python, Ruby and Rust will have separate repositories.
Both extraction branches are named `repository-split`, with no codex prefix.
The prior `mittel-labs/bibleit` staging branch and its existing project remain
recoverable, and its remote main is unchanged.

| Item | Status | Evidence |
| --- | --- | --- |
| Separate paths and dependency direction | Complete | CLI imports Go client; Go client has no CLI dependency |
| Independent Go client extraction | Complete | Own module, fixtures, example and native CI; local race/vet/build pass |
| Pin published client and wire CLI harness | Complete | Pinned `v0.0.0-20261010200950-9a52be34b0c3`; no replacement; both integration test binaries compile; fixture snapshots match |
| CLI extraction and local validation | Complete | Race/vet/build; real process/PTY and Bash/Zsh completion checks pass on local macOS/arm64 |
| Publish and verify both hosted matrices | Complete | Go client and CLI each passed Windows, Ubuntu and macOS; CLI Ubuntu also passed Fish/Bash/Zsh |

Progress: **5/5 repository-separation items complete**. Historical native results
above apply to the original implementation; new repository workflows must pass
independently. No package release tag or binary release has been published.

[Go client native matrix passed](https://github.com/bibleit-labs/bibleit-go/actions/runs/38082660896)
on the extracted client commit `9a52be34b0c3`. The CLI pins that commit using
a Go pseudo-version until a reviewed semantic-version release is published.
The new split integration harness compiles both binaries; the real-server
integration results above remain historical and were not rerun for this split.

[CLI native matrix passed](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38082843195)
on code commit `e46bfd6`. Both new repositories have `repository-split` as their
initial default branch. No branch uses the codex prefix.

All CLI and language-client repositories are public open source projects under
`bibleit-labs`. Use Conventional Commits for subsequent changes and maintain
contribution documentation, the existing MIT license, and independent release
notes. `bibleit-server` remains separate and is not moved or made public here.


## Real-provider follow-up — 11 October 2026

Assessment ledger a40d27d records actual human email delivery/verification and
fresh Turnstile acceptance. The original helper's replayfalse400/400 output is
inconclusive; no backend bypass is established. An improved browser helper checks
rendered localized feedback and CSRF/rate/continuity signals, with six offline
classification cases passing. A fresh human retry is requested in the assessment
chat; replay remains open until that real result is verified.

The later font-only source14628f3/native d6e31dc deployment is separately recorded:
workflow38157279015, registry digest28fb9a0b and validated image ID392fc186.
Published and running digest match. Its report verifies406 hashes/eight zero-exit
stages/native amd64/cleanup[0,0];185 APK comparisons have zero fixed alerts. Its
owner confirms CSS/signup/readiness. Earlier full client and boundary checks
remain attributed to4243; no silent inheritance or unnecessary protocol rerun.

`main` is now default in all three repos. The CLI's stale legacy origin/main ref
was not part of the recreated private repo; its unpublished merge was removed,
and reviewed CLI history was published directly. Older master refs remain
preserved. Main CI source bd66265 passes all seven native/candidate jobs in
[38160808491](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38160808491) and
[38160808600](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38160808600).


The improved real-token retry now confirms replay acceptance twice, rather than
only inconclusive response classification. Human-denial=false, invalid-email=true
on both HTTP400 responses; CSRF/rate/continuity checks rule out those alternatives.
The assessment is investigating deployed credentials/request handling and owns
the repair. Hold release publication, qualify the changed source/image as needed,
and rerun genuine fresh-token/replay acceptance after repair. Email acceptance
remains recorded separately; no raw challenge material is retained.
