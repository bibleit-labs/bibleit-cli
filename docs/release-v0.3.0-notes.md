# Bibleit CLI v0.3.0

Broadcast Studio now has matching HTTP and SSH controls for an explicit workspace.
Operators can list/export/create/save/duplicate saved Setups, inspect the connected
YouTube channel and events, and request testing, live, end and publisher stop.
Owners/admins can also delete saved Setups. Source capture and Google authorization
continue to require Studio in the browser.

HTTP commands require a personal `bt_` token with the relevant broadcast scope;
SSH commands use registered personal keys. Both retain existing workspace role
checks. SSH negotiates read/write/lifecycle capabilities and reports an actionable
error on older or unconfigured servers. Setups use revisions and private retained request
files; retries reconcile the original signed ticket before uploading again. Event
controls require retained operation UUIDs. One SSH connection is reused throughout
an invocation, including multipart export and upload.

Release publication is pending a published bibleit-go version containing the
broadcast contract; the current source uses the sibling SDK only for local
integration checks. The final release must pin that public version and pass
standalone `GOWORK=off` validation. Existing output schemas remain compatible. Packages include macOS, Linux and Windows amd64/arm64 executables,
completions, installers, checksums and required redistribution notices.

```sh
bibleit --transport ssh --identity ~/.ssh/bibleit broadcast <workspace-uuid> preflight
bibleit --transport ssh --identity ~/.ssh/bibleit broadcast <workspace-uuid> setup list
```

Production workspace admission and publisher capacity remain server policy. This
release does not grant workspace membership, expose ingest credentials, start
capture automatically, or bypass the existing one-publisher production cap.
