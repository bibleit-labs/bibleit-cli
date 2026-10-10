# Native runtime validation

The CI workflow is `.github/workflows/native.yml`. It pins Go **1.27.1** and runs
on native Ubuntu 24.04, Windows Server 2025 and macOS 15 runners. Each job runs
race tests, vet, a CGO-free executable build and the process smoke harness.
Linux also installs and runs Bash, Zsh and Fish completion checks. Push, pull
request and manual dispatch events trigger the workflow once it is published.

The hosted workflow passed all three native jobs on 2026-10-10 at code commit
`08e4a31d0ff468b07e01408788f8b94a42995dc0`:
[verified run](https://github.com/mittel-labs/bibleit/actions/runs/38081432573).
That run validates the pre-extraction implementation in the original staging
repository. Before publication was paused, the independent repositories also passed their native matrices:

- [Go client](https://github.com/bibleit-labs/bibleit-go/actions/runs/38082660896), commit `9a52be34b0c3`.
- [CLI](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38082843195), code commit `e46bfd6`.

Those historical branches were `repository-split`, with no codex prefix. The
user is clearing these repositories; hosted records may no longer be available.
Current changes stay local on `fixed-service-endpoint`, and use a local workspace. Each matrix targets runner-default architectures, not all six
build targets. The original repository's remote main is unchanged.

## Historical evidence before fixed-endpoint changes

| Environment | Result | Coverage |
| --- | --- | --- |
| macOS/arm64, Go 1.27.1 | Pass | Race/vet/build; native process/HTTP fixture; actual terminal confirmation; Bash/Zsh completion |
| Linux/arm64 in Docker VM, Go 1.27.1 | Pass | Race/vet/build; native process/HTTP fixture; actual PTY confirmation; Bash/Zsh/Fish 3.6.0 completion |
| Windows Server 2025 hosted runner | Pass | Race/vet/build; actual saved DACL before/after replacement; console handles/CRLF; native process/HTTP fixture and redirected confirmation guards |
| Windows/arm64 | Cross-build only | Native arm64 execution is outside the current hosted matrix |
| Hosted Ubuntu/Windows/macOS jobs | Pass | All jobs succeeded in the linked run; Ubuntu also executes Bash/Zsh/Fish completion |

Linux runs execute Linux binaries on the Linux VM's arm64 CPU. They do not prove
native Linux/amd64 behavior or Windows behavior. The latest checks use a local
loopback HTTP fixture, not a real private-service instance; the separate server
integration harness retains its previous results.

## Run locally

Unix:

```sh
go test -race -count=1 ./...
go vet ./...
CGO_ENABLED=0 go build -o build/bibleit .
python3 scripts/check_cli_runtime.py --binary build/bibleit --shells bash zsh fish
```

Use only installed shells in `--shells`; explicitly requested missing shells fail
instead of silently skipping. The harness copies the executable into a directory
containing spaces, creates disposable profiles, starts a loopback-only HTTP
fixture, and removes its temporary files. Completion checks verify nested verbs,
option values, profile names, empty prefixes, positional references and `--`.
They assert no fixture requests or profile writes occur during completion.

Windows PowerShell, with Go, Python, OpenSSH and a Go-compatible MinGW GCC on PATH:

```powershell
$env:GOTOOLCHAIN = 'local'
$env:CGO_ENABLED = '1'
go test -race -count=1 ./...
go vet ./...
$env:CGO_ENABLED = '0'
go build -o build/bibleit.exe .
python scripts/check_cli_runtime.py --binary build/bibleit.exe
```

The Windows-only Go tests verify a real console handle with GetConsoleMode and
confirmation input using CRLF, plus the actual saved configuration DACL before
and after replacement. The process smoke harness covers redirected handles and
the no-prompt/--yes paths. Unix PTY checks exercise the built executable's real
prompt, rejection and confirmed dispatch. Windows console input end-to-end is
still a manual check beyond the console-handle test.

## Windows credential storage

Unix config files retain mode `0600`. Windows mode bits cannot supply the same
privacy guarantee: the CLI replaces the entire DACL on the new empty temporary
file with a protected DACL granting the current user's SID full access before
writing token data. This removes both inherited and explicit grants from the
creation environment. ACL failures abort saving; no credential is written to that
file. Replacement keeps the protected file's ACL. Existing configuration is
hardened on its next successful write. This requires an ACL-capable Windows file
system; administrators retain OS-level authority. The Windows test uses icacls
to inspect the resulting DACL.

The implementation uses Microsoft's [SetNamedSecurityInfoW](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-setnamedsecurityinfow)
and [SDDL conversion](https://learn.microsoft.com/en-us/windows/win32/api/sddl/nf-sddl-convertstringsecuritydescriptortosecuritydescriptorw).
DACL inspection follows the [icacls documentation](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/icacls).
Console tests use [AllocConsole](https://learn.microsoft.com/en-us/windows/console/allocconsole)
and [GetConsoleMode](https://learn.microsoft.com/en-us/windows/console/getconsolemode).
Go installation in CI uses [actions/setup-go](https://github.com/actions/setup-go)
with an exact version and automatic toolchain switching disabled.


## Current fixed-endpoint draft (2026-10-10)

- Go client and CLI: local macOS/arm64 race tests, vet and build passed.
- CLI native macOS process/HTTP fixture, PTY confirmation and Bash/Zsh checks passed.
- Both modules: native Linux/arm64 race tests and vet passed in a disposable Docker VM.
- Linux CLI process/HTTP fixture, PTY confirmation, and Fish/Bash/Zsh checks passed.
- CLI cross-builds passed for darwin/linux/windows on amd64 and arm64.
- Both updated real-server integration test binaries compile; real-server tests were not rerun.
- Native Windows and hosted Actions were not rerun for these local-only changes.

The runtime fixture now uses IPv4 loopback port 8080 and needs that port free.
Its requests use the fixed `http://localhost:8080` service address. The fixture
does not stop or replace an existing listener. See [endpoint policy](endpoint-policy.md).

Publication and releases are paused. Before publication resumes, pin a real
Go client version and validate builds with `GOWORK=off`, then rerun hosted checks.
