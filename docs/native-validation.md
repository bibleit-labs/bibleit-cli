# Development and compatibility validation

Use Go 1.27.1. Normal CLI and SDK unit tests run without the private server or a
production account. See [CONTRIBUTING.md](../CONTRIBUTING.md) for standalone and
optional workspace setup. Go modules must be authenticated while the SDK is private.

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
python3 scripts/check_server_container.py --baseline contract/server-baseline.json
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
Homebrew and Scoop checks. Corrected v0.0.2 is staged privately. Windows
PowerShell 7/5.1 and Scoop checks passed on final package candidate f0c080c;
validate the final approved artifacts before release. A build or Linux PowerShell
parser test alone does not close those gates. Native code tests passed
on the three hosted platforms at the recorded candidate source; current artifact
and final native results are in [release readiness](release-v0.0.2-readiness.md).

Dated implementation and deployment records remain in [historical status](status.md).
They establish only their recorded revision/scope, not a new validation claim.
