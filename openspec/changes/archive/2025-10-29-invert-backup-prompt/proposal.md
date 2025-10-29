# Proposal: Invert Backup Prompt

## Why
The current backup prompt asks users to opt-in to creating backups, which makes the unsafe option (no backup) too easy to choose accidentally. Users who habitually press Enter or type 'n' skip the backup, risking data loss. Inverting the prompt makes backup creation the safe default behavior.

## What Changes
- Update Razdfile.yml modification prompt from "Create backup? [Y/n]" to "Modify WITHOUT backup? [Y/n]"
- Update mise.toml overwrite prompt from "Create backup? [Y/n]" to "Overwrite WITHOUT backup? [Y/n]"
- Invert the prompt logic so Enter/'n' creates backup and 'Y' skips backup
- No changes to auto-approve or config flags behavior

## Impact
- **Affected specs**: tool-integration (backup prompt behavior)
- **Affected code**: `src/config/mise_sync.rs` (two prompt sites at lines ~122 and ~185)
- **Breaking change**: Users familiar with current prompts must adapt to new behavior
- **User experience**: Safer default behavior for hasty or unfamiliar users
