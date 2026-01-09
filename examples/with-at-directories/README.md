# @ Directories Example

This example demonstrates how `@` directories merge their contents into the parent directory map, similar to `@` files.

## Structure

```
with-at-directories/
  entities/
    item1.yml
    @group1/              # Merges into entities map
      item2.yml
      item3.yml
    @group2/              # Merges into entities map
      item4.yml
```

## Usage

Pack this directory:

```bash
fyaml examples/with-at-directories
```

Or from this directory:

```bash
cd examples/with-at-directories
fyaml .
```

**Note:** `fyaml pack` is an alias and works identically for backward compatibility.

## Files

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: first item
    tags: []
```

**`entities/@group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: second item
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

## Expected Output

The `@group1` and `@group2` directory names do not appear in the output. All files from `@` directories merge directly into the `entities` map:

```yaml
entities:
  item1:
    entity:
      attributes:
        name: first item
        tags: []
      id: example1
  item2:
    entity:
      attributes:
        name: second item
        tags:
          - tag1
      id: example2
  item3:
    entity:
      attributes:
        name: third item
        tags:
          - tag2
      id: example3
  item4:
    entity:
      attributes:
        name: fourth item
        tags: []
      id: example4
```

**Key points:**

- The `@group1` and `@group2` directory names do not appear in the output
- All files from `@` directories merge directly into the parent map (`entities`)
- This is useful for organizing large numbers of files without creating deep nesting
- Files are organized in source but produce a flat structure in output
- In canonical mode (default), keys are sorted alphabetically
- In preserve mode, `@` directories are processed in alphabetical order by path (deterministic across operating systems)

## Notes

- `@` directories work similarly to `@` files but allow organizing multiple files
- Empty `@` directories are ignored (no keys created)
- Nested `@` directories are supported: `@group1/@shared/` merges into parent of `@group1/`
- This is an extension to the FYAML specification
