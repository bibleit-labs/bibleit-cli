# Service endpoint policy

Updated 2026-10-11. All language clients target the single project-managed Bibleit
service. The server repository is private; its eventual organizational move is
separate from publishing public client repositories.

The Go client declares `ServerEndpoint = "http://localhost:8080"` in `client.go`.
This is a development placeholder. Replace it with the actual production HTTPS
address before release. The CLI derives its HTTP and browser authentication
address from that constant. SSH currently uses the project-managed development
address `127.0.0.1:2222`; establish its production address before shipping SSH.

Create an authenticated client with `NewClient(apiToken)`. Use
`NewClientWithHTTPClient(apiToken, httpClient)` for custom transport/timeouts.
`NewAuthClient()` creates the browser-code exchange client; its custom transport
constructor is `NewAuthClientWithHTTPClient(httpClient)`. No constructor exposes
an endpoint. Both simple constructors retain an error return for validation. Commands, code exchange,
and revocation all construct URLs at the fixed service, and redirects remain
blocked. Credentials remain explicit; the library reads no environment variables.
Custom HTTP clients configure transport/timeouts, not a service URL option.
As with any open source library, callers can modify source or implement their own
network transport; this API policy does not replace server-side authorization.

CLI configuration stores one HTTP token and optional SSH identity. HTTP is the
default; --transport ssh selects SSH explicitly. Named profiles are removed. There is no
`--endpoint` option, completion candidate, or endpoint environment setting.
The persisted endpoint is a service-binding record. Its value must match the
current project-managed address. A mismatch fails locally; cached credentials
are never silently copied to a new service. Legacy bearer-only and version-1 profile configs are rejected. Schema version 2
binds both HTTP and SSH addresses; it never auto-selects an old profile. Users can retain the old config as a backup and
sign in with a fresh config after an address change.

Apply the same policy to future Python, Ruby and Rust clients: a project-owned
service constant, no constructor/environment/config endpoint override, and
matching authentication and data-request destinations.

## Local development and verification

Keep `bibleit-cli` and `bibleit-go` as sibling checkouts. A local ignored
`go.work` uses both while `go.mod` has an unreleased client version placeholder.
Do not publish standalone builds until a real client version is pinned and
`GOWORK=off` validation passes.

Unit tests use injected in-memory HTTP transports. Real-server integration now puts the disposable server at `localhost:8080`
and calls public constructors without destination remapping. The isolated
container runner shares a disposable PostgreSQL network namespace with the
fixture, leaving host port 8080 untouched. It rebuilds Linux NIFs from the frozen
source snapshot. No server files or developer databases are modified. The native process harness currently binds a disposable HTTP fixture on
IPv4 loopback port 8080 to test the development endpoint without a public override.
It requires that port to be free and fails if another service is listening; it
never stops or replaces an existing service. Adapt that fixture strategy before
switching the fixed endpoint to production.

Release tags and binary publication remain paused. The user recreated the client
repositories privately and authorized validation-branch pushes and CI; the
user authorized all prepared source branch pushes on 2026-10-11. Customer
release tags/assets remain gated on production readiness; the fixed-endpoint
rule remains in effect. Local Git bundles
in the CLI's ignored `.local-backups/` directory preserve the pre-change histories.
