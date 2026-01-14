# Common Patterns

## Verify Output Matches Existing File

Use the `--check` flag to verify that the generated output matches an existing file or stdin:

```bash
# Check against a file
fyaml -o output.yml --check

# Check against stdin (implicit)
fyaml --check < expected.yml

# Check against stdin (explicit)
fyaml --check --output - < expected.yml

# In CI/CD pipelines with pipes
cat expected.yml | fyaml --check
```

This is useful in CI/CD pipelines to ensure configuration hasn't changed unexpectedly.

**Exit codes:**

- `0` - Output matches the file or stdin
- `2` - Output differs from the file or stdin
- `1` - Error occurred during packing

**Note:** When using `--check` without `--output` or with `--output -`, fyaml reads from stdin. If stdin is a terminal (not piped or redirected), an error will be returned to prevent the program from blocking.

## Combine with Other Tools

fyaml works well with other command-line tools:

```bash
# Pipe to jq for JSON processing
fyaml --format json | jq '.entities'

# Use with envsubst for variable substitution
fyaml | envsubst > config-final.yml

# Replace placeholders with sed
fyaml | sed 's/{{VERSION}}/v1.0.0/g' > output.yml

# Validate YAML output
fyaml | yamllint

# Format with yq
fyaml | yq eval -P
```

**Note:** fyaml doesn't support templating, but you can pipe output to tools like `envsubst`, `sed`, or `yq` for post-processing.
