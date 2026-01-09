# Examples

This page provides complex examples showing how multiple fyaml concepts work together. For basic examples, see the [Usage Guide](usage.md).

## Combining Multiple Concepts

This example demonstrates how root-level files, @ files, and deep nesting work together:

### Directory Structure

```
config/
  @defaults.yml
  entities/
    @shared.yml
    item1/
      config.yml
      metadata.yml
    item2/
      settings.yml
  category1/
    item3.yml
    item4.yml
```

### Input Files

**`@defaults.yml`:**

```yaml
project: example
version: 1.0.0
```

**`entities/@shared.yml`:**

```yaml
environment: production
region: us-east-1
```

**`entities/item1/config.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags: []
```

**`entities/item1/metadata.yml`:**

```yaml
related:
  - id: example2
    attributes:
      name: related item
      tags:
        - tag1
```

**`entities/item2/settings.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: another item
    tags:
      - tag2
```

**`category1/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags: []
```

**`category1/item4.yml`:**

```yaml
entity:
  id: example4
  attributes:
    name: fourth item
    tags: []
```

### Command

```bash
fyaml config/
```

### Output

```yaml
category1:
  item3:
    entity:
      id: example3
      attributes:
        name: third item
        tags: []
  item4:
    entity:
      id: example4
      attributes:
        name: fourth item
        tags: []
entities:
  environment: production
  item1:
    config:
      entity:
        id: example1
        attributes:
          name: sample name
          tags: []
    metadata:
      related:
        - attributes:
            name: related item
            tags:
              - tag1
          id: example2
  item2:
    settings:
      entity:
        id: example2
        attributes:
          name: another item
          tags:
            - tag2
  region: us-east-1
project: example
version: 1.0.0
```

**Key points:**

- `@defaults.yml` at root merges `project` and `version` into the top level
- `entities/@shared.yml` merges `environment` and `region` into the `entities` map
- Deep nesting (3 levels: `entities/item1/config.yml`) works naturally
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
entity:
  id: example1
  attributes:
    name: deepest item
    tags: []
```

**`level1/level2/level3/another.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: level3 item
    tags: []
```

**`level1/level2/middle.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: middle item
    tags: []
```

**`level1/top.yml`:**

```yaml
entity:
  id: example4
  attributes:
    name: top item
    tags: []
```

### Output

```yaml
level1:
  level2:
    level3:
      another:
        entity:
          id: example2
          attributes:
            name: level3 item
            tags: []
      level4:
        deep:
          entity:
            id: example1
            attributes:
              name: deepest item
              tags: []
    middle:
      entity:
        id: example3
        attributes:
          name: middle item
          tags: []
  top:
    entity:
      id: example4
      attributes:
        name: top item
        tags: []
```

This demonstrates that fyaml handles arbitrary nesting depth - organize your configuration as deeply as needed.

## Root Files with Nested Structure

Combining root-level files with nested directories:

### Directory Structure

```
config/
  shared.yml
  settings.yml
  category1/
    entities/
      item1.yml
      item2.yml
    group1/
      item3.yml
```

### Input Files

**`shared.yml`:**

```yaml
name: myapp
version: 1.0.0
```

**`settings.yml`:**

```yaml
debug: false
log_level: info
```

**`category1/entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: first item
    tags: []
```

**`category1/entities/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: second item
    tags:
      - tag1
```

**`category1/group1/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags: []
```

### Output

```yaml
category1:
  entities:
    item1:
      entity:
        id: example1
        attributes:
          name: first item
          tags: []
    item2:
      entity:
        id: example2
        attributes:
          name: second item
          tags:
            - tag1
  group1:
    item3:
      entity:
        id: example3
        attributes:
          name: third item
          tags: []
debug: false
log_level: info
name: myapp
version: 1.0.0
```

**Note:** The root-level files (`shared.yml` and `settings.yml`) merge their contents directly into the output root, while nested directories create the hierarchical structure.

## File Includes Example

This example demonstrates using file includes to share common configuration and include script content:

### Directory Structure

```
config/
  entities/
    item1.yml
  shared/
    defaults.yml
  scripts/
    hello.sh
    validate.sh
```

### Input Files

**`entities/item1.yml`:**

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

**`shared/defaults.yml`:**

```yaml
timeout: 30
retries: 3
health_check:
  enabled: true
  interval: 60
```

**`scripts/hello.sh`:**

```bash
#!/bin/bash
echo "Hello World"
```

**`scripts/validate.sh`:**

```bash
#!/bin/bash
echo "Validating..."
```

### Command

```bash
fyaml config/ --enable-includes
```

### Output

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags: []
      config:
        timeout: 30
        retries: 3
        health_check:
          enabled: true
          interval: 60
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
    timeout: 30
    retries: 3
    health_check:
      enabled: true
      interval: 60
```

**Key points:**

- `!include` merges YAML structures (the `config` key contains the merged content from `defaults.yml`)
- `!include-text` and `<<include()>>` include raw text content (the script files are included as-is)
- All three include mechanisms can be used together
- Includes are processed relative to the file containing them

## @ Directories Example

