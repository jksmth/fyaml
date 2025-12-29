# fyaml

[![CI](https://github.com/jksmth/fyaml/actions/workflows/ci.yml/badge.svg)](https://github.com/jksmth/fyaml/actions/workflows/ci.yml)
[![CodeQL](https://github.com/jksmth/fyaml/actions/workflows/codeql.yml/badge.svg)](https://github.com/jksmth/fyaml/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/jksmth/fyaml/graph/badge.svg?token=YZOTQL769O)](https://codecov.io/gh/jksmth/fyaml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/jksmth/fyaml/badge)](https://securityscorecards.dev/viewer/?uri=github.com/jksmth/fyaml)
[![Signed Releases](https://img.shields.io/badge/releases-signed-green)](https://github.com/jksmth/fyaml#verification)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/jksmth/fyaml?include_prereleases&sort=semver)](https://github.com/jksmth/fyaml/releases)

**fyaml** compiles a directory tree of YAML or JSON files into a single deterministic document.

It exists to solve a common, recurring problem:

> Some tools expect configuration to live in a single YAML file, even as that file grows to thousands of lines.

fyaml lets you work with structure and small files, while still producing the single file those tools expect.

---

## What problem this solves

Many systems are designed around a **single YAML configuration file**:

- CI/CD platforms
- API specifications (such as OpenAPI)
- Tools that do not support includes or composition

As configurations grow, this becomes difficult to manage:

- Files reach thousands of lines
- Merge conflicts become common
- Reviews get harder, not easier
- Structure is implied by indentation and comments
- Confidence in changes drops over time

fyaml solves this by separating how configuration is **authored** (as files and directories) from how it is **consumed** (as a single document).

You organize configuration as directories and files.
fyaml compiles that structure into the single document the target system expects.

There is no logic, templating, or execution model involved.

---

## What fyaml does

fyaml is intentionally limited in scope to keep output predictable and diffs trustworthy.

- Maps directory structure directly to YAML structure
- Lets you split large configs into small, focused files
- Produces a single, predictable output document
- Runs as a build-time step

**How it works:**
- Directory names become map keys
- File names (without extension) become nested keys
- Files starting with `@` merge into the parent directory
- Root-level files merge directly into the output
- Output is deterministic with keys sorted alphabetically

---

## When to use this

Use fyaml when:

- You need to produce a single YAML or JSON file
- The configuration is large enough to benefit from structure
- Readable diffs and predictable output matter
- You want organization without adding logic

fyaml is not a good fit if you need:
- conditionals
- loops
- variable resolution
- runtime behavior

Those concerns are better handled by other tools.

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/jksmth/fyaml/main/install.sh | bash
```

### From Source (Go)

```bash
go install github.com/jksmth/fyaml@latest
```

### From Pre-built Binaries

Download the latest release from the [GitHub releases page](https://github.com/jksmth/fyaml/releases).

**Linux/macOS:**
```bash
curl -L https://github.com/jksmth/fyaml/releases/latest/download/fyaml_linux_amd64.tar.gz | tar xz
chmod +x fyaml
./fyaml pack examples/basic
```

**Windows:** Download the `.zip` file from releases and extract.

### Docker

**Run directly:**
```bash
docker run --rm -v $(pwd):/workspace ghcr.io/jksmth/fyaml:latest pack /workspace/examples/basic
```

**Use in multi-stage builds:**
```dockerfile
# Build stage - copy fyaml binary
FROM ghcr.io/jksmth/fyaml:latest AS fyaml

# Your application stage
FROM your-base-image:latest
COPY --from=fyaml /fyaml /usr/local/bin/fyaml

# Use fyaml in your build process
COPY config/ /config/
RUN fyaml pack /config > /app/config.yml
```

Releases are cryptographically signed. See [Verification](#verification) below.

## Verification

fyaml releases are signed with [cosign](https://github.com/sigstore/cosign) using keyless signing, providing cryptographic proof that artifacts are authentic and haven't been tampered with.

### Verify Binary Signatures

```bash
# Download the release and signature files
VERSION="v1.0.0"
wget https://github.com/jksmth/fyaml/releases/download/${VERSION}/checksums.txt
wget https://github.com/jksmth/fyaml/releases/download/${VERSION}/checksums.txt.sigstore

# Verify signature
cosign verify-blob --certificate-identity-regexp '^https://github.com/jksmth/fyaml' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt \
  --bundle checksums.txt.sigstore
```

### Verify Docker Images

```bash
cosign verify --certificate-identity-regexp '^https://github.com/jksmth/fyaml' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/jksmth/fyaml:v1.0.0
```

### Software Bill of Materials (SBOM)

SPDX SBOMs are included with each release, providing a complete inventory of all dependencies. Download from the [releases page](https://github.com/jksmth/fyaml/releases) (files ending in `.sbom.spdx.json`).

## Usage

### Basic Usage

```bash
# Pack current directory to stdout
fyaml

# Pack specific directory
fyaml pack /path/to/config

# Write to file
fyaml pack /path/to/config -o output.yml

# Output as JSON
fyaml pack /path/to/config --format json -o output.json

# Check if output matches existing file
fyaml pack /path/to/config -o output.yml --check
```

### Examples

Given a directory structure:

```
config/
  components/
    database.yml
    cache.yml
  features/
    auth.yml
    payments.yml
```

Running `fyaml pack config` produces:

```yaml
components:
  database: <contents of database.yml>
  cache: <contents of cache.yml>
features:
  auth: <contents of auth.yml>
  payments: <contents of payments.yml>
```

For complete runnable examples, see the [`examples/`](examples/) directory.

### Root Files

Files at the root level merge directly into the output (their filename is not used as a key):

```
config/
  metadata.yml     # Merges directly into root
  services/
    api.yml
```

Produces:

```yaml
services:
  api: <contents>
metadata: <contents of metadata.yml>
```

### @ Files

Files starting with `@` merge into their parent directory map:

```
config/
  services/
    @common.yml    # Merges into services map
    api.yml
```

Produces:

```yaml
services:
  <contents of @common.yml>
  api: <contents of api.yml>
```

### YAML Anchors and Aliases

YAML anchors (`&anchor`) and aliases (`*alias`) are resolved within each individual file during parsing. Anchors and aliases **cannot** reference values across different files—they only work within a single YAML document.

When files are processed, any anchors and aliases are expanded to their actual values. The final output contains the expanded values, not the anchor/alias references.

For example, if `config.yml` contains:
```yaml
defaults: &defaults
  timeout: 30
  retries: 3

service:
  <<: *defaults
  name: api
```

The output will contain the expanded values:
```yaml
defaults:
  timeout: 30
  retries: 3
service:
  timeout: 30
  retries: 3
  name: api
```

### Multi-Document YAML Files

YAML supports multiple documents in a single file, separated by `---`. However, **fyaml only processes the first document** in multi-document files. Subsequent documents are silently ignored.

**Why this limitation exists:**

fyaml uses the filesystem structure to organize YAML—directory names and file names become keys in the output. This design assumes **one file = one logical unit**. Multi-document files conflict with this model because:

- A single filename cannot represent multiple documents
- The filesystem structure (which fyaml relies on) already provides a way to organize multiple resources
- Supporting multi-document files would create ambiguity about which document gets the filename as a key

**What to do instead:**

Instead of using multi-document files, organize your resources using separate files:

```yaml
# Instead of this (multi-document):
resources.yml:
  ---
  kind: Deployment
  metadata:
    name: api
  ---
  kind: Service
  metadata:
    name: api

# Use this (fyaml's filesystem-based approach):
resources/
  deployment.yml    # Contains the Deployment
  service.yml       # Contains the Service
```

This approach:
- Works naturally with fyaml's directory structure model
- Makes each resource easy to find and edit
- Allows you to use directory names to group related resources
- Produces clear, predictable output structure

### File Content Requirements

**Each YAML/JSON file must contain a map (object/dictionary) at the top level.** The file itself must be a map, but nested values within that map can be any YAML type (scalars, arrays, nested maps, etc.).

This requirement exists because fyaml merges file contents into the parent directory map structure. Only maps can be merged into other maps.

**Examples:**

```yaml
# ✅ Supported - top-level is a map
name: api
items: [1, 2, 3]           # Array nested in map
value: 42                   # Scalar nested in map
settings:                   # Nested map
  timeout: 30
  retries: 3

# ❌ Not supported - top-level is not a map
hello                       # Scalar
- item1                     # Array
- item2
- item3
```

If you attempt to pack a file containing a top-level scalar or array, fyaml will return an error: `expected a map, got a <type> which is not supported at this time for "<filepath>"`.

## Exit Codes

- `0` - Success
- `1` - Pack or IO error
- `2` - `--check` mismatch

## About

fyaml implements the [FYAML specification](SPECIFICATION.md) (also available at [github.com/CircleCI-Public/fyaml](https://github.com/CircleCI-Public/fyaml/blob/master/fyaml-specification.md)).

It's a small, focused tool that:
- Works with any YAML-based system (CI/CD, Kubernetes, APIs, etc.)
- Produces deterministic output (identical input always produces identical output)
- Has a minimal surface area focused on one task
- Does not implement templating, variables, or conditionals

**Need templating or variable substitution?** Use external tools like `envsubst` alongside fyaml. For example:

If your `config/services/api.yml` contains:
```yaml
name: ${SERVICE_NAME}
replicas: ${REPLICA_COUNT}
image: ${IMAGE_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
```

Set environment variables and run:
```bash
export SERVICE_NAME=api
export REPLICA_COUNT=3
export IMAGE_REGISTRY=ghcr.io
export IMAGE_NAME=myapp
export IMAGE_TAG=v1.0.0

fyaml pack config/ | envsubst > config-final.yml
```

This keeps fyaml focused on structure compilation while allowing you to use specialized tools for templating.

**Extensions:** fyaml includes optional extensions (like JSON support) that enhance functionality while maintaining spec compliance. See the [Extensions](#extensions) section for details.

**Implementation Note:** fyaml sorts all map keys alphabetically to ensure deterministic output. The FYAML specification does not specify key ordering, so this is an implementation choice that provides reproducibility and makes output suitable for version control and comparison.

## Extensions

fyaml includes the following extensions beyond the FYAML specification. These features are opt-in and do not affect spec-compliant behavior.

### JSON Support

**Input:** fyaml accepts `.json` files in addition to `.yml` and `.yaml` files. JSON files are processed the same way as YAML files.

**Output:** Use the `--format` flag (or `-f`) to output JSON instead of YAML:

```bash
fyaml pack config/ --format json
fyaml pack config/ -f json -o output.json
```

The default output format is YAML. JSON output is formatted with 2-space indentation.

**Empty Output Behavior:**
- YAML format: Returns empty output (0 bytes) when no files found (aligns with yq)
- JSON format: Returns `null` when no files found (aligns with jq)

**Note:** These extensions are not part of the FYAML specification. For spec-compliant behavior, use only `.yml` and `.yaml` files with YAML output (default).

## License

MIT License - see [LICENSE](LICENSE) for details.

## Attribution

Portions of this code are derived from the [CircleCI CLI](https://github.com/CircleCI-Public/circleci-cli), which is also licensed under the MIT License. See [NOTICE](NOTICE) for details.

