# REST API migration

The CLI uses the pinned `https://api.bibleit.app/v1` service. The canonical
contract is maintained by bibleit-server and copied into bibleit-go's
`rest/openapi.json`; [the hosted reference](https://docs.bibleit.dev) describes
resource schemas and required scopes.

## Existing commands

HTTP covers identity, account and organization metadata, quotas, personal token
metadata, translation libraries and catalogues, book/chapter/verse reads, search,
and named or unnamed Live creation. Live resources include statistics, settings,
stack push/pop/clear, audience secrets, playback, display clear, individual and
bulk deletion, and event subscriptions through REST long polling.

Book names resolve through the canonical dictionary, including aliases and
localized names. SSH book names follow the selected edition's language pack;
canonical numeric book IDs work across transports. Collections follow every `next_cursor`. Whole-book reads
also follow bounded pages while preserving complete CLI output. Explicit book
`--limit` and/or `--cursor` requests preserve one page and its native cursor.
At the API, every whole-book request returns a bounded page (50 verses by
default, maximum 500) and `next_cursor`; use the SDK iterators or follow the
cursor to read the complete book. Search preserves the
server's bounded `has_more` flag. Conditional Live mutations read an ETag once
and pass it as `If-Match`; 412 and 428 map to conflict exit 6. Mutations are never
automatically retried. HTTP never uses the deprecated command API.

Personal organization membership and application status require a requesting
user; service identity is not a membership directory. Delegated credentials
cannot access metadata requiring unconsented `server.get`/`token.get`, or edit
personal libraries without a consented write scope. Help explains these
credential restrictions. Personal library edits use the personal translation
permission and quota; service edits use their organization permission.
Translation catalogues return installed chapter/verse coverage. Ping checks
REST OpenAPI availability; server info shows endpoint and transport metadata.

SSH and browser login/code exchange/logout keep their existing supported endpoints.

Use a finite personal `bt_` token or delegated `oa_` token for personal resources,
and a `svc_` service token for organization resources. Standard commands discover
the workspace through `/me` or `/organization`; delegated tokens consequently
need `identity.read`, and service tokens need `organization.get`, in addition
to the operation's own scopes. Infinite legacy personal tokens must be replaced
with finite credentials to use REST.

## CLI usage

The existing commands select the HTTP resources internally:

```sh
bibleit auth info
bibleit read web John 3 16
bibleit search web "love"
bibleit live list
bibleit live create "Sunday"
bibleit live "$LIVE_ID" pause
```

`BIBLEIT_TOKEN` overrides the cached token without saving it. HTTP raw output is
native REST JSON. HTTP JSON output uses `{schema_version:2,ok:true,data:...}` with
native JSON types, resource field names and nested metadata, including rights.
SSH JSON retains schema version 1. HTTP subscription JSON is NDJSON with a
schema-version-2 envelope around each native event. Stack pop needs no
confirmation; clear, delete and secret rotation remain guarded. SSH commands
continue through OpenSSH to `bibleit.app:22`, including Live subscriptions.

## Coordinated release

These CLI changes use the published bibleit-go v0.2.0 `rest` package, generated
with pinned `oapi-codegen` v2.8.0 through `go generate ./rest`.
The Python REST generator has been removed. Small Go templates preserve the
SDK client API and its endpoint, redirect and retry policies. HTTP dispatch accepts
CLI arguments directly and renders structured JSON; it has no command adapter.
During development use the sibling checkouts in the existing local Go workspace:

```sh
go test ./...
go test ../bibleit-go/...
```

The CLI pins the published bibleit-go v0.2.0 module without a local replacement.
SDK v0.2.0 is published; CLI binary publication remains a separate step.
Standalone validation uses:

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

Historical v0.1.0 real-server validation is recorded in [release evidence](releases/v0.1.0-integration.json).
Do not add sibling-path replacement directives or workspace files to a release.

The expanded command surface is qualified against the published SDK without a
workspace or replacement. Current integration results do not alter historical
v0.1.0 evidence. Homebrew and existing v0.1.0 published assets stay unchanged.
