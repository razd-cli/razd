# Specification: CLI Interface

## ADDED Requirements

### Requirement: CLI Entry Point

The CLI SHALL provide a single binary `razd` as the main entry point.

#### Scenario: Show version
- **GIVEN** razd is installed
- **WHEN** user runs `razd --version`
- **THEN** the version string is printed to stdout
- **AND** exit code is 0

#### Scenario: Show help
- **GIVEN** razd is installed
- **WHEN** user runs `razd --help` or `razd -h`
- **THEN** help text with available commands and options is printed
- **AND** exit code is 0

---

### Requirement: Global Flags

The CLI SHALL support the following global flags available for all commands.

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--version` | | bool | false | Show version |
| `--help` | `-h` | bool | false | Show help |
| `--verbose` | `-v` | bool | false | Enable verbose output |
| `--silent` | `-s` | bool | false | Disable all output |
| `--dir` | `-d` | string | `.` | Working directory |
| `--color` | | bool | auto | Enable colored output |
| `--yes` | `-y` | bool | false | Assume yes to all prompts |
| `--list` | | bool | false | List all available tasks |
| `--no-sync` | | bool | false | Skip Razdfile ↔ mise.toml sync |
| `--taskfile` | `-t` | string | | Path to taskfile/razdfile |
| `--razdfile` | | string | | Path to razdfile (priority over --taskfile) |

#### Scenario: Verbose mode
- **GIVEN** a valid Razdfile exists
- **WHEN** user runs `razd run mytask --verbose`
- **THEN** debug information is printed to stdout

#### Scenario: Silent mode
- **GIVEN** a valid Razdfile exists
- **WHEN** user runs `razd run mytask --silent`
- **THEN** only task output is shown, no razd messages

#### Scenario: Working directory
- **GIVEN** a Razdfile exists in `/path/to/project`
- **WHEN** user runs `razd run mytask -d /path/to/project`
- **THEN** razd reads Razdfile from that directory

#### Scenario: Custom razdfile path
- **GIVEN** a Razdfile exists at `/path/to/custom.yml`
- **WHEN** user runs `razd run mytask --razdfile /path/to/custom.yml`
- **THEN** razd uses that file as configuration

#### Scenario: Global list flag
- **GIVEN** a valid Razdfile exists
- **WHEN** user runs `razd --list`
- **THEN** all tasks are listed (equivalent to `razd list`)

---

### Requirement: Run Command

The CLI SHALL provide a `run` command to execute tasks.

#### Scenario: Run named task
- **GIVEN** a Razdfile with task "build"
- **WHEN** user runs `razd run build`
- **THEN** the "build" task is executed

#### Scenario: Run default task
- **GIVEN** a Razdfile with task "default"
- **WHEN** user runs `razd run` without task name
- **THEN** the "default" task is executed

#### Scenario: Pass CLI arguments
- **GIVEN** a Razdfile with task using `{{.CLI_ARGS}}`
- **WHEN** user runs `razd run mytask -- --flag value`
- **THEN** `--flag value` is available in CLI_ARGS template variable

#### Scenario: Task not found
- **GIVEN** a Razdfile without task "nonexistent"
- **WHEN** user runs `razd run nonexistent`
- **THEN** error message is shown
- **AND** exit code is 200

---

### Requirement: Up Command

The CLI SHALL provide an `up` command for one-command project setup.

#### Scenario: Set up local project
- **GIVEN** current directory has a Razdfile
- **WHEN** user runs `razd up`
- **THEN** trust is verified
- **AND** tools are installed (if mise/devbox configured)
- **AND** "default" task is executed

#### Scenario: Clone and set up remote project
- **GIVEN** a valid git repository URL
- **WHEN** user runs `razd up https://github.com/user/repo`
- **THEN** repository is cloned
- **AND** trust is verified
- **AND** `razd up` is run in cloned directory

#### Scenario: Initialize new Razdfile
- **GIVEN** no Razdfile exists in current directory
- **WHEN** user runs `razd up --init`
- **THEN** project type is detected (Node.js, Rust, Python, Go, Generic)
- **AND** appropriate Razdfile.yml template is created

#### Scenario: Non-interactive mode
- **GIVEN** prompts would normally appear
- **WHEN** user runs `razd up --yes`
- **THEN** all prompts are auto-confirmed

#### Scenario: No configuration found
- **GIVEN** no Razdfile, Taskfile, or mise.toml exists
- **WHEN** user runs `razd up`
- **THEN** user is prompted to create Razdfile.yml
- **OR** error message with hint is shown

