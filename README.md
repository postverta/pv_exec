## pv_exec

`pv_exec` is the root process of a development container. It exposes a few GRPC services
to the external world to perform operations within the container. The service descriptions
can be found under the `proto` directory. Here are some brief explanations:

### exec

The `exec` service allows the caller to execute a short-lived task within the
container. Conceptually, it is similar to the UNIX fork & exec routine, with
some advanced features that make task management easier.  The service is
inspired by the awesome [`runit` service](http://smarden.org/runit/), and takes
a configuration and controller pattern.

Each task needs to be pre-defined with a configuration directory in a similar
way as `runit`. The startup parameter `-exec-config-root` specifies the root
directory where the configurations are looked for. For all the task definitions
of Postverta, please refer to the `task/common` and `task/javascript`
directories in the `base-image` repository.

The name of the configuration directory defines the task name. The directory
must include an executable `run` script which will be executed with `bash` when
the task is triggered.

The directory can also optionally contain two files: `input_scopes` and
`output_scopes`. They describe the input and output scopes needed by the task,
and are used by the service to ensure the right concurrency. Basically, when a
task is executed, it will acquire a read-lock for each of the input scopes, as
well as a write-lock for each of the output scopes. This is critical to ensure
corretness when multiple tasks are accessing the same resource. For instance,
the `npm install` task will update the `node_modules` directory, which is
accessed when the `clone` task is executed to clone the entire workspace.

This part of the service is actually not related to Postverta in any
significant way, and can be used just as a simple init scheme for containers
(with some modifications).

### process

The `process` service is in charge of managing long running processes inside
the container. It is similar to `systemd` or any other service management
daemon.

One advanced feature of the service is the capability to detect whether the
running process is listening on a port, and expose the port to the host
machine. This is essentially how we allow external requests to be forwarded to
the running application process inside the container. This is also how we
expose the language server service to the frontend editor.

### worktree

The `worktree` service is used in conjunction of the `lazytree` in-memory file
system. The idea is to periodically take snapshots of the file system and
upload the data to blob storage. Please refer to the `lazytree` repository for
more details.

## Usage

Just build the binary and run it. It can be run as a daemon (with the `daemon`
sub-command) or as a test GRPC client (with the `exec` sub-command). The
typical deployment requires the binary to be deployed on each of the container
servers. The binary will be automatically run by `pv_agent` when a new
container is created.
