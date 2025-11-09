# Saggy Specification

## Overview

`saggy` is a secure secrets management tool that uses SOPS and age encryption to manage environment-specific configuration files in version control.

---

## Directory Structure

```
project-root/
├── .saggy/
│   ├── config.yaml              # Saggy configuration
│   ├── keys/                    # Public age keys (committed)
│   │   ├── ci-prod.pub
│   │   ├── ci-staging.pub
│   │   ├── alice.pub
│   │   └── bob.pub
│   └── secrets/                 # Encrypted secrets (committed)
│       ├── prod.env.sops
│       ├── staging.env.sops
│       ├── dev.env.sops
│       └── database.json.sops
├── .saggy.yaml                  # SOPS configuration (committed)
└── .gitignore                   # Should ignore private keys
```

**Private keys location** (NOT in repo):
- `~/.config/saggy/keys/<name>.key` (Linux/macOS)
- `%APPDATA%/saggy/keys/<name>.key` (Windows)
- Or via environment variable: `SAGGY_PRIVATE_KEY_PATH`

---

## Commands

### 1. `saggy init`

**Purpose**: Initialize current directory as a saggy repository.

**Behavior**:
```bash
saggy init [--key-dir <path>]
```

**Actions**:
1. Create `.saggy/` directory
2. Create `.saggy/keys/` subdirectory
3. Create `.saggy/secrets/` subdirectory
4. Generate `.saggy/config.yaml` with defaults
5. Generate `.saggy.yaml` (SOPS configuration template)
6. Update/create `.gitignore` to include:
   ```
   # Saggy private keys (NEVER COMMIT)
   *.key
   **/*.key
   .saggy/keys/*.key
   
   # Saggy temporary files
   .saggy/.tmp/
   ```
7. Print initialization success message with next steps

**`.saggy/config.yaml` template**:
```yaml
version: 1
default_key_dir: ~/.config/saggy/keys
age_recipients_file: .saggy/keys/
secrets_dir: .saggy/secrets
```

**`.saggy.yaml` template** (SOPS configuration):
```yaml
creation_rules:
  # Default: use all public keys in .saggy/keys/
  - path_regex: .saggy/secrets/.*\.sops$
    age: >-
      age1public_key_1,
      age1public_key_2
```

**Flags**:
- `--key-dir <path>`: Custom location for private keys (default: `~/.config/saggy/keys`)

**Exit codes**:
- `0`: Success
- `1`: Already initialized (`.saggy/` exists)
- `2`: Not in a git repository (warning, continues anyway)

**Example output**:
```
✓ Initialized saggy in /path/to/project
✓ Created .saggy/ directory structure
✓ Generated .saggy/config.yaml
✓ Generated .saggy.yaml (SOPS config)
✓ Updated .gitignore

Next steps:
1. Add public keys to .saggy/keys/
2. Update .saggy.yaml with key recipients
3. Create encrypted secrets: saggy create --from <file>
4. Commit .saggy/ to version control

Private keys should be stored in: ~/.config/saggy/keys/
NEVER commit private keys to the repository!
```

---

### 2. `saggy env <file> [..files] -- <command>`

**Purpose**: Load encrypted environment files and execute command with those variables.

**Behavior**:
```bash
saggy env <file1> [file2 ...] -- <command> [args...]
```

**Actions**:
1. Validate all files exist in `.saggy/secrets/`
2. Locate private key(s) from:
   - `$SAGGY_PRIVATE_KEY` (raw key material)
   - `$SAGGY_PRIVATE_KEY_PATH` (file path)
   - Default location: `~/.config/saggy/keys/*.key`
3. Decrypt each file using SOPS + age
4. Parse decrypted content as environment variables (`.env` format)
5. Merge variables (later files override earlier ones)
6. Execute command with merged environment
7. Clean up any temporary decrypted data in memory

**File format support**:
- `.env` format (KEY=VALUE)
- JSON format (if file ends in `.json.sops`)
- YAML format (if file ends in `.yaml.sops` or `.yml.sops`)

