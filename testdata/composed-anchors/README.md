# Composed Anchors Test

This test verifies that **alias resolution in anchor nodes** works correctly when anchors reference other anchors.

## Test Scenario

This test demonstrates that when an anchor node contains an alias reference to another anchor, yaml.v4 automatically resolves the alias when all anchors are prepended together in the same document.

### Test Files

1. **`shared/templates/base.yml`**: Defines a base anchor `&base_template` with common properties (timeout, retries, enabled, metadata).

2. **`shared/templates/item-template.yml`**: Defines a composed anchor `&item_template` that references the base anchor using a merge key (`<<: *base_template`). This tests that when we collect and prepend anchors, the `item_template` anchor correctly expands the `*base_template` alias to include the base template's content.

3. **`entities/item1.yml`**: Uses the composed anchor (`*item_template`) to verify that cross-file anchor resolution works with composed anchors.

## What This Tests

- **Automatic Alias Resolution**: When `item_template` anchor is collected, it contains `<<: *base_template`. When we prepend all anchors to the target file, yaml.v4 automatically resolves the `*base_template` alias because `&base_template` is also in the prepended text. No manual expansion is needed - yaml.v4 handles this automatically during parsing.

- **Cross-File Resolution**: The final output should show that `*item_template` resolves correctly with all the merged content from `base_template`.

## Expected Behavior

- **Canonical Mode**: Aliases are fully expanded. The `item_template` should contain all properties from `base_template` merged with its own properties.
- **Preserve Mode**: Anchor references are preserved, but the composed anchor should still be correctly expanded when prepended (so it can be marshaled without unresolved references).
