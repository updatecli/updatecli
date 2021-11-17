#!/bin/bash

set -eux -o pipefail

## Prints the version of grep
# Can be '2.5.1-FreeBSD' (macOS) or '3.4' (Ubuntu 20.04) for instance
grep --version 2>&1 | head -n1 | awk '{print $4}'
