# Service endpoint and credential policy

The Go client and CLI target the project-managed service `https://bibleit.app`.
SSH uses `bibleit.app:22`. Repository/download access and service authorization
are separate. The server repository remains private.

The SDK declares ServerEndpoint in client.go. HTTP commands, browser approval,
code exchange and revocation use that origin. HTTPS verification remains enabled
and redirects are blocked. NewClient(apiToken) validates explicit credentials;
NewClientWithHTTPClient customizes transport/timeouts without an endpoint option.
The library does not read environment variables or CLI configuration.

The CLI defaults to HTTP; `--transport ssh` selects OpenSSH explicitly.
`BIBLEIT_TOKEN` supplies an HTTP token; it is not an SSH credential. SSH identity
selection and host verification use OpenSSH, including the user's SSH config.
The CLI exposes no endpoint/hostname overrides or named profiles.
For host trust, the SSH adapter uses HostKeyAlias=legacy.bibleit.example to reuse
the verified legacy known_hosts entry; the network destination remains
bibleit.app:22. The alias is a trust lookup, not a destination override.

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
