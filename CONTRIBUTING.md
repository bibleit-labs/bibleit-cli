# Contributing to bibleit-cli

Bibleit CLI is a public MIT-licensed project in the `bibleit-labs` organization.
Bug reports, documentation improvements, and focused pull requests are welcome.
For larger API or behavior changes, open an issue describing the customer use
case before implementation. Never include credentials or private server data
in issues, commits, or test fixtures.

## Development

Use Go **1.27.1** and build with `go build -o build/bibleit .`.
The CLI uses the sibling `bibleit-go` checkout through an ignored local
`go.work` while publication is paused. The version in `go.mod` is an unreleased
placeholder; pin a real client release before standalone builds. Keep library behavior in that repository; profiles, terminal output,
browser orchestration, and OpenSSH process integration belong here.

For changes to executable behavior, run:

```sh
go test -race -count=1 ./...
go vet ./...
go build -o build/bibleit .
python3 scripts/check_cli_runtime.py --binary build/bibleit
```

Linux contributors can add `--shells bash zsh fish` to the runtime check.
CI runs Windows, Ubuntu, and macOS checks. Documentation-only changes need
link/content review rather than additional executable tests.

A local `go.work` must include the sibling `../bibleit-go` checkout for now.
Do not commit workspace files or local `replace` directives. Published CLI
builds must resolve the pinned client without access to a sibling checkout.
Real-server tests require the separate server checkout; see
[native validation](docs/native-validation.md) and `scripts/check_server.py`.

## Commits and pull requests

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):
`type(scope): description`, with an optional scope. Examples:

- `feat(profile): add named account profiles`
- `fix(windows): protect credential file permissions`
- `docs: explain Go client dependency updates`
- `ci: validate Fish completion`

Use `!` or a `BREAKING CHANGE:` footer for incompatible changes. Keep PRs
focused and explain the user-visible change and relevant validation.
Do not use the `codex/` prefix for branch names.

## Releases and compatibility

CLI releases and Go client releases are independent. Update the pinned client
explicitly and commit `go.mod` and `go.sum` together. Document supported server
operations, required Go version, and compatibility limitations. Protocol fixtures
are versioned snapshots; synchronize changes deliberately across client
repositories. A passing local fixture test is not a real-server compatibility
claim. Semantic-version tags and downloadable binary releases are separate
from publishing development branches.

The service endpoint is project-managed; see [the endpoint policy](docs/endpoint-policy.md).
Private validation-branch pushes and CI are now authorized in the recreated
private repositories. Release tags, public installation and binary releases
remain paused until explicitly authorized. The separate `homebrew-tap` stores
the formula; its local development candidate is not ready for distribution.

Local candidates are built with `scripts/package_release.py --local --version
v0.1.0-dev.1`; this creates no Git tags. Run `scripts/check_release.py --assets
build/releases` to check archives and exercise your platform's installer offline.
See [release preparation](docs/releases.md) for output paths, Windows checks,
package-manager manifests, and the production gates. The prepared candidate
workflow has read-only repository permissions and no publication steps.

Run `python3 scripts/check_homebrew.py --assets build/releases/<candidate>` on
macOS/Linux with Homebrew already installed. This clones a disposable prefix,
uses local archive URLs, tests install/completions/reinstall/uninstall, and removes
the fixture. To validate the separate tap, add `--formula ../homebrew-tap/Formula/bibleit.rb`.
On Windows, `scripts/check_scoop.py --assets <candidate> --scoop-source <source>`
uses an existing Scoop checkout, a disposable root, and loopback archive downloads.
Private CI needs `BIBLEIT_CLIENT_READ_TOKEN`, restricted to contents-read on
`bibleit-labs/bibleit-go`, to check out the exact client SHA. Workspace validation
does not close the standalone `GOWORK=off` production-release gate.

For real-server integration while the developer service occupies port 8080, run
`python3 scripts/check_server_container.py --baseline contract/server-baseline.json`.
This uses a disposable container namespace and database; it reads a frozen
server snapshot and rebuilds NIFs without editing the server checkout.
Use `NewClient(apiToken)` for standard client construction and the separate
`NewClientWithHTTPClient` constructor for transport/timeout customization.
