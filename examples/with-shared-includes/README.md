# Shared Includes Example

This example demonstrates using the `--chroot` flag to share include files across multiple configuration directories.

## Structure

```
with-shared-includes/
  shared-includes/
    common.yaml        # Shared configuration used by multiple configs
    defaults.yaml      # Shared defaults
  config1/
    app.yaml           # Config 1 using shared includes
  config2/
    app.yaml           # Config 2 using shared includes
```

## Usage

Pack a specific config directory with `--chroot` set to the project root:

```bash
# From project root (examples/with-shared-includes/)
fyaml config1 --chroot . --enable-includes
fyaml config2 --chroot . --enable-includes
```

Or from anywhere:

```bash
fyaml examples/with-shared-includes/config1 --chroot examples/with-shared-includes --enable-includes
fyaml examples/with-shared-includes/config2 --chroot examples/with-shared-includes --enable-includes
```

## Files

**`shared-includes/common.yaml`** - Shared configuration:

```yaml
timeout: 30
retries: 3
enabled: true
```

**`shared-includes/defaults.yaml`** - Shared defaults:

```yaml
environment: production
debug: false
log_level: info
```

**`config1/app.yaml`** - First configuration using shared includes:

```yaml
app:
  id: app1
  name: My First App
  config: !include ../shared-includes/common.yaml
  settings: !include ../shared-includes/defaults.yaml
```

**`config2/app.yaml`** - Second configuration using the same shared includes:

```yaml
app:
  id: app2
  name: My Second App
  config: !include ../shared-includes/common.yaml
  settings: !include ../shared-includes/defaults.yaml
```

## Expected Output

When packing `config1` with `--chroot . --enable-includes`:

```yaml
app:
  config:
    enabled: true
    retries: 3
    timeout: 30
  id: app1
  name: My First App
  settings:
    debug: false
    environment: production
    log_level: info
```

When packing `config2` with `--chroot . --enable-includes`:

```yaml
app:
  config:
    enabled: true
    retries: 3
    timeout: 30
  id: app2
  name: My Second App
  settings:
    debug: false
    environment: production
    log_level: info
```

## Key Points

- **Relative paths remain relative to the file containing the tag** - `../shared-includes/common.yaml` in `config1/app.yaml` is resolved relative to `config1/app.yaml`, not the current working directory
- **Only files from the pack directory appear in output** - `shared-includes/` files don't appear in the output structure, only their content is included
- **`--chroot` sets the security boundary** - Allows includes from outside the pack directory but within the chroot boundary
- **Portable** - Works the same regardless of where you run the command from

## Without `--chroot`

If you try to pack without `--chroot`:

```bash
fyaml config1 --enable-includes
```

This will fail with an error like:

```
include path ../shared-includes/common.yaml escapes chroot boundary
```

Because includes must be within the pack directory by default.

## See Also

- [File Includes](usage/file-includes.md) - Complete documentation on includes
- [Shared Includes Pattern](usage/file-includes.md#shared-includes-pattern) - More details on this use case
