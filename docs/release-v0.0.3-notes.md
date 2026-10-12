# v0.0.3

This release improves local help and installation guidance.

- Bare `bibleit auth` shows help without loading credentials, including unsupported legacy configurations.
- Document Homebrew completion setup for Bash, Zsh and Fish, and safe backup/recovery of legacy configuration.
- Update the CLI description and packaged README to describe Bibleit directly.
- Verify SSH host keys for `bibleit.app` directly. Existing trust saved only under the old hostname is not reused; OpenSSH may record the first key for the current hostname under the existing `accept-new` policy. Changed known keys remain rejected. Never disable host-key checking to migrate.

All six archives include the current README, shell completions, LICENSE and THIRD_PARTY_NOTICES. The SDK remains pinned to v0.0.1 with Go 1.27.1. Existing releases are unchanged.
