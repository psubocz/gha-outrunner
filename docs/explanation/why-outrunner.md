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

## Design Decisions

### Single binary, no daemon dependencies

outrunner has no external dependencies beyond the backends themselves (Docker, libvirtd, Tart). No database, no message queue, no coordinator. State is held in memory. If outrunner restarts, it cleans up orphans and starts fresh.

### Scaleset API, not polling

outrunner uses GitHub's [scaleset API](https://github.com/actions/scaleset) rather than polling the REST API for queued jobs. The scaleset API pushes job notifications to listeners via long-polling, giving faster response times and no wasted API calls. This is the same API that ARC uses.

### Guest agents, not SSH

VM provisioners use hypervisor-level guest agents instead of SSH or WinRM. This avoids network configuration, credential management, and firewall rules entirely. See [Architecture](architecture.md) for details.

### JIT registration, not pre-registered runners

Each runner gets a one-time JIT configuration token. It registers with GitHub on startup, runs one job, and is never reused. No runner tokens are baked into images, and no long-lived credentials exist inside environments.

## When to Use Something Else

- **GitHub-hosted runners** are the simplest option if they meet your needs (standard environments, no secrets on the runner, acceptable cost).
- **ARC** is the right choice if you already run Kubernetes and want mature, battle-tested autoscaling.
- **Static self-hosted runners** work if you only need one or two runners and don't care about isolation between jobs.

outrunner fills the gap: you want ephemeral isolation, you don't want Kubernetes, and you have a server (or a Mac) to run it on.
