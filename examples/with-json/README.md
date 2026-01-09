# JSON Input/Output Example

This example demonstrates fyaml's JSON support - accepting JSON input files and outputting JSON format.

## Structure

```
with-json/
  entities/
    item1.yml        # YAML file
    item2.json       # JSON file
  config/
    settings.json    # JSON file
```

## Usage

### Output as YAML (Default)

```bash
fyaml examples/with-json
```

This processes both YAML and JSON input files and outputs YAML:

```yaml
config:
  settings:
    database:
      host: localhost
      port: 5432
    logging:
      level: info
entities:
  item1:
    entity:
      attributes:
        name: sample name
        tags:
          - tag1
          - tag2
      id: example1
  item2:
    entity:
      attributes:
        name: another name
        tags: []
      id: example2
```

### Output as JSON

```bash
fyaml examples/with-json --format json
```

This processes both YAML and JSON input files and outputs JSON:

```json
{
  "config": {
    "settings": {
      "database": {
        "host": "localhost",
        "port": 5432
      },
      "logging": {
        "level": "info"
      }
    }
  },
  "entities": {
    "item1": {
      "entity": {
        "id": "example1",
        "attributes": {
          "name": "sample name",
          "tags": [
            "tag1",
            "tag2"
          ]
        }
      }
    },
    "item2": {
      "entity": {
        "id": "example2",
        "attributes": {
          "name": "another name",
          "tags": []
        }
      }
    }
  }
}
```

### Preserve Mode with JSON

```bash
fyaml examples/with-json --mode preserve --format json
```

In preserve mode with JSON output:

- Key order is maintained (JSON preserves object key order)
- Comments are lost (JSON doesn't support comments)

## Files

**`entities/item1.yml`** (YAML input):

```yaml
entity:
  id: example1
  attributes:
    name: sample name
    tags:
      - tag1
      - tag2
```

**`entities/item2.json`** (JSON input):

```json
{
  "entity": {
    "id": "example2",
    "attributes": {
      "name": "another name",
      "tags": []
    }
  }
}
```

**`config/settings.json`** (JSON input):

```json
{
  "database": {
    "host": "localhost",
    "port": 5432
  },
  "logging": {
    "level": "info"
  }
}
```

## Key Points

- **Mixed input**: fyaml accepts both `.yml`, `.yaml`, and `.json` files as input
- **JSON output**: Use `--format json` to output JSON instead of YAML
- **Preserve mode**: In preserve mode, key order is maintained in JSON output
- **Comments**: JSON doesn't support comments, so comments are lost regardless of mode
- **Deterministic**: Both YAML and JSON output are deterministic

## Notes

- JSON is a subset of YAML, so JSON files are parsed as YAML
- All input files (YAML or JSON) must contain a map/object at the top level
- JSON output uses 2-space indentation by default (customizable with `--indent`)
- Empty output returns `null` in JSON format (vs. empty string in YAML format)
