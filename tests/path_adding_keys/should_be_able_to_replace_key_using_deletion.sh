#!/bin/bash

## Setup

PUBLIC_KEYFILE="./secrets/public-age-keys.json"
PRIVATE_KEYFILE="./secrets/age.key"

if [ -f "$PUBLIC_KEYFILE" ]; then echo "Public keys already exist."; exit 0; fi
if [ -f "$PRIVATE_KEYFILE" ]; then echo "private key already exists."; exit 0; fi

## Should be able to generate a new key

## Should be able to replace key using deletion

# Generate the key for the first time
$SAGGY keygen

# Save the old public key
FIRST_PUBLIC_KEY="$(age-keygen -y "$PRIVATE_KEYFILE")"

# Delete the old key
rm -f "$PRIVATE_KEYFILE"

# Generate a new key
$SAGGY keygen

if [ ! -f "$PRIVATE_KEYFILE" ]; then echo "Should generate a replacement key."; exit 1; fi
if ! grep -iq "$(hostname)" "$PUBLIC_KEYFILE"; then echo "Should include the hostname in the public keyfile."; exit 1; fi
if ! grep -q "$(age-keygen -y "$PRIVATE_KEYFILE")" "$PUBLIC_KEYFILE"; then echo "Should include the public key in the public keyfile."; exit 1; fi
## the key will have overridden the hostname in the public keyfile
if grep -q "$FIRST_PUBLIC_KEY" "$PUBLIC_KEYFILE"; then echo "Should not include the old public key in the public keyfile."; exit 1; fi