**Environment variable parsing**:
```bash
# Supports standard .env format
DATABASE_URL=postgres://localhost/db
API_KEY=secret123

# Comments
# This is ignored

# Multi-line (quoted)
PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"

# Variable expansion (optional feature)
BASE_URL=https://api.example.com
API_ENDPOINT=${BASE_URL}/v1
```

**Flags**:
- `--key <path>`: Specify private key file explicitly
- `--key-env <var>`: Read private key from environment variable
- `--prefix <prefix>`: Add prefix to all loaded variables (e.g., `PROD_`)
- `--export`: Export variables to current shell (requires `eval`)
- `--dry-run`: Show what variables would be loaded without executing
- `--no-inherit`: Don't inherit existing environment variables

**Exit codes**:
- `0`: Command executed successfully
- `1`: File not found or decryption failed
- `2`: Private key not found
- `3`: Invalid file format
- `N`: Exit code of executed command

**Examples**:
```bash
# Load single environment
saggy env prod.env.sops -- npm start

# Load multiple (staging overrides defaults)
saggy env defaults.env.sops staging.env.sops -- ./deploy.sh

# Dry run to see variables
saggy env prod.env.sops -- --dry-run

# Use specific key
saggy env prod.env.sops --key ~/.secrets/prod.key -- node app.js

# In CI/CD (key from environment)
export SAGGY_PRIVATE_KEY="AGE-SECRET-KEY-1..."
saggy env prod.env.sops -- ./run-tests.sh
```

**Error handling**:
```
Error: File not found: .saggy/secrets/prod.env.sops
Hint: Available files: dev.env.sops, staging.env.sops

Error: Failed to decrypt .saggy/secrets/prod.env.sops
Hint: No valid private key found. Tried:
  - $SAGGY_PRIVATE_KEY (not set)
  - $SAGGY_PRIVATE_KEY_PATH (not set)
  - ~/.config/saggy/keys/*.key (no matching keys)

Error: Invalid .env format at line 15
Line: INVALID LINE HERE
Hint: Expected format: KEY=VALUE
```

---

### 3. `saggy file -I {} <file> -- <command>`

**Purpose**: Temporarily decrypt a file and make it available to a command via interpolation.

**Behavior**:
```bash
saggy file -I <placeholder> <file> -- <command with {placeholder}>
```

**Actions**:
1. Validate file exists in `.saggy/secrets/`
2. Locate private key (same as `saggy env`)
3. Create temporary directory: `.saggy/.tmp/<random>/`
4. Decrypt file to temporary location
5. Replace placeholder in command with temp file path
6. Execute command
7. **Securely delete** temporary file (overwrite with zeros, then delete)
8. Remove temporary directory

**Security considerations**:
- Temp files have `600` permissions (owner read/write only)
- Temp directory has `700` permissions (owner only)
- Files are overwritten with random data before deletion
- On Unix: use `shred` or `srm` if available
- On error/interrupt: ensure cleanup runs (signal handlers)

**Flags**:
- `-I <placeholder>`: Placeholder to replace (default: `{}`)
- `--keep`: Don't delete temp file after command (for debugging)
- `--temp-dir <path>`: Custom temporary directory
- `--key <path>`: Specify private key file

**Exit codes**:
- `0`: Command executed successfully  
- `1`: File not found or decryption failed
- `2`: Private key not found
- `3`: Cleanup failed (warning)
- `N`: Exit code of executed command

**Examples**:
```bash
# Pass decrypted config file to app
saggy file -I {} database.json.sops -- node app.js --config {}

# Multiple files (call saggy file multiple times, or extend spec)
saggy file -I {db} database.json.sops -- \
  saggy file -I {api} api-keys.json.sops -- \
  ./deploy.sh --db {db} --api {api}

# Custom placeholder
saggy file -I $CONFIG service-config.yaml.sops -- \
  kubectl apply -f $CONFIG

# Keep file for debugging
saggy file --keep -I {} debug.env.sops -- cat {}
# Prints: Temporary file kept at: .saggy/.tmp/abc123/debug.env
```

