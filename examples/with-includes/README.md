# File Includes Example

This example demonstrates fyaml's file include functionality using all three include mechanisms:

- `!include` - Include parsed YAML structures
- `!include-text` - Include raw text content
- `<<include()>>` - Backward-compatible alias for `!include-text`

## Structure

```
with-includes/
  shared/
    defaults.yml        # Shared YAML configuration
  scripts/
    hello.sh           # Simple script
    validate.sh        # Validation script
  entities/
    item1.yml          # Entity using includes
```

## Usage

Pack this directory with includes enabled:

```bash
fyaml examples/with-includes --enable-includes
```

Or from this directory:

```bash
cd examples/with-includes
fyaml . --enable-includes
```

**Note:** `fyaml pack` is an alias and works identically for backward compatibility.

## Files

**`shared/defaults.yml`** - Shared configuration included as YAML:

```yaml
timeout: 30
retries: 3
health_check:
  enabled: true
  interval: 60
```

**`scripts/hello.sh`** - Simple script included as text:

```bash
#!/bin/bash
echo "Hello World"
```

**`scripts/validate.sh`** - Validation script included as text:

```bash
#!/bin/bash
echo "Validating..."
```

**`entities/item1.yml`** - Entity configuration using all include mechanisms:

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
  config: !include ../shared/defaults.yml
  steps:
    - run:
        name: Greeting
        command: !include-text ../scripts/hello.sh
    - run:
        name: Validate
        command: <<include(../scripts/validate.sh)>>
```

## Expected Output

When packed with `--enable-includes`, the includes are processed:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
      config:
        health_check:
          enabled: true
          interval: 60
        retries: 3
        timeout: 30
      steps:
        - run:
            name: Greeting
            command: |
              #!/bin/bash
              echo "Hello World"
        - run:
            name: Validate
            command: |
              #!/bin/bash
              echo "Validating..."
shared:
  defaults:
    health_check:
      enabled: true
      interval: 60
    retries: 3
    timeout: 30
```

Note: The `shared/defaults.yml` file appears in the output because it's part of the directory structure, but its content is also merged into `entities.item1.entity.config` via the `!include` tag.

## Notes

- All three include mechanisms work together
- `!include` merges YAML structures
- `!include-text` and `<<include()>>` include raw text content
- Paths are relative to the file containing the include
- All includes must be within the pack root directory
