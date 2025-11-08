# Add Interactive Yes Flag

## Problem Statement

Users currently must manually respond to prompts during `razd` operations:
- Mise/Razdfile synchronization conflicts require manual confirmation
- Creating new Razdfile.yml requires user approval
- Backup operations require confirmation before overwriting files

This manual interaction prevents automation workflows, CI/CD pipelines, and scripted project setups from running unattended. Users need a way to automatically approve all prompts.

## Current Behavior

The tool prompts users in several scenarios:

1. **Mise sync operations** (in `mise_sync.rs`):
   - "Razdfile.yml has no mise config, but mise.toml exists. Sync mise.toml → Razdfile.yml? [Y/n]"
   - "mise.toml will be overwritten. Overwrite WITHOUT backup? [Y/n]"
   - "Razdfile.yml does not exist. Create it with mise config? [Y/n]"
   - "Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"
   - Three-option conflict resolution (Options 1-3)

2. **Up command** (in `up.rs`):
   - "Would you like to create a Razdfile.yml? [y/N]"

3. **Internal auto_approve mechanism**:
   - `SyncConfig` already has `auto_approve: bool` field
   - Used internally but not exposed via CLI
   - Currently always set to `false` in user-facing operations

## Proposed Solution

Add a global `-y, --yes` flag that automatically answers "yes" to all prompts, enabling:
- Unattended execution in automation scripts
- CI/CD pipeline integration without manual intervention
- Faster workflows for users who trust default choices

The flag will:
- Set `SyncConfig.auto_approve = true` for mise sync operations
- Auto-approve Razdfile.yml creation in `up` command
- Use sensible defaults for conflict resolution (prefer Razdfile.yml over mise.toml)
- Skip all user input prompts

## Success Criteria

1. `razd --yes up` runs without any user prompts
2. `razd -y list` and other commands work with short form flag
3. All existing prompts respect the yes flag
4. Default behavior (without flag) remains unchanged
5. Documentation updated to explain the flag's behavior

## Scope

**In Scope:**
- Add `-y, --yes` global CLI flag
- Update `SyncConfig` initialization to respect the flag
- Update `prompt_yes_no()` in `up.rs` to auto-approve when flag set
- Update conflict resolution in `mise_sync.rs` to choose Option 1 (Razdfile priority)
- Update all `prompt_user_approval()` calls to respect auto_approve
- Add tests for auto-approve behavior

**Out of Scope:**
- Interactive prompts in external tools (mise, task, git)
- Granular control over which prompts to auto-approve
- Configuration file option for default yes behavior
- Dry-run mode or preview of what would be approved
