# v0.3.0 release readiness

Released source: [`v0.3.0`](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.3.0),
commit `3e23adf6fe10a118c71b3da5583b37ae9b41ee91`. The release pins the published
`github.com/bibleit-labs/bibleit-go v0.5.0` with public proxy/checksum verification.
No development workspace or local replacement is a release input.

The source supports the same Broadcast Studio controls through REST and SSH.
Current workspace roles remain authoritative; personal REST token scopes add a
credential ceiling. Browser capture and Google consent remain browser actions.
Production admission must remain Mittel-only, with one publisher per host and
workspace.

Completed checks:

- Standalone Go 1.27.1 `GOWORK=off` race tests, vet and build passed with the public
  SDK. Native macOS runtime, TLS, PTY, confirmation and completion checks passed.
- Disposable real REST and SSH rehearsal at `12da155` passed full 1 MiB input,
  create/get/export/save, stale revision rejection, duplicate/delete, retained
  request retries/status and lifecycle reads; REST also checked browser-required
  takeover. All temporary Setups and publisher/transfer resources were removed.
  The published SDK retains the tested transport contract and exact-byte helpers.
- Scoped production REST and SSH smoke checks passed after the server deployment:
  Setup CRUD, exact-input replay/status, revision-bound exports, stale and
  cross-workspace denial, and browser-required takeover. No event or publisher
  was started; owned Setups, temporary credential and private request files were
  cleaned up. Admission remained Mittel-only with both publisher caps at one.
- Default SSH lifecycle tables now show public fields instead of transport
  base64. Native JSON/raw envelopes retain their existing shape; malformed and
  oversized payloads fail without displaying their contents.
- SDK MIT license text is unchanged; the notice bundle now records v0.5.0.
  Other dependency versions, Go 1.27.1 and the six CGO-disabled targets are unchanged.
- Go, Lua and Java SDKs share the canonical broadcast operation/model contract.
  The three SDK releases precede CLI publication.

Frozen candidate validation uses the existing three-OS native workflow and
six-target installer workflow, including both Windows PowerShell versions,
Scoop and notice/checksum retention. Review the workflow results and immutable
candidate manifest for the exact source commit; earlier v0.4.0/SSH-only candidate
results do not qualify a new source build for publication.

Verified release evidence:

- [Final main native checks](https://github.com/bibleit-labs/bibleit-cli/actions/runs/44559973164)
  passed Linux, macOS and Windows. [Final main installer qualification](https://github.com/bibleit-labs/bibleit-cli/actions/runs/44559973291)
  passed all six targets.
- The annotated immutable `v0.3.0` tag targets the released source above. Its
  [release workflow](https://github.com/bibleit-labs/bibleit-cli/actions/runs/44560568088)
  passed the three native platforms and six installer targets before creating
  the draft. All 12 uploaded draft files matched the verified tag candidate.
- [CLI v0.3.0](https://github.com/bibleit-labs/bibleit-cli/releases/tag/v0.3.0)
  was published on 2026-11-19. Anonymous retrieval verified all 12 public assets,
  all eight advertised checksums, source/toolchain and public SDK provenance.
- Actual explicit-version and latest online installs returned `bibleit CLI
  version 0.3.0` and working broadcast help. They used a private temporary prefix
  without credentials, leaving the existing user CLI and configuration untouched.
- The tag's eight archives/installers exactly match the candidate used for the
  production read-only SSH table smoke. Preflight, channel and output profiles
  were readable, while schema-version-1 JSON and raw output remained compatible.
- Machine-readable results are recorded in
  [the v0.3.0 release evidence](releases/v0.3.0-integration.json).

The maintained Homebrew tap has a separate formula/public-download qualification
and normal merge gate. Its distribution cutover does not replace the checks above.

Physical Windows/USB Broadcast Studio tests remain deferred; native CLI installer
and SDK runtime validation are separate release checks. The published tag and assets remain immutable; future code or packaged-content
changes require a new release.
