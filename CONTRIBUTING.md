# Contributing

## Getting Started

```bash
git clone https://github.com/psubocz/gha-outrunner.git
cd gha-outrunner
make build
make test
make lint    # requires golangci-lint
```

## Making Changes

1. Open an issue to discuss the change before writing code.
2. Fork the repo and create a branch.
3. Make your changes. Run `make test` and `make lint`.
4. Open a pull request.

## Code Style

- `gofmt -s` for formatting (enforced by CI).
- Keep functions focused and small.
- Error messages should be lowercase, no trailing punctuation.
