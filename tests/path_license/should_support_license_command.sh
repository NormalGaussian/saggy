#!/bin/bash

## Should support license command

LICENSE="./license.txt"
LICENSE_FULL="./license_full.txt"

$SAGGY license > "$LICENSE"
$SAGGY license --full > "$LICENSE_FULL"

## Verify
if [ ! -f "$LICENSE" ]; then echo "Should give output for license."; exit 1; fi
if [ ! -f "$LICENSE_FULL" ]; then echo "Should give output for full license."; exit 1; fi
if diff "$LICENSE" "$LICENSE_FULL" >/dev/null; then echo "Should give different output for license and full license."; exit 1; fi
