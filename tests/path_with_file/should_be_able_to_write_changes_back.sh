#!/bin/bash

## Setup

ENCRYPTED_FILE="$(mktemp -d)/testfile.sops"
DECRYPTED_FILE="$(mktemp -d)/testfile.decrypted"
PLAINTEXT_FILE="$(mktemp -d)/testfile.plaintext"
REPLACEMENT_PLAINTEXT_FILE="$(mktemp -d)/new_testfile.plaintext"

mkdir -p "$(dirname "$ENCRYPTED_FILE")"
mkdir -p "$(dirname "$DECRYPTED_FILE")"
mkdir -p "$(dirname "$PLAINTEXT_FILE")"
mkdir -p "$(dirname "$REPLACEMENT_PLAINTEXT_FILE")"

echo "test content" > "$PLAINTEXT_FILE"
echo "new content" > "$REPLACEMENT_PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be able to run a command with a decrypted file and write changes back

# Use with to write the new content to the encrypted file
$SAGGY with "$ENCRYPTED_FILE" -w -- cp "$REPLACEMENT_PLAINTEXT_FILE" {}

# Use with to extract the encrypted content
$SAGGY with "$ENCRYPTED_FILE" -- cat {} > "$DECRYPTED_FILE"

# Verify the content was written
if ! diff "$REPLACEMENT_PLAINTEXT_FILE" "$DECRYPTED_FILE"; then echo "Should contain the new content."; exit 1; fi
