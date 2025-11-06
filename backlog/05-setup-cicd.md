# Setup CI/CD for Automated Builds

**Priority:** MEDIUM
**Labels:** `ci-cd`, `infrastructure`, `medium-priority`

## Summary
Implement CI/CD pipeline to automatically build binaries and publish releases. This is listed as a TODO in the README.

## Current State
- ❌ No CI/CD pipeline
- ❌ No automated builds
- ❌ No binary releases on GitHub

## Proposed Implementation

### GitHub Actions Workflow
Create `.github/workflows/release.yml` to:
- [ ] Build binaries for multiple platforms (linux, darwin, windows)
- [ ] Build for multiple architectures (amd64, arm64)
- [ ] Run all tests before building
- [ ] Create GitHub releases automatically on version tags
- [ ] Upload binaries as release assets

### Platforms to Support
- Linux (amd64, arm64)
- macOS/Darwin (amd64, arm64)
- Windows (amd64)

### Additional Workflows
- [ ] `.github/workflows/test.yml` - Run tests on every PR
- [ ] `.github/workflows/lint.yml` - Run linters (golangci-lint)

## Benefits
- Users can download pre-built binaries
- Automated quality checks on every PR
- Consistent build environment
- Easy distribution

## Reference
README.md lines 37-38:
```
// TODO: Use CI to generate a binary
// TODO: publish binaries to releases
```

## Priority
MEDIUM - Needed for v1.0 release and user adoption
