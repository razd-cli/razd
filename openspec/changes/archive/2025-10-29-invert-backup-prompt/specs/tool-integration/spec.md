# tool-integration Spec Delta

## MODIFIED Requirements

### Requirement: File backup prompts during sync operations
The system SHALL prompt users to explicitly opt-out of backup creation, making backup the safe default behavior.

**Previous behavior**: System asked "Create backup? [Y/n]" where Y created backup and Enter/n skipped it.
**New behavior**: System asks "Modify WITHOUT backup? [Y/n]" where Y skips backup and Enter/n creates it.

#### Scenario: User declines to skip backup (creates backup)
**Given** Razdfile.yml exists and needs modification  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"  
**And** user types 'n', 'N', or presses Enter  
**Then** the system should:
- Create a backup file (Razdfile.yml.backup)
- Proceed with the modification
- Display confirmation "✓ Synced mise.toml → Razdfile.yml"

#### Scenario: User explicitly opts out of backup
**Given** Razdfile.yml exists and needs modification  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"  
**And** user types 'Y' or 'y'  
**Then** the system should:
- Skip creating a backup file
- Proceed with the modification
- Display confirmation "✓ Synced mise.toml → Razdfile.yml"

#### Scenario: mise.toml overwrite prompt uses inverted logic
**Given** mise.toml exists and will be overwritten  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  mise.toml will be overwritten. Overwrite WITHOUT backup? [Y/n]"  
**And** user types 'n', 'N', or presses Enter  
**Then** the system should:
- Create a backup file (mise.toml.backup)
- Overwrite mise.toml with new content
- Update tracking metadata
- Display confirmation "✓ Synced Razdfile.yml → mise.toml"

#### Scenario: Auto-approve mode still creates backups
**Given** user has enabled auto-approve mode  
**And** create_backups config is true  
**When** a file needs modification  
**Then** the system should:
- Skip the interactive prompt
- Create backup automatically
- Proceed with modification
- Not require user interaction

#### Scenario: create_backups config disabled
**Given** user has set create_backups config to false  
**When** a file needs modification  
**Then** the system should:
- Not prompt about backups
- Not create backup files
- Proceed directly with modification

#### Scenario: Default behavior is safe (backup creation)
**Given** a user is unfamiliar with the prompt  
**And** Razdfile.yml needs modification  
**When** the user presses Enter without typing anything  
**Then** the system should:
- Interpret empty input as 'n' (decline to skip backup)
- Create a backup file
- Proceed safely with modification

## Implementation Notes

### Affected Prompts
1. **Razdfile.yml modification**: "⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"
2. **mise.toml overwrite**: "⚠️  mise.toml will be overwritten. Overwrite WITHOUT backup? [Y/n]"

### Logic Inversion
The prompt logic must be inverted:
- **Before**: `if prompt_user_approval()? { create_backup() }`
- **After**: `if !prompt_user_approval()? { create_backup() }`

### Backward Compatibility
This is a breaking change in user interaction:
- Users accustomed to pressing 'n' to skip backups must now press 'Y'
- Users accustomed to pressing 'Y' or Enter to create backups can now just press Enter (easier)
- The new prompt wording clearly indicates the new behavior

### Safety Improvement
The inversion makes backup creation the path of least resistance:
- Hasty users who press Enter get the safe behavior
- Deliberate users who want to skip backups must type 'Y'
- This aligns with "secure by default" UX principles
