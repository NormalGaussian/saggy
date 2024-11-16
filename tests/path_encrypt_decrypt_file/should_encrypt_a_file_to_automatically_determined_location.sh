#!/bin/bash

## Setup

PLAINTEXT_FILE="./plaintext"

mkdir -p "$(dirname "$PLAINTEXT_FILE")"

echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen

## Should be able to encrypt a file to an automatically determined location

$SAGGY encrypt "$PLAINTEXT_FILE"

## Verify
if [ ! -f "$PLAINTEXT_FILE.sops" ]; then echo "Should create an encrypted file."; exit 1; fi
if sops --decrypt "$PLAINTEXT_FILE.sops" | diff - "$PLAINTEXT_FILE" >/dev/null; then echo "Should contain the encrypted content."; exit 1; fi

## decrypt tests verify file content