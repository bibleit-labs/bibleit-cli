# Public CI and repository settings preparation

> Launch update, 2026-10-12: the owner approved actions 1–6 and protection at
> public cutover. The SDK, CLI and tap are now public with protected main/release
> tags. See [actual launch evidence](public-launch.md) for executed actions and
> verified installation coverage. The preparation text below is a historical
> snapshot. bibleit-server remains private; its docs PR is not merged/deployed.


Prepared 2026-10-11. All repositories remain private. Workflow changes are on
private review branches. On 2026-10-12 SDK/CLI Actions PR creation/approval
capability was disabled, with read-only defaults retained. Tap permissions remain
unchanged. No rulesets, visibility or secrets were changed.

## Contributor and release workflows

CLI native/candidate workflows keep GOWORK=off. Private repositories use the
existing narrowly scoped SDK read secret and GOPRIVATE. Public repositories skip
that authentication step and use ordinary go mod download/verify without the
private secret; reusable workflows no longer require it. Private authentication
and module downloads share one step so the credential environment cannot leak
into an unauthenticated later helper invocation.

After cutover, public fork pull requests use read-only native tests. No pull_request_target or
workflow_run trigger executes contributor code. Release writes remain confined
to the draft job after native validation, with the existing main ancestry guard.
Tap release reads select the private cross-repository token only while private;
public mode uses the read-scoped GitHub token. The manually dispatched updater
retains repository write permissions solely to open reviewed formula PRs, without
auto-merge. It still requires non-draft, notice-complete release packages.

All external Actions references use full commit SHAs resolved from the upstream
versions already in use. Weekly GitHub Actions Dependabot configuration proposes
updates for review; no automatic update merge is enabled. Read-only checkout
steps disable credential persistence; the formula updater keeps checkout write
credentials for its scoped git push.

Official guidance: [GitHub Actions secure use](https://docs.github.com/en/actions/reference/security/secure-use).

## Repository settings proposal

The [reviewable JSON proposal](security/public-repository-settings-proposal-2026-10-11.json)
contains main-branch and v* tag rules for all three repositories. Main requires a
PR, current native/style checks, resolved discussions and linear history, and
blocks deletion/force pushes. Required checks bind to the verified GitHub Actions
application (15368). It requires zero approving reviews for the current
single-maintainer setup: a sole author cannot approve their own PR. Increase that
count when an independent reviewer is available. Tag rules block update/deletion;
new tags remain possible only through separately approved release preparation.
No bypass actor is proposed. Validate live check contexts and the final merge
method before enabling the rules, especially around the tap updater.

Read-only inspection found default workflow permissions read and PR creation/
approval capability enabled in all three repos. That capability was disabled for SDK and CLI on 2026-10-12 and read back from
the API. The tap needs it enabled to create formula PRs; its
workflow never approves a review. Configure available secret scanning, push
protection, dependency alerts and private vulnerability reporting at launch,
and verify each feature rather than inferring it from an omitted API field.

Ruleset inspection returned HTTP 403 for all three private repos: GitHub requires
an upgraded plan or public visibility for this feature. It did not prove whether
protections exist. Settings configuration remains gated on feature availability
and launch approval. [Ruleset availability](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets).

## Validation and remaining gates

Actionlint v1.7.12 passed all prepared workflows (embedded shell/Python lint tools
were not installed). The public dependency path cannot pass an empty-cache
anonymous fetch while the SDK remains private. Prove it after approved SDK
publication, using default proxy/sumdb and no Git/private-token credentials.
The CLI/tap public switch must follow SDK/CLI access in the approved launch order.

Retain private tokens until their replacement paths are verified, then revoke
obsolete tokens deliberately. Settings payloads are proposals, not applied state.
Native installation evidence is recorded in [release readiness](release-v0.0.2-readiness.md);
package validation does not establish public download access or a waived release
approval. Windows PowerShell 7/5.1 and Scoop checks passed, along with tap-owned native
macOS/Linux installation checks on the exact final candidate. Native SDK and CLI
code tests passed on all three hosted platforms. Exact run/source references
are recorded in release readiness.

## Immediate main protection request — 2026-10-12

The owner requested protection for all three main branches. Admin access and
private visibility were confirmed. Classic branch protection PUT requests were
attempted with PR requirements, strict GitHub Actions checks, admin enforcement,
linear history, resolved discussions and force-push/deletion blocking. All three
were rejected by GitHub with HTTP 403 requiring an eligible plan or public access.
No protection was applied; main remains unprotected.

[Exact payloads and results](security/main-branch-protection-2026-10-12.json) are
ready to apply without another main-protection approval once eligibility changes.
For organization-owned private repos, GitHub Team or Enterprise provides this
feature; personal GitHub Pro does not upgrade the organization. No plan purchase
or visibility change was made. [GitHub protection availability](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).
