# Bibleit CLI

Read and search Bible translations, manage Lives, and inspect your account from the terminal.
The `bibleit` executable uses the [Bibleit Go client](https://github.com/bibleit-labs/bibleit-go)
and uses the pinned REST service at `https://api.bibleit.app/v1`.
Browser authentication uses `https://bibleit.app`.
SSH commands connect to `bibleit.app:22`.

## REST integration

Existing HTTP commands use REST wherever a matching resource exists, with the
CLI commands shown below. HTTP uses REST exclusively; commands without a REST
operation fail locally. See [REST migration details](docs/rest-api.md) for
transport availability, structured output and the coordinated SDK release.

## Installation

Install with Homebrew or a script installer below, or build from source using
the Go toolchain specified in [go.mod](go.mod); see [Build from source](#build-from-source).

Repository access and Bibleit service access are separate: using the CLI still
requires the relevant account credentials, permissions and translation access.
Packages cover macOS, Linux and Windows on amd64/arm64; binaries require no Go or
Python runtime.

Homebrew (macOS/Linux):

```sh
brew install bibleit-labs/tap/bibleit-cli
```

Bash (macOS/Linux):

```sh
curl -fsSL https://github.com/bibleit-labs/bibleit-cli/releases/latest/download/install.sh | bash
```

PowerShell (Windows):

```powershell
Invoke-WebRequest -UseBasicParsing -Uri https://github.com/bibleit-labs/bibleit-cli/releases/latest/download/install.ps1 -OutFile install.ps1
& .\install.ps1
```

Script installers default to `~/.local/bin` or `%LOCALAPPDATA%\Bibleit\bin`;
add that directory to PATH if needed. They verify checksums, version and required
notices before replacing an executable and retain notices in `bibleit-licenses`.
They require no administrator access and do not edit shell profiles.
SSH commands need a separate OpenSSH client. Normal OS security protections remain
enabled; these packages are not platform-signed or notarized.

Release packages include a Scoop manifest, but no public Scoop bucket exists.
To update a script installation, rerun the installer. For Homebrew, run
`brew update` followed by `brew upgrade bibleit-labs/tap/bibleit-cli`.

Uninstall a script installation by removing the executable and its adjacent
`bibleit-licenses` directory; use `brew uninstall bibleit-cli` for Homebrew.
Saved credentials remain. Use `bibleit auth logout` before removal when you want
to revoke the saved HTTP credential.

### Build from source

Use the Go toolchain specified in [go.mod](go.mod). From the repository directory,
build and run the executable on macOS/Linux:

```sh
GOWORK=off go build -o build/bibleit .
./build/bibleit --help
```

On Windows PowerShell:

```powershell
$env:GOWORK = 'off'
go build -o build/bibleit.exe .
.\build\bibleit.exe --help
```

Go downloads the pinned client dependency automatically; a sibling SDK checkout
is optional. The executable is written to `build/` and reports `dev` as its
version. See [Shell completion](#shell-completion) to enable completion in your shell.

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
verifies the host key directly for bibleit.app. Older hostname trust entries
are not reused; OpenSSH records a newly encountered key and rejects changed keys.
HTTP and SSH credentials may belong to different accounts; `auth info` reports the identity for the selected transport.
`BIBLEIT_TOKEN` applies only to HTTP. There is no automatic transport fallback.

The CLI stores preferences and saved credentials together at `~/.bibleit/config.yaml`.
`BIBLEIT_CONFIG` selects another configuration file. Unix permissions and Windows
ACLs restrict saved credentials to the current user. SSH private keys stay in their original files.

`bibleit auth` shows authentication help locally, even if saved configuration
cannot be read. Use `bibleit auth login` to start browser sign-in.

### Recover an older configuration

An older configuration can cause `unsupported configuration` or `unsupported
configuration version` errors. The CLI requires a fresh sign-in rather than
silently reusing credentials from an older format or a different service address.

For the default configuration on macOS/Linux, keep a protected backup and sign in:

```sh
mv -n ~/.bibleit/config.yaml \
  ~/.bibleit/config.yaml.backup.$(date +%Y%m%d-%H%M%S)
bibleit auth login
```

On Windows PowerShell:

```powershell
$config = Join-Path $HOME '.bibleit/config.yaml'
Move-Item -LiteralPath $config -Destination "$config.backup.$(Get-Date -Format yyyyMMdd-HHmmss)"
bibleit auth login
```

If `BIBLEIT_CONFIG` is set, back up that selected file instead. Keep the backup
private: it may contain credentials. SSH users can sign in with
`bibleit auth login ssh` and their registered key. Moving the local file does not
revoke an old token; manage obsolete credentials through your account.

### Sign out

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
| `organization list [cursor]`, `organization application <id> status` | Inspect memberships and private verification applications; requires organization capability |
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

HTTP uses versioned REST resources for quotas, token metadata, library changes,
book reads, Live settings/statistics, retained stacks and scoped bulk deletion.
Personal organization membership/application commands require personal credentials;
a service organization identity is not a membership listing. Personal library
writes use `bt_` authority and quotas; delegated `oa_` credentials do not gain a
write scope from read consent. Server and SDK support for these operations is
required; see the [implementation status](docs/cli-ux-plan.md).

`live <id> subscribe` watches REST snapshots/heartbeats over bounded long polls
with HTTP, and streams connection events with explicit SSH. Neither promises
replay. Ctrl+C stops watching. Custom Live secrets and ineffective standalone
connection-secret authentication are unsupported by this CLI. Secret creation or
rotation returns a one-time secret: use `--secret-output <new-private-file>` to
keep it out of terminal/pipeline output. The file must not already exist.

The CLI version, server version, and protocol version are independent.
Use `bibleit version` for the CLI and `bibleit server info` for the service.
Ordinary source builds report `dev`.

## Output and automation

The default output is a human-readable table. HTTP `--format raw` returns native
REST JSON; `--format json` wraps that data in a schema-version-2 envelope. SSH
retains protocol replies in raw mode and its schema-version-1 JSON envelope.
Subscriptions support `--format ndjson`: HTTP uses a schema-2 envelope around
the native REST event; SSH retains its schema-1 stream contract.
Errors go to stderr and display control characters safely.
See the [output contract](docs/output.md).

Stack pop accepts a positive count (default 1) and never prompts. Commands that
delete Lives, clear content, or rotate/delete
secrets require confirmation. Interactive prompts require a terminal;
automation must explicitly supply `--yes`:

```sh
bibleit --yes --format json live "$LIVE_ID" delete
```

Use this command only when you intend to delete the selected Live. Confirmation does
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

### Homebrew shell setup

Homebrew installs and links Bash, Zsh, and Fish completion files with the
executable. Your shell must load them; installing the formula does not edit
your shell startup files.

For Zsh, evaluate Homebrew's environment before initializing completion in
`~/.zshrc` (or earlier in `~/.zprofile`):

```zsh
eval "$(brew shellenv)"
autoload -Uz compinit
compinit
```

Oh My Zsh initializes completion itself. Place the `brew shellenv` line before
`source "$ZSH/oh-my-zsh.sh"`; a second `compinit` call is unnecessary. Open a new
terminal after changing the order. To enable Bibleit immediately in an already
initialized Zsh session, run `source <(bibleit completion zsh)`.

For Bash, add the following to your shell startup file (`~/.bashrc`, or
`~/.bash_profile` for a macOS login shell):

```bash
source "$(brew --prefix)/etc/bash_completion.d/bibleit"
```

Homebrew's Fish discovers installed completions automatically. Other Fish
installations can use the per-user Fish setup above. Press Tab after `bibleit `
or `bibleit auth ` to complete commands. Completion lists static commands and
flags; it does not retrieve account data or suggest credentials.
See [Homebrew's shell completion guide](https://docs.brew.sh/Shell-Completion)
for general shell setup and completion-directory troubleshooting.

## Development

Use the Go toolchain specified in [go.mod](go.mod). The CLI pins the Go client;
a sibling checkout is optional.
Use ordinary public Go module proxy and checksum verification to download it.

```sh
GOWORK=off go mod download
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

See [Build from source](#build-from-source) to compile and run the executable.
For joint client/CLI development, an uncommitted local workspace can include both
checkouts: `go work init . ../bibleit-go`.
Native code tests cover macOS, Linux and Windows. See
[CONTRIBUTING.md](CONTRIBUTING.md) for development checks and the
[changelog](CHANGELOG.md) for changes.

Contributions use Conventional Commits. See [contributing](CONTRIBUTING.md).
Licensed under [MIT](LICENSE). Binary distributions also include
[third-party notices](THIRD_PARTY_NOTICES); retain both files when redistributing
the CLI. The software license does not grant rights to translations returned
by the service.

For contributions and bug reports, see [CONTRIBUTING.md](CONTRIBUTING.md).
Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).

## Preferences and offline configuration

Save reusable defaults in `~/.bibleit/config.yaml`:

```yaml
version: 2
transport: ssh
identity: ~/.ssh/bibleit
format: table
```

Use `bibleit config show` to see the file location and effective defaults without
revealing credentials. Use `config get <key>`,
`config set <key> <value>`, and `config unset <key>`. Only `transport`, `identity`,
and `format` are accepted. `config unset identity` explicitly returns to OpenSSH
agent/config selection without reviving an old cached identity. Flags override
`BIBLEIT_TRANSPORT`, `BIBLEIT_IDENTITY`, and `BIBLEIT_FORMAT`, which override YAML
and then built-in defaults. Set NDJSON explicitly for a subscription rather than
as a default for finite commands. A successful login remembers its transport.

Login saves credentials in this same YAML file. `BIBLEIT_CONFIG` selects a single
alternative file for all configuration. Separate JSON credential files are ignored;
there is no automatic import or merge. Sign in again if you need to save an HTTP
credential in the YAML file. `config path` has been removed; use `config show`.
Unix credential files must be owned by the current user, regular and private
(`chmod 600`); writes use private permissions or a protected Windows DACL.
Credential files are protected plaintext, not an OS credential vault. No token
copies or service rebinding are performed. Help, version and completion work even
when configuration is broken. No transport fallback changes the selected account.

Use `organization list --cursor <cursor>` for the next personal directory page.
SSH organization commands require the `organization` server capability.

HTTP whole-book reads follow bounded pages automatically and still return the
complete book. To inspect a single page, use:

```sh
bibleit read bpml Salmos --limit 50 --format json
bibleit read bpml Salmos --limit 50 --cursor '<next_cursor>' --format json
```

`--limit` accepts 1–500 verses. `--cursor` alone uses the server's default page
size of 50. Keep the same account, translation and book when continuing; page
size may change. A 3 MiB response cap can make pages smaller than the requested
verse limit. The final page returns `next_cursor: null`. If the translation
edition changes, restart without `--cursor`. Chapter and verse reads, SSH, and
Live snapshots keep their existing behavior.
`account quota list` and `account token list` are aliases for `account quotas` and
`account tokens`; `live <id> clear` remains an alias for stack clear. Human output
uses user terminology; machine schemas keep their documented transport fields.
