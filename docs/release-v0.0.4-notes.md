# v0.0.4

Browser sign-in now exits promptly when authorization is denied, including a
clear message when the account is blocked. Denied sign-in does not exchange a
token or change saved credentials. The browser service must deliver the denial
callback; account access remains enforced by the service for every client version.

Callbacks must carry the original authentication state. Unrelated callbacks and
invalid-state requests cannot terminate a legitimate login. Error descriptions
from callback input are not reflected into terminal or browser output.

The README now includes the source-build command and keeps internal validation
instructions in maintainer documentation. Homebrew completion and legacy config
recovery guidance remain available.

The SDK remains pinned to v0.0.1 with Go 1.27.1. All six archives include the README,
shell completions, LICENSE and THIRD_PARTY_NOTICES. Existing releases are unchanged.
