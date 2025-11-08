# CLI Interface Specification - Yes Flag

## ADDED Requirements

### Requirement: Global Yes Flag

The CLI SHALL provide a global `--yes` flag (with short form `-y`) that automatically answers "yes" to all interactive prompts.

**Rationale:** Enable unattended execution in automation scripts, CI/CD pipelines, and batch operations without manual intervention.

**Acceptance Criteria:**
- Flag is available globally across all commands
- Short form `-y` works identically to `--yes`
- Flag appears in `razd --help` output
- Default behavior (no flag) maintains interactive prompts

#### Scenario: User runs command with --yes flag

**Given:**
- User has razd installed
- A project directory exists without Razdfile.yml

**When:**
- User runs `razd --yes up`

**Then:**
- All prompts are automatically approved with "yes"
- Razdfile.yml is created without asking
- Mise sync operations proceed without confirmation
- Command completes without requiring user input
- Exit code is 0 (success)

#### Scenario: User runs command with -y short form

**Given:**
- User has razd installed
- A project with mise.toml but no Razdfile.yml

**When:**
- User runs `razd -y up`

**Then:**
- Behavior is identical to `--yes`
- All prompts auto-approved
- mise.toml synced to Razdfile.yml automatically
- No user input required

#### Scenario: User runs command without yes flag

**Given:**
- User has razd installed
- No Razdfile.yml exists

**When:**
- User runs `razd up` (without --yes)

**Then:**
- Interactive prompts are displayed
- User must manually answer [Y/n] questions
- Behavior unchanged from current implementation
- Backward compatibility maintained

#### Scenario: Yes flag with list command

**Given:**
- User has razd installed
- Valid Razdfile.yml exists

**When:**
- User runs `razd --yes list`

**Then:**
- Tasks are listed normally
- No prompts occur (list doesn't have prompts anyway)
- Flag is accepted but has no effect
- Exit code is 0

#### Scenario: Help text displays yes flag

**Given:**
- User has razd installed

**When:**
- User runs `razd --help`

**Then:**
- Output includes `-y, --yes` in global options
- Help text explains: "Automatically answer 'yes' to all prompts"
- Flag is documented before command-specific options

## MODIFIED Requirements

None - this is a new feature addition.

## REMOVED Requirements

None - backward compatibility maintained.
