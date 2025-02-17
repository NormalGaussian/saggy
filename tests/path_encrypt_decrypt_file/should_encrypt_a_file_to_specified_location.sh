#!/bin/bash

## Setup

ENCRYPTED_FILE="./testfile.sops"
PLAINTEXT_FILE="./testfile.plaintext"

mkdir -p "$(dirname "$ENCRYPTED_FILE")"
mkdir -p "$(dirname "$PLAINTEXT_FILE")"

echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen

## Should be able to encrypt a file to a specified location

$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Verify
if [ ! -f "$ENCRYPTED_FILE" ]; then echo "Should create an encrypted file."; exit 1; fi
if sops --decrypt "$ENCRYPTED_FILE" | diff - "$PLAINTEXT_FILE" >/dev/null; then echo "Should contain the encrypted content."; exit 1; fi

## decrypt tests verify file content
