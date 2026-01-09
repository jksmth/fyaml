# fyaml Documentation

**fyaml** compiles a directory tree of YAML or JSON files into a single deterministic document (with optional comment and key order preservation).

## What is fyaml?

fyaml solves a common, recurring problem:

> Some tools expect configuration to live in a single YAML file, even as that file grows to thousands of lines.

fyaml lets you work with structure and small files, while still producing the single file those tools expect.

## Quick Start

```bash
# Install fyaml
curl -sSL https://raw.githubusercontent.com/jksmth/fyaml/main/install.sh | bash

# Pack current directory to stdout
fyaml

# Pack specific directory
fyaml config/

# Write to a file
fyaml -o output.yml
```

## Key Features

- **Directory structure → YAML structure**: Maps directory structure directly to YAML structure
- **Split large configs**: Break down thousands of lines into small, focused files
- **Deterministic output**: Choose between canonical mode (sorted keys, no comments) or preserve mode (authored order and comments). See [Output Modes](usage.md#output-modes) for details.
- **No logic or templating**: Pure structure compilation, no execution model
- **JSON support**: Accepts JSON input and can output JSON

## How It Works

- Directory names become map keys
- File names (without extension) become nested keys
- Files starting with `@` merge into the parent directory
- Root-level files merge directly into the output
- Output is deterministic. See [Output Modes](usage.md#output-modes) for details on canonical vs preserve modes

## Example

Given this directory structure:

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
```

**`entities/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
```

Running `fyaml config/` produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
  item2:
    entity:
      id: example2
      attributes:
        name: another name
```

For more examples, see the [Usage Guide](usage.md#basic-usage) or [Examples](examples.md).

## When to Use fyaml

Use fyaml when:

- ✅ You need to produce a single YAML or JSON file
- ✅ The configuration is large enough to benefit from structure
- ✅ Readable diffs and predictable output matter
- ✅ You want organization without adding logic

fyaml is **not** a good fit if you need:

- ❌ Conditionals
- ❌ Loops
- ❌ Variable resolution
- ❌ Runtime behavior

Those concerns are better handled by other tools.

## Documentation

- **[Installation](installation.md)** - How to install fyaml
- **[Usage Guide](usage.md)** - Basic usage, common patterns, and limitations
- **[Examples](examples.md)** - Detailed examples with outputs
- **[Command Reference](reference.md)** - Complete command and flag reference

For quick installation without Go, see [Installation - Quick Install](installation.md#quick-install-linuxmacos) or [Installation - Docker](installation.md#docker).

## Learn More

- View the [FYAML Specification](https://github.com/CircleCI-Public/fyaml/blob/master/fyaml-specification.md)
- Check out the [examples directory](https://github.com/jksmth/fyaml/tree/main/examples) in the repository
- See the [README](../README.md) for project overview and quick start
- Report issues or contribute on [GitHub](https://github.com/jksmth/fyaml)
