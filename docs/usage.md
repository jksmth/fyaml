# Usage Guide

This guide covers basic usage patterns, commands, and common workflows with fyaml.

## Basic Usage

### Pack to stdout

The simplest usage is to pack a directory and output to stdout:

```bash
fyaml pack config/
```

This reads all YAML/JSON files in `config/` and outputs a single YAML document.

You can also run `fyaml` without arguments to pack the current directory:

```bash
fyaml
```

This is equivalent to `fyaml pack .`

### Pack to a File

Write the output to a file using the `-o` or `--output` flag:

```bash
fyaml pack config/ -o output.yml
```

### Pack Current Directory

If you don't specify a directory, fyaml packs the current directory:

```bash
cd config/
fyaml          # Packs current directory
fyaml pack     # Same as above
fyaml pack .   # Explicitly specify current directory
```

### Output as JSON

Use the `--format` or `-f` flag to output JSON instead of YAML:

```bash
fyaml pack config/ --format json -o output.json
```

## Common Patterns

### Verify Output Matches Existing File

Use the `--check` flag to verify that the generated output matches an existing file:

```bash
fyaml pack config/ -o output.yml --check
```

This is useful in CI/CD pipelines to ensure configuration hasn't changed unexpectedly.

**Exit codes:**
- `0` - Output matches the file
- `2` - Output differs from the file
- `1` - Error occurred during packing

### Combine with Other Tools

fyaml works well with other command-line tools:

```bash
# Pipe to jq for JSON processing
fyaml pack config/ --format json | jq '.services'

# Use with envsubst for variable substitution
fyaml pack config/ | envsubst > config-final.yml

# Validate YAML output
fyaml pack config/ | yamllint

# Format with yq
fyaml pack config/ | yq eval -P
```

## Directory Structure Rules

### Basic Structure

Directory names become map keys in the output:

```
config/
  services/
    api.yml
    db.yml
```

**`services/api.yml`:**
```yaml
name: api
port: 8080
```

**`services/db.yml`:**
```yaml
name: db
type: postgresql
```

Produces:

```yaml
services:
  api:
    name: api
    port: 8080
  db:
    name: db
    type: postgresql
```

### Root-Level Files

Files at the root level merge directly into the output (their filename is not used as a key):

```
config/
  metadata.yml     # Merges directly into root
  services/
    api.yml
```

**`metadata.yml`:**
```yaml
version: 1.0.0
environment: production
```

**`services/api.yml`:**
```yaml
name: api
port: 8080
```

Produces:

```yaml
environment: production
services:
  api:
    name: api
    port: 8080
version: 1.0.0
```

**Note:** The `metadata.yml` contents merge directly into the root. The filename `metadata` is not used as a key.

### @ Files

Files starting with `@` merge their contents into the parent directory map:

```
config/
  services/
    @common.yml    # Merges into services map
    api.yml
```

**`services/@common.yml`:**
```yaml
environment: production
region: us-east-1
```

**`services/api.yml`:**
```yaml
name: api
port: 8080
```

Produces:

```yaml
services:
  api:
    name: api
    port: 8080
  environment: production
  region: us-east-1
```

The `@` prefix is removed, and the file's contents are merged directly into the parent map.

**Note:** While the specification allows multiple `@` files in the same directory, it's recommended to use only one per directory. If multiple `@` files exist, they all merge into the parent map, and key collisions are resolved by the last file processed (order is not guaranteed).

### Nested Directories

Directories can be nested to any depth:

```
config/
  infrastructure/
    compute/
      servers.yml
      load-balancers.yml
    network/
      vpc.yml
```

**`infrastructure/compute/servers.yml`:**
```yaml
type: ec2
instances: 5
```

**`infrastructure/compute/load-balancers.yml`:**
```yaml
type: alb
count: 2
```

**`infrastructure/network/vpc.yml`:**
```yaml
cidr: 10.0.0.0/16
```

Produces:

```yaml
infrastructure:
  compute:
    load-balancers:
      count: 2
      type: alb
    servers:
      instances: 5
      type: ec2
  network:
    vpc:
      cidr: 10.0.0.0/16
```

## File Naming

### Supported Extensions

fyaml processes files with these extensions:

- `.yml`
- `.yaml`
- `.json`

You can mix these file types in the same directory structure. JSON files must be valid JSON (not JSON5 or other variants) and must have a top-level object (map), just like YAML files.

### Ignored Files

fyaml automatically ignores:

- Files and directories starting with `.` (dot files)
- Files without supported extensions

### File Name to Key Mapping

The filename (without extension) becomes the key in the output:

- `database.yml` → key: `database`
- `api-service.yaml` → key: `api-service`
- `config.json` → key: `config`

**Exceptions:**
- Root-level files: filename is ignored, contents merge directly
- `@` files: `@` prefix is removed, contents merge into parent

### Special Characters

File and directory names can contain hyphens, underscores, numbers, and mixed case. Special characters are preserved in the output key:

- `api-service.yml` → key: `api-service`
- `api_service.yml` → key: `api_service`
- `ServiceV2.yml` → key: `ServiceV2`

### File Name Collisions

If you have files with the same name but different extensions in the same directory (e.g., `service.yml`, `service.yaml`, `service.json`), they all produce the same key. The last one processed will overwrite previous ones, and processing order is not guaranteed. **Use a consistent extension throughout your project to avoid collisions.**

## Output Format

### YAML (Default)

Default output is YAML:

```bash
fyaml pack config/
```

YAML output uses standard YAML formatting with proper indentation.

### JSON

Output JSON using the `--format` flag:

```bash
fyaml pack config/ --format json
```

JSON output is formatted with 2-space indentation.

### Empty Output

When no files are found:

- **YAML format**: Returns empty output (0 bytes) - aligns with `yq` behavior
- **JSON format**: Returns `null` - aligns with `jq` behavior

A warning is printed to stderr in both cases.

## Deterministic Output

fyaml produces deterministic output:

- Identical directory structures always produce identical output
- Map keys are sorted alphabetically
- This makes output suitable for version control and comparison

This means you can safely commit the generated output to version control, and use `--check` in CI to ensure the source files and generated output stay in sync:

```bash
# In CI/CD pipeline - verify config hasn't changed
fyaml pack config/ -o config.yml --check
# Exits with code 2 if source files changed but output wasn't regenerated
```

## Working with Large Configurations

### Organize by Domain

Group related configuration by domain:

```
config/
  services/
    api.yml
    worker.yml
  infrastructure/
    database.yml
    cache.yml
  monitoring/
    alerts.yml
    dashboards.yml
```

### Use @ Files for Shared Configuration

Use `@` files to share common configuration:

```
config/
  services/
    @common.yml      # Shared settings
    api.yml          # API-specific
    worker.yml       # Worker-specific
```

The `@common.yml` might contain:

```yaml
environment: production
region: us-east-1
monitoring:
  enabled: true
```

## Integration with Templating

fyaml doesn't support templating, but you can combine it with other tools:

### Using envsubst

```bash
# Set environment variables
export SERVICE_NAME=api
export REPLICA_COUNT=3

# Pack and substitute
fyaml pack config/ | envsubst > config-final.yml
```

Your YAML files can contain `${VARIABLE}` syntax that `envsubst` will replace.

### Using sed

```bash
# Replace placeholders
fyaml pack config/ | sed 's/{{VERSION}}/v1.0.0/g' > output.yml
```

### Using yq

```bash
# Modify after packing
fyaml pack config/ | yq eval '.services.api.replicas = 5' - > output.yml
```

## Best Practices

1. **Keep files focused**: Each file should represent a single logical unit
2. **Use descriptive names**: File and directory names should clearly indicate their purpose
3. **Organize hierarchically**: Use directory structure to reflect configuration hierarchy
4. **Version control**: Commit both source directory structure and generated output
5. **Verify in CI**: Use `--check` flag in CI to catch unexpected changes
6. **Document structure**: Add README files in directories to explain organization

## Limitations

### File Content Requirements

Each YAML/JSON file must contain a map (object/dictionary) at the top level. The file itself must be a map, but nested values within that map can be any YAML type (scalars, arrays, nested maps, etc.).

**Supported:**
```yaml
# ✅ Top-level is a map
name: api
items: [1, 2, 3]           # Array nested in map
settings:                   # Nested map
  timeout: 30
```

**Not supported:**
```yaml
# ❌ Top-level is a scalar
hello

# ❌ Top-level is an array
- item1
- item2
```

If you attempt to pack a file containing a top-level scalar or array, fyaml will return an error: `expected a map, got a <type> which is not supported at this time for "<filepath>"`.

### YAML Anchors and Aliases

YAML anchors (`&anchor`) and aliases (`*alias`) are resolved **within each individual file** during parsing. Anchors and aliases **cannot** reference values across different files—they only work within a single YAML document.

If you need shared values across files, use `@` files to merge common configuration.

### Multi-Document YAML Files

YAML supports multiple documents in a single file, separated by `---`. However, **fyaml only processes the first document** in multi-document files. Subsequent documents are silently ignored.

Instead of using multi-document files, organize your resources using separate files:

```yaml
# Instead of this (multi-document):
config.yml:
  ---
  database:
    host: localhost
  ---
  cache:
    host: localhost

# Use this (fyaml's filesystem-based approach):
config/
  database.yml    # Contains the database config
  cache.yml       # Contains the cache config
```

### Large Files

fyaml processes files in memory. For very large files (hundreds of MB), this could consume significant memory. However, for typical configuration files (KB to low MB range), performance is excellent. Keep individual files focused and reasonably sized.

## Next Steps

- See [Examples](examples.md) for detailed examples
- Review [Command Reference](reference.md) for all options

