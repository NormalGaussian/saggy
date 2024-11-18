#!/bin/bash

## Setup

ENCRYPTED_DIR="./testdir.sops"
DECRYPTED_DIR="./decrypted"
PLAINTEXT_DIR="./plaintext"

mkdir -p "$ENCRYPTED_DIR"
mkdir -p "$DECRYPTED_DIR"
mkdir -p "$PLAINTEXT_DIR"

echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be transparent to command results

# This command should be successful
EXIT_CODE=0
if $SAGGY with "$ENCRYPTED_DIR" -- echo pipe 2> "$STDERR_FILE" > "$STDOUT_FILE"; then
  EXIT_CODE=$?
else 
  EXIT_CODE=$?
fi

if [ "$EXIT_CODE" -ne 0 ]; then echo "Should be successful."; exit 1; fi
if [ ! -f "$STDOUT_FILE" ]; then echo "Should create a stdout file."; exit 1; fi
if [ -s "$STDERR_FILE" ]; then echo "Should not create a stderr file."; exit 1; fi
if [ "$(cat "$STDOUT_FILE")" != "pipe" ]; then echo "Stdout file should contain 'pipe'."; exit 1; fi
