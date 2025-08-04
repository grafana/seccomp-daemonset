#!/usr/bin/env bats

load docker

function teardown() {
    _remove_docker
}

@test "docker image starts" {
    # We want the container to start and be healthy.
    run _docker run --health-start-period=1s --health-start-interval=0.1s --name "$(_container_name)" --volume "$BATS_TEST_TMPDIR":/seccomp-profiles:ro -d "$DOCKER_IMAGE" --source /seccomp-profiles
    [ "$status" -eq 0 ]
    [ -n "$output" ]

    # Wait for the container to be healthy
    _wait_for_healthy "$output"
    [ "$status" -eq 0 ]
}
