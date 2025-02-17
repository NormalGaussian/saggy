#!/bin/bash

## Setup

PLAINTEXT_DIR="./testdir.plaintext"

mkdir -p "$PLAINTEXT_DIR"
echo "test content 1" > "$PLAINTEXT_DIR/file1.txt"
echo "test content 2" > "$PLAINTEXT_DIR/file2.txt"

$SAGGY keygen

## Should be able to encrypt a directory to the automatically determined location

$SAGGY encrypt "$PLAINTEXT_DIR"

## Verify
if [ ! -d "$PLAINTEXT_DIR.sops" ]; then echo "Should create an encrypted directory."; exit 1; fi
if find "$PLAINTEXT_DIR.sops" -type f | grep -qv '.sops'; then echo "Should contain only encrypted files."; exit 1; fi

## Decrypt tests verify the content
