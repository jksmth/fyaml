# Command Reference

Complete reference for all fyaml commands, flags, and options.

## Commands

### `fyaml pack [DIR]`

Compile a directory of YAML/JSON files into a single document.

**Synopsis:**
```bash
fyaml pack [DIR] [flags]
```

**Arguments:**
- `DIR` - Directory to pack (default: current directory)

**Flags:**
- `-o, --output string` - Write output to file (default: stdout)
- `--check` - Compare generated output to `--output`, exit non-zero if different
- `-f, --format string` - Output format: `yaml` or `json` (default: `yaml`)

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

Compare the generated output with an existing file specified by `--output`. Exits with code 2 if they differ.

**Usage:**
```bash
fyaml pack config/ -o output.yml --check
```

**Behavior:**
- Requires `--output` to be specified
- Reads the existing file (if it exists)
- Compares byte-by-byte with generated output
- Exits with code 2 if different, 0 if same
- Useful in CI/CD to verify configuration hasn't changed

**Exit Codes:**
- `0` - Output matches the file
- `2` - Output differs from the file
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

### `--enable-includes`

Enable processing of `<<include(file)>>` directives. This is an extension to the FYAML specification.

**Usage:**
```bash
fyaml pack config/ --enable-includes
```

**Default:** `false` (disabled)

**Behavior:**
- When enabled, `<<include(path)>>` directives in YAML values are replaced with file contents
- **The directory passed to `pack` (the pack root) defines the include boundary; includes outside this directory are rejected.**
- File paths are resolved relative to the YAML file containing the directive
- Both absolute and relative paths are allowed, but must resolve to a path within the pack root directory
- The include directive must be the **entire value** (not embedded in other text)
- Only one include per value is allowed

**Examples:**
```bash
# Process includes
fyaml pack config/ --enable-includes

# Combine with other flags
fyaml pack config/ --enable-includes -o output.yml
fyaml pack config/ --enable-includes --format json
```

**Example YAML with include:**
```yaml
steps:
  - run:
      command: <<include(scripts/deploy.sh)>>
```

When packed with `--enable-includes`, the `<<include(...)>>` is replaced with the contents of `scripts/deploy.sh` (relative to the YAML file).

**Error Cases:**
- `echo <<include(f)>>` — "entire string must be include statement"
- `<<include(a)>> <<include(b)>>` — "multiple include statements"
- Missing file — "could not open path/to/file for inclusion"
- Path escapes pack root — "include path escapes pack root"
- Path escapes pack root — "include path escapes pack root"

**Note:** Without this flag, include directives are passed through unchanged. This preserves backward compatibility and keeps the default behavior spec-compliant.

## Exit Codes

fyaml uses the following exit codes:

| Code | Meaning | When It Occurs |
|------|---------|----------------|
| `0` | Success | Packing succeeded, or `--check` found no differences |
| `1` | Error | Pack error, IO error, invalid format, etc. |
| `2` | Mismatch | `--check` found differences between output and file |

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

**"pack error: failed to marshal tree: ..."**
- Invalid YAML/JSON in files
- File contains non-map top-level value

**"expected a map, got a <type> which is not supported at this time for \"<filepath>\"**
- File has top-level scalar or array instead of map
- See [Usage Guide](usage.md) for details

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

