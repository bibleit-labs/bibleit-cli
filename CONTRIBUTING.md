# Contributing to bibleit-cli

Bibleit CLI is MIT licensed. The repositories currently remain private;
public contributions and anonymous package installation await approved launch.
Authorized contributors can use issues and focused pull requests now. Discuss
larger behavior changes before implementation. Report security problems through
[SECURITY.md](SECURITY.md); redact credentials and private account/server data.

## Standalone development

Use Go **1.27.1**. The CLI pins `github.com/bibleit-labs/bibleit-go v0.0.1` in
go.mod/go.sum. A sibling checkout or workspace is optional. While that module is
private, authenticate Git with an authorized account and mark the module private:

```sh
export GOPRIVATE='github.com/bibleit-labs/*'
GOWORK=off go mod download
GOWORK=off go mod verify
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off go build -o build/bibleit .
python3 scripts/check_cli_runtime.py --binary build/bibleit
```

GOPRIVATE controls proxy/checksum lookup; it does not supply Git credentials.
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
optional isolated server integration. Contributors without server access can run
the normal unit/runtime tests. Protocol fixtures are versioned snapshots; update
them deliberately with the SDK and record the tested server baseline. Passing
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

## Releases

SDK and CLI versions are independent. Update go.mod/go.sum together when changing
the pinned SDK. Source builds report dev unless a version is injected. Release
packaging requires a clean committed standalone checkout and Go 1.27.1.
See [release maintenance](docs/releases.md) for candidate commands and gates.

v0.0.1 remains private with deficient binary notices. Corrected v0.0.2 is staged;
Windows installer/Scoop checks stay pending. Do not publish tags/assets, merge the
staged tap formula, withdraw old assets or change visibility without approval.
The separate tap owns formula review and package-manager validation. Keep LICENSE
and THIRD_PARTY_NOTICES with redistributed binaries. Software MIT does not grant
translation redistribution rights or service/account access.