---

### Requirement: Setup Command

The CLI SHALL provide a `setup` command to install project dependencies.

#### Scenario: Run setup task
- **GIVEN** Razdfile has "setup" task
- **WHEN** user runs `razd setup`
- **THEN** trust is verified
- **AND** "setup" task is executed

#### Scenario: Fallback to default setup
- **GIVEN** no "setup" task exists
- **WHEN** user runs `razd setup`
- **THEN** project dependencies are installed via detected method

---

### Requirement: List Command

The CLI SHALL provide a `list` command to show available tasks.

#### Scenario: List tasks with descriptions
- **GIVEN** a Razdfile with tasks having `desc` field
- **WHEN** user runs `razd list`
- **THEN** tasks with descriptions are printed in table format

#### Scenario: List all tasks
- **GIVEN** a Razdfile with internal tasks
- **WHEN** user runs `razd list --all`
- **THEN** all tasks including internal ones are printed

#### Scenario: JSON output
- **GIVEN** a Razdfile with tasks
- **WHEN** user runs `razd list --json`
- **THEN** tasks are printed as JSON with structure:
  - `tasks[]`: array of task objects
  - Each task: `name`, `task`, `desc`, `summary`, `aliases`, `location`, `internal`
  - `location`: Razdfile path

#### Scenario: Task location in JSON
- **GIVEN** a Razdfile with tasks
- **WHEN** user runs `razd list --json`
- **THEN** each task includes `location` with `taskfile`, `line`, `column`

---

### Requirement: Trust Command

The CLI SHALL provide a `trust` command to manage project trust status.

#### Scenario: Trust current directory
- **GIVEN** current directory is not trusted
- **WHEN** user runs `razd trust`
- **THEN** directory is added to trusted list
- **AND** mise trust is run if mise config exists

#### Scenario: Show trust status
- **GIVEN** any directory
- **WHEN** user runs `razd trust --show`
- **THEN** trust status is displayed (Trusted/Ignored/Unknown)

#### Scenario: Untrust directory
- **GIVEN** a trusted directory
- **WHEN** user runs `razd trust --untrust`
- **THEN** directory is removed from trusted list

#### Scenario: Ignore directory
- **GIVEN** any directory
- **WHEN** user runs `razd trust --ignore`
- **THEN** directory is added to ignored list
- **AND** user will not be prompted for this project

#### Scenario: Trust all parent directories
- **GIVEN** nested project directories with configs
- **WHEN** user runs `razd trust --all`
- **THEN** all parent directories with config are trusted

#### Scenario: Trust custom path
- **GIVEN** a path argument
- **WHEN** user runs `razd trust /path/to/project`
- **THEN** specified path is trusted instead of current directory

---

### Requirement: Trust Verification

The CLI SHALL verify trust before executing any task.

#### Scenario: Trusted project execution
- **GIVEN** project is in trusted list
- **WHEN** user runs any command
- **THEN** command executes without prompt

#### Scenario: Untrusted project prompt
- **GIVEN** project is not trusted
- **WHEN** user runs any command
- **THEN** user is prompted to trust the project
- **AND** on "yes", project is added to trusted list and command executes
- **AND** on "no", command is aborted

#### Scenario: Ignored project rejection
- **GIVEN** project is in ignored list
- **WHEN** user runs any command
- **THEN** error is shown and command is aborted

#### Scenario: Auto-trust with --yes flag
- **GIVEN** project is not trusted
- **WHEN** user runs command with `--yes` flag
- **THEN** project is automatically trusted and command executes

---

### Requirement: Init Command (via razd up --init)

The CLI SHALL provide initialization via `razd up --init`.

#### Scenario: Create new Razdfile
- **GIVEN** no Razdfile exists in current directory
- **WHEN** user runs `razd up --init`
- **THEN** a new Razdfile.yml is created with template content

#### Scenario: Auto-detect project type
- **GIVEN** project has `package.json`
- **WHEN** user runs `razd up --init`
- **THEN** Node.js template is used

#### Scenario: Prevent overwrite
- **GIVEN** Razdfile.yml already exists
- **WHEN** user runs `razd up --init`
- **THEN** error message is shown
- **AND** exit code is 101

#### Scenario: Project type templates
- **GIVEN** various project files
- **WHEN** user runs `razd up --init`
- **THEN** appropriate template is selected:
  - `package.json` → Node.js
  - `Cargo.toml` → Rust
  - `requirements.txt` or `pyproject.toml` → Python
  - `go.mod` → Go
  - Otherwise → Generic

