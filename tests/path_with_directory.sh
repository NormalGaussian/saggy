#!/bin/bash

## Setup

ENCRYPTED_DIR="./tmp/testdir.sops"
DECRYPTED_DIR="./tmp/testdir.decrypted"
PLAINTEXT_DIR="./tmp/testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be able to run a command with a decrypted directory

# Use with to make a copy of the decrypted directory
$SAGGY with "$ENCRYPTED_DIR" -- cp -r {} "$DECRYPTED_DIR"

# Verify
if [ ! -d "$DECRYPTED_DIR" ]; then echo "Should create a decrypted directory."; exit 1; fi
if ! diff -r "$PLAINTEXT_DIR" "$DECRYPTED_DIR" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi

# Clean up
rm -rf "$DECRYPTED_DIR"

## Should be able to run a command with a decrypted directory and write changes back

# Setup some new content
NEW_PLAINTEXT_DIR="./tmp/new_testdir.plaintext"
mkdir -p "$NEW_PLAINTEXT_DIR"
echo "new content 1" > "$NEW_PLAINTEXT_DIR/file1.txt"
echo "new content 2" > "$NEW_PLAINTEXT_DIR/file2.txt"

# Use with to write the new content to the encrypted directory
$SAGGY with "$ENCRYPTED_DIR" -w -- cp -r "$NEW_PLAINTEXT_DIR/" {}

# Use with to extract the encrypted content
$SAGGY with "$ENCRYPTED_DIR" -- cp -r {} "$DECRYPTED_DIR"

# Verify the content was written
if ! diff -r "$NEW_PLAINTEXT_DIR" "$DECRYPTED_DIR"; then echo "Should contain the new content."; exit 1; fi

# Clean up
rm -rf "$ENCRYPTED_DIR" "$DECRYPTED_DIR"