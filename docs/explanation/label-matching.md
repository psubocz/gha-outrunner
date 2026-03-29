# How Label Matching Works

## Overview

Labels connect workflows to runner environments. When a workflow specifies `runs-on: linux`, outrunner needs to know which image to provision. This happens at two levels:

1. **GitHub → outrunner:** GitHub routes jobs to the scale set based on registered labels.
2. **outrunner → provisioner:** outrunner matches the job to a specific image in the config.

## Scale Set Registration

On startup, outrunner registers a scale set with GitHub and declares which labels it handles. These labels come from the `label` field of each image in the configuration:

```yaml
images:
  - label: linux        # ← registered as a scale set label
    docker:
      image: runner:latest
  - label: windows      # ← registered as a scale set label
    libvirt:
      path: /images/win.qcow2
```

When a workflow uses `runs-on: linux`, GitHub sees that the `outrunner` scale set handles the `linux` label and routes the job to it.

## Image Matching

When outrunner receives a job, it matches the job's labels against the configured images. The first image whose label appears in the job's label set is selected.

```
Job labels: ["linux"]
Config images:
  1. label: linux    ← match
  2. label: windows
Result: image 1 (Docker)
```

### Fallback Behavior

The scaleset API does not currently expose job labels to the listener (see [actions/scaleset#20](https://github.com/actions/scaleset/issues/20)). When labels are empty, outrunner falls back to the **first image** in the configuration.

This means that in practice, if you only have one image, it always works. With multiple images, the first one acts as the default. This will improve as the scaleset API adds label forwarding.

## How `runs-on` Maps to Labels

In your workflow:

```yaml
jobs:
  build:
    runs-on: linux     # single label
```

or:

```yaml
jobs:
  build:
    runs-on: [self-hosted, linux, x64]   # multiple labels
```

GitHub matches the `runs-on` labels against scale sets. A scale set matches if it has **at least one** of the requested labels. outrunner registers all image labels on a single scale set, so any job requesting any configured label will be routed to it.

## Multiple Scale Sets

Each outrunner instance creates one scale set. If you need independent scaling for different backends (e.g., separate `--max-runners` for Docker vs. libvirt), run multiple outrunner instances with different `--name` values and non-overlapping labels:

```bash
# Instance 1: Docker runners
./outrunner --name docker-runners --config docker.yml --max-runners 8

# Instance 2: VM runners
./outrunner --name vm-runners --config vms.yml --max-runners 2
```

## Current Limitations

1. **No label forwarding from scaleset API:** Job labels aren't passed to the listener yet, so multi-image configs rely on the first-image fallback. Single-image configs are unaffected.

2. **No label update on reuse:** If outrunner finds an existing scale set (from a previous run), it reuses it without checking whether the registered labels still match the config. If you change labels in the config, delete the stale scale set first (or use a different `--name`).

3. **Single label per image:** Each image declares one label. If a workflow uses `runs-on: [self-hosted, linux]`, only the `linux` part is matched against image labels. The `self-hosted` label is ignored by outrunner (it's a GitHub convention, not an outrunner concept).
