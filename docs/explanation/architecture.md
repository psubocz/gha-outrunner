# Architecture

outrunner is a single binary that bridges GitHub's scaleset API with local infrastructure: Docker containers, KVM virtual machines, or Tart VMs.

## System Overview

```
GitHub Actions                        outrunner                          Infrastructure
─────────────                        ──────────                          ──────────────
                                ┌──────────────────┐
Workflow queues job ───────────►│    Listener       │
                                │  (scaleset API)   │
                                └────────┬─────────┘
                                         │ desired runner count
                                ┌────────▼─────────┐
                                │     Scaler        │
                                │  (JIT config,     │
                                │   runner tracking)│
                                └────────┬─────────┘
                                         │ start/stop
                                ┌────────▼─────────┐
                                │ MultiProvisioner  │──── label ─────► Docker
                                │  (label routing)  │──── label ─────► Libvirt
                                └───────────────────┘──── label ─────► Tart
```

## Components

### Listener

The listener connects to GitHub's scaleset service via a long-polling message session. When jobs arrive or complete, GitHub sends messages through this channel. The listener delegates to the Scaler.

This uses the [actions/scaleset](https://github.com/actions/scaleset) Go library, which handles authentication, session management, and message parsing.

### Scaler

The Scaler implements the `listener.Scaler` interface from the scaleset library. It has three responsibilities:

1. **HandleDesiredRunnerCount:** GitHub says "I need N runners." The Scaler generates JIT configs and calls the provisioner's `Start` for each new runner needed. It never exceeds `--max-runners`.

2. **HandleJobStarted:** A runner picked up a job. Currently just logs this.

3. **HandleJobCompleted:** A job finished. The Scaler calls the provisioner's `Stop` to tear down the environment.

The Scaler tracks active runners by name in a map, protected by a mutex. On shutdown, it stops all remaining runners.

### MultiProvisioner

The MultiProvisioner routes runner requests to the correct backend based on label matching. When a job arrives:

1. Match the job's labels against the configuration to find the right image.
2. Look up which backend type that image uses (docker, libvirt, or tart).
3. Forward the `Start` call to that backend's provisioner.

For `Stop`, it tries all registered backends since only one will have the runner.

### Provisioners

Each provisioner implements a simple interface:

```go
type Provisioner interface {
    Start(ctx context.Context, req *RunnerRequest) error
    Stop(ctx context.Context, name string) error
    Close() error
}
```

The `RunnerRequest` carries the runner name, JIT configuration token, job labels, and the matched image configuration. Each provisioner uses these to create an environment and launch the runner process inside it.

See the [Provisioner Reference](../reference/provisioners.md) for details on each backend's lifecycle.

## Scale Set Lifecycle

1. **Startup:** outrunner registers a scale set with GitHub, declaring what labels it can handle. GitHub starts routing matching jobs to it.

2. **Running:** The listener long-polls for messages. GitHub sends desired runner counts based on queued jobs. outrunner provisions environments to meet demand (up to `--max-runners`).

3. **Shutdown:** On SIGINT, outrunner stops all running environments, closes the message session, and deletes the scale set registration. This tells GitHub to stop routing jobs here.

If outrunner crashes without cleanup, the stale scale set persists. On next startup, outrunner detects it via get-or-create logic and reuses it. Each provisioner also cleans up orphaned environments from previous runs.

## Authentication Flow

```
PAT ──► scaleset.Client ──► Registration Token ──► Broker Session
                                                        │
                                                  Long-poll for messages
                                                        │
                                              JIT Config per runner
                                                        │
                                              Runner self-registers
                                              with GitHub on startup
```

The PAT is only used to create the scaleset client and generate JIT configs. It's never passed into containers or VMs. Each runner gets a short-lived JIT token that's valid for a single registration.

## Why Guest Agents Instead of SSH

The libvirt and Tart provisioners use guest agents (QEMU Guest Agent and Tart's built-in agent) instead of SSH or WinRM to execute commands inside VMs. This eliminates:

- **IP address discovery:** No need to wait for DHCP, query ARP tables, or configure static IPs.
- **Credential management:** No SSH keys or passwords to inject, rotate, or secure.
- **Network configuration:** The agent communicates over a host-guest channel (virtio-serial for QEMU, native for Tart), not the network.
- **Firewall concerns:** No ports to open. The channel is local to the hypervisor.

The tradeoff is that the guest agent must be installed in the base image, and the API is more limited than SSH (no interactive sessions, no file transfer). For running a single command (`run.sh --jitconfig ...`), it's more than sufficient.
