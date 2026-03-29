# CLI Reference

## Usage

```
outrunner [flags]
```

outrunner registers a GitHub Actions [scale set](https://github.com/actions/scaleset) on a repository or organization, then listens for jobs and provisions ephemeral runner environments for each one.

## Flags

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--url` | string | | Yes | Repository or organization URL. Examples: `https://github.com/owner/repo`, `https://github.com/org` |
| `--token` | string | | Yes | GitHub Personal Access Token (fine-grained). Needs **Administration** read/write permission on the target repository or organization. |
| `--config` | string | | Yes | Path to the YAML configuration file. See [Configuration Reference](configuration.md). |
| `--name` | string | `outrunner` | No | Scale set name. Also used as a prefix for runner names and orphan cleanup. |
| `--max-runners` | int | `2` | No | Maximum number of concurrent runners. outrunner will never provision more than this many environments at once, regardless of how many jobs are queued. |
| `-h`, `--help` | | | No | Show help. |

## Examples

Single-repo, Docker-only:

```bash
./outrunner \
  --url https://github.com/myorg/myrepo \
  --token ghp_xxx \
  --config outrunner.yml
```

Organization-wide, higher concurrency:

```bash
./outrunner \
  --url https://github.com/myorg \
  --token ghp_xxx \
  --config outrunner.yml \
  --name myorg-runners \
  --max-runners 8
```

## Behavior

- On startup, outrunner looks for an existing scale set with the given `--name`. If found, it reuses it. If not, it creates a new one with labels from the configuration file.
- On shutdown (Ctrl+C / SIGINT), outrunner stops all running environments and deletes the scale set.
- If outrunner is force-killed (SIGKILL), the scale set and any running environments may be left behind. On next startup, each provisioner cleans up orphaned environments whose names start with `--name`.

## GitHub PAT Requirements

Create a **fine-grained** Personal Access Token at [github.com/settings/tokens](https://github.com/settings/tokens?type=beta):

- **Resource owner:** The organization or user that owns the repository.
- **Repository access:** Select the target repository (or all repositories for org-wide use).
- **Permissions:** Administration → Read and write.

Classic tokens also work but fine-grained tokens are recommended for least-privilege.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Clean shutdown (Ctrl+C) |
| 1 | Error (invalid config, authentication failure, listener error) |
