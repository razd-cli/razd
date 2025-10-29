# Design: Invert Backup Prompt

## Overview
This change inverts the backup prompt logic to make backup creation the default behavior, requiring explicit user opt-out to skip backups.

## Current Implementation

### Current Flow
```rust
if self.config.create_backups {
    if !self.config.auto_approve {
        println!("⚠️  Razdfile.yml will be modified. Create backup? [Y/n]");
        if self.prompt_user_approval()? {
            self.create_backup(&razdfile_path)?;
        }
    } else {
        self.create_backup(&razdfile_path)?;
    }
}
```

### Current Behavior
- Prompt: "Create backup? [Y/n]"
- Y or empty → create backup
- n or N → skip backup

## Proposed Implementation

### New Flow
```rust
if self.config.create_backups {
    if !self.config.auto_approve {
        println!("⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]");
        if !self.prompt_user_approval()? {  // Inverted logic
            self.create_backup(&razdfile_path)?;
        }
    } else {
        self.create_backup(&razdfile_path)?;
    }
}
```

### New Behavior
- Prompt: "Modify WITHOUT backup? [Y/n]"
- Y → skip backup (explicit opt-out)
- n, N, or empty → create backup (safe default)

## Implementation Strategy

### Files to Modify
1. **`src/config/mise_sync.rs`**: Two prompt sites
   - Line ~122: mise.toml overwrite prompt
   - Line ~185: Razdfile.yml modification prompt

### Changes Required
1. Update prompt message text from "Create backup?" to "Modify WITHOUT backup?" / "Overwrite WITHOUT backup?"
2. Invert the logic: negate the result of `prompt_user_approval()?` before the conditional
3. Keep all other behavior (auto-approve, config flags) unchanged

### Testing Considerations
- Verify Y → skips backup
- Verify n → creates backup  
- Verify empty input (Enter) → creates backup
- Verify auto-approve still creates backups
- Verify no-sync flag behavior unchanged

## User Impact
Users will need to adapt to the new prompt wording. However, the change is self-documenting through the explicit question phrasing.

## Rationale
This design:
- **Improves safety**: Default behavior (pressing Enter) is now the safe option
- **Minimal code change**: Only requires message updates and logic inversion
- **Preserves control**: Users can still opt out by typing Y
- **Clear intent**: The prompt explicitly states what will happen if user types Y