**Error handling**:
```
Error: Placeholder '{}' not found in command
Command: node app.js --config /path/to/file
Hint: Use -I to specify placeholder that appears in your command

Warning: Failed to securely delete temporary file
File: .saggy/.tmp/xyz/secret.json
Hint: Please manually delete this file
```

---

### 4. `saggy create --from <file>`

**Purpose**: Create an encrypted secrets file from a plaintext file.

**Behavior**:
```bash
saggy create --from <plaintext-file> [--name <output-name>] [--type <type>]
```

**Actions**:
1. Validate plaintext file exists and is readable
2. Check if file is already encrypted (warn if so)
3. Load public keys from `.saggy/keys/` (or `.saggy.yaml` config)
4. Determine output filename:
   - If `--name` provided: use it
   - Otherwise: `basename(file).sops` in `.saggy/secrets/`
5. Encrypt file using SOPS with age recipients from config
6. Write encrypted file to `.saggy/secrets/`
7. Optionally delete plaintext source (with confirmation)
8. Print success message with next steps

**Flags**:
- `--from <file>`: Source plaintext file (required)
- `--name <name>`: Output filename (default: auto-generated)
- `--type <type>`: File type (env, json, yaml) - auto-detected if omitted
- `--keys <key1,key2,...>`: Specific public keys to encrypt for
- `--delete-source`: Delete plaintext file after encryption (prompts for confirmation)
- `--force`: Overwrite existing encrypted file
- `--stdin`: Read from stdin instead of file

**File type detection**:
```
.env, .env.* → env format
.json        → json format
.yaml, .yml  → yaml format
```

**Exit codes**:
- `0`: Success
- `1`: Source file not found
- `2`: Encryption failed
- `3`: Output file already exists (use --force)
- `4`: No public keys configured

**Examples**:
```bash
# Create from .env file
saggy create --from prod.env
# Output: .saggy/secrets/prod.env.sops

# Custom name
saggy create --from config.json --name production-db.json.sops

# Encrypt for specific keys only
saggy create --from secrets.env --keys ci-prod,alice

# Delete source after encryption
saggy create --from sensitive.env --delete-source
# Prompts: Delete sensitive.env after encryption? [y/N]

# From stdin (useful in scripts)
echo "API_KEY=secret" | saggy create --stdin --name api.env.sops

# Force overwrite
saggy create --from staging.env --force
```

**Interactive flow** (when `--delete-source` is used):
```
Encrypting: prod.env
Public keys: ci-prod, alice, bob
Output: .saggy/secrets/prod.env.sops

✓ Encrypted successfully

Delete source file 'prod.env'? [y/N] y
✓ Deleted prod.env

Next steps:
1. Verify decryption works: saggy env prod.env.sops -- env | grep KEY
2. Commit .saggy/secrets/prod.env.sops
3. NEVER commit prod.env (plaintext)
```

---

## Additional Commands

### 5. `saggy keys`

**Purpose**: Manage age keys.

**Subcommands**:

```bash
# List public keys
saggy keys list

# Generate new key pair
saggy keys generate <name>

# Add existing public key
saggy keys add <name> <public-key-or-file>

# Remove public key
saggy keys remove <name>

# Show public key
saggy keys show <name>
```

**Example**:
```bash
# Generate new key pair
saggy keys generate alice
# Output:
# ✓ Generated key pair for 'alice'
# Public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
# Private key saved to: ~/.config/saggy/keys/alice.key
# Public key saved to: .saggy/keys/alice.pub
# 
# Add this public key to your .saggy.yaml configuration

# List all keys
saggy keys list
# Output:
# Public keys in .saggy/keys/:
#   alice.pub    - age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7...
#   bob.pub      - age1zj2kg5sfn9aqmcac8pql3z7hjy54pw...
#   ci-prod.pub  - age1ayyfg7zqgvc7w3j2elw8zmrj2kg5sf...
```

---

### 6. `saggy edit <file>`

**Purpose**: Edit encrypted file in decrypted form (like `sops <file>`).

**Behavior**:
```bash
saggy edit <file>
```

