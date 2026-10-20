# Root SDK API migration

CLI v0.2.3 pins bibleit-go v0.3.0 and imports only
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
