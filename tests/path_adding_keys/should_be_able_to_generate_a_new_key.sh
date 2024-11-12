#!/bin/bash

## Setup

PUBLIC_KEYFILE="./secrets/public-age-keys.json"
PRIVATE_KEYFILE="./secrets/age.key"

if [ -f "$PUBLIC_KEYFILE" ]; then echo "Public keys already exist."; exit 0; fi
if [ -f "$PRIVATE_KEYFILE" ]; then echo "private key already exists."; exit 0; fi

## Should be able to generate a new key

# Generate the key for the first time
$SAGGY keygen

if [ ! -f "$PRIVATE_KEYFILE" ]; then echo "Should generate a key."; exit 1; fi
if [ ! -f "$PUBLIC_KEYFILE" ]; then echo "Should create a public keyfile."; exit 1; fi
if ! grep -iq "$(hostname)" "$PUBLIC_KEYFILE"; then echo "Should include the hostname in the public keyfile."; exit 1; fi
if ! grep -q "$(age-keygen -y "$PRIVATE_KEYFILE")" "$PUBLIC_KEYFILE"; then echo "Should include the public key in the public keyfile."; exit 1; fi
