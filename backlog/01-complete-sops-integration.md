# Complete SOPS Integration for Bundled Dependencies

**Priority:** HIGH
**Labels:** `enhancement`, `high-priority`, `bundled-dependencies`

## Summary
Complete the bundled dependencies migration by integrating the SOPS Go library directly into the binary. Currently, `age` is bundled but `sops` still requires an external binary, causing 86% of tests to fail in bundled mode.

## Current State
- ✅ Age is bundled (keygen works)
- ❌ SOPS is NOT bundled (encrypt/decrypt fail with "sops: executable file not found")

## Goal
Make Saggy a truly standalone single binary that doesn't require external `age` or `sops` installations.

## Implementation Tasks
- [ ] Add `github.com/getsops/sops/v3` as a dependency
- [ ] Refactor `saggy/encrypt.go` to use SOPS library instead of exec calls
- [ ] Refactor `saggy/decrypt.go` to use SOPS library instead of exec calls
- [ ] Refactor `saggy/with.go` to use SOPS library for encryption/decryption operations
- [ ] Update tests to verify bundled SOPS functionality
- [ ] Ensure all 22 integration tests pass with `SAGGY_USE_BUNDLED_DEPENDENCIES=true`

## References
- SOPS Go library: https://pkg.go.dev/github.com/getsops/sops/v3
- Current test results: 3/22 passing in bundled mode
- Related files: `saggy/encrypt.go:76` (current error location)

## Priority
HIGH - This is a blocker for v1.0 release and single-binary distribution
