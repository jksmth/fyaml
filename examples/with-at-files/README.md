# @ Files Example

This example demonstrates how `@` files merge into their parent directory map.

## Structure

```
with-at-files/
  services/
    @common.yml    # Merges into services map
    api.yml
```

## Usage

Pack this directory:

```bash
fyaml pack examples/with-at-files
```

Or from this directory:

```bash
cd examples/with-at-files
fyaml pack .
```

## Expected Output

The `@common.yml` file merges into the `services` map alongside `api`:

```yaml
services:
  api:
    endpoints:
      - /health
      - /status
    name: api
    version: v1
  environment: production
  monitoring:
    enabled: true
    interval: 60
  region: us-east-1
```

Note how the contents of `@common.yml` appear directly in the `services` map, not nested under a key. Keys are sorted alphabetically for deterministic output.

