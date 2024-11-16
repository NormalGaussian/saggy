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

# This command should fail with a specific exit code
EXIT_CODE=0
if $SAGGY with "$ENCRYPTED_FILE" -- grep 2> "$STDERR_FILE" > "$STDOUT_FILE"; then
  EXIT_CODE=$?
else 
  EXIT_CODE=$?
fi

if [ ! "$EXIT_CODE" -eq 2 ]; then echo "Should exit with 2."; exit 1; fi
if [ -s "$STDOUT_FILE" ]; then echo "Should not create a stdout file."; exit 1; fi
if [ ! -f "$STDERR_FILE" ]; then echo "Should create a stderr file."; exit 1; fi

rm "$STDOUT_FILE" "$STDERR_FILE"