# File Naming

## Supported Extensions

fyaml processes files with these extensions:

- `.yml`
- `.yaml`
- `.json`

You can mix these file types in the same directory structure. JSON files must be valid JSON (not JSON5 or other variants) and must have a top-level object (map), just like YAML files.

## Ignored Files

fyaml automatically ignores:

- Files and directories starting with `.` (dot files)
- Files without supported extensions

## File Name to Key Mapping

The filename (without extension) becomes the key in the output:

- `item1.yml` → key: `item1`
- `entity-item.yaml` → key: `entity-item`
- `config.json` → key: `config`

**Exceptions:**

- Root-level files: filename is ignored, contents merge directly
- `@` files: `@` prefix is removed, contents merge into parent

## Special Characters

File and directory names can contain hyphens, underscores, numbers, and mixed case. Special characters are preserved in the output key:

- `entity-item.yml` → key: `entity-item`
- `entity_item.yml` → key: `entity_item`
- `ItemV2.yml` → key: `ItemV2`

## File Name Collisions

If you have files with the same name but different extensions in the same directory (e.g., `item1.yml`, `item1.yaml`, `item1.json`), they all produce the same key. The last one processed will overwrite previous ones. Files are processed in alphabetical order (see [Directory Structure Rules](directory-structure.md) above). **Use a consistent extension throughout your project to avoid collisions.**
