#!/bin/bash

## Setup

ENCRYPTED_DIR="./tmp/testdir.encrypted"
DECRYPTED_DIR="./tmp/testdir.decrypted"
PLAINTEXT_DIR="./tmp/testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

mkdir -p "$(dirname ENCRYPTED_DIR)"

$SAGGY keygen

## Should be able to encrypt a directory to a specified location

$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

# Verify
if [ ! -d "$ENCRYPTED_DIR" ]; then echo "Should create an encrypted directory."; exit 1; fi
if find "$ENCRYPTED_DIR" -type f | grep -qv '.sops'; then echo "Should contain only encrypted files."; exit 1; fi

# Decrypt tests verify the content
