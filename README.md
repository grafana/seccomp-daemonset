# seccomp-daemonset

A Kubernetes `DaemonSet` application for installing seccomp profiles on nodes.

This aims to be as simplistic and minimal as possible while still fulfilling all our requirements at Grafana Labs.

## Requirements

The following are the goals of this application:

  * Install seccomp profiles on the host nodes.
    * **Why?** We need the seccomp profiles on the nodes to use them with our other pods.
      They are necessary to implement stronger security measures.
  * Provide metrics before and after doing so.
    * **Why?** We want to be able to know when a file can or cannot be inserted on a node, and why.
  * Minimal memory usage.
    * **Why?** We have options like Go and Rust, where memory usage is already tiny.
      The only long-running process that is actually necessary in this is whatever required to store and serve metrics.
    * **Bonus:** This directly correlates to costs.
  * Minimal CPU usage.
    * **Why?** We don't want to cost the node anything significant after initial setup.
      There is not much work to copying files.
    * **Bonus:** This directly correlates to costs.
  * Near-instant startup and teardown durations.
    * **Why?** The profiles must exist before any users of them start getting scheduled.

## Architecture

The app works like this in its lifecycle:

  * Start an HTTP server to serve `/metrics` and `/healthz`.
  * Loop over all files in the input directory:
    - Delete old file
    - Attempt to hard-link the file into the destination.
    - If fails, copy the file into the destination.
    - Mark a metric per file.
  * Wait for a signal to stop, while serving requests.

This is the entire application.
It is intended to be run in a Kubernetes `DaemonSet`, where it will consume near-zero resources while sleeping.

## License

This software is licensed under the GPL v3.0.
It is open-source software, available to the public.
Please read the terms of the license if you wish to use the code.
