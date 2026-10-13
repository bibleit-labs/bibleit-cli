# Changelog

## v0.0.5 — 2026-10-13

- Add organization list/application status commands, capability discovery, help and shell completion. Preserve pending/deleted application state and map pending/suspended organization denials to forbidden exit 4.
- Organization membership is invitation-only; verification requests remain browser-managed and approval remains in Bibleit admin.

## v0.0.4 — 2026-10-12

- Stop browser sign-in promptly when authorization is denied, with a clear
  blocked-account message and no token exchange or credential changes.
- Require the original authentication state for denial callbacks; ignore
  unrelated or invalid-state requests while the legitimate login continues.
- Add source-build instructions and keep maintainer validation commands in
  maintainer documentation rather than the README.

## v0.0.3 — 2026-10-12

- Show authentication help for bare `bibleit auth` before reading saved credentials.
- Explain Homebrew completion initialization and recovery from legacy configuration.
- Use Bibleit service descriptions and withhold private repository references from
  current client documentation and historical public snapshots.
- Verify SSH host keys directly for `bibleit.app`, without a legacy host-key alias.
  Existing trust under an older hostname is not reused. Changed keys remain rejected.
- Require an explicit `--server` checkout path for optional maintainer integration.

## v0.0.2 — 2026-10-12

- Include SDK, Go runtime/standard-library and embedded third-party notices in
  all six archives, and the project MIT notice in both downloadable installers.
- Preserve LICENSE and THIRD_PARTY_NOTICES during Bash/PowerShell, Homebrew and
  Scoop installation. New script installers reject packages missing notices,
  including v0.0.1 archives, before replacing an existing executable.
- Verify notice contents through native installation/reinstallation checks.

Published after owner approval. Its assets are immutable; new source and README
changes belong to a future release. See [launch evidence](docs/public-launch.md).

## v0.0.1 — 2026-10-11

Initial Bibleit CLI release, using bibleit-go v0.0.1 and Go 1.27.1.

- Fixed production endpoints: HTTPS `https://bibleit.app` and SSH `bibleit.app:22`.
- API-token authentication, browser login/revocation and local credential storage;
  no endpoint overrides or profiles. Existing configurations tied to earlier
  endpoints require fresh authentication; see docs/releases.md for migration.
- Translation metadata/catalogues, identity, reading/search, Live management and
  SSH subscriptions, with stable text/JSON output.
- Bash, Zsh and Fish completion and destructive-command confirmation.
- Six binary archives: macOS, Linux and Windows, each amd64/arm64; Bash and
  PowerShell installers, checksums, provenance, Homebrew formula and Scoop manifest.

SSH commands require an installed OpenSSH client and verified host trust. Never
disable host-key checks to migrate an older configuration. HTTP subscriptions
require SSH. Requests reject embedded double quotes/control characters and
custom Live SECRET SET is unsupported. Catalogue numbering preserves verse gaps.

The release workflow validates native tests and the exact archives on Linux,
macOS and Windows, including PowerShell 7/5.1, Homebrew and Scoop installation.
Binaries have SHA-256 integrity checks but are not platform-signed or notarized.
Keep normal operating-system protections enabled.

Initial preparation was private. The notice-deficient v0.0.1 downloads were
withdrawn after owner approval; its source/tag remain. Public installation uses
the corrected release recorded in [launch evidence](docs/public-launch.md).

Production deployment/provider acceptance is complete at server3359535/native
d6e31dc/image0bbb. Earlier full HTTP/SSH/import and focused catalogue compatibility
checks retain their original tested revisions in contract/.
