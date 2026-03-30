# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project

gha-outrunner provisions ephemeral GitHub Actions runners without Kubernetes. It creates fresh Docker containers or VMs (libvirt/Tart) for each job via GitHub's scaleset API, then destroys them on completion.

## Build & Test Commands

```bash
make build          # Build outrunner binary
make test           # Run all tests
make lint           # Run golangci-lint
make clean          # Remove built binaries
make cross          # Cross-compile for Linux/macOS
go test ./...       # Run all tests directly
go test -run TestName ./path/to/pkg  # Run a single test
```

## Architecture

**Entry point**: `cmd/outrunner/main.go` — Cobra CLI that parses flags (`--url`, `--token`, `--config`, `--max-runners`), loads YAML config, resolves GitHub token (CLI flag → env var → systemd-creds → token_file), then spawns a worker goroutine per runner definition using errgroup.

**Core flow per worker**: Create/update scale set → create message session → instantiate provisioner → start listener → run scaler.

**Key components**:

- `config.go` — YAML config parsing/validation. Each runner must have labels and exactly one provisioner type (Docker/libvirt/Tart).
- `scaler.go` — Implements `listener.Scaler` interface. Manages runner lifecycle phases: Provisioning → Idle → Running → Stopping. One goroutine per runner instance, mutex-protected state, 30-second graceful shutdown timeout.
- `runner.go` — Runner phase enum and per-instance state with done channel for signaling.
- `log.go` — Custom `SimpleHandler` for slog with human-readable format (`YYYY-MM-DD HH:MM:SS LEVEL message key=val`).
- `provisioner.go` — `Provisioner` interface: `Start(ctx, req)`, `Stop(ctx, name)`, `Close()`.

**Provisioner backends** (in `provisioner/`):
- `docker/` — Creates containers with auto-remove.
- `libvirt/` — Creates qcow2 overlay VMs, communicates via QEMU Guest Agent.
- `tart/` — Clones macOS/ARM64 VMs, uses `tart exec` for guest communication.

All backends perform orphan cleanup on startup to remove leftover resources from previous runs. Cleanup is opt-in: provisioners that support it implement a `cleaner` interface (`Cleanup(prefix string)`), which is checked via type assertion in `main.go:cleanupOrphans()` — it's not part of the `Provisioner` interface.

**Version**: Set via ldflags at build time by GoReleaser (`-X main.version={{.Version}}`). Defaults to `"dev"` during development.

## Testing

Tests use Go's standard `testing` package. Key mocks live in `scaler_test.go`: `mockClient` (implements `ScaleSetClient`) and `mockProvisioner` (implements `Provisioner`). Use these when testing scaler behavior.

Provisioner backends (docker/libvirt/tart) are intentionally not unit-tested — they require real infrastructure and are integration-tested via dogfooding (the project runs its own CI on outrunner). Don't try to mock Docker/libvirt/Tart internals.

## Packaging & Multi-Repo

The default config template exists in two places that must stay in sync:
- `packaging/config.yml` — bundled in deb/rpm packages, installed to `/etc/outrunner/config.yml`
- `NetwindHQ/homebrew-tap` repo → `Formula/outrunner.rb` — generates config on `brew install`

Other packaging files:
- `packaging/outrunner.service` — systemd unit
- `scripts/postinstall.sh` / `scripts/preremove.sh` — deb/rpm lifecycle scripts
- `scripts/update-homebrew.sh` — called by release workflow to update the Homebrew formula

When modifying config schema, update: `config.go`, `packaging/config.yml`, `examples/`, and the Homebrew formula.

## Code Style

- `gofmt -s` enforced by CI
- Tabs for Go, 2-space indent for YAML/Markdown (`.editorconfig`)
- Error messages: lowercase, no trailing punctuation, wrap with context via `fmt.Errorf("operation: %w", err)`
- Linter config in `.golangci.yml`: errcheck, govet, staticcheck, unused, gocritic, ineffassign

## CI/CD

- **CI** (`.github/workflows/ci.yml`): build, test, vet, lint on push/PR to main (Go 1.26.1)
- **Release** (`.github/workflows/release.yml`): GoReleaser on version tags, produces binaries + deb/rpm packages, updates Homebrew formula
