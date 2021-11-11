#!/bin/bash

set -eux -o pipefail

# Test if the command grep is present
command -v grep
# Check file attributes of the grep binary
ls -l "$(command -v grep)"
