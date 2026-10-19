# Service endpoint and credential policy

REST resources target the pinned `https://api.bibleit.app/v1` service. Browser
approval, CLI code exchange and logout use `https://bibleit.app`. SSH uses
`bibleit.app:22`. Repository/download access and service authorization are separate.

The SDK declares the REST endpoint in `rest/client.go` and the browser origin in
`client.go`. HTTP resource clients use `rest.NewClient`; the parent credential
client supports browser-issued token revocation only. Command execution requires
an explicit SSH executor. There is no command-style HTTP transport or REST adapter.
Custom HTTP clients change transport/timeouts while retaining the fixed origin,
TLS verification and blocked redirects. The SDK does not load environment variables.

The CLI defaults to HTTP; `--transport ssh` selects OpenSSH explicitly.
`BIBLEIT_TOKEN` supplies an HTTP token; it is not an SSH credential. SSH identity
selection and host verification use OpenSSH, including the user's SSH config.
The CLI exposes no endpoint/hostname overrides or named profiles.
OpenSSH verifies the service's host key directly under bibleit.app. The CLI
does not override HostKeyAlias. Existing trust entries for an older hostname
are not copied automatically; OpenSSH records a newly encountered service key
and rejects changes to a key already recorded for bibleit.app. Users with a
custom OpenSSH configuration should review any host aliases there as well.

Configuration schema version 2 binds saved credentials to both service addresses.
A mismatched address, legacy bearer-only config or older profile config fails
locally. Credentials are never silently copied to a new service. BIBLEIT_CONFIG
selects a credential file, not a destination. Keep an old file as a protected
backup if needed and authenticate into a fresh configuration after a migration.

Never disable TLS verification or SSH host-key checking to resolve a migration.
Confirm a new server fingerprint through a trusted project channel before
updating known_hosts. Source changes or custom network transports do not bypass
server authentication, permissions or quotas.

See [authentication](../README.md#authentication-and-service-address),
[SECURITY.md](../SECURITY.md), and [historical status](status.md) for dated migration
evidence. Historical records do not establish readiness for a new release.
