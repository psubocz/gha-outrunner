# Why outrunner

## The Problem

You want ephemeral GitHub Actions runners: a fresh environment for every job, destroyed after. This gives you isolation, reproducibility, and no state leaking between jobs.

GitHub's official solution is [Actions Runner Controller (ARC)](https://github.com/actions/actions-runner-controller). It works well but requires Kubernetes. If you're running on a bare metal server, a VPS, or a Mac mini under your desk, standing up a Kubernetes cluster just for CI runners is a lot of overhead.

## What outrunner Does Differently

outrunner is a single binary. No cluster, no operator, no CRDs. Install it, point it at a config file and a GitHub token, and it provisions runners on demand using whatever infrastructure you already have:

- **Docker** for Linux containers (fastest, simplest)
- **libvirt/KVM** for full VMs (Windows, or when you need real kernel isolation)
- **Tart** for macOS and ARM64 Linux VMs on Apple Silicon

You can mix all three backends in one config. A single outrunner instance can serve Docker containers for Linux jobs and KVM virtual machines for Windows jobs on the same host.

## Design Highlights

- **No daemon dependencies** - no database, no message queue, no coordinator. State is in memory. Cleans up orphans on restart.
- **Scaleset API, not polling** - same push-based API that ARC uses. No wasted API calls.
- **Guest agents, not SSH** - no network config, no credentials to manage. See [Architecture](architecture.md) for details.
- **JIT registration** - each runner gets a one-time token. No long-lived credentials in images.

## When to Use Something Else

- **GitHub-hosted runners** are the simplest option if they meet your needs (standard environments, acceptable cost).
- **ARC** is the right choice if you already run Kubernetes and want mature, battle-tested autoscaling.
- **Self-hosted runner with `container:`** is a decent option if you only need Docker and are OK with the runner process living outside the container. Each workflow must opt in to `container:` (or you enforce it globally with `ACTIONS_RUNNER_REQUIRE_JOB_CONTAINER=true`). With outrunner, every job gets a fresh environment by default.
- **Static self-hosted runners** work if you only need one or two runners and don't care about isolation between jobs.

outrunner fills the gap: you want ephemeral isolation, you don't want Kubernetes, and you have a server (or a Mac) to run it on.
