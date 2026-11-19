# v0.3.0 release readiness

Source branch: `broadcast-studio-cli`. The candidate pins the published
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
- SDK MIT license text is unchanged; the notice bundle now records v0.5.0.
  Other dependency versions, Go 1.27.1 and the six CGO-disabled targets are unchanged.
- Go, Lua and Java SDKs share the canonical broadcast operation/model contract.
  The three SDK releases precede CLI publication.

Frozen candidate validation uses the existing three-OS native workflow and
six-target installer workflow, including both Windows PowerShell versions,
Scoop and notice/checksum retention. Review the workflow results and immutable
candidate manifest for the exact source commit; earlier v0.4.0/SSH-only candidate
results do not qualify a new source build for publication.

Remaining release gates:

- Production REST deployment and scoped token/permission smoke checks, followed
  by the server release sign-off (the prior production SSH CRUD smoke passed).
- Green native and installer validation for the reviewed final source.
- Source merged normally to main, a fresh immutable v0.3.0 tag and validated
  draft assets; publication requires the coordinated release greenlight.
- Anonymous CLI archive/install verification and distribution follow-up after
  publication.

Physical Windows/USB Broadcast Studio tests remain deferred; native CLI installer
and SDK runtime validation are separate release checks. No CLI tag or release
publication is authorized by candidate packaging alone.
