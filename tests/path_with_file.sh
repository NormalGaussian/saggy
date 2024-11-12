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

# Clean up
rm -f "$DECRYPTED_FILE"

## Should be able to run a command with a decrypted file and write changes back

# Setup some new content
NEW_PLAINTEXT="new content"
NEW_PLAINTEXT_FILE="./tmp/new_testfile.plaintext"
mkdir -p "$(dirname "$NEW_PLAINTEXT_FILE")"
echo "$NEW_PLAINTEXT" > "$NEW_PLAINTEXT_FILE"

# Use with to write the new content to the encrypted file
$SAGGY with "$ENCRYPTED_FILE" -w -- cp "$NEW_PLAINTEXT_FILE" {}

# Use with to extract the encrypted content
$SAGGY with "$ENCRYPTED_FILE" -- cat {} > "$DECRYPTED_FILE"

# Verify the content was written
if ! diff "$NEW_PLAINTEXT_FILE" "$DECRYPTED_FILE"; then echo "Should contain the new content."; exit 1; fi

# Clean up
rm -f "$ENCRYPTED_FILE" "$DECRYPTED_FILE"