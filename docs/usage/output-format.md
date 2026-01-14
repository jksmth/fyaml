# Output Format

## YAML (Default)

Default output is YAML:

```bash
fyaml
```

YAML output uses standard YAML formatting with 2-space indentation by default. You can customize the indent using the `--indent` flag:

```bash
# Use 4-space indent
fyaml --indent 4
```

## JSON

Output JSON using the `--format` flag:

```bash
fyaml --format json
```

JSON output is formatted with 2-space indentation by default. You can customize the indent using the `--indent` flag:

```bash
# Use 4-space indent for JSON
fyaml --format json --indent 4
```

## Non-String Map Keys

YAML allows map keys to be non-string types (numbers, booleans, etc.). **fyaml preserves these key types for YAML output** but normalizes them to strings for JSON output (JSON format requirement).

- **YAML output** (both modes): Non-string keys are preserved. For example, `123: value` stays as `123: value`.
- **JSON output** (both modes): Non-string keys are converted to strings. For example, `123: value` becomes `"123": value`.

**Example:**

```yaml
# Input file
123: "numeric key"
true: "boolean key"
```

**YAML output (canonical or preserve mode):**

```yaml
123: numeric key
true: boolean key
```

**JSON output (canonical or preserve mode):**

```json
{
  "123": "numeric key",
  "true": "boolean key"
}
```

**Note:** When non-string keys are present, canonical mode sorts them in type order: booleans first, then numbers (sorted numerically), then strings (sorted alphabetically).

## Empty Output

When no files are found:

- **YAML format**: Returns empty output (0 bytes) - aligns with `yq` behavior
- **JSON format**: Returns `null` - aligns with `jq` behavior

A warning is printed to stderr in both cases.

## Deterministic Output

Both modes are deterministic, making output suitable for version control and comparison. This means you can safely commit the generated output to version control and use the `--check` flag (see [Verify Output Matches Existing File](common-patterns.md#verify-output-matches-existing-file)) in CI to ensure the source files and generated output stay in sync.

## Converting YAML 1.1 Booleans

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
