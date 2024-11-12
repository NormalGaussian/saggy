#!/bin/bash

## Setup

ENCRYPTED_DIR="./testdir.encrypted"
DECRYPTED_DIR="./testdir.decrypted"
PLAINTEXT_DIR="./testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

mkdir -p "$(dirname "$ENCRYPTED_DIR")"
mkdir -p "$(dirname "$DECRYPTED_DIR")"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be able to decrypt a directory to a specified location

$SAGGY decrypt "$ENCRYPTED_DIR" "$DECRYPTED_DIR"

## Verify
if [ ! -d "$DECRYPTED_DIR" ]; then echo "Should create a decrypted directory."; exit 1; fi
if ! diff -r "$PLAINTEXT_DIR" "$DECRYPTED_DIR" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi
