# fyaml Extensions

This document describes extensions to the [FYAML specification](SPECIFICATION.md) implemented by fyaml. These features enhance functionality while maintaining spec compliance.

**Important:** Extensions are opt-in and do not affect spec-compliant behavior. For spec-compliant usage, use only `.yml` and `.yaml` files with YAML output (default).

## Extension List

1. [JSON Support](#json-support)
2. [File Includes](#file-includes)
3. [Boolean Conversion](#boolean-conversion)
4. [@ Directory Support](#-directory-support)

---

## JSON Support

**Status:** Extension
**Opt-in:** Via file extension (`.json` files) and `--format json` flag

fyaml accepts `.json` files in addition to `.yml` and `.yaml` files. JSON files are processed the same way as YAML files.

**Input:** Use `.json` files alongside YAML files
**Output:** Use the `--format` flag to output JSON: `fyaml pack config/ --format json`

See [README.md](README.md#json-support) for complete documentation.

---

## File Includes

**Status:** Extension
**Opt-in:** Via `--enable-includes` flag

Process include directives to include content from other files. Supports three mechanisms:

- `!include` — Include parsed YAML structures
- `!include-text` — Include raw text content
- `<<include()>>` — Backward-compatible alias for `!include-text`

**Usage:** `fyaml pack config/ --enable-includes`

See [README.md](README.md#file-includes) for complete documentation.

---

## Boolean Conversion

**Status:** Extension
**Opt-in:** Via `--convert-booleans` flag

Convert YAML 1.1-style boolean values (`on`, `off`, `yes`, `no`) to YAML 1.2 booleans (`true`, `false`).

**Usage:** `fyaml pack config/ --convert-booleans`

**Note:** Quoted values like `"on"` remain as strings.

See [README.md](README.md#boolean-conversion) for complete documentation.

---

## @ Directory Support

**Status:** Extension
**Opt-in:** Via directory naming convention (directories starting with `@`)

Directories starting with `@` merge their contents into the parent directory map, similar to how `@` files work. This allows directories to be used for organization without creating additional nesting levels in the output.

**Behavior:**
- Directory names starting with `@` are collapsed into the parent map
- The directory name after `@` is ignored (organizational only)
- Contents of `@` directories merge directly into the parent map

**Example:**

Directory structure:
```
components/
  database.yml
  @infrastructure/
    cache.yml
    queue.yml
  @monitoring/
    metrics.yml
```

Output:
```yaml
components:
  cache: <contents>
  database: <contents>
  metrics: <contents>
  queue: <contents>
```

**Use Cases:**
- Organizing large numbers of files without creating deep nesting
- Grouping related files logically while keeping flat output structure
- Maintaining organization in source while producing simpler output

**Edge Cases:**
- Empty `@` directories are ignored (no keys created)
- If both `@services/` directory and `@services.yml` file exist, both merge into parent (order not guaranteed)
- Nested `@` directories are supported: `@services/@common/` merges into parent of `@services/`

**Note:** This extension mirrors the behavior of `@` files (spec rule 6) but applies to directories. It is not part of the FYAML specification.

See [docs/usage.md](docs/usage.md) for complete usage examples.

---

## Extension Philosophy

Extensions in fyaml follow these principles:

1. **Opt-in:** Extensions must be explicitly enabled or used (via flags, file extensions, or naming conventions)
2. **Non-breaking:** Extensions do not affect spec-compliant behavior when not used
3. **Well-documented:** All extensions are clearly marked and documented
4. **Enhancement, not replacement:** Extensions enhance functionality without replacing spec behavior

When in doubt, default behavior remains spec-compliant.

