#!/usr/bin/env bash
# this is a helper script for BATS tests. load it with `load kubectl`

if [ -z "${KUBECONFIG:-}" ]; then
    if [ -f /etc/rancher/k3s/k3s.yaml ]; then
        export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
    else
        echo "error: KUBECONFIG is not set and no default k3s kubeconfig file found" >&2
        # We do this because we don't want to run this on a production cluster on accident.
        exit 1
    fi
fi

if [ -z "${KUBECTL:-}" ]; then
    if command -v kubectl &>/dev/null; then
        KUBECTL="kubectl"
    elif command -v k3s &>/dev/null; then
        KUBECTL="k3s kubectl"
    else
        echo "error: kubectl or k3s is not installed" >&2
        exit 1
    fi
fi

_kubectl() {
    if ! $KUBECTL version &>/dev/null; then
        echo "error: can't get server version from kubectl" >&2
        return 1
    fi

    $KUBECTL "$@"
}

NAMESPACE_PREFIX="${NAMESPACE_PREFIX:-bats}"

_namespace() {
    echo "${NAMESPACE_PREFIX}-${BATS_SUITE_TEST_NUMBER}"
}

_remove_namespace() {
    _kubectl delete ns --force "$(_namespace)" &>/dev/null || true
    echo "info: removed namespace $(_namespace)" >&2
}

_helm() {
    if ! command -v helm &>/dev/null; then
        echo "error: helm is not installed" >&2
        return 1
    fi

    env HELM_NAMESPACE="$(_namespace)" helm "$@"
}