---

### Requirement: Install Command

The CLI SHALL provide an `install` command to install project tools.

#### Scenario: Install mise tools
- **GIVEN** Razdfile has `mise.tools` configured
- **WHEN** user runs `razd install`
- **THEN** trust is verified
- **AND** `mise install` is executed

#### Scenario: Install devbox packages
- **GIVEN** Razdfile has `devbox.packages` configured
- **WHEN** user runs `razd install`
- **THEN** trust is verified
- **AND** `devbox install` is executed

#### Scenario: Install workflow fallback
- **GIVEN** Razdfile has "install" task
- **WHEN** user runs `razd install`
- **THEN** "install" task is executed instead of mise/devbox directly

---

### Requirement: Dev Command

The CLI SHALL provide a `dev` command as shortcut for development.

#### Scenario: Run dev task
- **GIVEN** Razdfile has "dev" task
- **WHEN** user runs `razd dev`
- **THEN** trust is verified
- **AND** "dev" task is executed

#### Scenario: No dev task
- **GIVEN** Razdfile has no "dev" task
- **WHEN** user runs `razd dev`
- **THEN** error message suggests creating "dev" task or using `razd up --init`
- **AND** exit code is 200

---

### Requirement: Build Command

The CLI SHALL provide a `build` command as shortcut for building.

#### Scenario: Run build task
- **GIVEN** Razdfile has "build" task
- **WHEN** user runs `razd build`
- **THEN** trust is verified
- **AND** "build" task is executed

#### Scenario: No build task
- **GIVEN** Razdfile has no "build" task
- **WHEN** user runs `razd build`
- **THEN** error message suggests creating "build" task
- **AND** exit code is 200

---

### Requirement: Exit Codes

The CLI SHALL return consistent exit codes for different outcomes.

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Command completed successfully |
| 1 | Unknown | Unknown error occurred |
| 100 | NoRazdfile | No Razdfile found |
| 101 | RazdfileExists | Razdfile already exists (for init) |
| 102 | InvalidConfig | Invalid or unparseable Razdfile |
| 103 | ProjectNotTrusted | Project not trusted and user declined |
| 104 | ProjectIgnored | Project is in ignored list |
| 200 | TaskNotFound | Task not found |
| 201 | TaskFailed | Task execution failed |

#### Scenario: Task failure exit code
- **GIVEN** a task that exits with code 42
- **WHEN** user runs `razd run failing-task`
- **THEN** exit code is 201

#### Scenario: Pass-through exit code
- **GIVEN** a task that exits with code 42
- **WHEN** user runs `razd run failing-task --exit-code`
- **THEN** exit code is 42

---

### Requirement: Shell Completion

The CLI SHALL support shell completion generation.

#### Scenario: Generate bash completion
- **GIVEN** razd is installed
- **WHEN** user runs `razd --completion bash`
- **THEN** bash completion script is printed to stdout

#### Scenario: Generate zsh completion
- **GIVEN** razd is installed
- **WHEN** user runs `razd --completion zsh`
- **THEN** zsh completion script is printed to stdout

#### Scenario: Generate fish completion
- **GIVEN** razd is installed
- **WHEN** user runs `razd --completion fish`
- **THEN** fish completion script is printed to stdout

---

### Requirement: Colored Output

The CLI SHALL support colored terminal output.

#### Scenario: Auto-detect color support
- **GIVEN** stdout is a TTY
- **WHEN** user runs any razd command
- **THEN** output includes ANSI color codes

#### Scenario: Disable colors via NO_COLOR
- **GIVEN** `NO_COLOR` environment variable is set
- **WHEN** user runs any razd command
- **THEN** output does not include color codes

#### Scenario: Force colors via FORCE_COLOR
- **GIVEN** `FORCE_COLOR` environment variable is set
- **AND** stdout is not a TTY
- **WHEN** user runs any razd command
- **THEN** output includes ANSI color codes

---

### Requirement: Task Execution Integration

The CLI SHALL integrate with go-task for task execution.

#### Scenario: Execute task via go-task
- **GIVEN** a valid Razdfile with tasks
- **WHEN** user runs `razd run mytask`
- **THEN** go-task executor runs the task
- **AND** task output is streamed to stdout/stderr

#### Scenario: Task variables
- **GIVEN** a Razdfile with task using variables
- **WHEN** user runs `razd run mytask VAR=value`
- **THEN** VAR is available as template variable in task
