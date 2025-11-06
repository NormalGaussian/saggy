# Update Go Version Requirement

**Priority:** LOW
**Labels:** `dependencies`, `low-priority`

## Summary
Update the minimum Go version requirement in `go.mod` from 1.22.2 to 1.23.0 or later.

## Rationale
- Latest SOPS v3 library requires Go 1.23.0
- Current system has Go 1.24.7 installed
- Project is already using Go 1.22.2 (outdated)
- Issue #1 (bundled SOPS integration) will require Go 1.23+

## Changes Required
Update both go.mod files:
- `/go.mod` - change `go 1.22.2` to `go 1.23`
- `/saggy/go.mod` - change `go 1.22.2` to `go 1.23`

## Testing
- [ ] Verify project builds with Go 1.23
- [ ] Verify project builds with Go 1.24
- [ ] Run all tests
- [ ] Update CI/CD pipelines (when implemented)

## Dependencies
This should be done as part of Issue #1 (SOPS integration) since the SOPS library requires Go 1.23.

## Priority
LOW - Will be addressed naturally when completing SOPS integration
