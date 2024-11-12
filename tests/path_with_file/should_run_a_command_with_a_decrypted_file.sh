#!/bin/bash

## Setup

ENCRYPTED_FILE="./tmp/testfile.sops"
DECRYPTED_FILE="./tmp/testfile.decrypted"
PLAINTEXT_FILE="./tmp/testfile.plaintext"

mkdir -p "$(dirname "$DECRYPTED_FILE")"
echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be able to run a command with a decrypted file

# Use with to make a copy of the decrypted file
$SAGGY with "$ENCRYPTED_FILE" -- cat {} > "$DECRYPTED_FILE"

# Verify
if [ ! -f "$DECRYPTED_FILE" ]; then echo "Should create a decrypted file."; exit 1; fi
if ! diff "$PLAINTEXT_FILE" "$DECRYPTED_FILE" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi
