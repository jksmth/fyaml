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

### `fyaml pack [DIR]`

Compile a directory of YAML/JSON files into a single document.

**Synopsis:**

```bash
fyaml [global flags] pack [DIR] [flags]
```

**Arguments:**

- `DIR` - Directory to pack (default: current directory)

**Flags:**

- `-o, --output string` - Write output to file (default: stdout)
- `--check` - Compare generated output to `--output`, exit non-zero if different
- `-f, --format string` - Output format: `yaml` or `json` (default: `yaml`)
- `--indent int` - Number of spaces for indentation (default: `2`)
- `--enable-includes` - Process `<<include(file)>>` directives (extension)
- `--convert-booleans` - Convert unquoted YAML 1.1 booleans to `true`/`false`

**Examples:**

```bash
# Pack current directory to stdout
fyaml pack

# Pack specific directory
fyaml pack config/

# Write to file
fyaml pack config/ -o output.yml

# Output as JSON
fyaml pack config/ --format json -o output.json

# Verify output matches file
fyaml pack config/ -o output.yml --check

# With verbose output
fyaml -v pack config/
```

**Exit Codes:**

- `0` - Success
- `1` - Pack or IO error
- `2` - `--check` mismatch (output differs from file)

### `fyaml version`

Print version information.

**Synopsis:**

```bash
fyaml version
```

**Output:**

```
1.0.3 (commit: c56e30ab7375f56ea0a57944b1354b970e66d7b2, date: 2025-12-29T23:31:56Z)
```

**Exit Codes:**

- `0` - Success

## Flags Reference

### `--output`, `-o`

Specify the output file path. If not provided, output is written to stdout.

**Usage:**

```bash
fyaml pack config/ -o output.yml
fyaml pack config/ --output /path/to/output.yml
```

**Behavior:**

- File is written atomically (creates temp file, then renames)
- File permissions are set to `0644`
- If the file exists, it's overwritten
- Parent directories are not created automatically (must exist)

**Examples:**

```bash
# Write to current directory
fyaml pack config/ -o config.yml

# Write to specific path
fyaml pack config/ -o /tmp/output.yml

# Write to subdirectory (must exist)
fyaml pack config/ -o build/config.yml
```

### `--check`

Compare the generated output with an existing file specified by `--output`. Uses `os.Exit(2)` if they differ (terminates immediately).

**Usage:**

```bash
fyaml pack config/ -o output.yml --check
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
fyaml pack config/ -o config.yml --check

# This will fail if config.yml is out of date
if ! fyaml pack config/ -o config.yml --check; then
    echo "Configuration is out of date!"
    exit 1
fi
```

### `--format`, `-f`

Specify the output format. Valid values: `yaml` or `json`.

**Usage:**

```bash
fyaml pack config/ --format json
fyaml pack config/ -f yaml
```

**Default:** `yaml`

**Behavior:**

- `yaml` - Outputs YAML format (default)
- `json` - Outputs JSON format with 2-space indentation
- Empty output behavior differs by format (see below)

**Examples:**

```bash
# YAML output (default)
fyaml pack config/

# JSON output
fyaml pack config/ --format json

# JSON to file
fyaml pack config/ -f json -o config.json
```

**Empty Output:**

- YAML format: Returns empty output (0 bytes) when no files found
- JSON format: Returns `null` when no files found

**See also:** [Usage Guide - Output Format](usage.md#output-format) for more details on YAML and JSON output behavior.

### `--indent`

Specify the number of spaces used for indentation in both YAML and JSON output.

**Usage:**

```bash
fyaml pack config/ --indent 2
fyaml pack config/ --indent 4
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
fyaml pack config/

# 4-space indent (YAML)
fyaml pack config/ --indent 4

# 2-space indent (JSON)
fyaml pack config/ --format json --indent 2

# 4-space indent (JSON)
fyaml pack config/ --format json --indent 4
```

**Note:** The default indent of 2 spaces is the widely accepted convention for YAML and JSON. You can override this if your project uses a different indentation style.

**See also:** [Usage Guide - Output Format](usage.md#output-format) for more details on YAML and JSON formatting.

### `--enable-includes`

Enable processing of file includes. This is an extension to the FYAML specification.

**Usage:**

```bash
fyaml pack config/ --enable-includes
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
fyaml pack config/ --enable-includes

# Combine with other flags
fyaml pack config/ --enable-includes -o output.yml
fyaml pack config/ --enable-includes --format json
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
fyaml pack config/ --convert-booleans
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
fyaml pack config/ --convert-booleans

# Combine with other flags
fyaml pack config/ --convert-booleans --enable-includes
fyaml pack config/ --convert-booleans -o output.yml
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
fyaml pack config/  # exits 0

# Error - invalid directory
fyaml pack /nonexistent  # exits 1

# Mismatch - config out of date
fyaml pack config/ -o config.yml --check  # exits 2 if different
```

## Default Behavior

### Default Directory

If no directory is specified, fyaml packs the current working directory:

```bash
# These are equivalent
fyaml pack
fyaml pack .
```

### Default Output

If no output file is specified, output goes to stdout:

```bash
# Output to stdout
fyaml pack config/

# Output to file
fyaml pack config/ -o output.yml
```

### Default Format

Default output format is YAML:

```bash
# YAML output (default)
fyaml pack config/

# JSON output
fyaml pack config/ --format json
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

## See Also

- [Usage Guide](usage.md) - Common usage patterns and limitations
- [Examples](examples.md) - Detailed examples
