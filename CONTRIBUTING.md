# Contributing to fyaml

Thank you for your interest in contributing to fyaml!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/fyaml.git`
3. Create a branch: `git checkout -b your-feature-name`

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for convenience commands)
- dprint (for formatting Markdown, YAML, JSON, Dockerfiles) - [Install](https://dprint.dev/install/)
- pre-commit (optional, for git hooks) - `pip install pre-commit` or `brew install pre-commit`

### Building

```bash
go build .
```

Or use the Makefile:

```bash
make build
```

### Testing

```bash
go test ./...
```

Or:

```bash
make test
```

### Linting

The CI uses `golangci-lint`. Run locally:

```bash
golangci-lint run
```

Or:

```bash
make lint
```

### Formatting

Format Markdown, YAML, JSON, and Dockerfiles:

```bash
make format
```

Or check formatting without modifying files:

```bash
make format-check
```

### Pre-commit Hooks

Install pre-commit hooks to automatically format and lint before commits:

```bash
pre-commit install
```

This will run formatting checks, linting, and other validations before each commit.

## Code Style

- Follow standard Go formatting (`gofmt`/`goimports`)
- Format Markdown, YAML, JSON, and Dockerfiles with `dprint` (run `make format`)
- Run `go vet` before committing
- Ensure all tests pass
- Add tests for new functionality
- Keep code simple and focused (this is intentionally "boring" software)

## Submitting Changes

1. Ensure all tests pass
2. Run the linter: `golangci-lint run`
3. Format code and docs: `make format`
4. Update documentation if needed
5. Submit a pull request with a clear description

## Project Philosophy

fyaml is intentionally minimal and focused:

- **Spec-first** - Implements the FYAML specification exactly
- **Deterministic** - Identical input always produces identical output
- **No templating** - This is a structural compiler, not a programming language
- **Vendor-neutral** - Works with any YAML-based system

When proposing changes, please consider whether they align with these principles.

## Extensions

fyaml implements the FYAML specification exactly, but may include extensions that enhance functionality without breaking spec compliance.

**Guidelines for Extensions:**

- Extensions must be **opt-in** or clearly marked as non-spec behavior
- Extensions must not break spec-compliant behavior when disabled or not used
- Extensions should be clearly documented as such in the README
- When proposing extensions, explain how they enhance rather than replace spec behavior

**Example:** JSON support is an extension that:

- Accepts `.json` files in addition to `.yml`/`.yaml` (opt-in via file extension)
- Provides `--format json` output option (opt-in via flag)
- Default behavior remains spec-compliant (YAML-only input/output)

## Questions?

Open an issue for discussion.
