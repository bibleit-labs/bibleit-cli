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

## Parallel CI scheduling

Native validation runs once per pull request update and on pushes to `main`.
Feature-branch pushes do not duplicate the PR checks, and tags use the release
workflow's native validation. A new PR update cancels that PR's older native run;
manual, main and release runs use separate concurrency groups.

The three native platform jobs can run simultaneously. Candidate and public
installer matrices allow all six platforms to run simultaneously, with no shared
concurrency group between workflows. GitHub runner availability and account
limits can still queue jobs. Release packaging waits for native tests, installer
checks wait for their archives, and publication waits for all required checks.

[Contract coverage](../contract/features.md) records supported client operations
and limits. Maintainers keep local service integration procedures in private
operations documentation; they are not part of public contributor setup.

## Package installation validation

[Release maintenance](releases.md) owns archive/checksum, Bash/PowerShell,
Homebrew and Scoop checks. Validate the exact committed inputs and approved
artifacts before publication; a build or Linux PowerShell parser test alone does
not establish native installation behavior. [Public launch evidence](public-launch.md)
and [the latest release record](release-v0.0.3-readiness.md) contain completed
native installation results and their source/artifact hashes.

Dated implementation and deployment records remain in [historical status](status.md).
They establish only their recorded revision/scope, not a new validation claim.
