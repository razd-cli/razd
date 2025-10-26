# Support Local Project Setup with `razd up`

## Change ID
`support-local-razd-up`

## Type
Enhancement

## Status
Proposed

## Why
When developers work with already-cloned projects, they currently need to run separate `razd install` and `razd setup` commands to get everything ready. This creates friction because:
- Users must remember two commands instead of one intuitive "get everything ready" command
- The mental model breaks: `razd up` works for new projects but not for existing ones
- Onboarding documentation becomes more complex with multiple setup paths
- Consistency with other package managers (npm, bundle, etc.) where one command works everywhere is lost

Making `razd up` work without a URL in local projects simplifies the developer experience and makes the tool more intuitive.

## What Changes

This change modifies the `razd up` command to make the URL argument optional. When no URL is provided, the command will:

1. Detect if the current directory contains a project (Taskfile.yml, mise.toml, or Razdfile.yml)
2. Execute the up workflow directly in the current directory
3. Provide clear error messages if no project is detected

The implementation involves:
- Updating CLI argument parsing to make URL optional in `Commands::Up`
- Modifying `commands::up::execute()` to handle both URL and local project scenarios
- Enhancing error handling and user feedback for local project detection

## Summary
Enhance `razd up` to support running from within an already-cloned project directory (without requiring a git URL), making it behave like `razd install + razd setup` when no URL is provided.

## Problem Statement

Currently, `razd up <url>` always requires a git URL and performs these steps:
1. Clone the repository from URL
2. Change to the cloned directory
3. Execute the up workflow (or fallback to mise install + task setup)

However, developers often:
- Already have projects cloned locally
- Want to quickly set up an existing project they just received or switched to
- Need the same "full setup" workflow without cloning

When working with an already-cloned project, users must manually run `razd install` and `razd setup` separately, which:
- Requires knowing two separate commands instead of one
- Breaks the mental model of "razd up = get everything ready"
- Creates inconsistency between fresh clone and existing project workflows

## Proposed Solution

Make the URL argument optional for `razd up`:

**With URL (current behavior):**
```bash
razd up https://github.com/user/repo.git
# Clones repo, then runs up workflow
```

**Without URL (new behavior):**
```bash
cd my-existing-project
razd up
# Detects local project, runs up workflow in current directory
```

The command will:
1. Check if URL argument is provided
2. If URL: perform git clone → change directory → run up workflow (current behavior)
3. If no URL: detect project in current directory → run up workflow directly

This maintains backward compatibility while providing a more intuitive workflow for local projects.

## Impact

### Affected Components
- CLI argument parsing (make URL optional in `Commands::Up`)
- `commands::up::execute()` implementation
- Up command capability specification

### User-Facing Changes
- `razd up` can now be run without arguments in existing project directories
- Help text and documentation updated to reflect optional URL
- Error messages improved to guide users when no URL and no project detected

### Breaking Changes
None - existing behavior with URL remains unchanged

## Benefits
1. **Simplified workflow**: One command for all setup scenarios
2. **Mental model consistency**: "razd up" always means "get everything ready"
3. **Reduced friction**: No need to remember separate install/setup commands
4. **Better UX**: Matches user expectations from other tools (npm install, bundle install, etc.)

## Risks and Mitigation
- **Risk**: User runs `razd up` in wrong directory by mistake
  - **Mitigation**: Clear error messages when no project detected (no Taskfile.yml, mise.toml, or Razdfile.yml)
  
- **Risk**: Confusion about what `razd up` does without arguments
  - **Mitigation**: Update help text and provide clear output messages about detected project

## Alternatives Considered
1. **Create a new command** (e.g., `razd setup-local`)
   - Rejected: Adds unnecessary cognitive load, fragments the "setup" concept
   
2. **Keep current behavior** (require URL always)
   - Rejected: Doesn't address user pain point of setting up existing projects

3. **Auto-detect git remote and offer to clone**
   - Rejected: Over-complicated, users already have the project

## Dependencies
- No external dependencies
- No conflicts with existing changes (add-razd-yml-config and add-rust-cli-foundation are orthogonal)

## Related Changes
- Complements `add-razd-yml-config` by providing better workflow execution
- Works with `add-rust-cli-foundation` infrastructure

## Open Questions
None - implementation path is clear

## Approval
- [ ] Reviewed by maintainer
- [ ] Technical approach approved
- [ ] Ready for implementation
