---
title: 'Fix PR 17 CI Go Toolchain Drift'
type: 'bugfix'
created: '2026-09-21'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** PR #17 fails the backend generated-code drift check because CI runs Go 1.27.1 while the module and committed generated output use Go 1.25.11; the newer standard-library compressor rewrites only the embedded OpenAPI bytes.

**Approach:** Make CI derive its Go toolchain from `backend/go.mod` so generation and verification use the repository's declared version consistently.

</frozen-after-approval>

## Implementation Notes

- Updated the backend CI setup to read the Go version from `backend/go.mod`, eliminating the toolchain mismatch that rewrote only the generated embedded OpenAPI compression payload.
