#!/bin/bash

## Setup

DECRYPTED_DIR="./testdir.decrypted"
PLAINTEXT_DIR="./testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

mkdir -p "$(dirname "$DECRYPTED_DIR")"

$SAGGY keygen

## Should be able to decrypt a directory to the automatically determined location

$SAGGY encrypt "$PLAINTEXT_DIR"
mv "$PLAINTEXT_DIR.sops" "$DECRYPTED_DIR.sops"
$SAGGY decrypt "$DECRYPTED_DIR.sops"

# Verify
if [ ! -d "$DECRYPTED_DIR" ]; then echo "Should create a decrypted directory."; exit 1; fi
if ! diff -r "$PLAINTEXT_DIR" "$DECRYPTED_DIR" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi

