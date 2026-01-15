# Quick Start

Pack a directory of YAML/JSON files into a single document:

```bash
# Pack current directory to stdout
fyaml

# Pack specific directory
fyaml config/

# Write to a file
fyaml -o output.yml
```

## Example

Given this directory structure:

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
  name: First Item
```

**`entities/item2.yml`:**

```yaml
entity:
  id: example2
  name: Second Item
```

Running `fyaml config/` produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      name: First Item
  item2:
    entity:
      id: example2
      name: Second Item
```

## Next Steps

- **[Basic Usage](basic-usage.md)** - Learn more about packing directories and output options
- **[Output Format](output-format.md)** - Output YAML or JSON, customize indentation
- **[Output Modes](output-modes.md)** - Choose between canonical (sorted) or preserve (authored order)
- **[Directory Structure](directory-structure.md)** - Understand how files map to YAML structure
- **[Common Patterns](common-patterns.md)** - Verify output, combine with other tools
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
