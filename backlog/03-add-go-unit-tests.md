# Add Go Unit Tests

**Priority:** MEDIUM
**Labels:** `testing`, `enhancement`, `medium-priority`

## Summary
Add comprehensive Go unit tests to complement existing bash integration tests. Currently, the project has no `*_test.go` files.

## Current Test Coverage
- ✅ Bash integration tests (22 tests in `tests/` directory)
- ❌ Go unit tests (0 tests)

## Proposed Test Coverage Areas

### High Priority
- [ ] `saggy/keygen.go` - Key generation logic
- [ ] `saggy/keys.go` - Key loading and parsing
- [ ] `saggy/encrypt.go` - Encryption logic
- [ ] `saggy/decrypt.go` - Decryption logic
- [ ] `saggy/SafeWholeFileIO.go` - Safe file I/O operations

### Medium Priority
- [ ] `saggy/cli.go` - CLI argument parsing
- [ ] `saggy/with.go` - Command execution with decrypted secrets
- [ ] `saggy/utils.go` - Utility functions
- [ ] `saggy/error.go` - Error handling

### Test Types Needed
- Unit tests for individual functions
- Table-driven tests for edge cases
- Error handling tests
- Integration tests for end-to-end workflows

## Benefits
- Faster feedback during development
- Better code coverage visibility
- Easier refactoring with confidence
- Catch regressions early

## Priority
MEDIUM - Improves development velocity and code quality
