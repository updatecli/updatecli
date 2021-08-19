#!/bin/bash

set -eux -o pipefail

## This script outputs the provided argument only if DRY_RUN is "true"
if [[ "${DRY_RUN:-true}" == false ]]
then
  echo "$1"
fi
