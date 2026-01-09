# Command Reference

Complete reference for all fyaml commands, flags, and options.

## Global Flags

These flags apply to all commands:

### `-v, --verbose`

Show debug output. All verbose output is written to stderr, so it doesn't interfere with piped stdout.

**Usage:**

```bash
fyaml -v pack config/
fyaml --verbose pack config/ -o output.yml
```

**Default:** `false` (disabled)

**Behavior:**

- When enabled, shows `[DEBUG] Processing: <filepath>` for each YAML/JSON file processed
- Warnings (e.g., empty directory) are always shown with `[WARN]` prefix, regardless of verbose flag
- All output goes to stderr, so stdout remains clean for piping
- Useful for debugging which files are being processed

**Output Example:**

```bash
$ fyaml -v pack config/
[DEBUG] Processing: /path/to/config/entities/item1.yml
[DEBUG] Processing: /path/to/config/entities/item2.yml
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

## Commands

### `fyaml [DIR]`

Compile a directory of YAML/JSON files into a single document. This is the primary command.

**Synopsis:**

```bash
fyaml [global flags] [DIR] [flags]
```

**Arguments:**

- `DIR` - Directory to pack (default: current directory)

**Flags:**

- `--dir string` - Explicitly specify directory to pack (avoids subcommand conflicts)
- `-o, --output string` - Write output to file (default: stdout)
- `-c, --check` - Compare generated output to `--output`, exit non-zero if different
- `-f, --format string` - Output format: `yaml` or `json` (default: `yaml`)
- `-m, --mode string` - Output mode: `canonical` (sorted keys, no comments) or `preserve` (authored order and comments) (default: `canonical`)
- `--indent int` - Number of spaces for indentation (default: `2`)
- `--enable-includes` - Process file includes (`!include`, `!include-text`, `<<include()>>`) (extension)
- `--convert-booleans` - Convert unquoted YAML 1.1 booleans to `true`/`false`
- `-V, --version` - Print version information and exit

**Examples:**

```bash
# Pack current directory to stdout
fyaml

# Pack specific directory
fyaml config/

# Write to file
fyaml -o output.yml

# Output as JSON
fyaml config/ --format json -o output.json

# Preserve authored order and comments
fyaml config/ --mode preserve -o output.yml

# Verify output matches file
fyaml -o output.yml --check

# Pack directory with conflicting name (e.g., directory named "pack")
fyaml --dir pack

# With verbose output
fyaml -v config/

# Show version
fyaml --version
```

**Exit Codes:**

- `0` - Success
- `1` - Pack or IO error
- `2` - `--check` mismatch (output differs from file)

### `fyaml pack [DIR]` (Alias)

An alias for `fyaml [DIR]` that works identically. Maintained for backward compatibility.

**Synopsis:**

```bash
fyaml [global flags] pack [DIR] [flags]
```

All flags and behavior are identical to the main `fyaml [DIR]` command.

**Examples:**

```bash
# These are equivalent:
fyaml config/
fyaml config/

# Both work with all flags:
fyaml config/ --enable-includes
fyaml config/ --enable-includes
```

### `fyaml version`

Print version information. Both `fyaml version` (subcommand) and `fyaml --version` or `fyaml -V` (flag) work identically.

**Synopsis:**

```bash
fyaml version
# or
fyaml --version
```

**Output:**

```
1.0.3 (commit: c56e30ab7375f56ea0a57944b1354b970e66d7b2, date: 2025-12-29T23:31:56Z)
```

**Exit Codes:**

- `0` - Success

## Flags Reference

### `--dir`

Explicitly specify the directory to pack. This flag takes precedence over positional arguments and is useful for avoiding conflicts when a directory has the same name as a subcommand (e.g., `pack` or `version`).

**Usage:**

```bash
fyaml --dir config/
fyaml --dir pack
```

**Behavior:**

- Takes precedence over positional arguments
- If both `--dir` and a positional argument are provided, `--dir` is used
- Useful for packing directories named `pack` or `version` without ambiguity

**Examples:**

```bash
# Pack directory named "pack" (avoids subcommand conflict)
fyaml --dir pack

