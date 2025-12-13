# Proposal: Fix mise tool name validation

## Summary

Fix the overly restrictive tool name validation that rejects valid mise backend-prefixed tool names like `npm:@fission-ai/openspec`. The current validation only allows alphanumeric characters, hyphens, and underscores after the prefix, but mise backends support additional characters like `@`, `/`, and `.` for package managers and GitHub-style references.

## Problem Statement

**GitHub Issue**: https://github.com/razd-cli/razd/issues/19

When users configure tools with scoped npm packages or other backend-specific formats, razd fails with a validation error:

```yaml
mise:
  tools:
    node: "22"
    "npm:@fission-ai/openspec": latest
    pnpm: latest
    task: latest
```

Error:

```
⚠ Mise sync check failed: Configuration error: Invalid tool name 'npm:@fission-ai/openspec'.
Tool names must contain only alphanumeric characters, hyphens, and underscores.
Examples: 'node', 'python-3', 'my_tool'
```

## Root Cause

The `validate_tool_name` function in `src/config/mise_validator.rs` strips only a simple prefix (e.g., `npm:`) and then validates the remainder against a strict regex `^[a-zA-Z0-9_-]+$`.

For `npm:@fission-ai/openspec`:

- The prefix `npm:` is stripped correctly
- The remainder `@fission-ai/openspec` contains `@` and `/` which fail validation

## Solution

Update the tool name validation to support mise's full backend syntax:

1. **Recognize backend prefixes**: `npm:`, `pipx:`, `cargo:`, `go:`, `gem:`, `aqua:`, `ubi:`, `github:`, `gitlab:`, `spm:`, `asdf:`, `vfox:`, `http:`
2. **Allow backend-specific characters after prefix**:
   - npm/pipx: `@` for scoped packages, `/` for scope separator
   - go: Full go module paths like `github.com/owner/repo/cmd/tool`
   - aqua/github/gitlab/ubi: `owner/repo` format
   - asdf/vfox: `owner/repo` or plugin URLs
3. **Keep strict validation for standalone tool names** (no prefix): Only alphanumeric, hyphens, underscores

## Impact

- **Scope**: Bug fix - restores compatibility with valid mise configurations
- **Breaking changes**: None - only allows more valid configurations
- **Risk**: Low - adds valid patterns, doesn't remove existing valid patterns

## References

- [mise registry documentation](https://mise.jdx.dev/registry.html) - shows backend syntax
- [mise plugins documentation](https://mise.jdx.dev/plugins.html) - explains plugin:tool format
