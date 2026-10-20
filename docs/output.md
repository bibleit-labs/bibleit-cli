# CLI output contract

HTTP resource commands return native REST data with `--format raw`. Their
`--format json` output uses schema version 2:

```json
{"schema_version":2,"ok":true,"data":{"id":"live-id","paused":false,"sequence":3}}
```

Field names and native JSON types come from the REST resource, including nested
rights and attribution. Paged library/token/Live collections aggregate all pages and return `data` with
`next_cursor:null`. Personal organization directory results preserve one page and
its opaque `next_cursor`; request the next page explicitly with `--cursor`. No command-protocol records are generated. Local commands,
errors and SSH keep the schema-version-1 contract described below.

HTTP whole-book reads fetch bounded pages internally and combine all verses,
preserving the complete-book response shape without `next_cursor`. If any page
fails, stdout remains empty: a partial traversal is never reported as a complete
book. Use `read <translation> <book> --limit 1..500` and/or `--cursor <cursor>`
to request exactly one page. Explicit page JSON/raw output includes
`next_cursor` (an opaque string, or `null` on the final page). Human output shows
the continuation cursor when another page exists. These flags require HTTP and
a whole-book reference; chapter and verse reads retain their existing behavior.

## SSH and local commands, schema version 1

Finite `--format json` commands write one object to stdout:

```json
{
  "schema_version": 1,
  "ok": true,
  "records": [
    {"type": "ok", "fields": {"pong": true}}
  ]
}
```

Records retain server order and lowercase record types. END is framing and is
excluded. Fields use snake_case wire names. Object key order is not contractual.
Missing fields remain absent; an explicit empty permission/translation list is
`[]`. Unknown fields stay strings, even if their contents look like numbers.
IDs, names, statuses, source names, references and verse texts remain strings.
This record envelope is used for SSH and local commands.

Known field types:

| Type | Fields |
| --- | --- |
| Boolean | paused, owned, protected, retiring, pong; auth in OK/local auth records |
| Non-negative integer | protocol_version, count, verses, chapters, results, books, commands, connections, actor_connections, anonymous_connections, running_for_seconds, revision, stack_entries, position, issued_at, created_at, last_used_at, expires_at, connected_at, connected_for_seconds, plan_activated_at, lives, tokens, keys, used, book, chapter, verse |
| String array | roles, permissions, capabilities, scopes; translations except the count cases below |
| Integer or string | quota limit is an integer, or the literal string `unlimited` |
| Context-specific translations | Account summary (`lives` present) and verse-push acknowledgement (`event=verse`) use an integer count; other translations fields are string arrays |

`auth` in COMMAND help records remains the string authority category. Opaque
account notification terms remain strings. Numeric/boolean validation failures
produce an error rather than silently converting a value to zero or false.
These are typed protocol records; they are not a serialization of the Go client's
domain structs. Future output schema changes must use a new schema_version.

Named profile commands and PROFILE output records are removed before release.
Local version/help/auth actions use version/help/auth record types. Interactive
browser progress and fallback browser URLs are written to stderr.

## Subscription NDJSON

For SSH `live <id> subscribe`, `--format ndjson` uses the schema below. Every
acknowledgement/event is one compact JSON object followed by a newline:

```json
{"schema_version":1,"record":{"type":"ok","raw":"OK id=live paused=false","fields":{"id":"live","paused":false}}}
{"schema_version":1,"record":{"type":"event","event":"closed","raw":"EVENT closed","fields":{}}}
```

The initial OK retains the server snapshot. Positional event names are exposed
as `event`; the keyed Live update becomes `event: "live"`. Verse events contain
`payload`, the native JSON decoded from the server's base64 payload. Every stream
record retains the original wire record in `raw`, including future positional
content. Invalid base64/JSON payloads fail explicitly. NDJSON keeps verse text
with embedded newlines inside JSON strings, not additional output lines.

Normal closed events end successfully. Revoked events are emitted before an
authentication error/exit 3. Unexpected disconnects fail. Reconnection remains
explicit and receives a fresh snapshot; NDJSON does not promise event replay.