# Pack directory named "version"
fyaml --dir version

# --dir takes precedence over positional argument
fyaml --dir config/ other/  # Uses config/, ignores other/
```

**See also:** [Usage Guide - Subcommand and Directory Name Conflicts](usage.md#subcommand-and-directory-name-conflicts) for more details.

### `--output`, `-o`

Specify the output file path. If not provided, output is written to stdout.

**Usage:**

```bash
fyaml -o output.yml
fyaml --output /path/to/output.yml
```

**Behavior:**

- File is written atomically (creates temp file, then renames)
- File permissions are set to `0644`
- If the file exists, it's overwritten
- Parent directories are not created automatically (must exist)

**Examples:**

```bash
# Write to current directory
fyaml -o config.yml

# Write to specific path
fyaml -o /tmp/output.yml

# Write to subdirectory (must exist)
fyaml -o build/config.yml

# Pack specific directory to file
fyaml config/ -o output.yml
```

### `-c, --check`

Compare the generated output with an existing file specified by `--output`. Uses `os.Exit(2)` if they differ (terminates immediately).

**Usage:**

```bash
fyaml -o output.yml --check
```

**Behavior:**

- Requires `--output` to be specified
- Reads the existing file (if it exists)
- Compares byte-by-byte with generated output
- Uses `os.Exit(2)` if different (terminates immediately), returns normally with code 0 if same
- Useful in CI/CD to verify configuration hasn't changed

**Exit Codes:**

- `0` - Output matches the file
- `2` - Output differs from the file (exits immediately via `os.Exit(2)`)
- `1` - Error (file read error, etc.)

**Examples:**

```bash
# Verify in CI
fyaml -o config.yml --check

# This will fail if config.yml is out of date
if ! fyaml -o config.yml --check; then
    echo "Configuration is out of date!"
    exit 1
fi
```

### `--format`, `-f`

Specify the output format. Valid values: `yaml` or `json`.

**Usage:**

```bash
fyaml config/ --format json
fyaml config/ -f yaml
```

**Default:** `yaml`

**Behavior:**

- `yaml` - Outputs YAML format (default)
- `json` - Outputs JSON format with 2-space indentation
- Empty output behavior differs by format (see below)

**Examples:**

```bash
# YAML output (default)
fyaml

# JSON output
fyaml --format json

# JSON to file
fyaml -f json -o config.json

# Pack specific directory as JSON
fyaml config/ --format json
```

**Empty Output:**

- YAML format: Returns empty output (0 bytes) when no files found
- JSON format: Returns `null` when no files found

**See also:** [Usage Guide - Output Format](usage.md#output-format) for more details on YAML and JSON output behavior.

### `--indent`

Specify the number of spaces used for indentation in both YAML and JSON output.

**Usage:**

```bash
fyaml config/ --indent 2
fyaml config/ --indent 4
```

**Default:** `2`

**Behavior:**

- Applies to both YAML and JSON output formats
- Must be a positive integer (error message: "must be positive")
- YAML output uses the specified indent spacing
- JSON output uses the specified indent spacing

**Examples:**

```bash
# Default 2-space indent (YAML)
fyaml

# 4-space indent (YAML)
fyaml --indent 4

# 2-space indent (JSON)
fyaml --format json --indent 2

