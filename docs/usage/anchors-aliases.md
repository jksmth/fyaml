# YAML Anchors and Aliases

With the `--enable-anchors` flag, anchors defined in one file can be referenced as aliases in another file. This enables patterns like:

```yaml
# shared/templates/common.yml
template: &common_template
  timeout: 30
  retries: 3
  enabled: true

# entities/item1.yml
entity:
  id: example1
  config: *common_template # Cross-file reference
  attributes:
    name: sample name
```

## Preserve Mode vs Canonical Mode

The behavior differs between modes and output formats:

- **Canonical Mode** (`--mode canonical`): Aliases are **expanded** into their full content:
  ```yaml
  shared:
    templates:
      common:
        template:
          timeout: 30
          retries: 3
          enabled: true
  entities:
    item1:
      entity:
        id: example1
        config:
          enabled: true
          retries: 3
          timeout: 30
        attributes:
          name: sample name
  ```

- **Preserve Mode** (`--mode preserve`): In YAML output, both anchor definitions and alias references are preserved as written. This maintains the original YAML structure:
  ```yaml
  shared:
    templates:
      common:
        template: &common_template
          timeout: 30
          retries: 3
          enabled: true
  entities:
    item1:
      entity:
        id: example1
        config: *common_template # Alias reference preserved
        attributes:
          name: sample name
  ```

## JSON Output Behavior

When using `--format json`, aliases are **always expanded** regardless of mode, because JSON doesn't support YAML anchors and aliases natively. Additionally, JSON output always sorts keys alphabetically, so key order is not preserved even in preserve mode:

```bash
# Both produce the same JSON output (aliases expanded)
fyaml --enable-anchors --mode preserve --format json
fyaml --enable-anchors --mode canonical --format json
```

**JSON output example:**

```json
{
  "entities": {
    "item1": {
      "entity": {
        "id": "example1",
        "config": {
          "enabled": true,
          "retries": 3,
          "timeout": 30
        },
        "attributes": {
          "name": "sample name"
        }
      }
    }
  },
  "shared": {
    "templates": {
      "common": {
        "template": {
          "timeout": 30,
          "retries": 3,
          "enabled": true
        }
      }
    }
  }
}
```

Note that:

- Anchor markers (`&anchor_name`) are removed
- Alias references (`*alias_name`) are expanded to their values
- Keys are sorted alphabetically (same in both preserve and canonical modes)

## Merge Keys (`<<:`)

Cross-file anchors work with merge keys in both modes. The merged content is applied correctly:

```yaml
# shared/defaults.yml
defaults: &common_defaults
  timeout: 30
  retries: 3
  enabled: true

# entities/item1.yml
entity:
  id: example1
  config:
    <<: *common_defaults # Merge from cross-file anchor
    name: Custom Config
    description: Override with custom values
```

## Important Notes

- **Requires `--enable-anchors` flag** - Cross-file anchor resolution is opt-in
- **Files must parse successfully** - If a file fails to parse (invalid YAML or syntax errors), its anchors won't be available to other files
- **Anchor resolution order** - Anchors are collected from all files first, then resolved. This means you can reference anchors defined in files that appear later in the directory structure
- **Preserve mode behavior (YAML only)** - In preserve mode with YAML output, alias references (`*alias`) are preserved in the output rather than expanded (this is intentional to preserve the original structure). In JSON output, aliases are always expanded regardless of mode

## Preserve Mode Validation

When using `--mode preserve` with `--enable-anchors` and `--format yaml`, fyaml validates that the output YAML is valid. If aliases appear before their anchor definitions in the output, fyaml will return an error with guidance on how to fix it.

**To ensure valid YAML in preserve mode:**

1. **Use `@` files for anchor definitions** - Files starting with `@` are processed first alphabetically, ensuring anchors appear before aliases:
   ```
   config/
     @anchors.yml          # Defines all anchors (processed first)
     entities/
       item1.yml           # Uses anchors
   ```

2. **Organize files alphabetically** - Name directories/files containing anchor definitions to sort before those using them:
   ```
   config/
     _shared/              # Underscore sorts before letters
       defaults.yml        # Anchor definitions
     entities/
       item1.yml           # Uses anchors
   ```

3. **Use canonical mode** - If file organization is difficult, use `--mode canonical` which expands aliases, avoiding ordering issues

**Example error when anchors appear after aliases:**

```bash
$ fyaml config/ --mode preserve --enable-anchors
Error: output YAML is invalid: yaml: unknown anchor 'common_defaults' referenced
Enable --verbose flag to see the invalid YAML content
In preserve mode with --enable-anchors, anchor definitions must appear before their aliases.
Consider reorganizing files or using --mode canonical
```
