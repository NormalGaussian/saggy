#!/bin/bash

## Setup

ENCRYPTED_FILE="./testfile.encrypted"
PLAINTEXT_FILE="./testfile.plaintext"
STDOUT_FILE="./.stdout"
STDERR_FILE="./.stderr"

mkdir -p "$(dirname "$ENCRYPTED_FILE")"
mkdir -p "$(dirname "$PLAINTEXT_FILE")"
echo "test content" > "$PLAINTEXT_FILE"

$SAGGY keygen
$SAGGY encrypt "$PLAINTEXT_FILE" "$ENCRYPTED_FILE"

## Should be transparent to command results

# This command should be successful
EXIT_CODE=0
if $SAGGY with "$ENCRYPTED_FILE" -- echo pipe 2> "$STDERR_FILE" > "$STDOUT_FILE"; then
  EXIT_CODE=$?
else 
  EXIT_CODE=$?
fi

if [ "$EXIT_CODE" -ne 0 ]; then echo "Should be successful."; exit 1; fi
if [ ! -f "$STDOUT_FILE" ]; then echo "Should create a stdout file."; exit 1; fi
if [ -s "$STDERR_FILE" ]; then echo "Should not create a stderr file."; exit 1; fi
if [ "$(cat "$STDOUT_FILE")" != "pipe" ]; then echo "Stdout file should contain 'pipe'."; exit 1; fi
