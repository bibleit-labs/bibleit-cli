# Root SDK API migration

CLI v0.3.0 pins bibleit-go v0.5.0 and imports only
`github.com/bibleit-labs/bibleit-go`. Resource operations, models, errors,
pagination and Live polling derive from `bibleit.NewClient`. SSH executors use
`bibleit.NewSSHClient` and `bibleit.CommandResponse`.

The root resource client also supports browser-issued credential revocation.
HTTP/SSH service endpoints, CLI commands and wire output schemas are unchanged.

Validate the published dependency without a development workspace:

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

For development against the sibling SDK checkout, run
`go work init . ../bibleit-go`; `go.work` is local and ignored. Release packaging
uses the published SDK with `GOWORK=off` and no replacements.

After publishing CLI artifacts, update the Homebrew tap with its verified
release manifest and checksums using `scripts/update_formula.py` in the tap.

## Broadcast controls

The CLI delegates SSH broadcast wire validation to `bibleit.BroadcastCommand`
and uses the generated REST methods for the same operations. Setup mutations
retain a server-signed request ticket and exact JSON bytes;
`NewBroadcastWriteIntentInput` and `BroadcastWriteOptions` bind their checksum,
revision and raw body without re-encoding. Lost replies are reconciled with
the original ticket. No helper restarts capture or grants workspace authority.

The v0.5.0 public module is verified through the normal Go proxy and checksum
database. Release builds use `GOWORK=off`; local sibling workspaces and replacement
modules are not release inputs.