## Errors and raw/human output

For json/ndjson, errors produce one JSON object on stderr and a nonzero exit:

```json
{"schema_version":1,"ok":false,"exit_code":7,"error":{"code":"rate_limited","message":"rate_limited","fields":{"retry_after_ms":"250"}}}
```

Error fields contain allowlisted safe retry metadata only. HTTP errors
also include http_status. Generic codes match the documented exit category;
server error codes remain intact. Failed finite commands write no success object.
A stream can emit records before a later error. Local help/version bypass config
loading, so they still work when configuration is unreadable.

SSH `raw` prints the decoded native protocol records. The SDK validates reply
framing and consumes the terminating `END`; that marker is not printed. Local
auth/version/help actions use human output. Default Live subscriptions summarize state changes and show
decoded verse references, translations and text. Sequence-only state updates are
quiet. Default finite tables use stable presentation columns, escape control
characters, and keep verse text readable. Malformed output after a mutation does
not imply rollback; the CLI never retries mutations. Existing prerelease JSON
string fields are replaced by this documented, versioned typed contract.

## Confirmation errors and completion scripts

Guarded destructive commands require `--yes` with JSON output or when stdin or
stderr is not a terminal. No request is sent if confirmation is missing: stderr
contains the existing schema-version-1 error envelope with
`error.code = "confirmation_required"` and `exit_code = 2`; stdout is empty.
Interactive rejection/EOF uses `cancelled`, exit 1, and sends no request.
Prompts appear only on stderr in table/raw mode. `--yes` does not change the
successful command response or bypass server permissions.

`completion bash|zsh|fish` generates a shell script in table/raw mode. JSON and
NDJSON reject script generation; `help completion --format json` remains a
normal help record. The hidden `__complete` shell adapter emits newline-separated
candidates, ignores network credentials and never contacts the server.

## Updated human UX and HTTP watching

Human Scripture rows share translation/reference/text columns; retained stacks
add position. Numeric-only legacy SSH records retain a faithful fallback when
localized metadata is unavailable; the client never infers missing verse numbers
or parses coordinates out of arbitrary Scripture text. Terminal text wraps without
truncation using Unicode display widths; native JSON/raw data are unchanged.
Human receipts and empty states describe the action/resource; they do not echo
search queries. Table labels use user terminology without renaming wire fields.

HTTP Live watching performs bounded long polls and keeps the returned cursor
opaque. `--format ndjson` emits one compact object per response:

```json
{"schema_version":2,"event":{"schema_version":1,"type":"stream.heartbeat","resource_type":"live","resource_id":"L1","sequence":1,"cursor":"opaque","data":null}}
```

The inner schema belongs to the REST event contract; it is independent of CLI
schema 2. HTTP raw emits native event objects separated by newlines. Human output
suppresses heartbeats. Finite JSON mode rejects a subscription and recommends
NDJSON. Polls use a 40-second minimum client timeout to accommodate the 25-second
wait; cancellation stops the loop. Errors stop watching without retries, cursor
reset, replay or credential/transport fallback.

Errors retain schema 1 and exit categories. Safe descriptions replace bare codes;
only allowlisted numeric retry metadata is forwarded. Arbitrary server fields,
Problem Details titles/URLs, raw SSH stderr and credential values are not copied
into diagnostics. REST 412/428 require inspection and review before retrying.

With `--secret-output <new-private-file>` on secret create/rotate, the secret is
written once to a protected file and its output field becomes
`written_to_private_file`, including raw/JSON. This explicit delivery option
intentionally redacts that field while preserving its type and surrounding schema.
Without the option, the authorized secret-generation result is the deliberate
one-time delivery; avoid redirecting it into ordinary logs. Existing files are
never overwritten; failed operations remove their empty reserved destination.

Pop never prompts. Delete/clear and secret rotation/removal retain confirmation;
routine prompts identify scope/consequence without transport or endpoint noise.