**Actions**:
1. Decrypt file to temporary location
2. Open in `$EDITOR` (default: vim/nano)
3. On save, re-encrypt with same keys
4. Update file in `.saggy/secrets/`
5. Clean up temporary file

---

### 7. `saggy rotate`

**Purpose**: Re-encrypt files with new keys.

**Behavior**:
```bash
saggy rotate <file> --add-keys <keys> --remove-keys <keys>
```

**Use case**: When team members join/leave or keys are compromised.

---

### 8. `saggy validate`

**Purpose**: Validate all encrypted files can be decrypted.

**Behavior**:
```bash
saggy validate [file...]
```

**Actions**:
1. Attempt to decrypt each file
2. Report success/failure
3. Exit with non-zero if any fail

**Example output**:
```
Validating encrypted files...
✓ .saggy/secrets/prod.env.sops
✓ .saggy/secrets/staging.env.sops
✗ .saggy/secrets/dev.env.sops (no valid key)

2/3 files valid
```

---

## Configuration File: `.saggy/config.yaml`

```yaml
version: 1

# Default private key directory
default_key_dir: ~/.config/saggy/keys

# Alternative: use specific key file
# default_key_file: ~/.config/saggy/keys/my-key.key

# Alternative: use environment variable
# default_key_env: SAGGY_PRIVATE_KEY

# Where public keys are stored
public_keys_dir: .saggy/keys

# Where encrypted secrets are stored  
secrets_dir: .saggy/secrets

# File format defaults
format:
  env:
    # Support variable expansion (e.g., ${VAR})
    expand_variables: true
    # Allow unquoted values
    allow_unquoted: true

# Temporary files
temp_dir: .saggy/.tmp
secure_delete: true  # Overwrite before deleting

# Logging
log_level: info  # debug, info, warn, error
```

---

## SOPS Configuration: `.saggy.yaml`

```yaml
creation_rules:
  # Production: require 2 of 3 keys
  - path_regex: .saggy/secrets/prod\..*\.sops$
    key_groups:
      - age:
          - age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p  # alice
          - age1zj2kg5sfn9aqmcac8pql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmr  # bob
          - age1ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8pql3z7hjy54pw3hyww5  # charlie
        threshold: 2

  # Staging: any dev + CI
  - path_regex: .saggy/secrets/staging\..*\.sops$
    key_groups:
      - age:
          - age1dev1...  # dev-team
          - age1ci2...   # ci-staging

  # Development: anyone
  - path_regex: .saggy/secrets/dev\..*\.sops$
    key_groups:
      - age:
          - age1dev1...
          - age1dev2...
          - age1dev3...

  # Default fallback
  - path_regex: .saggy/secrets/.*\.sops$
    age: >-
      age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p,
      age1zj2kg5sfn9aqmcac8pql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmr
```

---

## Security Requirements

### 1. **Private Key Handling**
- NEVER log private keys
- NEVER include private keys in error messages
- Store with `600` permissions (owner read/write only)
- Support system keychains (macOS Keychain, Windows Credential Manager)

### 2. **Temporary File Security**
- Create with `600` permissions immediately
- Overwrite with zeros before deletion (use `shred -u` on Linux)
- Clean up on signals: SIGINT, SIGTERM, SIGHUP
- Use cryptographically secure temp directory names

### 3. **Error Messages**
- Never reveal decrypted content in errors
- Don't expose full file paths in public repos (relative paths only)
- Sanitize command output in logs

### 4. **Environment Variables**
- Clear sensitive env vars from process memory after use
- Don't leak env vars to child processes unless necessary

---

## Error Handling

### Exit Codes
```
0   - Success
1   - General error (file not found, invalid format)
2   - Authentication error (no valid private key)
3   - Permission error
4   - Configuration error
5   - Encryption/decryption failed
130 - Interrupted (Ctrl+C)
```

