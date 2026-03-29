# How to Set Up for an Organization

By default, outrunner registers a scale set on a single repository. You can also register it at the organization level to serve all repositories.

## Organization-Level PAT

Create a fine-grained PAT at [github.com/settings/tokens?type=beta](https://github.com/settings/tokens?type=beta):

- **Resource owner:** Your organization
- **Repository access:** All repositories (or select specific ones)
- **Permissions:** Administration → Read and write

Note: The organization must allow fine-grained PATs. An org admin may need to enable this in Settings → Developer settings → Personal access tokens.

## Point to the Organization URL

Use the organization URL instead of a repository URL:

```bash
./outrunner \
  --url https://github.com/your-org \
  --token ghp_xxx \
  --config outrunner.yml \
  --name org-runners
```

The scale set is registered at the organization level. Any repository in the organization can use its labels in `runs-on`.

## Workflow Usage

Workflows in any repo in the org can target the runner:

```yaml
jobs:
  build:
    runs-on: linux
```

No per-repo configuration is needed.

## Multiple Organizations or Repos

Run separate outrunner instances with different `--name` values:

```bash
# Org-wide runners
./outrunner --name org-pool --url https://github.com/your-org ...

# Extra runners for a specific repo with heavy CI
./outrunner --name heavy-ci --url https://github.com/your-org/big-repo ...
```

## Security Considerations

- An organization-level PAT has broader access than a repository-level one. Use the minimum scope needed.
- Organization runners can be restricted to specific repositories via GitHub's runner group settings, but outrunner doesn't manage runner groups directly (it uses group 1, the default).
- Consider using separate outrunner instances per team or set of repos if you need access control boundaries.
