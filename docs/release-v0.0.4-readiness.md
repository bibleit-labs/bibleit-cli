# v0.0.4 preparation

This is a candidate, not a published release. Publication and the live tap update
require owner approval. The planned release title is exactly **v0.0.4**.

The blocked-login fix is merged through PR #21 at
`a3f6f0037c52426567eac7e7538f9709e255dd4c`. Native Linux, macOS and Windows race,
process, build and vet checks passed. Private service validation covers legacy
HTTP requests and token exchange, browser denial, SSH revocation and login races;
its implementation and operational evidence stay in private maintainer docs.

Before approval, freeze a clean source commit, run the v0.0.4 six-target candidate,
record the artifact identity and all twelve asset hashes, verify the generated tap
formula, and complete native installation checks on all supported architectures.
Keep the staged tap formula unmerged until the published assets match it.

After approval: tag the frozen source, validate the draft assets against the
approved candidate, publish with the reviewed notes, verify anonymous downloads,
merge the exact tap update, and run anonymous installation checks. Do not replace
older tags or assets. See [release notes](release-v0.0.4-notes.md).