### User-Friendly Errors
```
❌ Error: No private key found

Saggy couldn't find a private key to decrypt the file.

Tried locations:
  ✗ Environment: $SAGGY_PRIVATE_KEY (not set)
  ✗ Environment: $SAGGY_PRIVATE_KEY_PATH (not set)
  ✗ Default: ~/.config/saggy/keys/*.key (no files)

Solutions:
  1. Generate a key: saggy keys generate <name>
  2. Set $SAGGY_PRIVATE_KEY_PATH to your key file
  3. Request access from your team

Need help? https://docs.saggy.dev/troubleshooting
```

---

## Installation & Setup

### Installation
```bash
# Via package manager
brew install saggy           # macOS
apt install saggy            # Debian/Ubuntu
cargo install saggy          # From source

# Verify
saggy --version
```

### First-Time Setup
```bash
# 1. Initialize repo
cd my-project
saggy init

# 2. Generate your personal key
saggy keys generate $(whoami)

# 3. Add your public key to the repo
# (it's automatically added to .saggy/keys/)

# 4. Create encrypted secrets
saggy create --from .env.prod

# 5. Commit encrypted files
git add .saggy/
git commit -m "Add encrypted secrets"

# 6. Use in CI/CD
# Add $SAGGY_PRIVATE_KEY to CI secrets
# Use: saggy env prod.env.sops -- npm run deploy
```

---

## CI/CD Integration

### GitHub Actions
```yaml
name: Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install saggy
        run: |
          curl -sSL https://get.saggy.dev | sh
      
      - name: Deploy with secrets
        env:
          SAGGY_PRIVATE_KEY: ${{ secrets.PROD_PRIVATE_KEY }}
        run: |
          saggy env prod.env.sops -- npm run deploy
```

### GitLab CI
```yaml
deploy:
  script:
    - saggy env prod.env.sops -- ./deploy.sh
  variables:
    SAGGY_PRIVATE_KEY: $PROD_KEY
```

---

## Implementation Notes

### Tech Stack Recommendations
- **Language**: Rust or Go (for security, performance, single binary)
- **Dependencies**:
  - `sops` (Mozilla SOPS library)
  - `age` encryption
  - `clap` or `cobra` (CLI parsing)
  - `toml`/`yaml` parsers

### Testing Requirements
- Unit tests for encryption/decryption
- Integration tests with real SOPS/age
- Security tests (key leakage, temp file cleanup)
- CI/CD simulation tests

### Documentation
- Man pages for each command
- Online docs at docs.saggy.dev
- Examples repo with common patterns
- Troubleshooting guide

---

## Future Enhancements

### Phase 2
- `saggy diff`: Show diff between encrypted versions
- `saggy audit`: Audit log of who decrypted what
- `saggy sync`: Sync from remote secret managers (Vault, AWS Secrets Manager)
- Shell completion (bash, zsh, fish)

### Phase 3
- GUI for managing keys and secrets
- Browser extension for viewing secrets
- Integration with VS Code, IntelliJ
- Webhook support for secret rotation

---

## Examples: Real-World Workflows

### Developer Onboarding
```bash
# New developer joins team
# 1. They generate their key
saggy keys generate alice

# 2. Share public key with team (via Slack/email)
cat .saggy/keys/alice.pub

# 3. Team member adds their key and re-encrypts
saggy rotate staging.env.sops dev.env.sops --add-keys alice
git commit -am "Add Alice to dev/staging access"

# 4. Alice can now use secrets
saggy env dev.env.sops -- npm start
```

### Production Deploy
```bash
# CI/CD pipeline
export SAGGY_PRIVATE_KEY="${CI_SECRET_KEY}"

# Load multiple config files
saggy env \
  defaults.env.sops \
  prod.env.sops \
  prod-db.env.sops \
  -- ./deploy.sh

# Or with temporary config files
saggy file -I {db} database.json.sops -- \
  saggy env prod.env.sops -- \
  node deploy.js --db-config {db}
```

### Emergency Key Rotation
```bash
# Developer leaves, rotate all secrets
for file in .saggy/secrets/prod*.sops; do
  saggy rotate "$file" --remove-keys departing-user
done

git commit -am "Rotate keys after team change"
git push
```

---

This spec provides a complete blueprint for implementing `saggy`. Let me know if you'd like me to elaborate on any section or add additional features!
