# Merge Strategies Example

This example demonstrates the difference between `shallow` and `deep` merge strategies when combining YAML files.

## Directory Structure

```
with-merge-strategies/
├── @base.yml       # Base configuration (merged first)
├── @override.yml   # Override configuration (merged second)
└── README.md
```

## How It Works

When multiple files merge into the same location (like `@` files), fyaml uses the merge strategy to determine how overlapping keys are handled:

- **Shallow merge** (default): Later values completely replace earlier values
- **Deep merge**: Nested maps are merged recursively; only leaf values are replaced

## Try It

**Shallow merge (default):**

```bash
fyaml examples/with-merge-strategies --merge shallow
```

Output:

```yaml
config:
  nested:
    c: 3
  setting3: value3
```

Notice that `setting1`, `setting2`, `nested.a`, and `nested.b` from `@base.yml` are gone - the entire `config` map was replaced.

**Deep merge:**

```bash
fyaml examples/with-merge-strategies --merge deep
```

Output:

```yaml
config:
  nested:
    a: 1
    b: 2
    c: 3
  setting1: value1
  setting2: value2
  setting3: value3
```

With deep merge, values from both files are preserved. Only individual keys are overwritten when they conflict.

## When to Use Each

- **Shallow merge**: When you want later files to completely replace sections
- **Deep merge**: When you want to layer configurations, adding or overriding specific keys
