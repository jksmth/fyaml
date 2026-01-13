# Cross-File Anchors Example

This example demonstrates fyaml's cross-file YAML anchor functionality, allowing anchors defined in one file to be referenced as aliases in another file.

## Structure

```
with-anchors/
  shared/
    templates/
      common.yml        # Defines &common_template anchor
    defaults.yml        # Defines &common_defaults anchor
  entities/
    item1.yml           # Uses *common_template and *common_defaults
```

## Usage

Pack this directory with anchors enabled:

```bash
fyaml examples/with-anchors --enable-anchors
```

Or from this directory:

```bash
cd examples/with-anchors
fyaml . --enable-anchors
```

**Note:** `fyaml pack` is an alias and works identically for backward compatibility.

## Files

**`shared/templates/common.yml`** - Defines a common template anchor:

```yaml
template: &common_template
  timeout: 30
  retries: 3
  enabled: true
```

**`shared/defaults.yml`** - Defines common defaults anchor:

```yaml
defaults: &common_defaults
  timeout: 30
  retries: 3
  enabled: true
```

**`entities/item1.yml`** - Uses both anchors:

```yaml
entity:
  id: example1
  config:
    <<: *common_defaults # Merge from cross-file anchor
    name: Custom Config
  template: *common_template # Reference cross-file anchor
  attributes:
    name: sample name
```

## Expected Output (Canonical Mode)

When packed with `--enable-anchors --mode canonical` (default), aliases are expanded:

```yaml
entities:
  item1:
    entity:
      attributes:
        name: sample name
      config:
        enabled: true
        name: Custom Config
        retries: 3
        timeout: 30
      id: example1
      template:
        enabled: true
        retries: 3
        timeout: 30
shared:
  defaults:
    defaults:
      enabled: true
      retries: 3
      timeout: 30
  templates:
    common:
      template:
        enabled: true
        retries: 3
        timeout: 30
```

## Expected Output (Preserve Mode)

When packed with `--enable-anchors --mode preserve`, alias references are preserved:

```yaml
entities:
  item1:
    entity:
      id: example1
      config:
        !!merge <<: *common_defaults
        name: Custom Config
      template: *common_template
      attributes:
        name: sample name
shared:
  defaults:
    defaults: &common_defaults
      timeout: 30
      retries: 3
      enabled: true
  templates:
    common:
      template: &common_template
        timeout: 30
        retries: 3
        enabled: true
```

## Notes

- Anchors defined in `shared/` files can be referenced from `entities/` files
- Merge keys (`<<:`) work with cross-file anchors
- In canonical mode, aliases are expanded into their full content
- In preserve mode, alias references are preserved as `*alias_name`
- Requires `--enable-anchors` flag