# 4-space indent (JSON)
fyaml --format json --indent 4
```

**Note:** The default indent of 2 spaces is the widely accepted convention for YAML and JSON. You can override this if your project uses a different indentation style.

**See also:** [Usage Guide - Output Format](usage.md#output-format) for more details on YAML and JSON formatting.

### `--enable-includes`

Enable processing of file includes. This is an extension to the FYAML specification.

**Usage:**

```bash
fyaml config/ --enable-includes
```

**Default:** `false` (disabled)

**Include Mechanisms:**

When enabled, fyaml processes three include mechanisms:

| Syntax          | Purpose                                    | Example                            |
| --------------- | ------------------------------------------ | ---------------------------------- |
| `!include`      | Include parsed YAML structures             | `config: !include defaults.yml`    |
| `!include-text` | Include raw text content                   | `command: !include-text script.sh` |
| `<<include()>>` | Alias for `!include-text` (CircleCI style) | `command: <<include(script.sh)>>`  |

**Processing Order:**

1. `!include` tags are processed first (YAML structures merged)
2. `!include-text` tags are processed (text content replaced)
3. `<<include()>>` directives are processed (backward compatibility)

**Behavior:**

- **Pack root boundary**: All includes must resolve to paths within the pack root directory
- **Relative paths**: File paths are resolved relative to the file containing the include
- **Absolute paths**: Allowed but must be within the pack root
- **Nested includes**: Supported — included files can contain their own includes
- **JSON file support**:
  - `<<include()>>` works in JSON files (standard JSON)
  - `!include` and `!include-text` tags work in JSON files (non-standard JSON, but supported by fyaml)
  - YAML files can include JSON files using `!include`

**Examples:**

```bash
# Process includes
fyaml config/ --enable-includes

# Combine with other flags
fyaml config/ --enable-includes -o output.yml
fyaml config/ --enable-includes --format json
```

**Example: Including YAML Structures**

```yaml
# entities/item1.yml
entity:
  id: example1
  config: !include ../shared/defaults.yml
  attributes:
    name: sample name
```

**Example: Including Text Content**

```yaml
# entities/item1.yml
entity:
  id: example1
  attributes:
    name: sample name
  steps:
    - run:
        command: !include-text scripts/hello.sh
```

**Example: CircleCI Style**

```yaml
# Equivalent to !include-text
command: <<include(scripts/hello.sh)>>
```

**Error Cases:**

- `!include` on non-scalar — "must be used on a scalar value"
- `echo <<include(f)>>` — "entire string must be include statement"
- `<<include(a)>> <<include(b)>>` — "multiple include statements"
- Missing file — "could not open path/to/file for inclusion"
- Path escapes pack root — "include path escapes pack root"
- Invalid YAML/JSON in included file — "failed to parse YAML/JSON in path"

**Note:** Without this flag, include directives and tags are passed through unchanged. This preserves backward compatibility and keeps the default behavior spec-compliant.

**See also:** [Usage Guide - File Includes](usage.md#file-includes) for complete usage documentation and examples.

### `--convert-booleans`

Convert `on`/`off` and `yes`/`no` values to `true`/`false` booleans.

**Usage:**

```bash
fyaml --convert-booleans
```

**Default:** `false` (disabled)

**When to use:**

If your YAML files use `on`/`off` or `yes`/`no` for boolean values, they'll be treated as strings by default. Use this flag to convert them to actual boolean values.

**Behavior:**

- Unquoted values (`on`, `off`, `yes`, `no`, `y`, `n`, etc.) are converted to `true`/`false`
- Quoted strings (`"on"`, `'yes'`) are preserved as strings
- Non-boolean strings are unchanged

**Conversions:**

| Input               | Output  |
| ------------------- | ------- |
| `on`, `On`, `ON`    | `true`  |
| `off`, `Off`, `OFF` | `false` |
| `yes`, `Yes`, `YES` | `true`  |
| `no`, `No`, `NO`    | `false` |
| `y`, `Y`            | `true`  |
| `n`, `N`            | `false` |

**Examples:**

```bash
# Convert on/off to true/false
fyaml --convert-booleans

# Combine with other flags
fyaml --convert-booleans --enable-includes
fyaml --convert-booleans -o output.yml

# Pack specific directory with conversion
fyaml config/ --convert-booleans
```

**Example transformation:**

Input (`config/settings.yml`):

```yaml
enabled: on
debug: off
entity:
  id: example1
  attributes:
    name: "sample_item"
