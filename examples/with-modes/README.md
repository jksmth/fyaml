# Output Modes Comparison Example

This example demonstrates the difference between canonical and preserve modes using the same input files.

## Structure

```
with-modes/
  entities/
    zebra.yml
    alpha.yml
    @group1/
      item-z.yml
    @group2/
      item-a.yml
```

## Usage

### Canonical Mode (Default)

```bash
fyaml examples/with-modes
# or explicitly
fyaml examples/with-modes --mode canonical
```

### Preserve Mode

```bash
fyaml examples/with-modes --mode preserve
```

## Files

**`entities/zebra.yml`:**

```yaml
# This entity has a name that sorts later alphabetically
entity:
  id: zebra
  attributes:
    name: Zebra
    tags:
      - animal
```

**`entities/alpha.yml`:**

```yaml
# This entity has a name that sorts earlier alphabetically
entity:
  id: alpha
  attributes:
    name: Alpha
    tags:
      - first
```

**`entities/@group1/item-z.yml`:**

```yaml
# Entity from @group1 - @ directories are processed in alphabetical order
entity:
  id: item-z
  attributes:
    name: Item Z
```

**`entities/@group2/item-a.yml`:**

```yaml
# Entity from @group2 - @ directories are processed in alphabetical order
entity:
  id: item-a
  attributes:
    name: Item A
```

## Expected Output

### Canonical Mode Output

```yaml
entities:
  alpha:
    entity:
      attributes:
        name: Alpha
        tags:
          - first
      id: alpha
  item-a:
    entity:
      attributes:
        name: Item A
      id: item-a
  item-z:
    entity:
      attributes:
        name: Item Z
      id: item-z
  zebra:
    entity:
      attributes:
        name: Zebra
        tags:
          - animal
      id: zebra
```

**Key points:**

- Keys are sorted alphabetically: `alpha`, `item-a`, `item-z`, `zebra`
- Comments are removed
- `@` directory order doesn't affect output (all keys sorted)

### Preserve Mode Output

```yaml
entities:
  item-z:
    # Entity from @group1 - @ directories processed first (alphabetically)
    entity:
      id: item-z
      attributes:
        name: Item Z
  item-a:
    # Entity from @group2 - @ directories processed in alphabetical order
    entity:
      id: item-a
      attributes:
        name: Item A
  alpha:
    # Regular file - processed after @ directories, in alphabetical order
    entity:
      id: alpha
      attributes:
        name: Alpha
        tags:
          - first
  zebra:
    # Regular file - processed after @ directories, in alphabetical order
    entity:
      id: zebra
      attributes:
        name: Zebra
        tags:
          - animal
```

**Key points:**

- Files and directories are processed in alphabetical order by canonical path (deterministic across operating systems)
- `@` directories are processed first (because `@` sorts before letters), then regular files
- Within each category, items are sorted alphabetically: `@group1` before `@group2`, `alpha` before `zebra`
- Comments are preserved
- Key order within each file is maintained (authored order)
- The order is deterministic and cross-platform, not dependent on filesystem order

## When to Use Each Mode

- **Canonical mode**: Use when you need sorted keys and don't need comments. Sorted keys make diffs more readable and predictable.
- **Preserve mode**: Use when you want to maintain documentation in comments or preserve the authored key order from your source files.

Both modes are deterministic (same input always produces same output) and suitable for version control and CI/CD. The difference is in key ordering (sorted vs. authored) and comment preservation.
