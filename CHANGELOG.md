# Changelog

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

These initial releases remain private. Ordinary curl/PowerShell/Homebrew download
URLs cannot authenticate to private release assets; authorized validation uses
GitHub CLI downloads and local installer inputs. Public customer installation
awaits a separate visibility or authenticated-delivery decision.

Historical service deployment evidence is maintained privately. Public client
compatibility fixtures remain in contract/.