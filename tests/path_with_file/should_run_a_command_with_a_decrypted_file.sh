#!/bin/bash

## Setup

ENCRYPTED_FILE="./testfile.sops"
DECRYPTED_FILE="./testfile.decrypted"
PLAINTEXT_FILE="./testfile.plaintext"

mkdir -p "$(dirname "$DECRYPTED_FILE")"
echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be able to run a command with a decrypted file

# Use with to make a copy of the decrypted file
$SAGGY with "$ENCRYPTED_FILE" -- cat {} > "$DECRYPTED_FILE"

# Verify
if [ ! -f "$DECRYPTED_FILE" ]; then echo "Should create a decrypted file."; exit 1; fi
if ! diff "$PLAINTEXT_FILE" "$DECRYPTED_FILE"; then echo "Should contain the decrypted content."; exit 1; fi
