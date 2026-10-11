# Service endpoint policy

Updated 2026-10-11. All language clients target the single project-managed Bibleit
service. The server repository is private; its eventual organizational move is
separate from publishing public client repositories.

The Go client declares `ServerEndpoint = "https://bibleit-server.fly.dev"` in
`client.go`. The CLI derives its HTTP and browser authentication address from
that constant. SSH uses `bibleit-server.fly.dev:22`, with standard OpenSSH host
verification. These match the assessment-qualified deployment at server source
7442010. Fly supplies the trusted PROXY envelope to the server; clients send
ordinary SSH and never inject a PROXY preamble.

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

Unit tests use injected in-memory HTTP transports. Real-server integration calls
public constructors at the fixed production URL with standard TLS verification.
The container runner uses an egress-disabled namespace (`--network none`), maps
only the service hostname to loopback, and trusts a disposable fixture certificate
only inside that container. HTTPS on port 443 relays to the real server fixture;
ordinary SSH uses the fixed port 22. The developer's ports and host trust store
are untouched. Direct host integration is rejected. Wrong hostname and untrusted
certificate checks must fail before integration proceeds.

The native process harness first checks the actual binary's version, then compiles
a Go test executable using the same CLI entry function. Its test-only HTTP transport
allows only the production destination, dials an ephemeral IPv4 loopback TLS
fixture and verifies the hostname against a fixture-only trust pool. This works
on Windows, macOS and Linux without installing certificates. The dialer is absent
from production binaries. Process, PTY and shell-completion tests run offline;
no public endpoint option or insecure TLS setting is added.

The runner rebuilds Linux NIFs from the frozen source snapshot and uses only a
unique disposable database. No server checkout or developer database is modified.
Both client repositories record the verified server/native fingerprints only
when all requested HTTP/SSH/import checks pass.

Release tags and binary publication remain paused. The user recreated the client
repositories privately and authorized validation-branch pushes and CI; the
user authorized all prepared source branch pushes on 2026-10-11. Customer
release tags/assets remain gated on production readiness; the fixed-endpoint
rule remains in effect. Local Git bundles
in the CLI's ignored `.local-backups/` directory preserve the pre-change histories.
