# file-sync Capability Specification

## Purpose
Enable razd to track, detect, and synchronize changes between Razdfile.yml and mise.toml, preventing configuration drift while respecting manual edits to either file.

## Requirements

### ADDED Requirement: Track file modification times
The system SHALL track modification times of Razdfile.yml and mise.toml to detect changes.

#### Scenario: Store initial file metadata
**Given** razd processes a project for the first time  
**And** both Razdfile.yml and mise.toml exist  
**When** razd completes configuration processing  
**Then** the system should:
- Record Razdfile.yml modification time
- Record mise.toml modification time
- Record current time as last sync time
- Store metadata in user data directory with project-specific path

#### Scenario: Update metadata after sync
**Given** razd has regenerated mise.toml from Razdfile.yml  
**When** mise.toml write completes successfully  
**Then** the system should:
- Update mise.toml modification time in metadata
- Update last sync time to current time
- Persist metadata to disk

#### Scenario: Handle missing tracking metadata
**Given** a project exists but has no tracking metadata  
**When** razd checks for file changes  
**Then** the system should:
- Treat as first-time processing
- Create new tracking metadata
- Assume no prior sync has occurred

### ADDED Requirement: Detect Razdfile.yml changes
The system SHALL detect when Razdfile.yml has been modified since last sync.

#### Scenario: Detect Razdfile modification
**Given** tracking metadata shows Razdfile.yml last modified at T1  
**And** current Razdfile.yml modification time is T2 where T2 > T1  
**When** razd checks for changes on command start  
**Then** the system should:
- Identify Razdfile.yml as modified
- Trigger mise.toml regeneration
- Not prompt user for confirmation

#### Scenario: No Razdfile changes detected
**Given** tracking metadata shows Razdfile.yml last modified at T1  
**And** current Razdfile.yml modification time equals T1  
**When** razd checks for changes  
**Then** the system should:
- Mark Razdfile.yml as unchanged
- Skip mise.toml regeneration
- Continue with minimal overhead

### ADDED Requirement: Detect mise.toml changes
The system SHALL detect when mise.toml has been manually modified since last sync.

#### Scenario: Detect mise.toml manual edit
**Given** tracking metadata shows mise.toml last modified at T1  
**And** current mise.toml modification time is T2 where T2 > T1  
**And** Razdfile.yml has not changed  
**When** razd checks for changes on command start  
**Then** the system should:
- Identify mise.toml as manually modified
- Display warning about manual modification
- Prompt user to sync changes back to Razdfile.yml

#### Scenario: Mise.toml unchanged
**Given** tracking metadata shows mise.toml last modified at T1  
**And** current mise.toml modification time equals T1  
**When** razd checks for changes  
**Then** the system should:
- Mark mise.toml as unchanged
- Skip sync prompts
- Continue command execution

### ADDED Requirement: Regenerate mise.toml from Razdfile changes
The system SHALL automatically regenerate mise.toml when Razdfile.yml mise configuration changes.

#### Scenario: Regenerate on Razdfile change
**Given** Razdfile.yml has been modified  
**And** Razdfile.yml contains valid mise configuration  
**When** razd detects the change  
**Then** the system should:
- Parse mise configuration from Razdfile.yml
- Generate new mise.toml content
- Backup existing mise.toml to mise.toml.backup
- Write new mise.toml
- Update tracking metadata
- Display success message

#### Scenario: Handle regeneration errors
**Given** Razdfile.yml has been modified  
**And** Razdfile.yml contains invalid mise configuration  
**When** razd attempts regeneration  
**Then** the system should:
- Display parse error with line numbers
- Not modify existing mise.toml
- Not update tracking metadata
- Fail gracefully with actionable error message

### ADDED Requirement: Prompt for mise.toml to Razdfile sync
The system SHALL prompt users to sync manual mise.toml edits back to Razdfile.yml.

#### Scenario: User accepts sync from mise.toml
**Given** mise.toml has been manually modified  
**And** razd detects the change  
**When** user confirms sync prompt with 'y'  
**Then** the system should:
- Parse mise.toml to extract tools and plugins
- Update Razdfile.yml mise section
- Preserve other Razdfile.yml sections (version, tasks)
- Update tracking metadata
- Display success message
- Continue command execution

#### Scenario: User declines sync
**Given** mise.toml has been manually modified  
**And** razd displays sync prompt  
**When** user declines with 'n' or dismisses  
**Then** the system should:
- Not modify Razdfile.yml
- Update tracking metadata to mark prompt as shown
- Continue command execution
- Not prompt again until mise.toml changes again

#### Scenario: Non-interactive mode skips prompt
**Given** mise.toml has been manually modified  
**And** razd is running in non-interactive mode (CI, scripted)  
**When** sync check occurs  
**Then** the system should:
- Display warning about mise.toml modification
- Not prompt for input
- Continue command execution without sync

