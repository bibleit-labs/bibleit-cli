# Development and compatibility validation

Use Go 1.27.1. Normal CLI and SDK unit tests run without the private server or a
production account. See [CONTRIBUTING.md](../CONTRIBUTING.md) for standalone and
optional workspace setup. Download the public SDK through normal Go module proxy
and checksum verification.

## Offline checks

```sh
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off go build -o build/bibleit .
python3 scripts/check_cli_runtime.py --binary build/bibleit
```

The runtime helper exercises local process/output/completion behavior using
synthetic fixtures. Linux can add `--shells bash zsh fish`; install those shells
first. Native CI exercises Linux, macOS and Windows, including credential-file
protection and terminal behavior. Only run tests relevant to the change:
documentation changes need link/command/content review, not a new live-service run.

Shared contract fixtures are versioned snapshots. The SDK uses in-memory
transports; the CLI's test transport allows only the production hostname and
routes it to isolated loopback TLS with a fixture-only CA. Fixture trust is absent
from production executables and is never installed in the host trust store.
Passing these tests does not prove production availability or provider behavior.

## Optional private-server integration

The server is a separate private repository. Authorized maintainers with Docker
and the sibling SDK/server checkouts can run the isolated container harness:

```sh
python3 scripts/check_server_container.py --server /path/to/private-service --baseline contract/server-baseline.json
```

Use `--imports` when import coverage is relevant. The harness copies the frozen
server snapshot, rebuilds native dependencies and uses a disposable database and
network namespace. It records source/native fingerprints and checks that inputs stay stable during
the run. The --baseline argument writes the successful run's provenance to the
selected file; review that diff before committing it. Read its options before
running against changed server inputs. Ordinary contributors do not need this harness for unit tests.

Integration targets the fixed service name inside the isolated namespace; it does
not add a customer endpoint override. Record the exact tested baseline and scoped
operations. Fixture, private integration and genuine provider/public-DNS checks
are separate evidence. [Contract coverage](../contract/features.md) records the
client's supported operations and limits.

## Package installation validation

[Release maintenance](releases.md) owns archive/checksum, Bash/PowerShell,
Homebrew and Scoop checks. Validate the exact committed inputs and approved
artifacts before publication; a build or Linux PowerShell parser test alone does
not establish native installation behavior. [Public launch evidence](public-launch.md)
and [the latest release record](release-v0.0.3-readiness.md) contain completed
native installation results and their source/artifact hashes.

Dated implementation and deployment records remain in [historical status](status.md).
They establish only their recorded revision/scope, not a new validation claim.
