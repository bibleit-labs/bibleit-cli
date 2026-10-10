# Contributing to bibleit-cli

Bibleit CLI is a public MIT-licensed project in the `bibleit-labs` organization.
Bug reports, documentation improvements, and focused pull requests are welcome.
For larger API or behavior changes, open an issue describing the customer use
case before implementation. Never include credentials or private server data
in issues, commits, or test fixtures.

## Development

Use Go **1.27.1** and build with `go build -o build/bibleit .`.
The CLI depends on `github.com/bibleit-labs/bibleit-go` at the version in
`go.mod`. Keep library behavior in that repository; profiles, terminal output,
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

An optional local `go.work` can include the sibling `../bibleit-go` checkout.
Do not commit workspace files or local `replace` directives. Published CLI
builds must resolve the pinned client without access to a sibling checkout.
Real-server tests require the separate server checkout; see
[native validation](docs/native-validation.md) and `scripts/check_server.py`.

## Commits and pull requests

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):
`type(scope): description`, with an optional scope. Examples:

- `feat(profile): add endpoint-bound profiles`
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