### ADDED Requirement: Store tracking metadata in user data directory
The system SHALL store file tracking metadata in a platform-appropriate user data directory.

#### Scenario: Store metadata on Windows
**Given** razd is running on Windows  
**When** storing tracking metadata  
**Then** the system should:
- Use `%LOCALAPPDATA%\razd\file_tracking\` directory
- Create project-specific subdirectory using path hash
- Store metadata as `tracking.json`

#### Scenario: Store metadata on Unix
**Given** razd is running on Unix (Linux, macOS)  
**When** storing tracking metadata  
**Then** the system should:
- Use `~/.local/share/razd/file_tracking/` directory
- Create project-specific subdirectory using path hash
- Store metadata as `tracking.json`

#### Scenario: Handle metadata directory creation
**Given** file tracking directory does not exist  
**When** razd needs to store metadata  
**Then** the system should:
- Create directory recursively
- Set appropriate permissions (user-only on Unix)
- Store metadata file

### ADDED Requirement: Generate project-specific tracking paths
The system SHALL use project path hashing to isolate tracking metadata between projects.

#### Scenario: Generate unique tracking path
**Given** a project at absolute path `/home/user/projects/myapp`  
**When** razd calculates tracking metadata path  
**Then** the system should:
- Hash absolute project path using SHA256
- Use hash as subdirectory name (e.g., `a3f7e9...`)
- Ensure same project always maps to same tracking path

#### Scenario: Different projects use different tracking
**Given** two projects at different paths  
**When** razd processes both projects  
**Then** the system should:
- Generate different hashes for each project path
- Store metadata in separate directories
- Prevent cross-project tracking conflicts

### ADDED Requirement: Handle concurrent razd executions
The system SHALL handle concurrent razd command executions in the same project safely.

#### Scenario: Concurrent reads of tracking metadata
**Given** two razd commands start simultaneously  
**When** both read tracking metadata  
**Then** the system should:
- Allow concurrent reads without locking
- Load consistent metadata state
- Not corrupt tracking file

#### Scenario: Concurrent writes to tracking metadata
**Given** two razd commands attempt to write metadata simultaneously  
**When** both complete sync operations  
**Then** the system should:
- Use atomic file writes (write to temp, then rename)
- Ensure last-write-wins semantics
- Avoid partial writes or corruption

### ADDED Requirement: Backup mise.toml before overwriting
The system SHALL create a backup of mise.toml before regenerating it.

#### Scenario: Create backup before sync
**Given** mise.toml exists and will be overwritten  
**When** razd regenerates mise.toml from Razdfile  
**Then** the system should:
- Copy existing mise.toml to mise.toml.backup
- Overwrite mise.toml.backup if it already exists
- Generate new mise.toml content
- Write new content to mise.toml

#### Scenario: Skip backup for new mise.toml
**Given** mise.toml does not exist  
**And** Razdfile.yml has mise configuration  
**When** razd generates mise.toml for first time  
**Then** the system should:
- Not create a backup file
- Create mise.toml directly
- Update tracking metadata

### ADDED Requirement: Display clear sync messages
The system SHALL display clear, actionable messages during sync operations.

#### Scenario: Inform about automatic regeneration
**Given** Razdfile.yml mise config has changed  
**When** razd regenerates mise.toml  
**Then** the system should display:
- "Detected changes in Razdfile.yml mise configuration"
- "Regenerating mise.toml..."
- "✓ Successfully updated mise.toml"

#### Scenario: Warn about manual mise.toml edit
**Given** mise.toml has been manually modified  
**When** razd detects the change  
**Then** the system should display:
- Warning: "mise.toml was modified manually"
- "Sync changes to Razdfile.yml? [y/N]"
- Clear explanation of what sync will do

#### Scenario: Confirm successful sync from mise.toml
**Given** user accepted sync from mise.toml to Razdfile  
**When** sync completes successfully  
**Then** the system should display:
- "✓ Synced mise.toml changes to Razdfile.yml"
- Brief summary of what was updated (tools, plugins)

### ADDED Requirement: Preserve Razdfile.yml formatting during sync
The system SHALL preserve YAML formatting and comments in Razdfile.yml when syncing from mise.toml.

#### Scenario: Update mise section without affecting other content
**Given** Razdfile.yml contains:
```yaml
version: '3'

# Important comment about tasks
tasks:
  default:
    desc: "Run dev server"
    cmds:
      - npm run dev

mise:
  tools:
    node: "20"
```
**And** mise.toml was manually updated to node 22  
**When** user syncs from mise.toml to Razdfile  
**Then** the system should:
- Update only the mise.tools.node value to "22"
- Preserve task comments and formatting
- Maintain YAML structure
- Keep version and tasks sections unchanged
