# Saggy Project Backlog

This directory contains issue templates and backlog items for the Saggy project.

## Issue Priority Guide

- **HIGH** - Blockers for v1.0 release, critical functionality
- **MEDIUM** - Important improvements, should be done before v1.0
- **LOW** - Nice to have, can be deferred to post-v1.0

## Current Backlog Items

1. **[Complete SOPS Integration](01-complete-sops-integration.md)** - HIGH
   - Complete bundled dependencies to eliminate external sops requirement
   - Blocker for single-binary distribution

2. **[Update Dependencies](02-update-dependencies.md)** - MEDIUM
   - Update age, crypto, and sys packages to latest versions
   - Important for security and bug fixes

3. **[Add Go Unit Tests](03-add-go-unit-tests.md)** - MEDIUM
   - Add comprehensive Go unit tests (currently 0 exist)
   - Improves development velocity and code quality

4. **[Update Go Version](04-update-go-version.md)** - LOW
   - Update minimum Go requirement to 1.23+
   - Will be addressed with SOPS integration

5. **[Setup CI/CD](05-setup-cicd.md)** - MEDIUM
   - Implement automated builds and releases
   - Listed in README TODOs

6. **[Passphrase Support](06-passphrase-support.md)** - LOW
   - Add support for passphrase-protected keys
   - Security enhancement, not critical

## Creating GitHub Issues

To create these as GitHub issues, you can:

1. **Via GitHub Web UI:**
   - Go to your repository's Issues page
   - Click "New Issue"
   - Copy and paste the content from each markdown file
   - Add the appropriate labels

2. **Via GitHub CLI (if available):**
   ```bash
   gh issue create --title "Complete SOPS Integration" --body-file backlog/01-complete-sops-integration.md --label "enhancement,high-priority,bundled-dependencies"
   ```

## Project Status Summary

- **Current Version:** v0.8.0 (pre-v1.0)
- **Completion:** ~60-70% toward v1.0
- **Critical Path:** Complete SOPS integration (#1) → Update dependencies (#2) → CI/CD (#5)
- **Test Status:** 3/22 tests passing in bundled mode (86% failing due to missing SOPS)
