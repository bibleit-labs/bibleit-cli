# Maintaining Bibleit CLI

Normal contributor setup and offline tests are in [CONTRIBUTING.md](../CONTRIBUTING.md).
This guide collects release, installation-validation and audit material.

## Releases and installation checks

SDK and CLI versions are independent. Update go.mod/go.sum together when changing
the pinned SDK. Source builds report `dev` unless a version is injected. Release
packaging uses the Go toolchain specified in go.mod and a clean committed checkout.

[Release maintenance](releases.md) describes clean standalone packaging, pinned
SDK/toolchain inputs, exact archive and notice checks, native Bash/PowerShell,
Homebrew and Scoop validation, and approval before new tags or publication.
[Native validation](native-validation.md) describes offline tests and fixture
isolation. Local service integration procedures are maintained privately.
Fixture checks alone do not establish anonymous public installation.

[Public launch evidence](public-launch.md) records verified installation coverage,
repository protections and the completed launch checklist. The
[latest release record](release-v0.0.5-readiness.md) binds the published inputs and
native installation results. Dated evidence applies only to its recorded revision.

README, notice, installer, completion or build-input changes require new packages
and matching tap hashes in a future approved release. Keep published tags and
assets immutable; a README edit on main does not update an existing package.

## Security, licensing and audit records

The [security and Git history audit](security-audit.md) describes scan coverage,
reproduction commands and classified findings. The [licensing audit](licensing-audit.md)
records software rights and required redistribution notices. These dated reviews
do not automatically qualify later source, dependency or data changes.
Keep raw credentials, private service data and unredacted audit output outside Git.
Report vulnerabilities through [SECURITY.md](../SECURITY.md).

[Historical compatibility records](maintenance/compatibility-history.md) retain
prior integration, deployment and release narratives moved out of the contract
README. [Implementation status](status.md) and [the original plan](plan.md) are
historical records; use current release evidence for completed publication gates.


## Current release evidence

[v0.0.4 publication evidence](release-v0.0.4-readiness.md) records the completed
approved launch: exact asset hashes, published latest stable release, merged tap,
all six anonymous native CLI targets and all four anonymous native Homebrew targets.
The private real Chrome blocked-login fixture also passed before publication.

[v0.0.5 evidence](release-v0.0.5-readiness.md) records the organization read-command release and its current installation validation results.
