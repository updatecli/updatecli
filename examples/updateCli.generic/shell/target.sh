#!/bin/bash

set -eux -o pipefail

# log writes the provided string to stderr
function log() {
  >&2 echo "${1}"
}

test -n "${1}" || { log 'ERROR: Missing argument'; exit 1; }

if [[ "${DRY_RUN}" == false ]]
then
  # updatecli apply
  if [[ "${1}" == 'change' ]]
  then
    log 'Change applied.'
    # Prints to stdout to let updatecli knows that there was a change
    command -v grep
    ls -l "$(command -v grep)"
  else
    log 'No changes applied.'
    exit 0
  fi
else
  # updatecli diff
  log 'Variable DRY_RUN is different than false'
  if [[ "${1}" == 'change' ]]
  then
    log 'Change but in dry run'
    # Prints to stdout to let updatecli knows what changed
    command -v grep
    ls -l "$(command -v grep)"
  else
    log "No changes applied in dry run"
    exit 0
  fi
fi
