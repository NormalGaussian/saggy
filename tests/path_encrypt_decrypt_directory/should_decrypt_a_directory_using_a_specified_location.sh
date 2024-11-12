#!/bin/bash

## Setup

ENCRYPTED_DIR="./tmp/testdir.encrypted"
DECRYPTED_DIR="./tmp/testdir.decrypted"
PLAINTEXT_DIR="./tmp/testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be able to decrypt a directory to a specified location

$SAGGY decrypt "$ENCRYPTED_DIR" "$DECRYPTED_DIR"

# Verify
if [ ! -d "$DECRYPTED_DIR" ]; then echo "Should create a decrypted directory."; exit 1; fi
if ! diff -r "$PLAINTEXT_DIR" "$DECRYPTED_DIR" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi
