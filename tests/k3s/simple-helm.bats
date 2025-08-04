#!/usr/bin/env bats

load kubectl

function teardown() {
    _remove_namespace
}

@test "helm chart can be installed" {
    run _kubectl create ns "$(_namespace)"
    [ "$status" -eq 0 ]
    [ -n "$output" ]

    run _helm install --values "$BATS_TEST_DIRNAME"/simple-helm.yaml seccomp-daemonset "$BATS_TEST_DIRNAME"/../../helm/
    [ "$status" -eq 0 ]
    [ -n "$output" ]
}