```

Output with `--convert-booleans`:

```yaml
debug: false
enabled: true
entity:
  attributes:
    name: sample_item
  id: example1
```

**Note:** fyaml always outputs YAML 1.2 format where only `true` and `false` are booleans.

**See also:** [Usage Guide - Converting on/off and yes/no to true/false](usage.md#converting-onoff-and-yesno-to-truefalse) for more details and examples.

### `-m, --mode`

Select the output mode that controls key ordering and comment preservation.

**Usage:**

```bash
fyaml --mode canonical    # Default: sorted keys, no comments
fyaml -m preserve         # Preserve order and comments
```

**Default:** `canonical`

**Valid Values:**

- `canonical` (default) - Keys are sorted alphabetically, comments are removed
- `preserve` - Maintains authored key order and preserves comments

**Behavior:**

**Canonical Mode (Default):**

- All map keys are sorted alphabetically
- Comments are removed from output
- Deterministic output
- Ideal for tools that don't care about key ordering or comments, and when sorted keys make diffs more readable

**Preserve Mode:**

- Keys maintain the order they appear in source files
- Comments are preserved in YAML output
- Deterministic output
- Ideal for maintaining documentation in comments and preserving the authored structure from source files

**Interaction with JSON Output:**

- **Key order**: In preserve mode, key order is maintained in JSON output (JSON preserves object key order)
- **Comments**: JSON doesn't support comments, so comments are lost regardless of mode
- Both modes produce deterministic JSON output

**Examples:**

```bash
# Canonical mode (default)
fyaml
fyaml --mode canonical

# Preserve mode
fyaml --mode preserve
fyaml -m preserve

# Preserve mode with JSON output (key order preserved, comments lost)
fyaml --mode preserve --format json -o output.json

# Combine with other flags
fyaml config/ --mode preserve --enable-includes -o output.yml
fyaml -m preserve --convert-booleans
```

**When to Use Each Mode:**

- **Use canonical mode when:**
  - You need sorted keys (makes diffs more readable)
  - The target tool doesn't care about key ordering or comments
  - You prefer a consistent, predictable key order

- **Use preserve mode when:**
  - You want to maintain documentation in comments
  - You want to preserve the authored key order from source files
  - The target tool or your workflow benefits from maintaining source structure

**Note:** Both modes are deterministic and suitable for version control and CI/CD. The difference is in key ordering (sorted vs. authored) and comment preservation.

**See also:** [Usage Guide - Output Modes](usage.md#output-modes) for detailed documentation and examples.

### `--merge`

Control how maps are merged when multiple files contribute to the same key.

**Usage:**

```bash
fyaml --merge shallow    # Default: later files completely replace earlier ones
fyaml --merge deep       # Recursively merge nested maps
```

**Default:** `shallow`

**Valid Values:**

- `shallow`: Later file's value completely replaces earlier one (default, backward compatible)
- `deep`: Nested maps are merged recursively, only replacing values at the leaf level

**When to use:**

- **Shallow merge** (default): Use when you want later files to completely override earlier ones. This is the default behavior and maintains backward compatibility.
- **Deep merge**: Use when you want to combine configuration from multiple files, preserving values from earlier files that aren't overridden.

**Behavior:**

- **Shallow merge**: If two files define the same key, the entire value from the later file replaces the earlier one, even for nested maps.
- **Deep merge**: If both sides are maps, they are merged recursively. Non-map values or type mismatches use shallow behavior (replace).

**Important Notes:**

- Arrays always use "replace" behavior (last wins) even in deep merge mode
- Deep merge only affects nested maps - scalar values and arrays are always replaced
- Applies to all merging scenarios: root-level files, `@` files, and `@` directories

**Examples:**

```bash
# Use shallow merge (default)
fyaml config/

# Use deep merge to combine nested configurations
fyaml config/ --merge deep

