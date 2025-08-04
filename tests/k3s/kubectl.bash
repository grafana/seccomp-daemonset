#!/usr/bin/env bash
# this is a helper script for BATS tests. load it with `load kubectl`

_kubectl() {
    KUBECTL="${KUBECTL:-kubectl}"
    if ! command -v "$KUBECTL" &>/dev/null; then
        if ! command -v k3s &>/dev/null; then
            echo "error: kubectl or k3s is not installed" >&2
            return 1
        fi
        KUBECTL="k3s kubectl"
    fi

    if [ -z "${KUBECONFIG:-}" ]; then
        if [ -f /etc/rancher/k3s/k3s.yaml ]; then
            export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
        else
            echo "error: KUBECONFIG is not set and no default k3s kubeconfig file found" >&2
            # We do this because we don't want to run this on a production cluster on accident.
            return 1
        fi
    fi

    SUDO="${SUDO:-}"
    if [ -z "$SUDO" ]; then
        if ! $KUBECTL version &>/dev/null; then
            if command -v sudo &>/dev/null; then
                SUDO="sudo"
            elif command -v doas &>/dev/null; then
                SUDO="doas"
            else
                echo "error: kubectl seemingly requires elevated privileges but sudo and doas are not available" >&2
                return 1
            fi
        fi
    fi

    # shellcheck disable=SC2086 # we intentionally want to expand KUBECTL here
    if ! $SUDO $KUBECTL version &>/dev/null; then
        echo "error: can't get server version from kubectl" >&2
        return 1
    fi

    # shellcheck disable=SC2086 # we intentionally want to expand KUBECTL here
    $SUDO $KUBECTL "$@"
}

NAMESPACE_PREFIX="${NAMESPACE_PREFIX:-bats}"

_namespace() {
    echo "${NAMESPACE_PREFIX}-${BATS_SUITE_TEST_NUMBER}"
}

_remove_namespace() {
    _kubectl delete ns --force "$(_namespace)" &>/dev/null || true
    echo "info: removed namespace $(_namespace)" >&2
}
