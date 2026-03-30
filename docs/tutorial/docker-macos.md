# Tutorial: Docker Runner on macOS

In this tutorial we will set up outrunner on a Mac to run GitHub Actions jobs in ephemeral Docker containers using Colima. By the end, you'll trigger a workflow and watch it run in a container on your Mac.

## Prerequisites

- A Mac (Intel or Apple Silicon)
- [Homebrew](https://brew.sh) installed
- A GitHub repository you own

## 1. Install Docker via Colima

macOS doesn't have native Docker support. Colima runs a Linux VM that hosts the Docker daemon. If you already have Docker Desktop or another Docker runtime, skip to step 2.

```bash
brew install colima docker
colima start
```

Verify Docker is working:

```bash
docker run --rm hello-world
```

You should see "Hello from Docker!"

## 2. Build the Runner Image

Clone the repository and build the runner image:

```bash
git clone https://github.com/NetwindHQ/gha-outrunner.git
cd gha-outrunner
docker build -t outrunner-runner runner/
```

Note: On Apple Silicon, this builds a linux/arm64 image. The official `ghcr.io/actions/actions-runner` base image supports both amd64 and arm64.

## 3. Install outrunner

```bash
brew tap NetwindHQ/tap
brew install outrunner
```

## 4. Create a GitHub PAT

Go to [github.com/settings/tokens?type=beta](https://github.com/settings/tokens?type=beta) and create a fine-grained token:

- **Token name:** outrunner
- **Resource owner:** Your user or organization
- **Repository access:** Select the repository you want to use
- **Permissions:** Administration → Read and write

Copy the token.

## 5. Write a Configuration File

Create `outrunner.yml`:

```yaml
runners:
  linux:
    labels: [self-hosted, linux]
    docker:
      image: outrunner-runner:latest
```

## 6. Start outrunner

```bash
outrunner \
  --url https://github.com/YOUR_USER/YOUR_REPO \
  --token ghp_YOUR_TOKEN \
  --config outrunner.yml
```

You should see:

```
2026-03-30 14:05:09 INFO Auto-detected Docker host scaleSet=linux docker.host=unix:///Users/you/.colima/default/docker.sock
2026-03-30 14:05:09 INFO Loaded config runners=1
2026-03-30 14:05:10 INFO Scale set ready scaleSet=linux id=3
2026-03-30 14:05:10 INFO Listening for jobs scaleSet=linux maxRunners=2
```

Notice outrunner auto-detected the Colima Docker socket. This works with Docker Desktop too.

## 7. Create a Test Workflow

In your repository, create `.github/workflows/test-outrunner.yml`:

```yaml
name: Test Outrunner

on:
  workflow_dispatch:

jobs:
  hello:
    runs-on: [self-hosted, linux]
    steps:
      - run: echo "Hello from an ephemeral container!"
      - run: uname -a
```

Push this file and trigger it from GitHub → Actions → "Test Outrunner" → "Run workflow".

## 8. Watch It Work

In the outrunner terminal:

```
2026-03-30 14:06:12 INFO Spawning runner scaleSet=linux scaler.name=linux-a1b2c3d4 scaler.runnerID=1
2026-03-30 14:06:13 INFO Container started scaleSet=linux docker.name=linux-a1b2c3d4 docker.image=outrunner-runner:latest
2026-03-30 14:06:18 INFO Job completed scaleSet=linux scaler.runnerName=linux-a1b2c3d4 scaler.result=succeeded
```

The workflow on GitHub should show a green checkmark.

## 9. Clean Up

Press Ctrl+C to stop outrunner.

If you're done with Colima too:

```bash
colima stop
```

## What's Different from Linux

The only difference is the Docker runtime. On Linux, Docker runs natively. On macOS, Docker runs inside a Linux VM (Colima, Docker Desktop, etc.). outrunner auto-detects the Docker socket regardless of which runtime you use.

Performance is slightly lower due to the VM layer, but for CI workloads this is rarely noticeable.

## Next Steps

- [Tutorial: Tart Linux runner on macOS](tart-linux-runner.md)
- [Tutorial: Tart macOS runner on macOS](tart-macos-runner.md)
- [How to build a custom Docker runner image](../howto/custom-docker-image.md)
