#!/usr/bin/env bash
set -euo pipefail

usage() {
    {
        echo "usage: $0 <directory> <image>"
        echo
        echo "Runs all BATS tests for the image given. It must already be loaded into the container runtime."
        echo "Requires bats-core to be installed (bats on Debian/Ubuntu/Arch, bats-core on Homebrew)."
    } >&2
    exit 1
}

if [ $# -ne 2 ]; then
    usage
fi

cd "$1" || {
    echo "fatal: could not change directory to $1" >&2
    exit 1
}

export DOCKER_IMAGE="$2"
if [ -z "${DOCKER_IMAGE:-}" ]; then
    usage
fi

if ! command -v bats &>/dev/null; then
    echo "fatal: bats is not installed" >&2
    exit 1
fi

PARALLEL=""
if command -v parallel &>/dev/null || command -v rush &>/dev/null; then
    PARALLEL="--jobs $(nproc)"
fi

# Find and run .bats files in $1
# shellcheck disable=SC2086 # we intentionally want to expand $PARALLEL
bats --formatter pretty $PARALLEL "$1"
