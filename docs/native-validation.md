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
repositories were recreated; earlier hosted records may no longer be available.
At the publication pause, changes stayed local with a sibling workspace. Private
CI is now authorized; see the latest validation section below. Each matrix targets runner-default architectures, not all six
build targets. The original repository's remote main is unchanged.

## Single-configuration validation — 11 October 2026

The CLI removes named profiles and uses schema-v2 credentials with HTTP default
and explicit SSH transport. Local macOS/arm64 race/vet, process/loopback HTTP,
Unix credential permissions, actual PTY confirmation and Bash/Zsh completion pass.
Fresh isolated server integration also passes HTTP/auth/SSH/output checks and
25 import cases; the baseline contains 63 source and 59 test fingerprints.

The [client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38091438398)
passed Windows/Ubuntu/macOS at 17e146b. Tap syntax/style and seven updater
regressions passed at d08561b in
[tap CI](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38091389888).
The manual release updater remains uninvoked pending a real CLI release and
its documented private-read/PR permission setup.

The new [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481226)
passed Windows/Ubuntu/macOS at fe719db with exact client 17e146b, including native
Windows config ACL checks, process fixtures and Linux Bash/Zsh/Fish completion.
The [candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38091481159)
also passed at those inputs: six-target packaging, Linux/macOS installers,
PowerShell 7/5.1 replacement/failure checks, Scoop and Homebrew lifecycle checks.
Later documentation-only commits do not change the tested runtime code. These
are development-endpoint/workspace candidates; production packaging and fresh
published-download validation remain release gates. Later assessment changes
are outside the frozen integration snapshot. Windows/arm64 is cross-build only.

## Readiness refresh — 11 October 2026

