#!/bin/bash

## Setup

PUBLIC_KEYFILE="./secrets/public-age-keys.json"
PRIVATE_KEYFILE="./secrets/age.key"

if [ -f "$PUBLIC_KEYFILE" ]; then echo "Public keys already exist."; exit 0; fi
if [ -f "$PRIVATE_KEYFILE" ]; then echo "private key already exists."; exit 0; fi

## Should not be able to override the key

# Generate the key for the first time
$SAGGY keygen

# Attempt to generate the key again without deleting the old one
if $SAGGY keygen; then
    echo "Should not be able to override a key OR should not succeed without generating a key."
    exit 1
fi
