# v0.3.0 release readiness

Source branch: `broadcast-studio-cli` (no branch prefix). Includes the current main
SDK v0.4.0 update and the Broadcast Studio integration originally staged at c0c6e30.

Completed source checks:

- Full Go 1.27.1 pinned-SDK race tests and `go vet ./...` passed.
- Disposable real-SSH staging passed 1 MiB upload, hash/revision export, CRUD,
  same-ticket retries/status, stale revision rejection and lifecycle reads.
- Server staging passed 31 integration and 41 unit/authentication checks,
  including synthetic-provider testing/live/end and resource cleanup.
- No dependency or redistribution-notice change beyond the already-merged SDK
  update; packaging retains LICENSE and THIRD_PARTY_NOTICES.

Publication gates still to record:

- Updated production REST deployment and token/permission smoke checks (the
  earlier production SSH permission/CRUD smoke check passed).
- Six-target candidate packaging and native installer workflow checks.
- Release source merged into main, fresh v0.3.0 tag and validated draft assets.
- Public release, anonymous archive retrieval, and distribution follow-up.

Production admission must remain Mittel-only and host/workspace publisher limits
must remain one. Physical Windows/USB Broadcast Studio tests stay deferred; native
CLI installer workflow checks are separate release packaging checks.

## REST and reusable SDK parity gate

The user expanded v0.3.0 to include REST, SSH and reusable Go/Lua/Java SDK parity.
The CLI source now includes HTTP broadcast controls with the same command grammar,
explicit revisions, private request files and status-before-retry behavior as SSH.
HTTP sends exact signed body bytes; it does not re-encode Setup JSON. SSH wire
construction now delegates to bibleit-go rather than maintaining a second grammar.

These additions are tested against an ignored temporary workspace containing the
SDK feature branch. Publication remains blocked until that SDK is publicly
released, the CLI pins the public version without replacements, and fresh
standalone race/vet/native packaging checks pass. Earlier v0.4.0 validation does
not qualify the new REST implementation for release.
