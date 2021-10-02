#!/bin/sh

set -eu

echo "$1" | cut -f1,2 -d'.'
