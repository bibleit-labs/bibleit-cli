# Bibleit CLI

Read and search Bible translations, manage Lives, and inspect your account from the terminal.
The `bibleit` executable uses the [Bibleit Go client](https://github.com/bibleit-labs/bibleit-go)
and connects to the project-managed service at `https://bibleit.app`.
SSH commands connect to `bibleit.app:22`.

## Installation and current availability

The CLI, Go client and Homebrew tap are **private**. Private CLI v0.0.1 exists,
but its binary packages lack required third-party notices. Corrected **v0.0.2**
packages and a matching tap formula are staged for review; they are not a
published release. Native Windows installer/Scoop checks remain pending.
See [release readiness](docs/release-v0.0.2-readiness.md) for the remaining gates.

Authorized contributors can build from source using Go 1.27.1 and authenticated
Git access to the pinned private SDK; see [Development](#development).
Repository access and Bibleit service access are separate: using the CLI still
requires the relevant account credentials, permissions and translation access.

### Planned public installation

The following commands apply **after the corrected release and public access are
approved**. They are not available to unauthenticated users today. Packages cover
macOS, Linux and Windows on amd64/arm64; binaries require no Go or Python runtime.

Homebrew (macOS/Linux):

```sh
brew install bibleit-labs/tap/bibleit-cli
```

Bash (macOS/Linux):

```sh
curl -fsSL https://github.com/bibleit-labs/bibleit-cli/releases/download/v0.0.2/install.sh | bash -s -- v0.0.2
```

PowerShell (Windows; validation pending):

```powershell
Invoke-WebRequest -UseBasicParsing -Uri https://github.com/bibleit-labs/bibleit-cli/releases/download/v0.0.2/install.ps1 -OutFile install.ps1
& .\install.ps1 -Version v0.0.2
```

Script installers default to `~/.local/bin` or `%LOCALAPPDATA%\Bibleit\bin`;
add that directory to PATH if needed. They verify checksums, version and required
notices before replacing an executable and retain notices in `bibleit-licenses`.
They require no administrator access and do not edit shell profiles.
SSH commands need a separate OpenSSH client. Normal OS security protections remain
enabled; these packages are not platform-signed or notarized.

Release packages include a Scoop manifest, but no public Scoop bucket exists.
See [distribution and maintenance](docs/releases.md) for offline candidate checks,
updates, uninstalling and package-manager validation.

## Quickstart

Sign in through your browser, then inspect the translations available to your account:

```sh
bibleit auth login
bibleit auth info
bibleit translation list
```

Use an enabled translation slug from that list. For example, if `web` is enabled:

```sh
bibleit translation catalog web
bibleit read web John 3 16
bibleit search web "love"
bibleit --format json account info
```

Book names come from the translation catalogue and may be localized. Book identifiers
from the catalogue also work. Quote arguments containing spaces.

For automation, provide an **API token** through your secret store as `BIBLEIT_TOKEN`.
The CLI uses it for HTTP commands without saving or printing it.

## Authentication and service address

HTTP is the default transport. Browser login saves an HTTP credential; an environment
API token takes precedence for data commands and `auth info`.

SSH requires a key registered with your Bibleit account and OpenSSH on your machine:

```sh
bibleit auth login ssh --identity ~/.ssh/bibleit_key
bibleit --transport ssh auth info
bibleit --transport ssh live list
```

An explicitly selected identity requires its corresponding `.pub` file. OpenSSH
verifies the server host key. HTTP and SSH credentials may belong to different
accounts; `auth info` reports the identity for the selected transport.
`BIBLEIT_TOKEN` applies only to HTTP. There is no automatic transport fallback.

The CLI stores one credential configuration at `~/.bibleit/config.json`.
`BIBLEIT_CONFIG` selects another configuration file. Unix permissions and Windows
ACLs restrict saved credentials to the current user. SSH private keys stay in their original files.

```sh
bibleit auth logout
```

Logout revokes the saved HTTP credential and clears saved HTTP/SSH selections.
Failed revocation preserves credentials for retry. It does not revoke an environment
API token, remove SSH keys, or unregister them from your account.

The service address is fixed; there are no endpoint overrides or named profiles.
HTTPS certificate and SSH host verification remain enabled.
See the [endpoint and credential policy](docs/endpoint-policy.md) for binding and migration details.

## Commands

Use `bibleit help` or `bibleit <command> --help` for usage.

| Command | Purpose |
| --- | --- |
| `version` | Show the installed CLI version |
| `auth login`, `auth info`, `auth whoami`, `auth logout` | Manage and inspect authentication |
| `ping`, `server info`, `server help [topic]` | Inspect service availability and capabilities |
| `account info`, `account quotas`, `account tokens` | Inspect account details and active token metadata |
| `translation list` | List account translations |
| `translation add\|remove\|info\|catalog <translation>` | Manage translations and inspect metadata/references |
| `read <translation> <book> [chapter] [verse]` | Read a reference |
| `search <translation> <query>` | Search a translation |
| `live list`, `live create [name]` | Discover and create Lives |
| `live <id> info\|stats\|start\|stop\|pause\|resume` | Inspect or control a Live |
| `live <id> set name\|reference\|translations <value...>` | Update a Live |
| `live <id> secret create\|rotate\|delete` | Manage a server-generated Live secret |
| `live <id> stack info\|clear\|pop [count]` | Inspect or remove stack entries |
| `live <id> stack push [translation] <book> [chapter] [verse]` | Add a reference to the stack |
| `live <id> clear\|delete`, `live delete all` | Clear or delete Lives |
| `--transport ssh live <id> subscribe` | Stream Live events |

Custom Live secrets are unsupported. Live subscriptions require SSH, start with a
snapshot, and do not offer a replay cursor. Run the command again for a fresh snapshot
after a disconnect. Ctrl+C closes the subscription.

The CLI version, server version, and protocol version are independent.
Use `bibleit version` for the CLI and `bibleit server info` for the service.
Ordinary source builds report `dev`.

## Output and automation

The default output is a human-readable table. Select `--format raw` for finite
protocol replies or `--format json` for the versioned JSON envelope.
SSH subscriptions support `--format ndjson`, with one event per line.
Errors go to stderr and display control characters safely.
See the [output contract](docs/output.md).

Commands that delete Lives, clear content, pop stack entries, or rotate/delete
secrets require confirmation. Interactive prompts require a terminal;
automation must explicitly supply `--yes`:

```sh
bibleit --yes --format json live delete all
```

Use this command only when you intend to delete all your Lives. Confirmation does
not bypass server permissions. Mutations are not retried automatically.
Use `--` before literal arguments that resemble global options.

| Exit code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Network, server, local failure, or cancellation |
| 2 | Usage or confirmation required |
| 3 | Authentication failure |
| 4 | Permission denied |
| 5 | Not found |
| 6 | Conflict or quota limit |
| 7 | Rate limited |

## Shell completion

Completion generation runs locally without authentication or network access.

Bash:

```sh
source <(bibleit completion bash)
```

Zsh:

```sh
autoload -Uz compinit && compinit
source <(bibleit completion zsh)
```

Fish:

```fish
mkdir -p ~/.config/fish/completions
bibleit completion fish > ~/.config/fish/completions/bibleit.fish
```

Homebrew installs the completion files with the executable.

## Development

Use **Go 1.27.1**. The CLI pins the released Go client; a sibling checkout is optional.
While the Go repository is private, configure Git access and set
`GOPRIVATE=github.com/bibleit-labs/*` when downloading private modules.

```sh
GOWORK=off go mod download
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go build -o build/bibleit .
```

For joint client/CLI development, an uncommitted local workspace can include both
checkouts: `go work init . ../bibleit-go`.
Native code tests cover macOS/Linux/Windows. Corrected-release Windows installer
and Scoop checks remain pending; fixture checks do not prove public downloads.
See [native validation](docs/native-validation.md), [release maintenance](docs/releases.md),
and the [changelog](CHANGELOG.md).

Contributions use Conventional Commits. See [contributing](CONTRIBUTING.md).
Licensed under [MIT](LICENSE). Binary distributions also include
[third-party notices](THIRD_PARTY_NOTICES); retain both files when redistributing
the CLI. The software license does not grant rights to translations returned
by the service.

For contributions and bug reports, see [CONTRIBUTING.md](CONTRIBUTING.md).
Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).
