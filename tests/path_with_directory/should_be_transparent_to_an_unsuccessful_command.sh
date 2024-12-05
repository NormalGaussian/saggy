#!/bin/bash

## Setup

ENCRYPTED_DIR="./testdir.sops"
DECRYPTED_DIR="./decrypted"
PLAINTEXT_DIR="./plaintext"
STDOUT_FILE="./.stdout"
STDERR_FILE="./.stderr"

mkdir -p "$ENCRYPTED_DIR"
mkdir -p "$DECRYPTED_DIR"
mkdir -p "$PLAINTEXT_DIR"

echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_DIR" "$ENCRYPTED_DIR"

## Should be transparent to command results

# This command should fail

EXIT_CODE=0
if $SAGGY with "$ENCRYPTED_DIR" -- false 2> "$STDERR_FILE" > "$STDOUT_FILE"; then
  EXIT_CODE=$?
else 
  EXIT_CODE=$?
fi

if [ "$EXIT_CODE" -eq 0 ]; then echo "Should fail."; exit 1; fi
if [ -s "$STDOUT_FILE" ]; then echo "Should not create a stdout file."; exit 1; fi
if [ -s "$STDERR_FILE" ]; then echo "Should not create a stderr file."; exit 1; fi

rm "$STDOUT_FILE" "$STDERR_FILE"
