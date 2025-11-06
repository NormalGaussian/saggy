# Add Support for Keys with Passphrases

**Priority:** LOW
**Labels:** `enhancement`, `security`, `low-priority`

## Summary
Add support for age keys protected with passphrases. Currently listed as a missing feature in README.

## Current State
From README.md:
> Support keys with passphrases
> - Saggy doesn't currently support asking for a passphrase to decrypt a key. This is wholly untested.

## Requirements
- [ ] Detect if a key file is passphrase-protected
- [ ] Prompt user for passphrase when needed
- [ ] Support environment variable for passphrase (for CI/CD)
- [ ] Handle passphrase securely (don't log, clear from memory)
- [ ] Update documentation with passphrase usage

## Implementation Considerations
- Use `filippo.io/age` library's passphrase support
- Consider `golang.org/x/term` for secure password input
- Support both interactive and non-interactive modes

## Testing
- [ ] Test with passphrase-protected keys
- [ ] Test passphrase from stdin/env var
- [ ] Test incorrect passphrase handling
- [ ] Test timeout for passphrase entry

## Priority
LOW - Nice to have for security, but not blocking core functionality

## References
- README.md "Missing features" section
- "The path forwards" section
