# CLI consistency and UX contract

The CLI applies one grammar, help and presentation policy across its commands.
This document describes the implementation and qualification gates; historical
published assets keep the behavior documented at their source revision.

## Command grammar and compatibility

The command registry owns syntax, examples, aliases, offline help and static
completion. Validate grammar before requesting credentials or connecting: a
missing Live ID, unknown option or invalid stack-pop count returns usage guidance.
The registry enumerates optional book/chapter/verse forms and compact references.
SSH names follow the selected edition's language pack; numeric canonical book IDs
are portable. HTTP resolves dictionary aliases before sending canonical IDs.

Use user terminology in human output and help. Preserve machine wire fields
needed by existing clients. Keep documented aliases, including `live <id> clear`
and the singular account quota/token aliases, with a canonical spelling in help.
Standalone audience-secret authorization cannot persist across SSH connections;
explain that limit instead of accepting ineffective syntax.

## Human and machine output

Scripture uses shared translation/reference/text rows, with position on stacks.
Prefer explicit localized references and source metadata. Preserve legacy text
when that metadata is unavailable; never guess coordinates from display text.
Wrap full text by display-cell width and preserve Unicode and multiline content.
Receipts report actual affected counts and empty states describe the result.

HTTP JSON uses a schema-version-2 envelope around native REST data. SSH/local
JSON and errors retain schema version 1. HTTP raw output remains native REST JSON.
These transport differences are intentional compatibility contracts. Keep
rights/attribution metadata and native numeric/null values in machine output.

## Errors and confirmations

Explain authentication, permissions/scopes, quotas, conflicts, missing resources
and service availability. Expose only safe error metadata; never print credentials,
one-time secrets, arbitrary response bodies, raw SSH diagnostics or pasted invalid
preference values as error details. Usage mistakes should identify the corrected
syntax rather than returning a generic resource-not-found code.

Stack pop needs no confirmation. Clear, delete, removal and secret rotation remain
guarded; bulk deletion names the whole selected workspace. Prompts describe the
action and target. Routine endpoint details belong in server info. Noninteractive
commands require the documented explicit confirmation option where applicable.

## Shared preferences and private delivery

`~/.bibleit/config.yaml` stores versioned nonsecret transport, identity and format
preferences. Explicit flags override environment values, then YAML, then defaults.
Keep credentials in their existing protected, service-bound file. Reject unknown
YAML keys/types and malformed documents. An explicitly cleared identity must not
resurrect an older credential-file identity.

`--secret-output <new-private-file>` delivers a server-issued secret once, using
exclusive creation and platform-appropriate private permissions. Never overwrite
a destination or echo the delivered value to normal output. If no destination is
selected, successful create/rotate output is the deliberate one-time delivery.
Protected credential files remain plaintext; OS credential-vault integration is
separate future work, not a claimed property of this implementation.

## REST, SSH and credentials

Use typed resource operations directly. Never fall back to command HTTP, silently
switch transport/account, retry a mutation, or treat an unsupported operation as
a successful parity test. Workspace discovery, ETags and current scopes remain
explicit. Service organization identity is not a personal membership directory.
Delegated credentials cannot gain unconsented metadata or library-write scopes.

The resource surface includes account metadata, translation libraries/catalogues,
book/chapter/verse reads, search, Live lifecycle/settings/statistics/stack/secrets,
workspace-scoped bulk deletion and subscriptions. Server help over HTTP combines
local syntax with contract discovery; it does not claim effective credential grants.

Whole-book REST reads always return bounded pages: default 50, maximum 500
verses, with `next_cursor` on every response and no unbounded escape. The SDK
offers both single-page methods and lazy bounded page iterators. The
CLI follows pages for a normal whole-book read; explicit HTTP book `--limit`
and/or `--cursor` returns one page. Pages retain rights and stable source verse
coordinates. Reject cursor loops, duplicate/out-of-order coordinates and mixed
editions; never emit a partial book as a successful complete read. Cursors are
opaque continuations, not authorization grants, Live cursors or ETags.

HTTP subscriptions use bounded long polling, opaque cursors and snapshots plus
heartbeats; SSH subscriptions retain connection events. Neither implies replay.
HTTP NDJSON wraps native events in schema version 2; raw uses native event objects.
Ctrl-C cancels outstanding work, and immediate heartbeat replies must not hot-poll.
Finite calls retain bounded deadlines; long polls use a deadline above server wait.

## Verification and release gates

The inventory separates command binding, fixtures, real server-domain assertions,
native processes and deployed read-only checks. A binding or a rejected command
is not evidence of successful domain behavior. Explicit credential exclusions
are reported as inapplicable, never counted as passed operations.

The command matrix contains 216 transport/credential cases: 204 applicable cases
and 12 credential exclusions. Fifteen local/auth variants, four subscription
credential cells and one compatibility-only connection-secret form have separate
proof. Validate local help/config/completion, browser authorization and logout,
registered-key login, Scripture source metadata, retained-stack effects, actual
settings, secret protection and cleanup. Test reads and mutations against isolated
synthetic data. Production verification originally covered read-only commands.
The user subsequently authorized production mutations on their account, including
deletion of its existing Live to free the single available slot. Run this expanded
campaign with temporary test Lives, explicit ownership checks, request pacing and
cleanup. Collection deletion requires an initially empty scope containing only
Lives created by that run. Preserve the original credential and translation library;
test browser login/logout with a separately issued, subsequently revoked credential.

Pagination checks compare every source verse from paged traversal with reference
chapter reads for personal, delegated and service credentials. Cover default and explicit page sizes,
continuation with a different limit, invalid/mismatched/stale cursors,
cancellation, and unchanged chapter/verse behavior. Production checks traverse a short
book and verify safe cursor errors without changing content or recording cursors.

[Machine-readable validation](releases/unreleased-cli-ux-validation.json) records
the qualified snapshots, counts and remaining gates. Native CI must cover Linux,
macOS and Windows; installation checks cover supported architectures, shells,
private files, replacement and package-manager lifecycle. API and client versions
are independent: publish a new immutable SDK, pin the CLI dependency and rerun
standalone checks without a workspace/replacement before a release candidate.

Keep the installed Homebrew formula and existing v0.1.0 assets unchanged. Candidate
validation, SDK publication and a later CLI/Homebrew release are separate stages.
