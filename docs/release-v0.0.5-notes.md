Read organization memberships and your verification requests from the terminal:

- `bibleit organization list [cursor]` lists one directory page.
- `bibleit organization application <id> status` shows your request status.
- `bibleit server help organization` discovers the service's read commands.

The CLI checks service support before sending organization reads and explains how
to proceed when an older server lacks the capability. JSON output keeps schema
version 1 and retains application states, including history after an organization
is deleted. Pending and suspended organization denials use forbidden exit code 4.

Requests and corrections remain in the browser, membership is invitation-only,
and approval remains in Bibleit admin. Personal Bible/verse Live commands remain
available. This release uses the separately published Go SDK v0.0.2.
