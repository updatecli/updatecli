#!/usr/bin/env bash

## This script expects 'updatecli manifest validate' to fail, so the exit code is
## inverted: the test passes when the invalid manifest is rejected.

set -ux

: "${VENOM_VAR_binpath:? Please set VENOM_VAR_binpath to updatecli binary dirname}"
: "${VENOM_VAR_rootpath:=../..}"

## In order for the Updatecli manifest to work, we must be at the root of the git repository
pushd "$VENOM_VAR_rootpath" || exit 1

"$VENOM_VAR_binpath/updatecli" manifest validate --experimental --config e2e/updatecli.d/invalid.d/

if [[ $? -eq 0 ]]; then
  echo "ERROR: the invalid manifest should have been rejected"
  exit 1
fi

echo "The invalid manifest was rejected as expected"
exit 0