# Combine with other flags
fyaml config/ --merge deep --mode preserve
fyaml config/ --merge deep --format json
```

**Example transformation:**

Input files:

**`@base.yml`:**

```yaml
config:
  setting1: value1
  nested:
    a: 1
    b: 2
```

**`@override.yml`:**

```yaml
config:
  setting2: value2
  nested:
    c: 3
```

Output with `--merge shallow` (default):

```yaml
config:
  setting2: value2
  nested:
    c: 3
```

Output with `--merge deep`:

```yaml
config:
  setting1: value1
  setting2: value2
  nested:
    a: 1
    b: 2
    c: 3
```

**See also:** [Usage Guide - Merge Behavior](usage.md#merge-behavior) for more details and examples.

## Exit Codes

fyaml uses the following exit codes:

| Code | Meaning  | When It Occurs                                       |
| ---- | -------- | ---------------------------------------------------- |
| `0`  | Success  | Packing succeeded, or `--check` found no differences |
| `1`  | Error    | Pack error, IO error, invalid format, etc.           |
| `2`  | Mismatch | `--check` found differences between output and file  |

**Examples:**

```bash
# Success
fyaml  # exits 0

# Error - invalid directory
fyaml /nonexistent  # exits 1

# Mismatch - config out of date
fyaml -o config.yml --check  # exits 2 if different
```

## Default Behavior

### Default Directory

If no directory is specified, fyaml packs the current working directory:

```bash
# These are equivalent
fyaml
fyaml .
```

### Default Output

If no output file is specified, output goes to stdout:

```bash
# Output to stdout
fyaml

# Output to file
fyaml -o output.yml

# Pack specific directory to stdout
fyaml config/
```

### Default Format

Default output format is YAML:

```bash
# YAML output (default)
fyaml

# JSON output
fyaml --format json

# Pack specific directory as JSON
fyaml config/ --format json
```

## Error Messages

### Common Errors

**"pack error: failed to build filetree: ..."**

- Directory doesn't exist
- Permission denied
- Invalid path
- Error message includes the underlying error details

**"pack error: failed to marshal tree: ..."**

- Invalid YAML/JSON in files
- File contains non-map top-level value
- Error message includes the underlying error details (often includes file path and line/column information)

**"expected a map, got a `<type>` which is not supported at this time for \"<filepath>\"**

- File has top-level scalar or array instead of map
- `<type>` is the Go type (e.g., `string`, `[]interface{}`)
- `<filepath>` is the full path to the problematic file
- See [Usage Guide - File Content Requirements](usage.md#file-content-requirements) for details

**"invalid format: <format> (must be 'yaml' or 'json')"**

- Invalid `--format` value
- Use `yaml` or `json` only

**"--check requires --output to be specified"**

- `--check` flag used without `--output`
- Must specify output file when using `--check`

**"failed to read output file: ..."**

- Error reading file for `--check` comparison
- File may be unreadable (permission issue)

**"warning: no YAML/JSON files found in directory: <path>"**

- Directory contains no `.yml`, `.yaml`, or `.json` files
- Not an error, but output will be empty/null

## Performance Considerations

- Files are processed in memory
- Suitable for typical configuration files (KB to low MB)
- Very large files (hundreds of MB) may consume significant memory
- Processing is single-threaded

## Troubleshooting

### Subcommand and Directory Name Conflicts

If you have a directory named `pack` or `version`, use the `--dir` flag to explicitly specify the directory:

```bash
# Pack directory named "pack"
fyaml --dir pack

# Pack directory named "version"
fyaml --dir version
```

The `--dir` flag takes precedence over positional arguments and avoids any ambiguity. See [Usage Guide - Subcommand and Directory Name Conflicts](usage.md#subcommand-and-directory-name-conflicts) for more details.

## See Also

- [Usage Guide](usage.md) - Common usage patterns and limitations
- [Examples](examples.md) - Detailed examples
