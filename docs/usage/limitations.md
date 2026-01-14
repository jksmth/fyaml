# Limitations

## File Content Requirements

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

## YAML Anchors and Aliases

YAML anchors (`&anchor`) and aliases (`*alias`) are resolved **within each individual file** during parsing by default. For cross-file anchor support, see the [YAML Anchors and Aliases](anchors-aliases.md) section.

**Preserve Mode Constraint:** When using `--mode preserve` with `--enable-anchors`, anchor definitions must appear before their aliases in the output. If aliases appear first, fyaml will return an error. See [Preserve Mode Validation](anchors-aliases.md#preserve-mode-validation) for details on how to organize files to avoid this issue.

## Large Files

fyaml processes files in memory. For very large files (hundreds of MB), this could consume significant memory. However, for typical configuration files (KB to low MB range), performance is excellent. Keep individual files focused and reasonably sized.
