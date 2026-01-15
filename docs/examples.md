# Examples

This page provides complex examples showing how multiple fyaml concepts work together. For basic examples and detailed documentation, see the [Usage Guide](usage/index.md).

All examples shown here correspond to runnable examples in the [examples directory](https://github.com/jksmth/fyaml/tree/main/examples) in the repository.

## Combining Multiple Concepts

This example demonstrates how root-level files, @ files, and deep nesting work together:

### Directory Structure

```
config/
  defaults.yml
  entities/
    @shared.yml
    item1/
      config.yml
      metadata.yml
    item2/
      settings.yml
  category1/
    item3.yml
    item4.yml
```

### Input Files

**`defaults.yml`:**

```yaml
project: example
version: 1.0.0
```

**`entities/@shared.yml`:**

```yaml
environment: production
region: us-east-1
```

**`entities/item1/config.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

**`entities/item1/metadata.yml`:**

```yaml
related:
  - id: example2
    attributes:
      name: related item
      tags:
        - tag1
```

**`entities/item2/settings.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another item
    tags:
      - tag2
```

**`category1/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags: []
```

**`category1/item4.yml`:**

```yaml
entity:
  id: example4
  attributes:
    name: fourth item
    tags: []
```

### Command

```bash
fyaml config/
```

### Output

```yaml
category1:
  item3:
    entity:
      id: example3
      attributes:
        name: third item
        tags: []
  item4:
    entity:
      id: example4
      attributes:
        name: fourth item
        tags: []
entities:
  environment: production
  item1:
    config:
      entity:
        id: example1
        attributes:
          name: sample name
          tags: []
    metadata:
      related:
        - attributes:
            name: related item
            tags:
              - tag1
          id: example2
  item2:
    settings:
      entity:
        id: example2
        attributes:
          name: another item
          tags:
            - tag2
  region: us-east-1
project: example
version: 1.0.0
```

**Key points:**

- `defaults.yml` at root merges `project` and `version` into the top level (root-level files merge directly)
- `entities/@shared.yml` merges `environment` and `region` into the `entities` map (uses `@` to merge into parent directory)
- Deep nesting (3 levels: `entities/item1/config.yml`) works naturally
- All concepts work together in a single structure

## Runnable Examples

The following examples demonstrate individual concepts. Each has a corresponding runnable example in the repository:

- **[Basic Example](https://github.com/jksmth/fyaml/tree/main/examples/basic)** - Simple directory structure
  ```bash
  fyaml examples/basic
  ```
  See [Basic Usage](usage/basic-usage.md) for details.

- **[@ Files Example](https://github.com/jksmth/fyaml/tree/main/examples/with-at-files)** - Using `@` files to merge content
  ```bash
  fyaml examples/with-at-files
  ```
  See [Directory Structure - @ Files](usage/directory-structure.md#-files) for details.

- **[@ Directories Example](https://github.com/jksmth/fyaml/tree/main/examples/with-at-directories)** - Using `@` directories for organization
  ```bash
  fyaml examples/with-at-directories
  ```
  See [Directory Structure - @ Directories](usage/directory-structure.md#-directories) for details.

- **[File Includes Example](https://github.com/jksmth/fyaml/tree/main/examples/with-includes)** - Including YAML structures and text files
  ```bash
  fyaml examples/with-includes --enable-includes
  ```
  See [File Includes](usage/file-includes.md) for details.

- **[Shared Includes Example](https://github.com/jksmth/fyaml/tree/main/examples/with-shared-includes)** - Sharing includes across multiple config directories using `--chroot`
  ```bash
  fyaml examples/with-shared-includes/config1 --chroot examples/with-shared-includes --enable-includes
  fyaml examples/with-shared-includes/config2 --chroot examples/with-shared-includes --enable-includes
  ```
  See [File Includes - Shared Includes Pattern](usage/file-includes.md#shared-includes-pattern) for details.

- **[Cross-File Anchors Example](https://github.com/jksmth/fyaml/tree/main/examples/with-anchors)** - Using YAML anchors across files
  ```bash
  fyaml examples/with-anchors --enable-anchors
  ```
  See [YAML Anchors and Aliases](usage/anchors-aliases.md) for details.

- **[Output Modes Example](https://github.com/jksmth/fyaml/tree/main/examples/with-modes)** - Comparing canonical vs preserve modes
  ```bash
  fyaml examples/with-modes              # Canonical mode
  fyaml examples/with-modes --mode preserve  # Preserve mode
  ```
  See [Output Modes](usage/output-modes.md) for details.

See the [examples directory](https://github.com/jksmth/fyaml/tree/main/examples) for all available examples.
