# Usage Guide

This guide covers basic usage patterns, commands, and common workflows with fyaml. fyaml supports two output modes: canonical (default, sorted keys, no comments) and preserve (authored order and comments). Use the `--mode` flag to select a mode.

## Quick Start

Get started quickly with common tasks:

- **[Basic Usage](#basic-usage)** - Pack directories and output to files
- **[Output Modes](#output-modes)** - Choose between canonical (sorted) or preserve (authored order)
- **[Directory Structure](#directory-structure-rules)** - How files map to YAML structure
- **[File Includes](#file-includes)** - Include content from other files
- **[Troubleshooting](#troubleshooting)** - Common issues and solutions

For a complete example, see the [README Quick Example](../README.md#quick-example).

## Basic Usage

### Pack to stdout

The simplest usage is to pack the current directory and output to stdout:

```bash
fyaml
fyaml --mode preserve    # Preserve authored order and comments
```

This reads all YAML/JSON files in the current directory and outputs a single YAML document.

You can also specify a directory explicitly:

```bash
fyaml config/
```

**Note:** `fyaml pack [DIR]` is an alias for `fyaml [DIR]` and works identically for backward compatibility. Both forms pack the current directory when no directory is specified.

### Pack to a File

Write the output to a file using the `-o` or `--output` flag:

```bash
fyaml -o output.yml
```

### Pack Current Directory

If you don't specify a directory, fyaml packs the current directory:

```bash
cd config/
fyaml          # Packs current directory
fyaml pack     # Same as above (alias)
fyaml .        # Explicitly specify current directory
```

### Output as JSON

Use the `--format` or `-f` flag to output JSON instead of YAML:

```bash
fyaml config/ --format json -o output.json
fyaml config/ --format json --mode preserve -o output.json    # Key order preserved (JSON doesn't support comments)
```

**Note:** JSON format doesn't support comments, but in preserve mode, key order is maintained in the JSON output.

## Common Patterns

### Verify Output Matches Existing File

Use the `--check` flag to verify that the generated output matches an existing file:

```bash
fyaml -o output.yml --check
```

This is useful in CI/CD pipelines to ensure configuration hasn't changed unexpectedly.

**Exit codes:**

- `0` - Output matches the file
- `2` - Output differs from the file
- `1` - Error occurred during packing

### Combine with Other Tools

fyaml works well with other command-line tools:

```bash
# Pipe to jq for JSON processing
fyaml --format json | jq '.entities'

# Use with envsubst for variable substitution
fyaml | envsubst > config-final.yml

# Replace placeholders with sed
fyaml | sed 's/{{VERSION}}/v1.0.0/g' > output.yml

# Validate YAML output
fyaml | yamllint

# Format with yq
fyaml | yq eval -P
```

**Note:** fyaml doesn't support templating, but you can pipe output to tools like `envsubst`, `sed`, or `yq` for post-processing.

## Output Modes

fyaml supports two output modes that control how keys are ordered and whether comments are preserved:

### Canonical Mode (Default)

Canonical mode is the default and provides:

- **Sorted keys**: All map keys are sorted alphabetically
- **No comments**: Comments are removed from the output
- **Deterministic output**

This mode is ideal for:

- Tools that don't care about key ordering or comments
- When sorted keys make diffs more readable and predictable

```bash
fyaml --mode canonical    # Explicit canonical mode (default)
fyaml                     # Same as above
```

### Preserve Mode

Preserve mode maintains the authored structure:

- **Authored key order**: Keys appear in the order they were written in source files
- **Comments preserved**: All YAML comments are maintained in the output
- **Deterministic output**

This mode is ideal for:

- Maintaining documentation in comments
- Preserving the authored key order from source files

```bash
fyaml --mode preserve     # Preserve order and comments
fyaml -m preserve         # Shorthand
```

### Mode Comparison Example

Given the same input files, here's how the output differs:

**Input (`entities/item1.yml`):**

```yaml
# This is a comment
zebra: value-z
alpha: value-a
```

**Canonical mode output:**

```yaml
alpha: value-a
zebra: value-z
```

**Preserve mode output:**

```yaml
# This is a comment
zebra: value-z
alpha: value-a
```

### JSON Output and Modes

When using JSON output format:

- **Key order**: In preserve mode, key order is maintained in JSON output (JSON preserves object key order)
- **Comments**: JSON doesn't support comments, so comments are lost regardless of mode
- Both modes produce deterministic JSON output

```bash
fyaml --format json --mode preserve    # Key order preserved, comments lost
```

## Directory Structure Rules

**Note:** Files and directories are processed in alphabetical order (deterministic across operating systems). This means `@` directories are processed before regular files (because `@` sorts before letters), and within each category, items are sorted alphabetically.

### Merge Behavior

When multiple files contribute to the same key (e.g., root-level files, `@` files, or `@` directories), fyaml uses **shallow merge** semantics: the later file's value completely replaces the earlier one. Nested maps are not merged recursively.

**Example:**

If you have two files that both define `config`:

**`@shared1.yml`:**

```yaml
config:
  database:
    host: localhost
    port: 5432
```

**`@shared2.yml`:**

```yaml
config:
  database:
    port: 3306
```

The result is `config: {database: {port: 3306}}` (the entire nested map from `@shared2.yml` replaces the one from `@shared1.yml`), not `config: {database: {host: localhost, port: 3306}}`.

This behavior applies to all merging scenarios: root-level files, `@` files, and `@` directories.

### Basic Structure

Directory names become map keys in the output:

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
```

**`entities/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
    tags: []
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
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags: []
```

### Root-Level Files

Files at the root level merge directly into the output (their filename is not used as a key):

```
config/
  shared.yml       # Merges directly into root
  entities/
    item1.yml
```

**`shared.yml`:**

```yaml
version: 1.0.0
environment: production
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
environment: production
version: 1.0.0
```

**Note:** The `shared.yml` contents merge directly into the root. The filename `shared` is not used as a key. **Merging is shallow**: if multiple root-level files or `@` files define the same key, the later file's value completely replaces the earlier one (nested maps are not merged recursively).

### @ Files

Files starting with `@` merge their contents into the parent directory map:

```
config/
  entities/
    @shared.yml    # Merges into entities map
    item1.yml
```

**`entities/@shared.yml`:**

```yaml
environment: production
region: us-east-1
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
  environment: production
  region: us-east-1
```

The `@` prefix is removed, and the file's contents are merged directly into the parent map.

**Note:** While the specification allows multiple `@` files in the same directory, it's recommended to use only one per directory. If multiple `@` files exist, they all merge into the parent map, and key collisions are resolved by the last file processed. **Merging is shallow**: if two files define the same key with nested maps, the entire nested map from the later file replaces the earlier one (nested values are not merged recursively). Files are processed in alphabetical order (see [Directory Structure Rules](#directory-structure-rules) above).

### @ Directories

Directories starting with `@` merge their contents into the parent directory map, similar to `@` files. This allows directories to be used for organization without creating additional nesting levels in the output.

**Example:**

Directory structure:

```
config/
  entities/
    item1.yml
    @group1/              # Merges into entities map
      item2.yml
      item3.yml
    @group2/              # Merges into entities map
      item4.yml
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

**`entities/@group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
    tags:
      - tag1
```

**`entities/@group1/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags:
      - tag2
```

**`entities/@group2/item4.yml`:**

```yaml
entity:
  id: example4
  attributes:
    name: fourth item
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags:
          - tag1
  item3:
    entity:
      id: example3
      attributes:
        name: third item
        tags:
          - tag2
  item4:
    entity:
      id: example4
      attributes:
        name: fourth item
        tags: []
```

**Key Points:**

- The `@group1` and `@group2` directory names do not appear as keys in the output
- All files from `@` directories merge directly into the parent map (`entities` in this example)
- This is useful for organizing large numbers of files without creating deep nesting
- Empty `@` directories are ignored (no keys created)

**Use Cases:**

- Organizing large numbers of files without creating deep nesting
- Grouping related files logically while keeping flat output structure
- Maintaining organization in source while producing simpler output

**Edge Cases:**

- Empty `@` directories are ignored (no keys created)
- If both `@group1/` directory and `@group1.yml` file exist, both merge into parent. Files are processed in alphabetical order (see [Directory Structure Rules](#directory-structure-rules) above).
- Nested `@` directories are supported: `@group1/@shared/` merges into parent of `@group1/`

**Note:** This is an extension to the FYAML specification. See [EXTENSIONS.md](https://github.com/jksmth/fyaml/blob/main/EXTENSIONS.md) for information about extensions to the specification.

### Nested Directories

Directories can be nested to any depth:

```
config/
  category1/
    group1/
      item1.yml
      item2.yml
    group2/
      item3.yml
```

**`category1/group1/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: first item
    tags: []
```

**`category1/group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: second item
    tags:
      - tag1
```

**`category1/group2/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags: []
```

Produces:

```yaml
category1:
  group1:
    item1:
      entity:
        id: example1
        attributes:
          name: first item
          tags: []
    item2:
      entity:
        id: example2
        attributes:
          name: second item
          tags:
            - tag1
  group2:
    item3:
      entity:
        id: example3
        attributes:
          name: third item
          tags: []
```

## File Naming

### Supported Extensions

fyaml processes files with these extensions:

- `.yml`
- `.yaml`
- `.json`

You can mix these file types in the same directory structure. JSON files must be valid JSON (not JSON5 or other variants) and must have a top-level object (map), just like YAML files.

### Ignored Files

fyaml automatically ignores:

- Files and directories starting with `.` (dot files)
- Files without supported extensions

### File Name to Key Mapping

The filename (without extension) becomes the key in the output:

- `item1.yml` → key: `item1`
- `entity-item.yaml` → key: `entity-item`
- `config.json` → key: `config`

**Exceptions:**

- Root-level files: filename is ignored, contents merge directly
- `@` files: `@` prefix is removed, contents merge into parent

### Special Characters

File and directory names can contain hyphens, underscores, numbers, and mixed case. Special characters are preserved in the output key:

- `entity-item.yml` → key: `entity-item`
- `entity_item.yml` → key: `entity_item`
- `ItemV2.yml` → key: `ItemV2`

### File Name Collisions

If you have files with the same name but different extensions in the same directory (e.g., `item1.yml`, `item1.yaml`, `item1.json`), they all produce the same key. The last one processed will overwrite previous ones. Files are processed in alphabetical order (see [Directory Structure Rules](#directory-structure-rules) above). **Use a consistent extension throughout your project to avoid collisions.**

## Output Format

### YAML (Default)

Default output is YAML:

```bash
fyaml
```

YAML output uses standard YAML formatting with 2-space indentation by default. You can customize the indent using the `--indent` flag:

```bash
# Use 4-space indent
fyaml --indent 4
```

### JSON

Output JSON using the `--format` flag:

```bash
fyaml --format json
```

JSON output is formatted with 2-space indentation by default. You can customize the indent using the `--indent` flag:

```bash
# Use 4-space indent for JSON
fyaml --format json --indent 4
```

### Empty Output

When no files are found:

- **YAML format**: Returns empty output (0 bytes) - aligns with `yq` behavior
- **JSON format**: Returns `null` - aligns with `jq` behavior

A warning is printed to stderr in both cases.

## Deterministic Output

Both modes are deterministic, making output suitable for version control and comparison.

This means you can safely commit the generated output to version control, and use `--check` in CI to ensure the source files and generated output stay in sync:

```bash
# In CI/CD pipeline - verify config hasn't changed
fyaml -o config.yml --check
# Exits with code 2 if source files changed but output wasn't regenerated
```

## Working with Large Configurations

### Organize by Domain

Group related configuration by domain:

```
config/
  entities/
    item1.yml
    item2.yml
  category1/
    item3.yml
    item4.yml
  category2/
    item5.yml
    item6.yml
```

### Use @ Files for Shared Configuration

Use `@` files to share common configuration:

```
config/
  entities/
    @shared.yml      # Shared settings
    item1.yml        # Item-specific
    item2.yml        # Item-specific
```

The `@shared.yml` might contain:

```yaml
environment: production
region: us-east-1
monitoring:
  enabled: true
```

## File Includes

fyaml supports including content from other files using the `--enable-includes` flag. This is an extension to the FYAML specification. See [EXTENSIONS.md](https://github.com/jksmth/fyaml/blob/main/EXTENSIONS.md) for information about extensions.

This feature is useful for:

- Sharing common configuration across multiple files
- Keeping scripts and commands in separate files for better organization
- Reusing YAML fragments without duplication

### Include Mechanisms

fyaml supports three include mechanisms:

| Syntax          | Purpose                        | Use Case                          |
| --------------- | ------------------------------ | --------------------------------- |
| `!include`      | Include parsed YAML structures | Shared config, reusable fragments |
| `!include-text` | Include raw text content       | Scripts, SQL queries, commands    |
| `<<include()>>` | Alias for `!include-text`      | CircleCI style syntax             |

### Including YAML Structures (`!include`)

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

### Including Text Content (`!include-text`)

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

### CircleCI Style (`<<include()>>`)

The `<<include()>>` directive syntax is supported as an alias for `!include-text`. This syntax was inspired by CircleCI's orb pack implementation:

```yaml
# Both are equivalent:
command: !include-text scripts/hello.sh
command: <<include(scripts/hello.sh)>>
```

### Combining Include Mechanisms

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

### Nested Includes

Included files can contain their own includes:

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

### JSON File Support

fyaml supports includes in JSON files, with some limitations:

#### Using `<<include()>>` in JSON Files (Recommended)

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

#### Using YAML Tags in JSON Files

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

#### Including JSON Files from YAML

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

### Security

All includes are confined to the pack root directory:

- Paths are resolved relative to the file containing the include
- Absolute paths are allowed but must be within the pack root
- Attempts to escape the pack root (e.g., `../../etc/passwd`) are rejected

## Best Practices

1. **Keep files focused**: Each file should represent a single logical unit
2. **Use descriptive names**: File and directory names should clearly indicate their purpose
3. **Organize hierarchically**: Use directory structure to reflect configuration hierarchy
4. **Version control**: Commit both source directory structure and generated output
5. **Verify in CI**: Use `--check` flag in CI to catch unexpected changes
6. **Document structure**: Add README files in directories to explain organization
7. **Use includes sparingly**: Prefer filesystem structure for organization; use includes for reusable fragments and text content

## Limitations

### File Content Requirements

Each YAML/JSON file must contain a map (object/dictionary) at the top level. The file itself must be a map, but nested values within that map can be any YAML type (scalars, arrays, nested maps, etc.).

**Supported:**

```yaml
# ✅ Top-level is a map
entity:
  id: example1
  attributes:
    name: sample name
    tags: [tag1, tag2] # Array nested in map
    settings: # Nested map
      timeout: 30
```

**Not supported:**

```yaml
# ❌ Top-level is a scalar
hello

# ❌ Top-level is an array
- item1
- item2
```

If you attempt to pack a file containing a top-level scalar or array, fyaml will return an error with the exact format:

```
expected a map, got a <type> which is not supported at this time for "<filepath>"
```

Where `<type>` is the Go type (e.g., `string`, `[]interface{}`) and `<filepath>` is the full path to the problematic file.

### YAML Anchors and Aliases

YAML anchors (`&anchor`) and aliases (`*alias`) are resolved **within each individual file** during parsing. Anchors and aliases **cannot** reference values across different files—they only work within a single YAML document.

If you need shared values across files, use the `!include` feature (with `--enable-includes`) to include YAML content from other files at specific locations in your structure. This provides similar functionality to cross-file anchors:

```yaml
# shared/defaults.yml
timeout: 30
retries: 3

# entities/item1.yml
entity:
  id: example1
  config: !include ../shared/defaults.yml
  attributes:
    name: sample name
```

**Note:** `@` files can also be used to merge common configuration at the directory level, but `!include` is more flexible as it allows you to include content at any point in your YAML structure, not just at the directory level.

### Multi-Document YAML Files

YAML supports multiple documents in a single file, separated by `---`. However, **fyaml only processes the first document** in multi-document files. Subsequent documents are silently ignored.

Instead of using multi-document files, organize your resources using separate files:

```yaml
# Instead of this (multi-document):
config.yml:
  ---
  entity:
    id: example1
    attributes:
      name: first item
  ---
  entity:
    id: example2
    attributes:
      name: second item

# Use this (fyaml's filesystem-based approach):
config/
  item1.yml    # Contains the first entity config
  item2.yml    # Contains the second entity config
```

### Non-String Map Keys

YAML allows map keys to be non-string types (numbers, booleans, etc.), but **fyaml normalizes all keys to strings** in certain modes:

- **Canonical mode** (default): All non-string keys are converted to strings during processing. For example, `123: value` becomes `"123": value`.
- **Preserve mode** (`--mode preserve`): Non-string keys are preserved in YAML output, but **must be converted to strings for JSON output** (JSON format requirement).

**Example:**

```yaml
# Input file
123: "numeric key"
true: "boolean key"
```

**Canonical mode output:**
```yaml
"123": numeric key
"true": boolean key
```

**Preserve mode output (YAML):**
```yaml
123: numeric key
true: boolean key
```

**Preserve mode output (JSON):**
```json
{
  "123": "numeric key",
  "true": "boolean key"
}
```

### Converting `on`/`off` and `yes`/`no` to `true`/`false`

If your YAML files use `on`/`off` or `yes`/`no` for boolean values, fyaml will treat them as strings by default. This can cause issues if your tools expect actual boolean values.

**The Problem:**

If your source file contains:

```yaml
entity:
  id: example1
  attributes:
    enabled: on
    active: yes
```

The output will be:

```yaml
entity:
  attributes:
    active: "yes" # String, not boolean
    enabled: "on" # String, not boolean
  id: example1
```

This can cause validation errors or unexpected behavior in tools that expect boolean values.

**The Solution:**

Use the `--convert-booleans` flag to automatically convert these values:

```bash
fyaml --convert-booleans
```

Now the output will be:

```yaml
entity:
  attributes:
    active: true # Boolean, not string
    enabled: true # Boolean, not string
  id: example1
```

**What Gets Converted:**

The flag converts unquoted values to booleans:

| Input          | Output (with `--convert-booleans`) |
| -------------- | ---------------------------------- |
| `enabled: on`  | `enabled: true`                    |
| `enabled: off` | `enabled: false`                   |
| `enabled: yes` | `enabled: true`                    |
| `enabled: no`  | `enabled: false`                   |
| `enabled: y`   | `enabled: true`                    |
| `enabled: n`   | `enabled: false`                   |

**Important:** Quoted values are preserved as strings. If you want a value to remain a string, quote it:

| Input           | Output                   |
| --------------- | ------------------------ |
| `enabled: "on"` | `enabled: "on"` (string) |
| `name: on_call` | `name: on_call` (string) |

**Best Practice:**

For new files, use `true`/`false` directly in your source files. This avoids the need for the conversion flag:

```yaml
entity:
  id: example1
  attributes:
    enabled: true # Recommended: use true/false directly
    active: true
```

**Technical Note:** fyaml outputs YAML 1.2 format, which only recognizes `true`/`false` as booleans. Values like `on`/`off` and `yes`/`no` were valid booleans in YAML 1.1 but are treated as strings in YAML 1.2. The `--convert-booleans` flag converts these legacy values to their YAML 1.2 equivalents.

### Large Files

fyaml processes files in memory. For very large files (hundreds of MB), this could consume significant memory. However, for typical configuration files (KB to low MB range), performance is excellent. Keep individual files focused and reasonably sized.

## Subcommand and Directory Name Conflicts

### The Issue

If you have a directory named `pack` or `version`, you might encounter a conflict because these are also subcommand names. By default, Cobra (the CLI framework fyaml uses) checks for subcommands before processing positional arguments, so `fyaml pack` will always invoke the pack subcommand, not try to pack a directory named "pack".

### The Solution: `--dir` Flag

Use the `--dir` flag to explicitly specify the directory you want to pack. This avoids any ambiguity:

```bash
# Pack a directory named "pack"
fyaml --dir pack

# Pack a directory named "version"
fyaml --dir version

# The --dir flag takes precedence over positional arguments
fyaml --dir config/ config/  # Uses --dir value (config/)
```

### Alternative Solutions

If you prefer not to use the `--dir` flag, you can use relative or absolute paths:

```bash
# Use relative path with ./
fyaml ./pack

# Use absolute path
fyaml /path/to/pack
```

However, the `--dir` flag is the recommended approach as it's explicit and clear.

### Both Forms Work

Remember that both `fyaml` and `fyaml pack` work identically:

```bash
# These are equivalent:
fyaml config/
fyaml pack config/

# These are also equivalent:
fyaml --dir pack
fyaml pack --dir pack
```

## Troubleshooting

### Common Error Messages

**"pack error: failed to build filetree: ..."**

- Directory doesn't exist or path is incorrect
- Permission denied accessing the directory
- Invalid path format

**Solution:** Verify the directory path exists and you have read permissions.

**"pack error: failed to marshal tree: ..."**

- Invalid YAML/JSON syntax in one or more files
- File contains non-map top-level value (scalar or array)

**Solution:** Check the error message for the specific file path, then validate that file's YAML/JSON syntax.

**"expected a map, got a `<type>` which is not supported at this time for \"<filepath>\"**

- File has top-level scalar (e.g., just `hello`) or array (e.g., `- item1`) instead of a map
- Each file must start with a map/object structure

**Solution:** Wrap the content in a map. For example, change `hello` to `value: hello` or `- item1` to `items: [item1]`.

**"invalid format: <format> (must be 'yaml' or 'json')"**

- Invalid `--format` value provided

**Solution:** Use only `yaml` or `json` as the format value.

**"--check requires --output to be specified"**

- `--check` flag used without `--output`

**Solution:** Always specify `-o` or `--output` when using `--check`.

**"failed to read output file: ..."**

- Error reading file for `--check` comparison
- File may be unreadable (permission issue)

**Solution:** Check file permissions and ensure the file is readable.

**"warning: no YAML/JSON files found in directory: <path>"**

- Directory contains no `.yml`, `.yaml`, or `.json` files
- Not an error, but output will be empty/null

**Solution:** Verify you're pointing to the correct directory and that it contains YAML/JSON files.

### Debugging Tips

**Use verbose mode to see which files are processed:**

```bash
fyaml -v config/
```

This shows `[DEBUG] Processing: <filepath>` for each file, helping you identify:

- Which files are being read
- If files are being ignored (not shown in debug output)
- The order files are processed

**Verify directory structure:**

```bash
# List all YAML/JSON files
find config/ -type f \( -name "*.yml" -o -name "*.yaml" -o -name "*.json" \)

# Check file permissions
ls -la config/
```

**Check file syntax:**

```bash
# Validate YAML syntax
yamllint config/**/*.yml

# Or use yq to check syntax
yq eval . config/entities/item1.yml
```

**Understand unexpected output:**

1. Check if files are being processed: use `-v` flag
2. Verify file structure matches expected output structure
3. Remember that keys are sorted alphabetically in output (in canonical mode)
4. Check for `@` files that might be merging unexpectedly
5. Verify root-level files are merging as expected

**Verify includes are working:**

If using `--enable-includes`, check:

- Include paths are relative to the file containing the include
- Included files exist and are within the pack root
- Include syntax is correct (`!include`, `!include-text`, or `<<include()>>`)

**Check for file name collisions:**

Files with the same name but different extensions (e.g., `item1.yml` and `item1.yaml`) produce the same key. The last one processed overwrites previous ones. Use a consistent extension throughout your project.

## Next Steps

- See [Examples](examples.md) for detailed examples
- Review [Command Reference](reference.md) for all options
