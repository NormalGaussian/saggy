#!/bin/bash

## Setup

PUBLIC_KEYFILE="./secrets/public-age-keys.json"
PRIVATE_KEYFILE="./secrets/age.key"

if [ -f "$PUBLIC_KEYFILE" ]; then echo "Public keys already exist."; exit 0; fi
if [ -f "$PRIVATE_KEYFILE" ]; then echo "private key already exists."; exit 0; fi

## Should be able to add an additional key with a different hostname

SAGGY_KEYNAME=oldkey $SAGGY keygen

# Save the old public key
OLD_PUBLIC_KEY=$(age-keygen -y "$PRIVATE_KEYFILE")
# generate a new private key
rm -f "$PRIVATE_KEYFILE"
SAGGY_KEYNAME=newkey $SAGGY keygen

# Both the new and old public keys should be in the public keyfile
if [ ! -f "$PRIVATE_KEYFILE" ]; then echo "Should generate a new key."; exit 1; fi
if ! grep -q "oldkey" "$PUBLIC_KEYFILE"; then echo "Should include the old key name in the public keyfile."; exit 1; fi
if ! grep -q "newkey" "$PUBLIC_KEYFILE"; then echo "Should include the new key name in the public keyfile."; exit 1; fi
if ! grep -q "$(age-keygen -y "$PRIVATE_KEYFILE")" "$PUBLIC_KEYFILE"; then echo "Should include the new public key in the public keyfile."; exit 1; fi
if ! grep -q "$OLD_PUBLIC_KEY" "$PUBLIC_KEYFILE"; then echo "Should include the old public key in the public keyfile."; exit 1; fi
