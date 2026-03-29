# How to Run Multiple Backends Together

A single outrunner instance can serve Docker containers, libvirt VMs, and Tart VMs simultaneously. Jobs are routed to the right backend based on their `runs-on` label.

## Configuration

List all images in one config file. Each image declares its label and backend:

```yaml
images:
  - label: linux
    docker:
      image: outrunner-runner:latest

  - label: windows
    libvirt:
      path: /var/lib/libvirt/images/windows-builder.qcow2
      runner_cmd: 'C:\actions-runner\run.cmd'
      cpus: 4
      memory: 8192

  - label: macos
    tart:
      image: ghcr.io/cirruslabs/macos-sequoia-base:latest
      runner_cmd: /Users/admin/actions-runner/run.sh
      cpus: 4
      memory: 8192
```

outrunner initializes only the backends that are needed. If no image uses Docker, the Docker client is never created.

## Start outrunner

```bash
./outrunner \
  --url https://github.com/your/repo \
  --token ghp_xxx \
  --config mixed.yml \
  --max-runners 4
```

The `--max-runners` limit applies globally across all backends.

## Workflow Usage

```yaml
jobs:
  build-linux:
    runs-on: linux
    steps:
      - run: echo "Docker container"

  build-windows:
    runs-on: windows
    steps:
      - run: echo "KVM virtual machine"

  build-macos:
    runs-on: macos
    steps:
      - run: echo "Tart virtual machine"
```

## Multiple Images on the Same Backend

You can define multiple images for the same backend with different labels:

```yaml
images:
  - label: linux-small
    docker:
      image: runner:latest

  - label: linux-gpu
    docker:
      image: runner-with-cuda:latest

  - label: windows-2022
    libvirt:
      path: /images/win2022.qcow2
      runner_cmd: 'C:\actions-runner\run.cmd'

  - label: windows-2025
    libvirt:
      path: /images/win2025.qcow2
      runner_cmd: 'C:\actions-runner\run.cmd'
```

## Independent Scaling

If you need different `--max-runners` per backend, run separate outrunner instances with non-overlapping labels:

```bash
# Terminal 1: Docker (high concurrency)
./outrunner --name docker-pool --config docker.yml --max-runners 8 ...

# Terminal 2: VMs (limited concurrency)
./outrunner --name vm-pool --config vms.yml --max-runners 2 ...
```

## Platform Constraints

Not all backends work on all hosts:

| Backend | Linux host | macOS host |
|---------|-----------|------------|
| Docker  | Yes       | Yes (via Colima/Docker Desktop) |
| Libvirt | Yes       | No |
| Tart    | No        | Yes (Apple Silicon only) |

A single host can run Docker + libvirt (Linux) or Docker + Tart (macOS), but not all three.
