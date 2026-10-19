# REST API migration

The CLI uses the pinned `https://api.bibleit.app/v1` service. The canonical
contract is maintained by bibleit-server and copied into bibleit-go's
`rest/openapi.json`; [the hosted reference](https://docs.bibleit.dev) describes
resource schemas and required scopes.

## Existing commands

HTTP identity, account info, translation lists/metadata, chapter
and verse reads, search, named Live creation/list/info, lifecycle/playback, display clear,
name edits and individual deletion now use REST. Book names resolve through the
canonical book dictionary, including aliases and localized names. Collections
follow every `next_cursor`. Search preserves the server's bounded `has_more`
flag. Conditional Live mutations read an ETag once and pass it as `If-Match`;
412 and 428 map to conflict exit 6. Mutations are never automatically retried.

HTTP never uses the deprecated command API and does not translate command strings.
Commands without a supported REST resource fail locally, without a request.
Quotas, personal token metadata/library editing, organization applications,
unnamed Live creation, book-wide reads, Live statistics, stack commands, Live
secrets, bulk deletion and reference/translation settings remain SSH-only.
Translation catalogues over HTTP expose reviewed book coverage; SSH retains its
chapter/verse catalogue. Ping checks REST OpenAPI availability; server info shows
API metadata. Service tokens can list their own organization and service tokens.

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
SSH JSON retains schema version 1. Destructive confirmations remain in place. SSH commands
continue through OpenSSH to `bibleit.app:22`, including Live subscriptions.

## Coordinated release

This change depends on the new `bibleit-go/rest` package. HTTP dispatch accepts
CLI arguments directly and renders structured JSON; it has no command adapter.
During development use the sibling checkouts in the existing local Go workspace:

```sh
go test ./...
go test ../bibleit-go/...
```

The v0.1.0 CLI pins the published bibleit-go v0.1.0 module without a local
replacement. Release order is SDK first, then CLI. Standalone validation uses:

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

Real-server validation is recorded in [release evidence](releases/v0.1.0-integration.json).
Do not add sibling-path replacement directives or workspace files to a release.
