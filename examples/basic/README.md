# Basic Example

This example demonstrates the basic FYAML packing functionality with a simple directory structure.

## Structure

```
basic/
  components/
    database.yml
    cache.yml
```

## Usage

Pack this directory:

```bash
fyaml pack examples/basic
```

Or from this directory:

```bash
cd examples/basic
fyaml pack .
```

## Expected Output

```yaml
components:
  cache:
    name: cache
    settings:
      host: localhost
      port: 6379
      ttl: 3600
    type: redis
  database:
    name: database
    settings:
      host: localhost
      pool_size: 10
      port: 5432
    type: postgresql
```

Note: Keys are sorted alphabetically for deterministic output.