Prepared source branches are pushed to the private repositories. The fresh
[Go-client native matrix](https://github.com/bibleit-labs/bibleit-go/actions/runs/38090370470)
passed Windows, Ubuntu and macOS at client `310259e1acb5dbd1b07378bfb41462fdddfcda2c`.
The fresh [CLI native matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38090456759)
passed all three platforms at CLI `0ce1c4732526c2af6303a0074094e9670f9bcbd7`,
using that exact client SHA. CLI checks include race/vet/build, native process
fixtures, and Linux Bash/Zsh/Fish runtime checks. These draft runs use the
localhost endpoint and an ephemeral private-client workspace; production addresses,
a released client and standalone GOWORK=off validation remain separate gates.
The hosted matrix exercises runner architectures; Windows/arm64 remains
cross-build only, not native execution certification.

The refreshed [release-candidate matrix](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38090456683)
also passed all three platforms at the same CLI/client revisions: six-target
packaging, native Bash/PowerShell installers, PowerShell 7/5.1 replacement/failure
paths, Scoop lifecycle and macOS Homebrew lifecycle. Artifacts are private
development candidates, not production releases or verified public downloads.

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


## Simple-client real-server checks (2026-10-10)

The Go client and CLI passed local race tests, vet, and build after switching to
`NewClient(apiToken)` and `NewAuthClient()`. The real-server harness also passed
against the fixed service in a disposable Linux network namespace:

```sh
python3 scripts/check_server_container.py --baseline contract/server-baseline.json
```

The fixture server binds IPv4 loopback port 8080, so public constructors connect
to their actual project-managed address. There is no client URL or destination
remapping. The runner shares only the temporary database container's namespace;
no ports are exposed on the host. NIFs are rebuilt for Linux from frozen source,
and the original server checkout and existing host service remain untouched.

Passing coverage: 27 decoder fixtures, 18 encoder fixtures, 7 targeted security
cases, Go HTTP/domain/PKCE/token checks, actual CLI/OpenSSH subscription lifecycle,
and 8 authentication/SSH EUnit cases. The baseline records the server working-tree
fingerprints and fixed HTTP endpoint. Native Windows was not rerun for that increment.

## Private package-manager validation (2026-10-10)

The user recreated the Go-client and CLI repositories privately and authorized
validation-branch pushes and CI. Releases/tags remain paused. The Go-client matrix
passed Windows, Ubuntu and macOS at `3d15ec6f54eba822129863e86e006d51deb07725` in
[run 38086747099](https://github.com/bibleit-labs/bibleit-go/actions/runs/38086747099).

Real Homebrew macOS/arm64 install/test/completion/reinstall/uninstall checks pass
using the generated development formula in a disposable Homebrew prefix. The
normal user installation is unchanged. Report: `build/validation/homebrew.json`.
The tap is a separate `bibleit-labs/homebrew-tap` repository, prepared locally.

The native CLI matrix passed Windows/Ubuntu/macOS at
`1e58c54adb21cb1bfb00e9c7e021b38278f72828` in
[run 38088193147](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38088193147).
Ubuntu also executed Bash/Zsh/Fish completion checks.

The distribution matrix passed at the same CLI revision in
[run 38088193158](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38088193158):
Linux/macOS Bash installation/replacement/failure paths; native Windows
PowerShell 7 and Windows PowerShell 5.1 installation/replacement/failure paths;
PowerShell parser checks; real macOS Homebrew lifecycle; real native Windows
Scoop install, executable shim, reset, reinstall, uninstall and hash rejection.
Scoop engine: `e6aa3b366bdee8ed138c1e0f7b85192ebdd35d0f`. Hash rejection checks
require the explicit failure and absence of an executable, current link and shim,
independently of Scoop's pipeline exit status. Windows/arm64 remains cross-build
only. Local evidence is under `build/validation/ci-38088193158`.

The scoped checkout secret is configured; these workflows use exact Go-client
revision `3d15ec6f54eba822129863e86e006d51deb07725`, ephemeral go.work and
development-only artifacts. They do not validate GOWORK=off production builds
or hosted public download URLs. The subsequent local client fixture/comment
refresh for the server's LIVE STATS repair is outside these pinned CI runs.

## Fresh full HTTP/SSH integration (2026-10-10)

Local branch `integration-refresh`, CLI source `045fd20`, Go client `9332d35`:
`python3 scripts/check_server_container.py --baseline contract/server-baseline.json`
passed against server HEAD `[private digest withheld]` with uncommitted assessment changes.
The baseline fingerprints identify the actual tested source; this is not a claim
against a published server release. The assessment continues editing the server; source
fingerprints identify the tested snapshot rather than its moving working tree.

The harness passed 27 decoder and 18 encoder fixtures, seven targeted security
cases, all 11 HTTP integration subtests, real OpenSSH subscription events,
disconnect/reconnect and deletion closure, CLI profile/stable-output checks and
eight authentication/SSH EUnit tests. The LIVE STATS repair passed both canonical
encoding and real subscriber statistics. No regressions or new fixes were needed.
Negative SSH algorithm tests intentionally produce rejected-handshake logs.

Linux fixture NIFs were rebuilt from frozen source. The temporary database and
server shared a disposable network namespace; no host ports were exposed, and
the database container was removed. Full output is saved locally at
`build/validation/full-integration-refresh-2026-10-10.log`. Both full runs passed;
the second included a fetcher edit discovered after the first snapshot. Subsequent
translation/registry edits from the ongoing assessment are outside this baseline.
This refresh does not repeat
native Windows/package-manager CI, validate production TLS/download addresses,
or replace the server's broader security assessment.

## Latest assessment translation validation (2026-10-10)

`python3 scripts/check_server_container.py --imports --baseline contract/server-baseline.json`
passed with CLI harness `26de771`, Go client `5138077` and the server working tree
at HEAD `[private digest withheld]`. The baseline now fingerprints
all 63 server source and 57 test files. Both sets matched the server at completion;
future assessment edits remain outside that snapshot.

In addition to the complete existing HTTP/SSH harness (27 decoder/18 encoder
fixtures, 7 targeted security cases, 11 HTTP subtests, real OpenSSH lifecycle and
profile/output checks, 8 auth/key EUnit tests), all 10 import-access and 14
publication cases pass. These exercise immutable generations, concurrent
installations and archives, retained native handles, idempotent provenance and
response-loss retries, registry outage, legacy upgrade, install/delete races,
path containment, staged failures and process death before/after publication.

The runner executes the actual suite cases with setup/teardown callbacks; it does
not run the full Common Test campaign. Access cases use real HTTP with loopback
archive fixtures, while publication fault cases use the assessment's production
module clone with filesystem/NIF boundary substitutions. The optional flag keeps
the normal client/CLI gate available independently.

Log: `build/validation/assessment-integration-2026-10-10.log`. The isolated
database was removed, no host ports were exposed and no server files were changed.
Native Windows/package-manager evidence remains the preceding private CI runs;
this increment stays entirely local.
