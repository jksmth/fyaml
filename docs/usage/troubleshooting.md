# Troubleshooting

## Subcommand and Directory Name Conflicts

If you have a directory named `pack` or `version`, you might encounter a conflict because these are also subcommand names. By default, `fyaml pack` will always invoke the pack subcommand, not try to pack a directory named "pack".

**Solution:** Use the `--dir` flag to explicitly specify the directory:

```bash
# Pack a directory named "pack"
fyaml --dir pack

# Pack a directory named "version"
fyaml --dir version
```

Alternatively, use relative or absolute paths:

```bash
fyaml ./pack
fyaml /path/to/pack
```

## Common Error Messages

**"pack error: failed to build filetree: ..."**

- Directory doesn't exist or path is incorrect
- Permission denied accessing the directory
- Invalid path format

**Solution:** Verify the directory path exists and you have read permissions.

**"pack error: failed to marshal tree: ..."**

- Invalid YAML/JSON syntax in one or more files
- File contains non-map top-level value (scalar or array)

**Solution:** Check the error message for the specific file path, then validate that file's YAML/JSON syntax.

**"expected a map, got a `<type>` which is not supported at this time for \"<filepath>\"**

- File has top-level scalar (e.g., just `hello`) or array (e.g., `- item1`) instead of a map
- Each file must start with a map/object structure

**Solution:** Wrap the content in a map. For example, change `hello` to `value: hello` or `- item1` to `items: [item1]`.

**"invalid format: <format> (must be 'yaml' or 'json')"**

- Invalid `--format` value provided

**Solution:** Use only `yaml` or `json` as the format value.

**"--check requires --output to be specified"**

- `--check` flag used without `--output`

**Solution:** Always specify `-o` or `--output` when using `--check`.

**"failed to read output file: ..."**

- Error reading file for `--check` comparison
- File may be unreadable (permission issue)

**Solution:** Check file permissions and ensure the file is readable.

**"warning: no YAML/JSON files found in directory: <path>"**

- Directory contains no `.yml`, `.yaml`, or `.json` files
- Not an error, but output will be empty/null

**Solution:** Verify you're pointing to the correct directory and that it contains YAML/JSON files.

### Invalid YAML: Anchor Ordering in Preserve Mode

**Error:** `"output YAML is invalid: yaml: unknown anchor '<anchor_name>' referenced"`

**When it occurs:**
- Using `--mode preserve` with `--enable-anchors` and `--format yaml`
- An alias reference (`*alias_name`) appears in the output before its anchor definition (`&anchor_name`)
- YAML requires anchors to be defined before they can be referenced

**Debugging:**
- Use `--verbose` flag to see the invalid YAML content in the error output
- This helps identify exactly where the anchor ordering issue occurs

**Solutions:**

1. **Use `@` files for anchor definitions** - Files starting with `@` are processed first alphabetically:
   ```
   config/
     @anchors.yml          # Defines all anchors (processed first)
     entities/
       item1.yml           # Uses anchors
   ```

2. **Organize files alphabetically** - Name directories/files containing anchors to sort before those using them:
   ```
   config/
     _shared/              # Underscore sorts before letters
       defaults.yml        # Anchor definitions
     entities/
       item1.yml          # Uses anchors
   ```

3. **Use canonical mode** - If file organization is difficult, use `--mode canonical` which expands aliases, avoiding ordering issues

See [Preserve Mode Validation](anchors-aliases.md#preserve-mode-validation) for detailed guidance.

## Debugging Tips

**Use verbose mode to see which files are processed:**

```bash
fyaml -v config/
```

This shows `[DEBUG] Processing: <filepath>` for each file, helping you identify:

- Which files are being read
- If files are being ignored (not shown in debug output)
- The order files are processed

**Verify directory structure:**

```bash
# List all YAML/JSON files
find config/ -type f \( -name "*.yml" -o -name "*.yaml" -o -name "*.json" \)

# Check file permissions
ls -la config/
```

**Check file syntax:**

```bash
# Validate YAML syntax
yamllint config/**/*.yml

# Or use yq to check syntax
yq eval . config/entities/item1.yml
```

**Understand unexpected output:**

1. Check if files are being processed: use `-v` flag
2. Verify file structure matches expected output structure
3. Remember that keys are sorted alphabetically in output (in canonical mode)
4. Check for `@` files that might be merging unexpectedly
5. Verify root-level files are merging as expected

**Verify includes are working:**

If using `--enable-includes`, check:

- Include paths are relative to the file containing the include
- Included files exist and are within the pack root
- Include syntax is correct (`!include`, `!include-text`, or `<<include()>>`)

**Check for file name collisions:**

Files with the same name but different extensions (e.g., `item1.yml` and `item1.yaml`) produce the same key. The last one processed overwrites previous ones. Use a consistent extension throughout your project.

## Next Steps

- See [Examples](../examples.md) for detailed examples
- Review [Command Reference](../reference.md) for all options
