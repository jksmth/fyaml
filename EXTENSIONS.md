# fyaml Extensions

This document describes extensions to the [FYAML specification](SPECIFICATION.md) implemented by fyaml. These features enhance functionality while maintaining spec compliance.

**Important:** Extensions are opt-in and do not affect spec-compliant behavior. For spec-compliant usage, use only `.yml` and `.yaml` files with YAML output (default).

## Extension List

1. [JSON Support](#json-support)
2. [File Includes](#file-includes)
3. [Boolean Conversion](#boolean-conversion)
4. [@ Directory Support](#-directory-support)

---

## JSON Support

**Status:** Extension
**Opt-in:** Via file extension (`.json` files) and `--format json` flag

fyaml accepts `.json` files in addition to `.yml` and `.yaml` files. JSON files are processed the same way as YAML files. The `--format json` flag allows outputting JSON instead of YAML.

**Note:** This extension is not part of the FYAML specification. For spec-compliant behavior, use only `.yml` and `.yaml` files with YAML output (default).

See [docs/usage.md#output-format](docs/usage.md#output-format) and [docs/reference.md#--format--f](docs/reference.md#--format--f) for usage details.

---

## File Includes

**Status:** Extension
**Opt-in:** Via `--enable-includes` flag

Process include directives to include content from other files. Supports three mechanisms: `!include` (parsed YAML structures), `!include-text` (raw text content), and `<<include()>>` (backward-compatible alias for `!include-text`).

**Note:** This extension is not part of the FYAML specification. Without the flag, include directives are passed through unchanged, preserving spec-compliant behavior.

See [docs/usage.md#file-includes](docs/usage.md#file-includes) for complete usage documentation.

---

## Boolean Conversion

**Status:** Extension
**Opt-in:** Via `--convert-booleans` flag

Convert YAML 1.1-style boolean values (`on`, `off`, `yes`, `no`, `y`, `n`) to YAML 1.2 booleans (`true`, `false`). Quoted values like `"on"` remain as strings.

**Note:** This extension is not part of the FYAML specification. fyaml outputs YAML 1.2 format where only `true` and `false` are booleans. The flag converts legacy YAML 1.1 boolean values to their YAML 1.2 equivalents.

See [docs/usage.md#converting-onoff-and-yesno-to-truefalse](docs/usage.md#converting-onoff-and-yesno-to-truefalse) for complete usage documentation.

---

## @ Directory Support

**Status:** Extension
**Opt-in:** Via directory naming convention (directories starting with `@`)

Directories starting with `@` merge their contents into the parent directory map, similar to how `@` files work (spec rule 6). The directory name after `@` is ignored (organizational only), and contents merge directly into the parent map.

**Note:** This extension mirrors the behavior of `@` files but applies to directories. It is not part of the FYAML specification.

See [docs/usage.md#-directories](docs/usage.md#-directories) for complete usage documentation, examples, use cases, and edge cases.

---

## Extension Philosophy

Extensions in fyaml follow these principles:

1. **Opt-in:** Extensions must be explicitly enabled or used (via flags, file extensions, or naming conventions)
2. **Non-breaking:** Extensions do not affect spec-compliant behavior when not used
3. **Well-documented:** All extensions are clearly marked and documented
4. **Enhancement, not replacement:** Extensions enhance functionality without replacing spec behavior

When in doubt, default behavior remains spec-compliant.
