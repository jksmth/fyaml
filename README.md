# fyaml

[![CI](https://github.com/jksmth/fyaml/actions/workflows/ci.yml/badge.svg)](https://github.com/jksmth/fyaml/actions/workflows/ci.yml)
[![CodeQL](https://github.com/jksmth/fyaml/actions/workflows/codeql.yml/badge.svg)](https://github.com/jksmth/fyaml/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/jksmth/fyaml/graph/badge.svg?token=YZOTQL769O)](https://codecov.io/gh/jksmth/fyaml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/jksmth/fyaml/badge)](https://securityscorecards.dev/viewer/?uri=github.com/jksmth/fyaml)
[![Signed Releases](https://img.shields.io/badge/releases-signed-green)](https://github.com/jksmth/fyaml#verification)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/jksmth/fyaml?include_prereleases&sort=semver)](https://github.com/jksmth/fyaml/releases)

**Docs:** https://jksmth.github.io/fyaml/ | **Install:** https://jksmth.github.io/fyaml/installation/ | **Usage:** https://jksmth.github.io/fyaml/usage/ | **Issues:** https://github.com/jksmth/fyaml/issues

**fyaml** compiles a directory tree of YAML or JSON files into a single deterministic document.

It exists to solve a common, recurring problem:

> Some tools expect configuration to live in a single YAML file, even as that file grows to thousands of lines.

fyaml lets you work with structure and small files, while still producing the single file those tools expect.

## Quick Example

Given a directory structure:

```
config/
  entities/
    item1.yml
    item2.yml
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags:
      - tag1
      - tag2
```

**`entities/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
    tags: []
```

Run:

```bash
fyaml pack config/
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags:
          - tag1
          - tag2
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags: []
```

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

- **Organize as you want** - Split large configs into small, focused files organized in directories
- **Predictable output** - Identical input always produces identical output, making diffs meaningful
- **No surprises** - Pure structure compilation with no logic, templating, or execution model
- **Build-time tool** - Runs as a build step, producing the single file your tools expect

For technical details on how directory structure maps to YAML, see [How It Works](https://jksmth.github.io/fyaml/#how-it-works).

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

**Note:** This downloads and executes a script. For verification steps, see [Verification](#verification) below.

### From Source (Go)

```bash
go install github.com/jksmth/fyaml@latest
```

### Docker

**Run directly:**

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/jksmth/fyaml:latest pack /workspace/examples/basic
```

For more installation options (pre-built binaries, Windows, multi-stage Docker), see the [Installation Guide](https://jksmth.github.io/fyaml/installation/).

## Verification

fyaml releases are signed with [cosign](https://github.com/sigstore/cosign) using keyless signing.

For verification steps (binaries, Docker images, SBOMs), see [Installation - Verification](https://jksmth.github.io/fyaml/installation/#verification).

## Usage

### Basic Usage

```bash
# Pack current directory to stdout (fyaml is equivalent to fyaml pack .)
fyaml

# Pack specific directory
fyaml pack /path/to/config

# Write to file
fyaml pack /path/to/config -o output.yml

# Output as JSON
fyaml pack /path/to/config --format json -o output.json

# Check if output matches existing file
fyaml pack /path/to/config -o output.yml --check

# Verbose output (show files being processed)
fyaml -v pack /path/to/config
```

### More examples and patterns

- See the [Usage Guide](https://jksmth.github.io/fyaml/usage/) for:
  - [Directory structure rules](https://jksmth.github.io/fyaml/usage/#directory-structure-rules) - Basic structure, root files, `@` files and directories
  - [File includes](https://jksmth.github.io/fyaml/usage/#file-includes) - Using `!include`, `!include-text`, and `<<include()>>`
  - [Limitations](https://jksmth.github.io/fyaml/usage/#limitations) - File content requirements, YAML anchors, multi-document files
- See the [Examples](https://jksmth.github.io/fyaml/examples/) and the [`examples/`](examples/) directory for runnable examples.
- See the [Command Reference](https://jksmth.github.io/fyaml/reference/) for:
  - [Commands](https://jksmth.github.io/fyaml/reference/#commands) - `pack`, `version`
  - [Flags reference](https://jksmth.github.io/fyaml/reference/#flags-reference) - All available flags and options

## Exit Codes

- `0` - Success
- `1` - Pack or IO error
- `2` - `--check` mismatch (exits immediately via `os.Exit(2)`)

## About

fyaml implements the [FYAML specification](SPECIFICATION.md) (also available at [github.com/CircleCI-Public/fyaml](https://github.com/CircleCI-Public/fyaml/blob/master/fyaml-specification.md)).

It's a small, focused tool that:

- Works with any YAML-based system
- Produces deterministic output (identical input always produces identical output)
- Has a minimal surface area focused on one task
- Does not implement templating, variables, or conditionals

**Need templating or variable substitution?** Use external tools like `envsubst` alongside fyaml. This keeps fyaml focused on structure compilation while allowing you to use specialized tools for templating. See [Usage Guide - Integration with Templating](https://jksmth.github.io/fyaml/usage/#integration-with-templating) for examples.

**Extensions:** fyaml includes optional extensions (like JSON support) that enhance functionality while maintaining spec compliance. See the [Extensions](#extensions) section for details.

**Implementation Note:** fyaml sorts all map keys alphabetically to ensure deterministic output. The FYAML specification does not specify key ordering, so this is an implementation choice that provides reproducibility and makes output suitable for version control and comparison.

## Extensions

fyaml includes the following extensions beyond the FYAML specification. These features are opt-in and do not affect spec-compliant behavior.

- **JSON Support** - Accept `.json` files and output JSON format. See [Usage Guide - Output Format](https://jksmth.github.io/fyaml/usage/#output-format) and [Command Reference - --format/-f](https://jksmth.github.io/fyaml/reference/#format--f) for details.
- **File Includes** - Use `!include`, `!include-text`, and `<<include()>>` directives to include content from other files. See [Usage Guide - File Includes](https://jksmth.github.io/fyaml/usage/#file-includes) for complete documentation.
- **Boolean Conversion** - Convert YAML 1.1 booleans (`on`/`off`, `yes`/`no`) to YAML 1.2 (`true`/`false`). See [Usage Guide - Converting on/off and yes/no to true/false](https://jksmth.github.io/fyaml/usage/#converting-onoff-and-yesno-to-truefalse) for details.
- **@ Directory Support** - Directories starting with `@` merge into parent map, similar to `@` files. See [Usage Guide - @ Directories](https://jksmth.github.io/fyaml/usage/#-directories) for details.

For complete documentation on all extensions, see the [Usage Guide](https://jksmth.github.io/fyaml/usage/) and [Command Reference](https://jksmth.github.io/fyaml/reference/).

## License

MIT License - see [LICENSE](LICENSE) for details.

## Attribution

Portions of this code are derived from the [CircleCI CLI](https://github.com/CircleCI-Public/circleci-cli), which is also licensed under the MIT License. See [NOTICE](NOTICE) for details.
