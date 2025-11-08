# Preserve Razdfile Content During Mise Sync

## Problem

Currently, when synchronizing from `mise.toml` to `Razdfile.yml`, the sync process completely overwrites the entire Razdfile, causing critical data loss:

1. **Platform-specific commands are lost**: Commands with `platform: windows` or other platform specifiers get removed
2. **Task formatting is destroyed**: Original YAML formatting and comments are lost
3. **Other sections are unnecessarily touched**: `env`, `vars`, and unrelated configuration gets reformatted
4. **Task order changes**: Tasks get reordered alphabetically, losing intended structure

### Example Data Loss

**Before sync (Razdfile.yml):**
```yaml
tasks:
  install:
    desc: Install tools and dependencies
    cmds:
    - cmd: scoop install gcc make
      platform: windows
    - mise install
```

**After sync from mise.toml:**
```yaml
tasks:
  install:
    desc: Install tools and dependencies
    cmds:
    - scoop install gcc make  # ❌ Platform specifier lost!
    - mise install
```

## Solution

Implement **surgical synchronization** that:

1. **Only modifies the `mise:` section** when syncing from mise.toml
2. **Preserves all other content exactly** including formatting, comments, and structure
3. **Maintains platform-specific command metadata** in tasks
4. **Respects original YAML structure** and field ordering

## Benefits

- ✅ No more data loss during synchronization
- ✅ Platform-specific workflows remain intact
- ✅ Developer-authored formatting is preserved
- ✅ Sync becomes a safe, non-destructive operation
- ✅ Enables reliable bidirectional sync workflow

## Scope

This change affects the mise-to-Razdfile synchronization path only. The Razdfile-to-mise sync direction already works correctly since it only reads the `mise:` section.

## Why

This change addresses a critical data integrity issue that affects developer workflows. Currently, mise.toml synchronization is destructive and unpredictable, leading to:

1. **Lost platform-specific configurations** that developers carefully crafted
2. **Broken cross-platform workflows** that depend on platform specifiers
3. **Formatting churn** that pollutes git diffs with unrelated changes
4. **Developer frustration** from having to restore lost configuration

By implementing surgical synchronization, we make the mise integration reliable and trustworthy for production use.

## Implementation Strategy

Use YAML parsing and selective updating to modify only the `mise:` section while preserving the rest of the document structure.