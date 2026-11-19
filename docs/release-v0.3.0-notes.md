# Bibleit CLI v0.3.0

Broadcast Studio now has an SSH command family for an explicit workspace.
Operators can list/export/create/save/duplicate saved Setups, inspect the connected
YouTube channel and events, and request testing, live, end and publisher stop.
Owners/admins can also delete saved Setups. Source capture and Google authorization
continue to require Studio in the browser.

Commands use registered personal SSH keys and existing server workspace permissions.
They negotiate read/write/lifecycle capabilities and report an actionable error on
older or unconfigured servers. Setups use revisions and private retained request
files; retries reconcile the original signed ticket before uploading again. Event
controls require retained operation UUIDs. One SSH connection is reused throughout
an invocation, including multipart export and upload.

The release pins bibleit-go v0.4.0. Existing HTTP commands and output schemas stay
compatible. Packages include macOS, Linux and Windows amd64/arm64 executables,
completions, installers, checksums and required redistribution notices.

```sh
bibleit --transport ssh --identity ~/.ssh/bibleit broadcast <workspace-uuid> preflight
bibleit --transport ssh --identity ~/.ssh/bibleit broadcast <workspace-uuid> setup list
```

Production workspace admission and publisher capacity remain server policy. This
release does not grant workspace membership, expose ingest credentials, start
capture automatically, or bypass the existing one-publisher production cap.
