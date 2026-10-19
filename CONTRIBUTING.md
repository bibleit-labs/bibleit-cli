# Contributing to bibleit-cli

Bibleit CLI is MIT licensed and public. Use issues and focused pull requests for
contributions. Discuss
larger behavior changes before implementation. Report security problems through
[SECURITY.md](SECURITY.md); redact credentials and private account/server data.

## Standalone development

Use Go **1.27.1**. The CLI pins the released `github.com/bibleit-labs/bibleit-go` version in
go.mod/go.sum. A sibling checkout or workspace is optional. Download the public
module using normal Go proxy/checksum verification:

```sh
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off go build -o build/bibleit .
python3 scripts/check_cli_runtime.py --binary build/bibleit
```

These unit/runtime checks do not need the private server or a production account.
Linux contributors can add `--shells bash zsh fish` to the runtime check.
Native code CI covers Linux, macOS and Windows. Windows race tests need a supported
C compiler. Documentation-only changes require content/link review rather than
new executable tests.

For joint SDK/CLI changes, use an ignored local workspace:

```sh
go work init . ../bibleit-go
```

If go.work already exists, use `go work use . ../bibleit-go`. Keep library behavior
in the SDK; terminal output, credential storage, browser login and OpenSSH process
integration belong here. Never commit workspace files or local replace directives.
Verify release inputs with GOWORK=off and no replacement modules.

## Compatibility and validation

[Native validation](docs/native-validation.md) describes offline checks and
fixture isolation. Protocol fixtures are versioned snapshots; update them
deliberately with the SDK and record compatibility evidence. Passing
fixtures do not establish live-service behavior or anonymous distribution access.
The [output contract](docs/output.md) defines JSON/NDJSON and exit codes.

The project manages [service addresses](docs/endpoint-policy.md). Keep strict
HTTPS/SSH verification and explicit credentials; there are no customer endpoint
or profile overrides. Never log bearer values, browser codes, Live secrets, or
private keys. Use synthetic fixtures for account and translation data.

## Commits and reviews

Use Conventional Commits, for example `fix(auth): preserve credentials on failed
revocation` or `docs: clarify source builds`. Use `!` or a BREAKING CHANGE footer
for incompatible changes. Branches must not use a codex prefix. Explain changed
behavior and relevant validation; avoid unrelated work in a pull request.

## Maintainer documentation

See the [maintainer guide](docs/maintenance.md) for releases, installation checks,
security/licensing audits and historical compatibility records. The separate tap
owns formula review and package-manager validation. Keep LICENSE and
THIRD_PARTY_NOTICES with redistributed binaries. Software MIT does not grant
translation redistribution rights or service/account access.
