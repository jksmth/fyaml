# Examples

This page provides complex examples showing how multiple fyaml concepts work together. For basic examples, see the [Usage Guide](usage.md).

## Combining Multiple Concepts

This example demonstrates how root-level files, @ files, and deep nesting work together:

### Directory Structure

```
config/
  @defaults.yml
  services/
    @common.yml
    api/
      server.yml
      routes.yml
    worker/
      config.yml
  infrastructure/
    database.yml
    cache.yml
```

### Input Files

**`@defaults.yml`:**
```yaml
project: example
version: 1.0.0
```

**`services/@common.yml`:**
```yaml
environment: production
region: us-east-1
```

**`services/api/server.yml`:**
```yaml
port: 8080
timeout: 30
```

**`services/api/routes.yml`:**
```yaml
paths:
  - /health
  - /api
```

**`services/worker/config.yml`:**
```yaml
workers: 5
queue: default
```

**`infrastructure/database.yml`:**
```yaml
type: postgresql
host: localhost
port: 5432
```

**`infrastructure/cache.yml`:**
```yaml
type: redis
host: localhost
port: 6379
```

### Command

```bash
fyaml pack config/
```

### Output

```yaml
environment: production
infrastructure:
  cache:
    host: localhost
    port: 6379
    type: redis
  database:
    host: localhost
    port: 5432
    type: postgresql
project: example
region: us-east-1
services:
  api:
    routes:
      paths:
        - /api
        - /health
    server:
      port: 8080
      timeout: 30
  worker:
    config:
      queue: default
      workers: 5
version: 1.0.0
```

**Key points:**

- `@defaults.yml` at root merges `project` and `version` into the top level
- `services/@common.yml` merges `environment` and `region` into the `services` map
- Deep nesting (3 levels: `services/api/server.yml`) works naturally
- All concepts work together in a single structure

## Deep Nesting Example

This example shows how deeply nested structures are handled:

### Directory Structure

```
config/
  level1/
    level2/
      level3/
        level4/
          deep.yml
        another.yml
      middle.yml
    top.yml
```

### Input Files

**`level1/level2/level3/level4/deep.yml`:**
```yaml
value: deepest
```

**`level1/level2/level3/another.yml`:**
```yaml
value: level3
```

**`level1/level2/middle.yml`:**
```yaml
value: middle
```

**`level1/top.yml`:**
```yaml
value: top
```

### Output

```yaml
level1:
  level2:
    level3:
      another:
        value: level3
      level4:
        deep:
          value: deepest
    middle:
      value: middle
  top:
    value: top
```

This demonstrates that fyaml handles arbitrary nesting depth - organize your configuration as deeply as needed.

## Root Files with Nested Structure

Combining root-level files with nested directories:

### Directory Structure

```
config/
  metadata.yml
  settings.yml
  app/
    services/
      api.yml
      worker.yml
    database/
      postgres.yml
```

### Input Files

**`metadata.yml`:**
```yaml
name: myapp
version: 1.0.0
```

**`settings.yml`:**
```yaml
debug: false
log_level: info
```

**`app/services/api.yml`:**
```yaml
port: 8080
```

**`app/services/worker.yml`:**
```yaml
workers: 3
```

**`app/database/postgres.yml`:**
```yaml
host: localhost
port: 5432
```

### Output

```yaml
app:
  database:
    postgres:
      host: localhost
      port: 5432
  services:
    api:
      port: 8080
    worker:
      workers: 3
debug: false
log_level: info
metadata:
  name: myapp
  version: 1.0.0
```

**Note:** The root-level files (`metadata.yml` and `settings.yml`) merge their contents directly into the output root, while nested directories create the hierarchical structure.

## Try It Yourself

All examples in this documentation are based on the examples in the repository. Try them out:

```bash
# Clone the repository
git clone https://github.com/jksmth/fyaml.git
cd fyaml

# Try the basic example
fyaml pack examples/basic

# Try the @ files example
fyaml pack examples/with-at-files
```

See the [examples directory](https://github.com/jksmth/fyaml/tree/main/examples) for more runnable examples.
