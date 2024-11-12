#!/bin/bash

## Setup

ENCRYPTED_DIR="$(mktemp -d)/testdir.sops"
DECRYPTED_DIR="$(mktemp -d)/decrypted"
PLAINTEXT_DIR="$(mktemp -d)/plaintext"
REPLACEMENT_PLAINTEXT_DIR="$(mktemp -d)/replacement_plaintext"

mkdir -p "$ENCRYPTED_DIR"
mkdir -p "$DECRYPTED_DIR"
mkdir -p "$PLAINTEXT_DIR"
mkdir -p "$REPLACEMENT_PLAINTEXT_DIR"

echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

echo "replacement content 1" > "$REPLACEMENT_PLAINTEXT_DIR/file1.txt"
echo "replacement content 2" > "$REPLACEMENT_PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be able to run a command with a decrypted directory and write changes back

# Use with to write the new content to the encrypted directory
$SAGGY with "$ENCRYPTED_DIR" -w -- cp -r "$REPLACEMENT_PLAINTEXT_DIR"/* {}

# Use with to extract the encrypted content
$SAGGY with "$ENCRYPTED_DIR" -- cp -r {}/* "$DECRYPTED_DIR"

# Verify the content was written
if ! diff -r "$REPLACEMENT_PLAINTEXT_DIR" "$DECRYPTED_DIR"; then echo "Should contain the new content."; exit 1; fi
