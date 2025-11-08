# Interactive Prompts Specification - Auto-Approval

## ADDED Requirements

### Requirement: Auto-approve mise sync prompts

When `--yes` flag is active, all mise synchronization prompts SHALL be automatically approved without user input.

**Rationale:** Mise sync is the primary source of interactive prompts and must work unattended for CI/CD workflows.

**Acceptance Criteria:**
- `SyncConfig.auto_approve` set to `true` when flag present
- All `prompt_user_approval()` calls bypassed
- Backup creation proceeds automatically
- Conflict resolution uses default choice (Razdfile priority)

#### Scenario: Auto-approve sync from Razdfile to mise.toml

**Given:**
- Razdfile.yml exists with mise configuration
- mise.toml exists with different content
- User runs with `--yes` flag

**When:**
- Sync process detects conflict

**Then:**
- No prompt "mise.toml will be overwritten" shown
- Backup created automatically (if create_backups=true)
- Razdfile.yml content synced to mise.toml
- Process completes without user input

#### Scenario: Auto-approve sync from mise.toml to Razdfile

**Given:**
- mise.toml exists
- Razdfile.yml has no mise config
- User runs with `--yes` flag

**When:**
- Sync process detects missing mise config in Razdfile

**Then:**
- No prompt "Sync mise.toml → Razdfile.yml?" shown
- Razdfile.yml updated automatically
- Process completes successfully

#### Scenario: Auto-approve Razdfile creation

**Given:**
- mise.toml exists
- No Razdfile.yml exists
- User runs with `--yes` flag

**When:**
- Sync process needs to create Razdfile.yml

**Then:**
- No prompt "Create it with mise config?" shown
- Razdfile.yml created with minimal structure
- mise config imported automatically

### Requirement: Auto-approve conflict resolution

When `--yes` flag is active and sync conflict is detected, the system SHALL automatically choose Option 1 (Razdfile.yml priority).

**Rationale:** Razdfile.yml is the canonical source of truth in razd, so it should take priority in automated scenarios.

**Acceptance Criteria:**
- Three-option conflict prompt is bypassed
- Option 1 (Use Razdfile.yml) selected automatically
- Info message logged about automatic choice
- No stdin reading occurs

#### Scenario: Conflict resolved automatically

**Given:**
- Both Razdfile.yml and mise.toml exist
- Both have different mise configurations
- Semantic conflict detected
- User runs with `--yes` flag

**When:**
- `handle_sync_conflict()` is called

**Then:**
- No prompt with Options 1-3 shown
- Option 1 automatically selected
- Razdfile.yml synced to mise.toml
- Info message: "Auto-approved: Using Razdfile.yml (overwriting mise.toml)"

### Requirement: Auto-approve Razdfile creation in up command

When `--yes` flag is active, the system SHALL automatically approve Razdfile.yml creation during `razd up`.

**Rationale:** The up command workflow should run fully unattended when yes flag is set.

**Acceptance Criteria:**
- `prompt_yes_no()` returns true when RAZD_AUTO_YES=1
- No stdin reading occurs
- Razdfile.yml created with default template
- Up workflow continues execution

#### Scenario: Up command creates Razdfile automatically

**Given:**
- Project directory has no configuration files
- User runs `razd --yes up`

**When:**
- Up command detects no project configuration

**Then:**
- No prompt "Would you like to create a Razdfile.yml?" shown
- Razdfile.yml created with default structure
- Up workflow executes automatically
- Success message displayed

#### Scenario: Environment variable controls auto-approval

**Given:**
- RAZD_AUTO_YES environment variable set to "1"

**When:**
- `prompt_yes_no()` is called

**Then:**
- Function returns `true` immediately
- No output to stdout
- No reading from stdin
- Execution continues without pause

## MODIFIED Requirements

### Requirement: prompt_yes_no SHALL respect environment variable

The `prompt_yes_no()` function SHALL check RAZD_AUTO_YES environment variable before prompting user.

**Before:** `prompt_yes_no()` always prompts user for input.

**After:** 
- If RAZD_AUTO_YES is "1": return true immediately without prompting
- Otherwise: show prompt and read input as before

**Migration:** Existing behavior preserved when environment variable not set.

#### Scenario: Backward compatibility maintained

**Given:**
- RAZD_AUTO_YES not set (or set to "0")

**When:**
- `prompt_yes_no()` is called

**Then:**
- Prompt displayed normally
- User input required
- Behavior identical to previous version

## REMOVED Requirements

None - all existing prompt functionality preserved.
