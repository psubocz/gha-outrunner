# Tutorial: Docker Runner on Linux

In this tutorial we will set up outrunner on a Linux server to run GitHub Actions jobs in ephemeral Docker containers. By the end, you'll trigger a workflow and watch it run inside a container that's created on demand and destroyed after.

## Prerequisites

- A Linux server (Ubuntu, Debian, Fedora, etc.) with Docker installed
- A GitHub repository you own

## 1. Build the Runner Image

First, we need a Docker image that has the GitHub Actions runner agent. outrunner includes a minimal Dockerfile.

Clone the repository and build the image:

```bash
git clone https://github.com/NetwindHQ/gha-outrunner.git
cd gha-outrunner
docker build -t outrunner-runner runner/
```

You should see the image build successfully:

```
Successfully tagged outrunner-runner:latest
```

## 2. Install outrunner

```bash
curl -LO https://github.com/NetwindHQ/gha-outrunner/releases/latest/download/outrunner_amd64.deb
sudo dpkg -i outrunner_amd64.deb
```

Or from source: `go install github.com/NetwindHQ/gha-outrunner/cmd/outrunner@latest`

## 3. Create a GitHub PAT

Go to [github.com/settings/tokens?type=beta](https://github.com/settings/tokens?type=beta) and create a fine-grained token:

- **Token name:** outrunner
- **Resource owner:** Your user or organization
- **Repository access:** Select the repository you want to use
- **Permissions:** Administration → Read and write

Copy the token. You'll need it in the next step.

## 4. Write a Configuration File

Create `outrunner.yml`:

```yaml
runners:
  linux:
    labels: [self-hosted, linux]
    docker:
      image: outrunner-runner:latest
```

This tells outrunner: create a scale set named "linux" with the labels `self-hosted` and `linux`, and provision Docker containers from the `outrunner-runner:latest` image for jobs that match.

## 5. Start outrunner

```bash
outrunner \
  --url https://github.com/YOUR_USER/YOUR_REPO \
  --token ghp_YOUR_TOKEN \
  --config outrunner.yml
```

You should see output like:

```
2026-03-30 14:05:09 INFO Loaded config runners=1
2026-03-30 14:05:09 INFO Looking for scale set scaleSet=linux
2026-03-30 14:05:10 INFO Creating scale set scaleSet=linux
2026-03-30 14:05:10 INFO Scale set ready scaleSet=linux id=3
2026-03-30 14:05:10 INFO Listening for jobs scaleSet=linux maxRunners=2
```

outrunner is now listening for jobs. Leave it running.

## 6. Create a Test Workflow

In your GitHub repository, create `.github/workflows/test-outrunner.yml`:

```yaml
name: Test Outrunner

on:
  workflow_dispatch:

jobs:
  hello:
    runs-on: [self-hosted, linux]
    steps:
      - run: echo "Hello from an ephemeral container!"
      - run: hostname
      - run: cat /etc/os-release
```

Push this file to the repository.

## 7. Trigger the Workflow

Go to your repository on GitHub → Actions → "Test Outrunner" → "Run workflow" → click the green button.

Back in the outrunner terminal, you should see:

```
2026-03-30 14:06:12 INFO Spawning runner scaleSet=linux scaler.name=linux-a1b2c3d4 scaler.runnerID=1
2026-03-30 14:06:13 INFO Container started scaleSet=linux docker.name=linux-a1b2c3d4 docker.image=outrunner-runner:latest docker.id=e3f4a5b6c7d8
2026-03-30 14:06:14 INFO Job started scaleSet=linux scaler.runnerName=linux-a1b2c3d4
2026-03-30 14:06:18 INFO Job completed scaleSet=linux scaler.runnerName=linux-a1b2c3d4 scaler.result=succeeded
2026-03-30 14:06:18 INFO Stopping runner scaleSet=linux scaler.name=linux-a1b2c3d4
```

The workflow run on GitHub should show a green checkmark.

## 8. Clean Up

Press Ctrl+C in the outrunner terminal. It will stop cleanly:

```
2026-03-30 14:06:19 INFO All runners shut down scaleSet=linux
2026-03-30 14:06:19 INFO Shut down cleanly
```

Any running containers are stopped. The scale set is kept on GitHub for reuse on next startup.

## What Happened

1. outrunner created a scale set named "linux" with the labels `self-hosted` and `linux`.
2. When you triggered the workflow, GitHub matched the `runs-on: [self-hosted, linux]` labels to this scale set and notified outrunner that a runner was needed.
3. outrunner created a Docker container from `outrunner-runner:latest` with a JIT registration token.
4. The runner inside the container registered with GitHub, picked up the job, ran it, and exited.
5. outrunner detected the job completion and stopped the container (which auto-removed itself).

## Next Steps

- [How to build a custom Docker runner image](../howto/custom-docker-image.md)
- [How to deploy as a systemd service](../howto/systemd-service.md)
- [How to run multiple backends together](../howto/mixed-backends.md)
