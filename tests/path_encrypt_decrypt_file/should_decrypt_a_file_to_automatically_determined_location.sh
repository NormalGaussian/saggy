#!/bin/bash

## Setup

ENCRYPTED_FILE="$(mktemp -d)/testfile.sops"
PLAINTEXT_FILE="$(mktemp -d)/testfile.plaintext"

mkdir -p "$(dirname "$ENCRYPTED_FILE")"
mkdir -p "$(dirname "$PLAINTEXT_FILE")"

echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be able to decrypt a file to an automatically determined location

$SAGGY decrypt "$ENCRYPTED_FILE"

## Verify
if [ ! -f "$PLAINTEXT_FILE" ]; then echo "Should create a decrypted file."; exit 1; fi
if ! diff "$PLAINTEXT_FILE" "$PLAINTEXT_FILE" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi
