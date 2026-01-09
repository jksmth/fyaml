# Boolean Conversion Example

This example demonstrates the `--convert-booleans` flag, which converts YAML 1.1 boolean values (`on`/`off`, `yes`/`no`) to YAML 1.2 boolean values (`true`/`false`).

## Structure

```
with-booleans/
  config/
    settings.yml
```

## Usage

### Without Boolean Conversion (Default)

```bash
fyaml examples/with-booleans
```

This treats `on`/`off` and `yes`/`no` as strings:

```yaml
config:
  settings:
    debug: "off"
    enabled: "on"
    feature1: "yes"
    feature2: "no"
```

### With Boolean Conversion

```bash
fyaml examples/with-booleans --convert-booleans
```

This converts YAML 1.1 booleans to YAML 1.2 booleans:

```yaml
config:
  settings:
    debug: false
    enabled: true
    feature1: true
    feature2: false
```

## Files

**`config/settings.yml`:**

```yaml
debug: off
enabled: on
feature1: yes
feature2: no
```

## Conversions

The `--convert-booleans` flag converts the following values:

| Input               | Output  |
| ------------------- | ------- |
| `on`, `On`, `ON`    | `true`  |
| `off`, `Off`, `OFF` | `false` |
| `yes`, `Yes`, `YES` | `true`  |
| `no`, `No`, `NO`    | `false` |
| `y`, `Y`            | `true`  |
| `n`, `N`            | `false` |

**Note:** Quoted strings (e.g., `"on"`, `'yes'`) are preserved as strings and not converted.

## Key Points

- **Unquoted values only**: Only unquoted values are converted. Quoted strings remain strings.
- **Case insensitive**: Conversion is case-insensitive (`on`, `On`, `ON` all convert to `true`)
- **YAML 1.2 compliance**: fyaml always outputs YAML 1.2 format where only `true` and `false` are booleans
- **Useful for legacy configs**: Helps migrate YAML 1.1 configurations to YAML 1.2

## Example with Quoted Values

If your file contains:

```yaml
debug: off # Converts to false
enabled: "on" # Stays as string "on"
feature1: "yes" # Stays as string "yes"
feature2: no # Converts to false
```

With `--convert-booleans`, the output is:

```yaml
config:
  settings:
    debug: false
    enabled: "on"
    feature1: "yes"
    feature2: false
```

## Notes

- This flag is useful when working with legacy YAML files that use YAML 1.1 boolean syntax
- Quoted values are intentionally preserved as strings to allow explicit string values when needed
- The conversion only applies to scalar values, not to keys or other YAML structures
