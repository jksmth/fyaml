# Comment Preservation Example

This example demonstrates the `--mode preserve` feature, which preserves comments and authored key order from source YAML files in the output.

## Structure

```
with-comments/
  database.yml        # Root-level file (merges directly)
  logging.yml         # Root-level file (merges directly)
  shared.yml          # Root-level file (contains shared key)
  overrides.yml       # Root-level file (overrides shared key)
  entities/
    @shared.yml       # @ file (merges into entities map)
    item1.yml         # Regular file
    item2.yml         # Regular file
```

## Usage

### Without Comment Preservation (Default)

```bash
fyaml examples/with-comments
```

This produces standard output without comments:

```yaml
database:
  host: localhost
  port: 5432
entities:
  item1:
    entity:
      id: example1
      name: First Item
  item2:
    entity:
      id: example2
      name: Second Item
  shared_config:
    enabled: true
    timeout: 30
logging:
  level: info
shared:
  environment: production
  region: us-east-1
  timeout: 60
```

Note: The `timeout` value is `60` because `overrides.yml` comes after `shared.yml` alphabetically, so it overwrites the value. In preserve mode, the same alphabetical ordering applies.

### With Comment Preservation

```bash
fyaml examples/with-comments --mode preserve
```

This preserves all comments from source files in a clean, readable format:

```yaml
# Database configuration
# This is important for production
database:
  host: localhost # Database hostname
  port: 5432 # Database port
entities:
  entities:
    # Shared configuration for all entities
    shared_config:
      enabled: true # Feature flag
      timeout: 30 # Timeout in seconds
  item1:
    # First entity item
    entity:
      id: example1 # Unique identifier
      name: First Item # Display name
  # Footer comment for item1
  item2:
    # Second entity item
    entity:
      id: example2 # Unique identifier
      name: Second Item # Display name
# Footer comment for item2
logging:
  level: info # Logging level
# Override settings
# These override shared settings
# Global shared settings
# These apply to all components
shared:
  environment: production # Environment name
  region: us-east-1 # AWS region
  timeout: 30 # Connection timeout
  # Overridden timeout value
```

**Key observations:**

1. **Comments are preserved** - All header, inline, and footer comments from source files are included
2. **Multiple inline comments converted to block** - When the same key has inline comments from multiple files, they're converted to block form (head comments) with only the winning source's comment kept inline
3. **Deterministic ordering** - In canonical mode (default), keys are sorted alphabetically. In preserve mode, keys maintain authored order from source files. Comments are merged deterministically.
4. **All comment types** - Header comments (above keys), inline comments (on same line), and footer comments (below keys) are all preserved
5. **Value overwriting** - The `timeout` value is `30` (from `shared.yml` which comes after `overrides.yml` alphabetically), but comments from both files are preserved

## Comment Types Demonstrated

This example shows all three types of YAML comments:

1. **Header comments** - Comments above keys/values (e.g., `# Database configuration`)
2. **Inline comments** - Comments on the same line as values (e.g., `host: localhost # Database hostname`)
3. **Footer comments** - Comments below keys/values (e.g., `# Footer comment for item1`)

## Merging Behavior

### Shallow Merge

fyaml uses **shallow merge** semantics: when multiple files contribute to the same key, the later file's value completely replaces the earlier one. Nested maps are not merged recursively.

**Example:**

If `shared.yml` contains:

```yaml
shared:
  database:
    host: localhost
    port: 5432
```

And `overrides.yml` contains:

```yaml
shared:
  database:
    port: 3306
```

The result is `shared: {database: {port: 3306}}` (the entire nested map from `overrides.yml` replaces the one from `shared.yml`), **not** `shared: {database: {host: localhost, port: 3306}}`.

### Comment Merging

When multiple files contribute to the same output key, comments are merged intelligently:

- The `shared` key appears in both `shared.yml` and `overrides.yml`
- Comments from both files are preserved, but **multiple inline comments are converted to block form** for readability
- The winning source's (last file processed) inline comment is kept inline
- Other sources' inline comments are moved to head comments (above the key)
- The `timeout` value from `shared.yml` overwrites the value from `overrides.yml` (per FYAML merge rules - files processed later overwrite earlier ones), but **comments from both sources are preserved**

## Notes

- Comments are only preserved when using the `--mode preserve` flag
- Multiple inline comments are automatically converted to block form for readability
- Comments are merged deterministically based on file processing order (alphabetical)
- JSON output does not support comments (key order is preserved in preserve mode, but comments are lost)
- Orphan comments (comments in files with no actual data) are dropped by default
