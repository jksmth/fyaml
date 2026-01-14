# Directory Structure Rules

**Note:** Files and directories are processed in alphabetical order (deterministic across operating systems). This means `@` directories are processed before regular files (because `@` sorts before letters), and within each category, items are sorted alphabetically.

## Merge Behavior

When multiple files contribute to the same key (e.g., root-level files, `@` files, or `@` directories), fyaml supports two merge strategies:

- **Shallow merge** (default): The later file's value completely replaces the earlier one. Nested maps are not merged recursively.
- **Deep merge**: Nested maps are merged recursively, only replacing values at the leaf level. Non-map values are replaced (shallow behavior).

Use the `--merge` flag to control merge behavior:

```bash
fyaml --merge shallow    # Default: later files completely replace earlier ones
fyaml --merge deep       # Recursively merge nested maps
```

**Shallow Merge Example (default):**

If you have two files that both define `config`:

**`@shared1.yml`:**

```yaml
config:
  database:
    host: localhost
    port: 5432
```

**`@shared2.yml`:**

```yaml
config:
  database:
    port: 3306
```

With `--merge shallow` (default), the result is:

```yaml
config:
  database:
    port: 3306
```

The entire nested map from `@shared2.yml` replaces the one from `@shared1.yml`. It does not merge to `config: {database: {host: localhost, port: 3306}}`.

**Deep Merge Example:**

With the same files and `--merge deep`, the result is:

```yaml
config:
  database:
    host: localhost
    port: 3306
```

Nested maps are merged recursively.

**Important Notes:**

- Arrays always use "replace" behavior (last wins) even in deep merge mode
- If both sides are maps, they are merged recursively in deep mode. Otherwise, the source replaces the target (shallow behavior)
- This behavior applies to all merging scenarios: root-level files, `@` files, and `@` directories

## Basic Structure

Directory names become map keys in the output:

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
  attributes:
    name: sample name
    tags:
      - tag1
```

**`entities/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags:
          - tag1
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags: []
```

## Root-Level Files

Files at the root level merge directly into the output (their filename is not used as a key):

```
config/
  shared.yml       # Merges directly into root
  entities/
    item1.yml
```

**`shared.yml`:**

```yaml
version: 1.0.0
environment: production
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
environment: production
version: 1.0.0
```

**Note:** The `shared.yml` contents merge directly into the root. The filename `shared` is not used as a key. **Merging is shallow**: if multiple root-level files or `@` files define the same key, the later file's value completely replaces the earlier one (nested maps are not merged recursively).

## @ Files

Files starting with `@` merge their contents into the parent directory map:

```
config/
  entities/
    @shared.yml    # Merges into entities map
    item1.yml
```

**`entities/@shared.yml`:**

```yaml
environment: production
region: us-east-1
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
  environment: production
  region: us-east-1
```

The `@` prefix is removed, and the file's contents are merged directly into the parent map.

**Note:** While the specification allows multiple `@` files in the same directory, it's recommended to use only one per directory. If multiple `@` files exist, they all merge into the parent map, and key collisions are resolved by the last file processed. **Merging is shallow**: if two files define the same key with nested maps, the entire nested map from the later file replaces the earlier one (nested values are not merged recursively). Files are processed in alphabetical order (see [Directory Structure Rules](#directory-structure-rules) above).

## @ Directories

Directories starting with `@` merge their contents into the parent directory map, similar to `@` files. This allows directories to be used for organization without creating additional nesting levels in the output.

**Example:**

Directory structure:

```
config/
  entities/
    item1.yml
    @group1/              # Merges into entities map
      item2.yml
      item3.yml
    @group2/              # Merges into entities map
      item4.yml
```

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

**`entities/@group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another name
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

Produces:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags:
          - tag1
  item3:
    entity:
      id: example3
      attributes:
        name: third item
        tags:
          - tag2
  item4:
    entity:
      id: example4
      attributes:
        name: fourth item
        tags: []
```

**Key Points:**

- The `@group1` and `@group2` directory names do not appear as keys in the output
- All files from `@` directories merge directly into the parent map (`entities` in this example)
- This is useful for organizing large numbers of files without creating deep nesting
- Empty `@` directories are ignored (no keys created)

**Use Cases:**

- Organizing large numbers of files without creating deep nesting
- Grouping related files logically while keeping flat output structure
- Maintaining organization in source while producing simpler output

**Edge Cases:**

- Empty `@` directories are ignored (no keys created)
- If both `@group1/` directory and `@group1.yml` file exist, both merge into parent. Files are processed in alphabetical order (see [Directory Structure Rules](#directory-structure-rules) above).
- Nested `@` directories are supported: `@group1/@shared/` merges into parent of `@group1/`

**Note:** This is an extension to the FYAML specification. See [EXTENSIONS.md](https://github.com/jksmth/fyaml/blob/main/EXTENSIONS.md) for information about extensions to the specification.

## Nested Directories

Directories can be nested to any depth:

```
config/
  category1/
    group1/
      item1.yml
      item2.yml
    group2/
      item3.yml
```

**`category1/group1/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: first item
    tags: []
```

**`category1/group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: second item
    tags:
      - tag1
```

**`category1/group2/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags: []
```

Produces:

```yaml
category1:
  group1:
    item1:
      entity:
        id: example1
        attributes:
          name: first item
          tags: []
    item2:
      entity:
        id: example2
        attributes:
          name: second item
          tags:
            - tag1
  group2:
    item3:
      entity:
        id: example3
        attributes:
          name: third item
          tags: []
```

## Multi-Document YAML Files

YAML supports multiple documents in a single file, separated by `---`. **fyaml merges all documents in a file** using the same semantics as file merging (later values overwrite earlier values, respects `--merge`).

**Example:**

```yaml
# config.yml
timeout: 30
retries: 3
---
timeout: 60 # Overwrites previous value
debug: true # New key added
```

When processed, this produces:

```yaml
timeout: 60 # From second document (overwrites first)
retries: 3 # From first document (preserved)
debug: true # From second document (new key)
```

**Key behaviors:**

- All documents must be maps (consistent with file requirement)
- Documents merge in order (first to last, "later wins")
- Respects `--merge` (shallow/deep)
- Works identically for regular files, `@` files, and root files

**Note:** While multi-document files are supported, fyaml's filesystem-based approach (organizing resources as separate files) is generally recommended for better organization and maintainability.
