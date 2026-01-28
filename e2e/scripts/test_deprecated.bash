#!/usr/bin/env bash

set -eux

: "${VENOM_VAR_binpath:? Please set VENOM_VAR_binpath to updatecli binary dirname}"
: "${VENOM_VAR_rootpath:=../..}"

## In order for the Updatecli manifest to work, we must be at the root of the git repository 
pushd "$VENOM_VAR_rootpath"

# Test deprecated diff command (should show deprecation warning but still work)
"$VENOM_VAR_binpath/updatecli" pipeline diff --config  e2e/updatecli.d/deprecated.d

# Test that deprecated diff command with valid xml config still works with a warning
"$VENOM_VAR_binpath/updatecli" diff --config  e2e/updatecli.d/success.d/xml.yaml