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

- Production server deployment and SSH permission/CRUD smoke checks.
- Six-target candidate packaging and native installer workflow checks.
- Release source merged into main, fresh v0.3.0 tag and validated draft assets.
- Public release, anonymous archive retrieval, and distribution follow-up.

Production admission must remain Mittel-only and host/workspace publisher limits
must remain one. Physical Windows/USB Broadcast Studio tests stay deferred; native
CLI installer workflow checks are separate release packaging checks.
