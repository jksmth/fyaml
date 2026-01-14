# Basic Usage

## Pack to stdout

The simplest usage is to pack the current directory and output to stdout:

```bash
fyaml
fyaml --mode preserve    # Preserve authored order and comments
```

This reads all YAML/JSON files in the current directory and outputs a single YAML document.

You can also specify a directory explicitly:

```bash
fyaml config/
```

**Note:** `fyaml pack [DIR]` is an alias for `fyaml [DIR]` and works identically for backward compatibility. Both forms pack the current directory when no directory is specified.

## Pack to a File

Write the output to a file using the `-o` or `--output` flag:

```bash
fyaml -o output.yml
```

## Pack Current Directory

If you don't specify a directory, fyaml packs the current directory:

```bash
cd config/
fyaml          # Packs current directory
fyaml pack     # Same as above (alias)
fyaml .        # Explicitly specify current directory
```

## Output as JSON

Use the `--format` or `-f` flag to output JSON instead of YAML:

```bash
fyaml config/ --format json -o output.json
```

**Note:** JSON format doesn't support comments. JSON output always sorts keys alphabetically regardless of mode, so `--mode preserve` has no effect on JSON output.
