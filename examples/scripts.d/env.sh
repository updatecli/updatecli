#!/usr/bin/env bash

set -eux

echo "PATH: $PATH"

if [ "$UPDATECLI_PIPELINE_STAGE" == "target" ]; then
  echo "DRY_RUN: $DRY_RUN"
fi

