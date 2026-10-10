# Service endpoint policy

Updated 2026-10-10. All language clients target the single project-managed Bibleit
service. The server repository is private; its eventual organizational move is
separate from publishing public client repositories.

The Go client declares `ServerEndpoint = "http://localhost:8080"` in `client.go`.
This is a development placeholder. Replace it with the actual production HTTPS
address before release. The CLI derives its HTTP and browser authentication
address from that constant. SSH currently uses the project-managed development
address `127.0.0.1:2222`; establish its production address before shipping SSH.

`Config` accepts a token and optional HTTP client. `AuthConfig` accepts an
optional HTTP client. Neither exposes an endpoint. Commands, code exchange,
and revocation all construct URLs at the fixed service, and redirects remain
blocked. Credentials remain explicit; the library reads no environment variables.
Custom HTTP clients configure transport/timeouts, not a service URL option.
As with any open source library, callers can modify source or implement their own
network transport; this API policy does not replace server-side authorization.

CLI profiles select accounts/credentials and HTTP or SSH transport. There is no
`--endpoint` option, completion candidate, or endpoint environment setting.
The persisted endpoint is a service-binding record. Its value must match the
current project-managed address. A mismatch fails locally; cached credentials
are never silently copied to a new service. Legacy bearer-only configs lack
that binding and are rejected. Users can retain the old config as a backup and
sign in with a fresh config after an address change.

Apply the same policy to future Python, Ruby and Rust clients: a project-owned
service constant, no constructor/environment/config endpoint override, and
matching authentication and data-request destinations.

## Local development and verification

Keep `bibleit-cli` and `bibleit-go` as sibling checkouts. A local ignored
`go.work` uses both while `go.mod` has an unreleased client version placeholder.
Do not publish standalone builds until a real client version is pinned and
`GOWORK=off` validation passes.

Unit tests use injected in-memory HTTP transports. Real-server integration test
binaries have private fixture routing in `_test.go` files, excluded from customer
builds. The native process harness currently binds a disposable HTTP fixture on
IPv4 loopback port 8080 to test the development endpoint without a public override.
It requires that port to be free and fails if another service is listening; it
never stops or replaces an existing service. Adapt that fixture strategy before
switching the fixed endpoint to production.

No release tags, binary publication, or Git pushes are authorized at this stage.
Repository deletion/recreation is being handled by the user. Local Git bundles
in the CLI's ignored `.local-backups/` directory preserve the pre-change histories.
