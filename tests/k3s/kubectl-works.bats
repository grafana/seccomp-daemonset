#!/usr/bin/env bats

load kubectl

function teardown() {
    _remove_namespace
}

@test "kubectl is accessible from the CLI" {
    run _kubectl version &>/dev/null
    [ "$status" -eq 0 ]
    [ -n "$output" ]
}
