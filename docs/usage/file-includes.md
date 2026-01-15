# File Includes

fyaml supports including content from other files using the `--enable-includes` flag. This is an extension to the FYAML specification. See [EXTENSIONS.md](https://github.com/jksmth/fyaml/blob/main/EXTENSIONS.md) for information about extensions.

This feature is useful for:

- Sharing common configuration across multiple files
- Keeping scripts and commands in separate files for better organization
- Reusing YAML fragments without duplication

## Include Mechanisms

fyaml supports three include mechanisms:

| Syntax          | Purpose                        | Use Case                          |
| --------------- | ------------------------------ | --------------------------------- |
| `!include`      | Include parsed YAML structures | Shared config, reusable fragments |
| `!include-text` | Include raw text content       | Scripts, SQL queries, commands    |
| `<<include()>>` | Alias for `!include-text`      | CircleCI style syntax             |

## Security

All includes are confined to the chroot boundary (security boundary):

- **Default behavior**: Includes must be within the pack directory (DIR)
- **With `--chroot` flag**: Includes can come from outside DIR but must be within the chroot boundary
- Paths are resolved relative to the file containing the include
  - For nested includes, paths in included files are resolved relative to the included file's location
- Absolute paths are allowed but must be within the chroot boundary
- Attempts to escape the chroot boundary (e.g., `../../etc/passwd`) are rejected

**Note:** The chroot boundary is a security/confinement boundary (like Unix `chroot`). It controls where includes can come from, but does NOT affect what gets packed - that's still controlled by DIR.

## Including YAML Structures (`!include`)

Use `!include` to include and merge YAML content from another file:

```
config/
  entities/
    item1.yml
  shared/
    defaults.yml
```

**`shared/defaults.yml`:**

```yaml
timeout: 30
retries: 3
enabled: true
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  config: !include ../shared/defaults.yml
  attributes:
    name: sample name
    tags: []
```

Running `fyaml config/ --enable-includes`:

```yaml
entities:
  item1:
    entity:
      id: example1
      config:
        timeout: 30
        retries: 3
        enabled: true
      attributes:
        name: sample name
        tags: []
```

## Including Text Content (`!include-text`)

Use `!include-text` to include raw file content as a string value. This is ideal for scripts and commands:

```
config/
  entities/
    item1.yml
    scripts/
      hello.sh
```

**`entities/scripts/hello.sh`:**

```bash
#!/bin/bash
echo "Hello, World!"
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
  steps:
    - run:
        name: Greeting
        command: !include-text scripts/hello.sh
```

Running `fyaml config/ --enable-includes`:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
      steps:
        - run:
            name: Greeting
            command: |
              #!/bin/bash
              echo "Hello, World!"
```

## CircleCI Style (`<<include()>>`)

The `<<include()>>` directive syntax is supported as an alias for `!include-text`. This syntax was inspired by CircleCI's orb pack implementation:

```yaml
# Both are equivalent:
command: !include-text scripts/hello.sh
command: <<include(scripts/hello.sh)>>
```

## Combining Include Mechanisms

You can use all three mechanisms in the same project:

```yaml
# entities/item1.yml
entity:
  id: example1
  metadata: !include ../shared/metadata.yml # YAML structure
  attributes:
    name: sample name
  steps:
    - run:
        name: Greeting
        command: !include-text scripts/hello.sh # Text (tag syntax)
    - run:
        name: Farewell
        command: <<include(scripts/goodbye.sh)>> # Text (directive syntax)
```

## Nested Includes

Included files can contain their own includes. **Important:** Only `!include` supports nested includes. Text includes (`!include-text` and `<<include()>>`) are single-level only.

**Basic Nested Include:**

```yaml
# common/defaults.yml
base: !include base-defaults.yml
custom:
  timeout: 30
```

```yaml
# common/base-defaults.yml
retries: 3
debug: false
```

**Nested Includes with Relative Paths:**

When a file included via `!include` contains its own includes, paths in that included file are resolved relative to the included file's location, not the original file:

```
config/
  definitions/
    templates/
      worker.yml          # Includes ../../scripts/worker.sh
    services.yml          # Includes templates/worker.yml
  scripts/
    worker.sh
```

**`definitions/services.yml`:**

```yaml
services:
  worker: !include templates/worker.yml
```

**`definitions/templates/worker.yml`:**

```yaml
name: Worker Service
script: !include-text ../../scripts/worker.sh
```

When `services.yml` includes `templates/worker.yml`, the path `../../scripts/worker.sh` in `worker.yml` is resolved relative to `definitions/templates/` (the included file's location), correctly resolving to `config/scripts/worker.sh`.

**Text Includes are Single-Level:**

Text includes (`!include-text` and `<<include()>>`) load raw text content and do not process that text for further includes:

```yaml
# This works - YAML structure includes can nest
config: !include shared.yml # shared.yml can contain !include, !include-text, or <<include()>>

# This is single-level - the included text is not processed
command: !include-text script.sh # script.sh content is loaded as-is, even if it contains <<include()>>
```

## JSON File Support

fyaml supports includes in JSON files, with some limitations:

### Using `<<include()>>` in JSON Files (Recommended)

The `<<include()>>` directive works in JSON files since it's processed as a string value:

```json
{
  "entity": {
    "id": "example1",
    "attributes": {
      "name": "sample name",
      "command": "<<include(../scripts/hello.sh)>>"
    }
  }
}
```

This is **standard JSON** and will work with any JSON parser.

### Using YAML Tags in JSON Files

YAML tags (`!include`, `!include-text`) can also be used in JSON files:

```json
{
  "entity": {
    "id": "example1",
    "config": !include ../shared/defaults.json,
    "attributes": {
      "name": "sample name",
      "command": !include-text ../scripts/hello.sh
    }
  }
}
```

**Note:** This is **not standard JSON syntax**. Standard JSON parsers will reject files with YAML tags. However, since fyaml uses `yaml.Unmarshal` to parse JSON files (treating JSON as a subset of YAML), these tags will work when processed by fyaml.

### Including JSON Files from YAML

YAML files can include JSON files using `!include`:

```yaml
# entities/item1.yml
entity:
  id: example1
  config: !include ../shared/defaults.json
  attributes:
    name: sample name
```

The included JSON file will be parsed and merged into the YAML structure.

## Shared Includes Pattern

A common use case for the `--chroot` flag is sharing include files across multiple configuration directories within a project.

### Problem

Without `--chroot`, includes must be within the pack directory. This means:

- Shared includes must be duplicated in each config directory
- Or shared includes must use dot-prefix (`.shared/`) to exclude them from output

### Solution: Using `--chroot`

With the `--chroot` flag, you can place shared includes outside the pack directory but within a project root:

```
project-root/
  shared-includes/
    common.yaml
    defaults.yaml
  config1/
    app.yaml  # !include ../shared-includes/common.yaml
  config2/
    app.yaml  # !include ../shared-includes/common.yaml
```

**Usage:**

```bash
# From project-root/
fyaml config1 --chroot . --enable-includes
fyaml config2 --chroot . --enable-includes
```

**Key Points:**

- Relative paths in `!include` tags remain relative to the file containing the tag (portable)
- `DIR` (config1 or config2) still controls what gets packed - only files in that directory appear in output
- `--chroot` sets the security boundary, allowing includes from `shared-includes/` directory
- You can run `fyaml` from anywhere - relative paths work the same regardless of CWD

**Example:**

**`shared-includes/common.yaml`:**

```yaml
timeout: 30
retries: 3
enabled: true
```

**`config1/app.yaml`:**

```yaml
app:
  id: app1
  config: !include ../shared-includes/common.yaml
  name: My App
```

**`config2/app.yaml`:**

```yaml
app:
  id: app2
  config: !include ../shared-includes/common.yaml
  name: Another App
```

Running `fyaml config1 --chroot . --enable-includes` produces:

```yaml
app:
  id: app1
  config:
    timeout: 30
    retries: 3
    enabled: true
  name: My App
```

The shared `common.yaml` is included, but only files from `config1/` appear in the output structure.