This example demonstrates using `@` directories to organize files without creating additional nesting levels:

### Directory Structure

```
config/
  entities/
    item1.yml
    @group1/              # Merges into entities map
      item2.yml
      item3.yml
    @group2/              # Merges into entities map
      item4.yml
```

### Input Files

**`entities/item1.yml`:**

```yaml
entity:
  id: example1
  attributes:
    name: first item
    tags: []
```

**`entities/@group1/item2.yml`:**

```yaml
entity:
  id: example2
  attributes:
    name: second item
    tags:
      - tag1
```

**`entities/@group1/item3.yml`:**

```yaml
entity:
  id: example3
  attributes:
    name: third item
    tags:
      - tag2
```

**`entities/@group2/item4.yml`:**

```yaml
entity:
  id: example4
  attributes:
    name: fourth item
    tags: []
```

### Command

```bash
fyaml config/
```

### Output

```yaml
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: first item
        tags: []
  item2:
    entity:
      id: example2
      attributes:
        name: second item
        tags:
          - tag1
  item3:
    entity:
      id: example3
      attributes:
        name: third item
        tags:
          - tag2
  item4:
    entity:
      id: example4
      attributes:
        name: fourth item
        tags: []
```

**Key points:**

- The `@group1` and `@group2` directory names do not appear in the output
- All files from `@` directories merge directly into the parent map (`entities`)
- This is useful for organizing large numbers of files without creating deep nesting
- Files are organized in source but produce a flat structure in output

## Output Modes Comparison

This example demonstrates the difference between canonical and preserve modes using the same input files:

### Directory Structure

```
config/
  entities/
    zebra.yml
    alpha.yml
    @group1/
      item-z.yml
    @group2/
      item-a.yml
```

### Input Files

**`entities/zebra.yml`:**

```yaml
# This entity has a name that sorts later alphabetically
entity:
  id: zebra
  attributes:
    name: Zebra
    tags:
      - animal
```

**`entities/alpha.yml`:**

```yaml
# This entity has a name that sorts earlier alphabetically
entity:
  id: alpha
  attributes:
    name: Alpha
    tags:
      - first
```

**`entities/@group1/item-z.yml`:**

```yaml
# Entity from @group1 - @ directories are processed in alphabetical order
entity:
  id: item-z
  attributes:
    name: Item Z
```

**`entities/@group2/item-a.yml`:**

```yaml
# Entity from @group2 - @ directories are processed in alphabetical order
entity:
  id: item-a
  attributes:
    name: Item A
```

### Canonical Mode Output (Default)

```bash
fyaml config/ --mode canonical
# or simply
fyaml config/
```

**Output:**

```yaml
entities:
  alpha:
    entity:
      attributes:
        name: Alpha
        tags:
          - first
      id: alpha
  item-a:
    entity:
      attributes:
        name: Item A
      id: item-a
  item-z:
    entity:
      attributes:
        name: Item Z
      id: item-z
  zebra:
    entity:
      attributes:
        name: Zebra
        tags:
          - animal
      id: zebra
```

**Key points:**

- Keys are sorted alphabetically: `alpha`, `item-a`, `item-z`, `zebra`
- Comments are removed
- `@` directory order doesn't affect output (all keys sorted)

### Preserve Mode Output

```bash
fyaml config/ --mode preserve
```

**Output:**

```yaml
entities:
  item-z:
    # Entity from @group1 - @ directories processed first (alphabetically)
    entity:
      id: item-z
      attributes:
        name: Item Z
  item-a:
    # Entity from @group2 - @ directories processed in alphabetical order
    entity:
      id: item-a
      attributes:
        name: Item A
  alpha:
    # Regular file - processed after @ directories, in alphabetical order
    entity:
      id: alpha
      attributes:
        name: Alpha
        tags:
          - first
  zebra:
    # Regular file - processed after @ directories, in alphabetical order
    entity:
      id: zebra
      attributes:
        name: Zebra
        tags:
          - animal
```

**Key points:**

- Files and directories are processed in alphabetical order by canonical path (deterministic across operating systems)
- `@` directories are processed first (because `@` sorts before letters), then regular files
- Within each category, items are sorted alphabetically: `@group1` before `@group2`, `alpha` before `zebra`
- Comments are preserved
- Key order within each file is maintained (authored order)
- The order is deterministic and cross-platform, not dependent on filesystem order

### When to Use Each Mode

- **Canonical mode**: Use when you need sorted keys and don't need comments. Sorted keys make diffs more readable and predictable.
- **Preserve mode**: Use when you want to maintain documentation in comments or preserve the authored key order from your source files.

Both modes are deterministic (same input always produces same output) and suitable for version control and CI/CD. The difference is in key ordering (sorted vs. authored) and comment preservation.

## Try It Yourself

All examples in this documentation are based on the examples in the repository. Try them out:

```bash
# Clone the repository
git clone https://github.com/jksmth/fyaml.git
cd fyaml

# Try the basic example
fyaml examples/basic

# Try the @ files example
fyaml examples/with-at-files
```

See the [examples directory](https://github.com/jksmth/fyaml/tree/main/examples) for more runnable examples.
