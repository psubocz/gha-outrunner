# gha-outrunner

Ephemeral GitHub Actions runners — no Kubernetes required.

![How gha-outrunner works](docs/info.png)

outrunner provisions fresh Docker containers or VMs for each GitHub Actions job, then destroys them when the job completes. It uses GitHub's [scaleset API](https://github.com/actions/scaleset) to register as an autoscaling runner group.

**Why?** GitHub's [Actions Runner Controller (ARC)](https://github.com/actions/actions-runner-controller) requires Kubernetes. If you're running on bare metal or a simple VPS, you shouldn't need a cluster just to get ephemeral runners. outrunner gives you the same isolation guarantees with Docker, libvirt, or Tart.

## Status

Early development. All three provisioners tested end-to-end.

## Quick Start

Create a config file:

```yaml
# outrunner.yml
images:
  - label: linux
    docker:
      image: my-runner:latest

  - label: windows
    libvirt:
      path: /var/lib/libvirt/images/ci-runners/windows-builder.qcow2
      runner_cmd: 'C:\actions-runner\run.cmd'
      cpus: 4
      memory: 8192

  - label: macos
    tart:
      image: ghcr.io/cirruslabs/macos-sequoia-base:latest
      runner_cmd: /actions-runner/run.sh
      cpus: 4
      memory: 8192
```

Build and run:

```bash
go build -o outrunner ./cmd/outrunner

./outrunner \
  --url https://github.com/your/repo \
  --token ghp_xxx \
  --config outrunner.yml \
  --max-runners 2
```

Then in your workflow:

```yaml
jobs:
  build-linux:
    runs-on: linux
    steps:
      - run: echo "Running in an ephemeral container!"

  build-windows:
    runs-on: windows
    steps:
      - run: echo "Running in an ephemeral VM!"

  build-macos:
    runs-on: macos
    steps:
      - run: echo "Running in a macOS VM!"
```

## Provisioners

| Provisioner | Platform | How it works |
|-------------|----------|--------------|
| Docker | Any (Linux, macOS, Windows) | Creates a container per job. Fastest startup. |
| libvirt | Linux servers (KVM/QEMU) | Boots a VM from a qcow2 golden image using copy-on-write overlays. Uses QEMU Guest Agent for command execution — no SSH/WinRM needed. |
| Tart | macOS (Apple Silicon) | Clones a VM from an OCI image. Uses Tart's guest agent (`tart exec`) for command execution. |

All provisioners follow the same pattern: create environment → start runner → destroy after job.

### Guest Agents

The libvirt and Tart provisioners use guest agents instead of SSH or WinRM to execute commands inside VMs. This means no network configuration, no credentials, no IP discovery — the agent communicates over a host-guest channel (virtio-serial for QEMU, native for Tart).

- **libvirt**: [QEMU Guest Agent](https://wiki.qemu.org/Features/GuestAgent) — must be installed in the golden image (included by default in [rgl/windows-vagrant](https://github.com/rgl/windows-vagrant) images)
- **Tart**: [Tart Guest Agent](https://github.com/cirruslabs/tart-guest-agent) — pre-installed in all [Cirrus Labs](https://github.com/cirruslabs) images

## License

MIT
