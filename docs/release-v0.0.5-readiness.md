# v0.0.5 release evidence

The organization read commands use the published, checksummed Go SDK v0.0.2.
The production server advertises `organization_status`. Older servers receive
upgrade guidance before the CLI sends an unsupported organization command.

Frozen CLI source: `562798043308e3f1f6f2eafc4ebdad52ed6758d3`, merged through
[PR #25](https://github.com/bibleit-labs/bibleit-cli/pull/25). Go 1.27.1 packages
use no workspace or replacement modules. All six executable headers and twelve
asset hashes are recorded in [the validation record](security/v0.0.5-release-validation.json).

Native Linux, macOS and Windows race tests, vet, builds and process checks passed.
The [guarded release workflow](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38328017189)
is checking packaging and all six native installer targets. Draft asset verification,
publication, anonymous installation and Homebrew validation remain pending.
