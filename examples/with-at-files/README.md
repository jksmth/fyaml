# @ Files Example

This example demonstrates how `@` files merge into their parent directory map.

## Structure

```
with-at-files/
  entities/
    @shared.yml    # Merges into entities map
    item1.yml
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

The `@shared.yml` file merges into the `entities` map alongside `item1`:

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags:
          - tag1
          - tag2
  environment: production
  monitoring:
    enabled: true
    interval: 60
  region: us-east-1
```

Note how the contents of `@shared.yml` appear directly in the `entities` map, not nested under a key. Keys are sorted alphabetically for deterministic output.
