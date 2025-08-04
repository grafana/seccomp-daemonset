#!/usr/bin/env bash
set -euo pipefail
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE"/../test.sh "$HERE" "$@"
