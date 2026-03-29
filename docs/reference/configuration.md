# Configuration Reference

outrunner uses a YAML configuration file to define runner environments. Each image entry maps a workflow label to a provisioner backend.

## Schema

```yaml
images:
  - label: <string>          # Required. The runs-on label this image satisfies.
    docker:                   # Use Docker backend.
      image: <string>         # Docker image name or tag.
    libvirt:                  # Use libvirt/KVM backend.
      path: <string>          # Path to the base qcow2 disk image.
      runner_cmd: <string>    # Command to start the runner. Default: /actions-runner/run.sh
      cpus: <int>             # vCPU count. Default: 4
      memory: <int>           # Memory in MiB. Default: 8192
    tart:                     # Use Tart backend.
      image: <string>         # OCI image URL or local VM name.
      runner_cmd: <string>    # Command to start the runner. Default: /actions-runner/run.sh
      cpus: <int>             # vCPU count. Default: 4
      memory: <int>           # Memory in MiB. Default: 8192
```

## Rules

- Each image must have a `label`.
- Each image must specify exactly one of `docker`, `libvirt`, or `tart`.
- Multiple images can use the same backend with different labels.
- Labels are registered on the GitHub scale set. Workflows use `runs-on: <label>` to target a specific image.

## Fields

### `images[].label`

**Required.** The label that workflows use in `runs-on` to request this environment. Must be unique across all images.

```yaml
# In config:
images:
  - label: linux-x64
    docker:
      image: runner:latest

# In workflow:
jobs:
  build:
    runs-on: linux-x64
```

### `images[].docker`

Provisions a Docker container per job.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `image` | string | (required) | Docker image reference. Can be a local tag (`runner:latest`) or a registry reference (`ghcr.io/org/runner:v1`). |

The container runs `./run.sh --jitconfig <config>` as its command. The image must have the GitHub Actions runner installed at the working directory. See [Image Requirements](image-requirements.md).

### `images[].libvirt`

Provisions a KVM/QEMU virtual machine per job using a copy-on-write overlay.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `path` | string | (required) | Absolute path to the base qcow2 disk image. This image is never modified; each job gets a CoW overlay. |
| `runner_cmd` | string | `/actions-runner/run.sh` | Command to execute inside the VM via the QEMU Guest Agent. For Windows: `C:\actions-runner\run.cmd` |
| `cpus` | int | `4` | Number of vCPUs allocated to the VM. |
| `memory` | int | `8192` | Memory in MiB allocated to the VM. |

### `images[].tart`

Provisions a Tart virtual machine per job by cloning from a base image.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `image` | string | (required) | Tart image source. Can be an OCI registry reference (`ghcr.io/cirruslabs/macos-sequoia-base:latest`) or a local VM name. |
| `runner_cmd` | string | `/actions-runner/run.sh` | Command to execute inside the VM via `tart exec`. |
| `cpus` | int | `4` | Number of vCPUs allocated to the VM. |
| `memory` | int | `8192` | Memory in MiB allocated to the VM. |

## Examples

Docker only:

```yaml
images:
  - label: linux
    docker:
      image: outrunner-runner:latest
```

Mixed backends:

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
      runner_cmd: /actions-runner/run.sh
      cpus: 4
      memory: 8192
```

Multiple images on the same backend:

```yaml
images:
  - label: linux-small
    docker:
      image: runner:latest

  - label: linux-large
    docker:
      image: runner-with-tools:latest
```

## Label Matching

When a job is assigned to outrunner, the system matches it to an image by label. If no labels are available from the scaleset API (a known limitation), the first image in the list is used as a fallback. See [How Label Matching Works](../explanation/label-matching.md).
