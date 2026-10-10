# bibleit CLI

The CLI module is `github.com/bibleit-labs/bibleit-cli` and requires Go 1.27.1.
Development uses local checkouts and authorized private CI validation branches.
Public releases and tags remain paused. The independent Go client is the sibling
`../bibleit-go` checkout. Create a local workspace with `go work init . ../bibleit-go`
if `go.work` does not already exist, then build with `go build -o build/bibleit .`.
`go.mod` uses an unreleased `v0.0.0` placeholder; it must be replaced with a real
client release before publishing standalone CLI builds. No tags or binary releases
are being published yet.

bibleit is the Go CLI for Bibleit Server protocol version 1, built on the
[Go client](https://github.com/bibleit-labs/bibleit-go). It
implements a useful, typed subset of the API instead of providing a raw-command
escape hatch, so shell input cannot accidentally become protocol input.

The service address is controlled by the project, currently **`http://localhost:8080`**.
Customers cannot select another endpoint through constructor options, CLI flags,
environment variables, or saved configuration. The production HTTPS address will replace
this placeholder in the Go client's `ServerEndpoint` constant. See the
[endpoint policy](docs/endpoint-policy.md).

The CLI stores one credential configuration for the fixed service:

```sh
bibleit auth login
bibleit auth info
bibleit --transport ssh auth login --identity ~/.ssh/bibleit_key
bibleit --transport ssh live list
bibleit auth logout
```

HTTP is always the default. `--transport ssh` explicitly selects OpenSSH; there
is no automatic fallback. `auth login ssh` is an alias for SSH login. SSH login
caches an optional identity path without removing the saved HTTP token or changing
the default transport. HTTP and SSH identities can belong to different accounts;
`auth info` reports the actor for the selected transport. No named profiles,
`--profile` or account-switching commands are provided. `BIBLEIT_PROFILE` is ignored.
`--endpoint` is rejected and `BIBLEIT_ENDPOINT` is ignored.

`auth login` opens browser authorization. Use `--identity` only with SSH.
Logout revokes the cached HTTP credential and clears both cached credential
selections; failed revocation preserves them for retry. Registered SSH keys and
the agent remain under your control. Login/logout operate on saved credentials,
not an environment token.

For automation, set `BIBLEIT_TOKEN` through your environment/secret store. It
overrides the HTTP token for data commands and `auth info`, without being saved
or printed. Explicit SSH commands ignore this HTTP-only token. Use `--` before
literal command arguments that would otherwise be interpreted as global flags.

Use Go **1.27.1**. With the local workspace configured, build and test from the repository root:

    go build -o cli .
    go test ./...

## Versioning

The CLI and server are released independently:

- Local CLI builds report `dev`; version tags and downloadable releases are paused.
- The independent server version is reported by `bibleit server info`.
- Protocol compatibility is reported separately as `protocol_version`.

Use the local version command to inspect the installed client, and server info
to inspect the server to which it is authenticated:

    bibleit version
    # bibleit CLI version dev

    bibleit server info
    bibleit server help [topic]
    bibleit account info
    bibleit account quotas
    bibleit account tokens
    bibleit translation info <translation>
    bibleit translation catalog <translation>
    # includes version=<server version> and protocol_version=1

The CLI version is independent of the server version. Local candidate packaging
injects the chosen version; ordinary source builds continue to report `dev`.

## Installation and release candidates

Public installation is not available yet. Local packaging prepares six binary
archives (macOS/Linux/Windows, amd64/arm64), SHA-256 checksums, Homebrew and Scoop
manifests, and Bash/PowerShell installers. Everything stays local while
publication is paused:

```sh
python3 scripts/package_release.py --local --version v0.1.0-dev.1 --output build/releases/local-dev-1
python3 scripts/check_release.py --assets build/releases/local-dev-1
```

See [release packaging and distribution](docs/releases.md) for local installer
tests, future Homebrew/cURL/PowerShell/Scoop commands, and remaining release gates.

Commands:

    bibleit version
    bibleit auth login
    bibleit auth login ssh [--identity <private-key>]
    bibleit auth logout|info|whoami
    bibleit ping
    bibleit server info
    bibleit account info|quotas|tokens
    bibleit translation list
    bibleit translation add|remove <translation>
    bibleit read <translation> <book> [chapter] [verse]
    bibleit search <translation> <query>
    bibleit live list|create [name]|delete all
    bibleit live <id> info|stats|start|stop|pause|resume|subscribe|clear|delete
    bibleit live <id> set name|reference|translations <value...>
    bibleit live <id> secret <secret>|create|rotate|delete
    bibleit live <id> stack info|clear|pop [count]
    bibleit live <id> stack push [translation] <book> [chapter] [verse]

For web authentication, the CLI starts a temporary loopback callback, opens
the project-managed service URL in the default browser, and asks the signed-in
account to approve the terminal. The authorization code is protected with PKCE
and exchanged for a persistent account token; the token never appears in the
browser URL.

For SSH authentication, the CLI first delegates identity selection to OpenSSH,
including matching `~/.ssh/config` entries and keys loaded in your SSH agent.
When `--identity` is supplied, the private key and its `.pub` file must both
exist. The CLI uses `ssh-keygen` to validate Ed25519, RSA (at least 2048 bits),
and ECDSA P-256/P-384/P-521 public keys. Its fingerprint must already be
registered with an account. The SSH transport username is `bibleit-cli`; the
server identifies the account from the key fingerprint, not that username.

For the local SSH development server, generate a dedicated key and run:

    ssh-keygen -t ed25519 -f ~/.ssh/bibleit_local_dev -C bibleit-local-dev
    make -C ../bibleit-server run
    go build -o cli .
    ./cli auth login

After web sign-in, open the dashboard and add the contents of
`~/.ssh/bibleit_local_dev.pub` under **SSH keys**. Only then will SSH login
succeed:

    cat ~/.ssh/bibleit_local_dev.pub
    ./cli --transport ssh auth login --identity ~/.ssh/bibleit_local_dev
    ./cli --transport ssh live list

On the first connection, OpenSSH saves a previously unseen server host key in
`known_hosts`; it still rejects a host whose saved key later changes. Every
SSH-backed command creates a fresh authenticated SSH connection using the
cached settings.

With email or OAuth configured, the equivalent web flow is:

    ./cli auth login
    ./cli live list

Credentials are stored in `~/.bibleit/config.json`, with newly created
directories mode `0700` and file writes mode `0600` on Unix. On Windows, each
new config file receives a current-user-only ACL before credentials are written;
failed ACL setup aborts saving. Atomic replacements retain these protections.
`BIBLEIT_CONFIG` selects another file. Schema version 2 stores fixed HTTP/SSH
service bindings, an optional access_token and SSH identity path. It contains
no profiles or saved transport choice; private keys are never stored here.
Foreign service bindings, old configuration files and legacy credential files are
rejected before network access and never rewritten. Back up an old file and
sign in with a fresh config; there is no automatic account selection or migration.
Backing up an old file does not revoke its server credentials. Revoke obsolete CLI
credentials in account settings when replacing the old configuration.

Output defaults to a readable table. Use `--format raw` to preserve finite
protocol records, `--format json` for a version-1 typed record envelope, or
`--format ndjson` for an SSH subscription with one object per record. Finite
commands reject ndjson; subscriptions reject json. Error envelopes go to stderr
and retain server metadata. Human tables escape embedded control characters.
See [the stable output contract](docs/output.md) for types and examples.

```sh
bibleit --format json account tokens
bibleit --transport ssh --format ndjson live <id> subscribe
bibleit --format raw live list
```

Exit codes are stable:

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Network, server, or local failure |
| 2 | Invalid CLI usage |
| 3 | Authentication required or credential rejected |
| 4 | Permission denied |
| 5 | Requested resource not found |
| 6 | Conflict or quota rejection |
| 7 | Rate limited |

## Client development

Keep `bibleit-go` beside this repository for joint development. The local workspace
uses both modules (`go work init . ../bibleit-go`) and is ignored by Git. It is
required while the client is unreleased. When publication resumes, pin a released
client in `go.mod` and validate with `GOWORK=off`; published builds must not need
a sibling checkout or local replacement.

The real-server harness accepts `--go-client ../bibleit-go`; client integration
tests compile from that checkout, while CLI integration tests compile here.
It exercises public `NewClient(apiToken)` and `NewAuthClient()` constructors
against an actual server at the fixed endpoint. If host port 8080 is busy, run
`python3 scripts/check_server_container.py --baseline contract/server-baseline.json`.
That runner freezes the sources and builds Linux NIFs in a disposable namespace,
with no published ports and no access to the developer's running service.

See the [item-by-item progress](docs/status.md),
[operation inventory](contract/features.md), and [revised implementation plan](docs/plan.md), reviewed against the server
working tree on 2026-10-09. The current CLI is a development draft; organization
management and bearer-authenticated event watching are not implemented.

The reusable [Go client](https://github.com/bibleit-labs/bibleit-go) owns validated customer commands,
HTTP transport, and finite response decoding. The CLI owns browser login,
credential storage, terminal output, and its OpenSSH adapter. The shared protocol contract
and response fixtures are documented in [contract/README.md](contract/README.md).

The CLI depends on the independent `github.com/bibleit-labs/bibleit-go` module.
Python, Ruby, and Rust clients will live in `bibleit-python`, `bibleit-ruby`, and
`bibleit-rust` repositories under `bibleit-labs`. They have not been implemented yet.

Run a standalone client example with an existing account token:

```sh
cd ../bibleit-go
BIBLEIT_TOKEN=your-token go run ./examples/read
```

`secret set <value>` is rejected until the server command API supports custom
secret setting. Use `secret create` or `secret rotate` to
generate a server-managed secret. Live subscriptions require explicit SSH transport.

Subscriptions begin with a current Live snapshot, then emit live changes. Live
closure ends the watch successfully; revoked access and unexpected disconnects
return errors. Ctrl-C stops the SSH child cleanly. After a disconnect, run the
subscription command again to obtain a fresh snapshot. Protocol v1 does not
provide event replay or a resume cursor, so intervening events are not replayed.
Arguments containing embedded double quotes are rejected because the server's
current request tokenizer cannot represent them reliably.

## Shell completion

Generate scripts with `bibleit completion bash`, `bibleit completion zsh`, or
`bibleit completion fish`. Generation works without authentication or a valid
configuration file. It prints the script and does not change your shell settings.

For the current Bash session:

```sh
source <(bibleit completion bash)
```

For Zsh, initialize completion first:

```sh
autoload -Uz compinit
compinit
source <(bibleit completion zsh)
```

Alternatively, save Zsh output as `_bibleit` in a directory on `$fpath` before
running `compinit`. For Fish:

```fish
mkdir -p ~/.config/fish/completions
bibleit completion fish > ~/.config/fish/completions/bibleit.fish
```

Completion covers command words, flags and format/transport values. It never
reads credentials, contacts the server, writes configuration or suggests tokens,
secrets, Live IDs or translation/reference data. It works even when configuration
is missing or invalid. Script generation accepts normal or raw format;
JSON/NDJSON are not script formats.

## Destructive commands

Live deletion (including `live delete all`), stack clear/pop, and secret
rotation/removal require confirmation before a request is sent. A terminal
prompt shows the transport, endpoint and affected target. Type `yes`
for a single target, or `delete all` for bulk deletion. An empty answer, another
answer, EOF or an overlong answer cancels without sending a request.

Both stdin and stderr must be terminals for prompting. JSON output and
noninteractive use require `--yes`:

```sh
bibleit live session-id delete --yes
bibleit --format json live delete all --yes
```

`--yes` confirms the command once; server authentication and permissions still
apply, and failed mutations are not retried. It takes no value (`--yes=false`
is rejected). Options after `--` are literal command arguments.

Cancellation exits **1**. Missing confirmation exits **2**, with error code
`confirmation_required` in JSON; cancellation uses `cancelled`. Errors go to
stderr and leave stdout empty. Read commands, reversible Live controls/settings,
translation library removal, logout
retain their existing behavior. Creating a secret cannot replace an existing
secret; use guarded rotation to replace one.


## Native platform checks

The [native CI workflow](.github/workflows/native.yml) pins Go 1.27.1 and runs
race tests, vet, builds and executable smoke checks on Ubuntu, Windows and macOS.
Linux also runs actual Bash/Zsh/Fish completion. See
[native validation](docs/native-validation.md) for commands, coverage and verification evidence. Native macOS/arm64 and Linux/arm64 checks
pass locally for the fixed-endpoint draft, including Fish/Bash/Zsh on Linux.
All six Linux/macOS/Windows amd64/arm64 targets cross-build. Hosted results are tied to the exact revisions in the validation records.
Current private source pushes are authorized, and workflows pin the reviewed
Go client commit. Fresh native and package-candidate checks must pass on the
final release inputs; historical green runs do not certify later changes.


## Contributing and license

Bibleit CLI is a public open source project maintained in `bibleit-labs`.
See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, validation and Conventional
Commits. This project is distributed under the [MIT license](LICENSE).
The Go, Python, Ruby and Rust clients have independent repositories and releases.
