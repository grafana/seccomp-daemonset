#!/usr/bin/env bash
set -euo pipefail

HERE="$(dirname "$(readlink -f "$0")")"
if [ -z "$HERE" ]; then
    HERE="."
fi

usage() {
    {
        echo "usage: $0 <image> [flags to bats]"
        echo
        echo "Runs all BATS tests for the image given. It must already be loaded into the container runtime."
        echo "Requires bats-core to be installed (bats on Debian/Ubuntu/Arch, bats-core on Homebrew)."
    } >&2
    exit 1
}

if [ $# -lt 1 ]; then
    usage
fi

export DOCKER_IMAGE="$1"
if [ -z "${DOCKER_IMAGE:-}" ]; then
    usage
fi
shift

if ! command -v bats &>/dev/null; then
    echo "fatal: bats is not installed" >&2
    exit 1
fi
if ! command -v docker &>/dev/null; then
    echo "fatal: docker is not installed" >&2
    exit 1
fi

PARALLEL=""
if command -v parallel &>/dev/null || command -v rush &>/dev/null; then
    PARALLEL="--jobs $(nproc)"
fi

# Find and run .bats files in $HERE
# shellcheck disable=SC2086 # we intentionally want to expand $PARALLEL
bats --formatter pretty $PARALLEL "$HERE" "$@"
