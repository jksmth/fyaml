# Output Modes

fyaml supports two output modes that control how keys are ordered and whether comments are preserved:

## Canonical Mode (Default)

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

## Preserve Mode

Preserve mode maintains the authored structure where possible:

- **Authored key order**: Keys appear in the order they were written in source files
- **Comments preserved**: All YAML comments are maintained in the output
- **Deterministic output**: Same input always produces same output

**What Preserve Mode Guarantees:**

- Structure and comments are preserved where possible
- Deterministic output (same input = same output)
- Key ordering from source files is maintained
- Comments are preserved in YAML output

**What Preserve Mode Does NOT Guarantee:**

- Exact byte-for-byte preservation of YAML syntax
- Some YAML syntax may be normalized (e.g., merge keys may be emitted with explicit `!!merge` tags, even if not present in source)

This mode is ideal for:

- Maintaining documentation in comments
- Preserving the authored key order from source files
- When you need deterministic output but can accept some YAML syntax normalization

```bash
fyaml --mode preserve     # Preserve order and comments
fyaml -m preserve         # Shorthand
```

## Mode Comparison Example

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

## JSON Output and Modes

When using JSON output format:

- **Key order**: JSON output always sorts keys alphabetically regardless of mode
- **Comments**: JSON doesn't support comments, so comments are lost regardless of mode
- Both modes produce deterministic JSON output

```bash
fyaml --format json --mode preserve    # Keys sorted alphabetically, comments lost
```
