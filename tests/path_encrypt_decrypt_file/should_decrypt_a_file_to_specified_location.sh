#!/bin/bash

## Setup

ENCRYPTED_FILE="$(mktemp -d)/testfile.sops"
DECRYPTED_FILE="$(mktemp -d)/testfile.decrypted"
PLAINTEXT_FILE="$(mktemp -d)/testfile.plaintext"

mkdir -p "$(dirname "$DECRYPTED_FILE")"
mkdir -p "$(dirname "$PLAINTEXT_FILE")"

echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be able to decrypt a file to a specified location

$SAGGY decrypt "$ENCRYPTED_FILE" "$DECRYPTED_FILE"

## Verify
if [ ! -f "$DECRYPTED_FILE" ]; then echo "Should create a decrypted file."; exit 1; fi
if ! diff "$PLAINTEXT_FILE" "$DECRYPTED_FILE" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi
