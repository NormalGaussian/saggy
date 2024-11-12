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

## Should not be able to override the key

# Attempt to generate the key again without deleting the old one
if $SAGGY keygen; then
    echo "Should not be able to override a key OR should not succeed without generating a key."
    exit 1
fi

## Should be able to replace key using deletion

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

## Should be able to add an additional key with a different hostname

# Save the old public key
SECOND_PUBLIC_KEY=$(age-keygen -y "$PRIVATE_KEYFILE")
# generate a new private key
rm -f "$PRIVATE_KEYFILE"
SAGGY_KEYNAME=newkey $SAGGY keygen

# Both the new and old public keys should be in the public keyfile
if [ ! -f "$PRIVATE_KEYFILE" ]; then echo "Should generate a new key."; exit 1; fi
if ! grep -q "newkey" "$PUBLIC_KEYFILE"; then echo "Should include the new key name in the public keyfile."; exit 1; fi
if ! grep -q "$(age-keygen -y "$PRIVATE_KEYFILE")" "$PUBLIC_KEYFILE"; then echo "Should include the new public key in the public keyfile."; exit 1; fi
if ! grep -iq "$(hostname)" "$PUBLIC_KEYFILE"; then echo "Should include the original public key name in the public keyfile."; exit 1; fi
if ! grep -q "$SECOND_PUBLIC_KEY" "$PUBLIC_KEYFILE"; then echo "Should include the original public key in the public keyfile."; exit 1; fi
