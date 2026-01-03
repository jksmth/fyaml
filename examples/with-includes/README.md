# File Includes Example

This example demonstrates fyaml's file include functionality using all three include mechanisms:

- `!include` - Include parsed YAML structures
- `!include-text` - Include raw text content
- `<<include()>>` - Backward-compatible alias for `!include-text`

## Structure

```
with-includes/
  common/
    defaults.yml        # Shared YAML configuration
  scripts/
    deploy.sh          # Deployment script
    validate.sh        # Validation script
  services/
    api.yml            # Service using includes
```

## Usage

Pack this directory with includes enabled:

```bash
fyaml pack examples/with-includes --enable-includes
```

Or from this directory:

```bash
cd examples/with-includes
fyaml pack . --enable-includes
```

## Files

**`common/defaults.yml`** - Shared configuration included as YAML:
```yaml
timeout: 30
retries: 3
health_check:
  enabled: true
  interval: 60
```

**`scripts/deploy.sh`** - Deployment script included as text:
```bash
#!/bin/bash
echo "Deploying application..."
kubectl apply -f manifests/
```

**`scripts/validate.sh`** - Validation script included as text:
```bash
#!/bin/bash
echo "Validating configuration..."
kubectl validate -f manifests/
```

**`services/api.yml`** - Service configuration using all include mechanisms:
```yaml
name: api
version: v1
config: !include ../common/defaults.yml
steps:
  - run:
      name: Deploy
      command: !include-text ../scripts/deploy.sh
  - run:
      name: Validate
      command: <<include(../scripts/validate.sh)>>
```

## Expected Output

When packed with `--enable-includes`, the includes are processed:

```yaml
common:
  defaults:
    health_check:
      enabled: true
      interval: 60
    retries: 3
    timeout: 30
services:
  api:
    config:
      health_check:
        enabled: true
        interval: 60
      retries: 3
      timeout: 30
    name: api
    steps:
      - run:
          name: Deploy
          command: |
            #!/bin/bash
            echo "Deploying application..."
            kubectl apply -f manifests/
      - run:
          name: Validate
          command: |
            #!/bin/bash
            echo "Validating configuration..."
            kubectl validate -f manifests/
    version: v1
```

Note: The `common/defaults.yml` file appears in the output because it's part of the directory structure, but its content is also merged into `services.api.config` via the `!include` tag.

## Notes

- All three include mechanisms work together
- `!include` merges YAML structures
- `!include-text` and `<<include()>>` include raw text content
- Paths are relative to the file containing the include
- All includes must be within the pack root directory

