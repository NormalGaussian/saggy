#!/bin/bash

## Setup

ENCRYPTED_DIR="$(mktemp -d)/testdir.sops"
DECRYPTED_DIR="$(mktemp -d)/decrypted"
PLAINTEXT_DIR="$(mktemp -d)/plaintext"

mkdir -p "$ENCRYPTED_DIR"
mkdir -p "$DECRYPTED_DIR"
mkdir -p "$PLAINTEXT_DIR"

echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be able to run a command with a decrypted directory

# Use with to make a copy of the decrypted directory
$SAGGY with "$ENCRYPTED_DIR" -- cp -r {}/* "$DECRYPTED_DIR"

# Verify
if [ ! -d "$DECRYPTED_DIR" ]; then echo "Should create a decrypted directory."; exit 1; fi
if ! diff -r "$PLAINTEXT_DIR" "$DECRYPTED_DIR" >/dev/null; then echo "Should contain the decrypted content."; exit 1; fi

