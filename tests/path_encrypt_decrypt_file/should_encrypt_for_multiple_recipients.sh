#!/bin/bash

## Setup

PLAINTEXT_FILE="./plaintext"
SECRETS_DIR="./secrets"
PUBLIC_KEYFILE="$SECRETS_DIR/public-age-keys.json"
PRIVATE_KEYFILE="$SECRETS_DIR/age.key"

mkdir -p "$SECRETS_DIR"

echo "test content" > "$PLAINTEXT_FILE"

# Generate key for machine "alpha"
SAGGY_KEYNAME=alpha $SAGGY keygen

# Save alpha's private key
cp "$PRIVATE_KEYFILE" "$SECRETS_DIR/alpha.key"

# Generate key for machine "beta" (adds to public keyfile, replaces private key)
rm -f "$PRIVATE_KEYFILE"
SAGGY_KEYNAME=beta $SAGGY keygen

# Save beta's private key
cp "$PRIVATE_KEYFILE" "$SECRETS_DIR/beta.key"

## Encrypt with both recipients

$SAGGY encrypt "$PLAINTEXT_FILE"

ENCRYPTED_FILE="$PLAINTEXT_FILE.sops"
if [ ! -f "$ENCRYPTED_FILE" ]; then echo "Should create an encrypted file."; exit 1; fi

## Verify alpha can decrypt

SOPS_AGE_KEY_FILE="$SECRETS_DIR/alpha.key" sops --decrypt "$ENCRYPTED_FILE" > ./decrypted_alpha
if ! diff -q ./decrypted_alpha "$PLAINTEXT_FILE" >/dev/null; then echo "Alpha should be able to decrypt."; exit 1; fi

## Verify beta can decrypt

SOPS_AGE_KEY_FILE="$SECRETS_DIR/beta.key" sops --decrypt "$ENCRYPTED_FILE" > ./decrypted_beta
if ! diff -q ./decrypted_beta "$PLAINTEXT_FILE" >/dev/null; then echo "Beta should be able to decrypt."; exit 1; fi
