# Basic Example

This example demonstrates the basic FYAML packing functionality with a simple directory structure.

## Structure

```
basic/
  entities/
    item1.yml
    item2.yml
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
entities:
  item1:
    entity:
      id: example1
      attributes:
        name: sample name
        tags:
          - tag1
          - tag2
  item2:
    entity:
      id: example2
      attributes:
        name: another name
        tags: []
```

Note: Keys are sorted alphabetically for deterministic output.
