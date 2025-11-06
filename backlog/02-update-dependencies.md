# Update Dependencies to Latest Versions

**Priority:** MEDIUM
**Labels:** `dependencies`, `maintenance`, `medium-priority`

## Summary
Update project dependencies to their latest stable versions to benefit from bug fixes, security patches, and improvements.

## Current Versions vs Latest

| Package | Current | Latest | Gap |
|---------|---------|--------|-----|
| filippo.io/age | v1.2.0 | v1.2.1 | Patch update |
| golang.org/x/crypto | v0.24.0 | v0.43.0 | Major update |
| golang.org/x/sys | v0.21.0 | v0.37.0 | Major update |

## Update Commands
```bash
go get filippo.io/age@v1.2.1
go get golang.org/x/crypto@latest
go get golang.org/x/sys@latest
go mod tidy
```

## Testing Requirements
- [ ] Run full test suite after updates
- [ ] Verify bundled age functionality still works
- [ ] Test all CLI commands (keygen, encrypt, decrypt, with)
- [ ] Ensure backward compatibility with existing encrypted files

## Notes
- Age v1.2.1 is a patch release (likely safe)
- golang.org/x/crypto and golang.org/x/sys are typically backward compatible despite version jumps
- All updates should be tested before merging

## Priority
MEDIUM - Important for security and stability, but not blocking core functionality
