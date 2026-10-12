# v0.0.4 publication evidence

Published 2026-10-12 after owner approval of actions 1–6. The live CLI and tap
now use **v0.0.4**. The release title is exactly **v0.0.4**. Earlier tags and
asset identities, sizes and digests were verified unchanged.

## Frozen inputs

- CLI source: `df1904650266cac7d15fad254589f0e729f9ef22`, merged through protected main.
- Go 1.27.1; pinned, checksummed public SDK `github.com/bibleit-labs/bibleit-go v0.0.1`.
- Twelve clean standalone assets; all six executable headers match their targets.
- GitHub and local packages match byte-for-byte. Every archive includes the exact
  README, completions, LICENSE and THIRD_PARTY_NOTICES.
- Tap PR #15 formula matches the verified candidate after stable-version normalization.
  Native tap validation used `faa603f4c3e7ffecbc7c7ea0d4f8f3a30ae8f720`; evidence-only
  updates preserve the formula, workflow and helper bytes.

[Candidate 38275552347](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38275552347)
and [tap validation 38276244465](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38276244465)
completed successfully. [Tap PR #15](https://github.com/bibleit-labs/homebrew-tap/pull/15)
was merged as `370620ac2728cdd4c9d2eabdae3f0058d2d723e9`. [The validation record](security/v0.0.4-release-validation.json)
contains all twelve asset hashes, executable hashes/headers, artifact identity,
formula/helper hashes and native reports.

## Completed gates

| Gate | Targets | Result |
| --- | --- | --- |
| Native race, build, process and vet CI | Linux, macOS, Windows | Passed |
| Script installers, replacement and failure paths | Linux/macOS/Windows AMD64 and ARM64 | Passed |
| PowerShell 7 and 5.1; Scoop lifecycle and wrong-hash rejection | Windows AMD64/ARM64 | Passed |
| Homebrew install/test/reinstall/uninstall, notices, completions and legacy config preservation | macOS ARM64/Intel; Linux AMD64/ARM64 | Passed |
| Real-browser blocked-login notification | Headless Chrome, fresh profile and disposable accounts | Passed |

The browser fixture tested both an existing blocked session and blocked email
login during a pending CLI flow. The browser received HTTP 403 and followed the
denial callback; the CLI exited without requesting a token or creating a config.
Private service enforcement was separately verified using legacy wire requests,
SSH revocation and login-race tests. Implementation and operational evidence
remain private. No production accounts were changed, and no existing local server
was restarted. Chrome evidence does not establish Firefox/Safari compatibility.

## Completed launch actions

1. Created the annotated protected tag `v0.0.4` at the frozen source above.
   [Release workflow 38280448836](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38280448836)
   passed native tests and all six candidate installer targets before creating the draft.
2. Downloaded all twelve draft assets. Every hash and the exact provenance manifest
   matched the approved candidate, including GitHub's asset digests.
3. Published [v0.0.4](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.0.4)
   as latest stable with the reviewed title and notes.
4. Downloaded all twelve assets anonymously, verified their hashes and anonymous
   latest-release resolution, then merged tap PR #15 with required CI passing and
   the approved formula/helper bytes intact.
5. [Anonymous CLI run 38281676238](https://github.com/bibleit-labs/bibleit-cli/actions/runs/38281676238)
   passed Linux/macOS/Windows AMD64 and ARM64, using latest and explicit release URLs,
   with both PowerShell 7 and 5.1 on Windows.
   [Anonymous Homebrew run 38281509948](https://github.com/bibleit-labs/homebrew-tap/actions/runs/38281509948)
   passed macOS ARM64/Intel and Linux AMD64/ARM64 at standard prefixes with empty
   download caches, covering install/test/reinstall/uninstall, notices and completions.
6. Downloaded all ten reports, independently compared installed binary hashes with
   the approved executable evidence, and recorded release/validation artifact identities.
   The private service repository remains private; no production accounts were changed
   and no existing local server was restarted.

The first anonymous CLI run, 38281506837, failed on all six targets after installation
because the evidence record lacked the checker's `binary_sha256` lookup field.
The existing approved executable hashes were supplied in that field and independently
verified against extracted anonymous archives. All six subsequent checks passed;
release assets and validation helpers were unchanged.

All six approved launch actions are complete. Signing/notarization, a public Scoop
bucket and Firefox/Safari browser compatibility remain outside this release's scope.
