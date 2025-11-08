# cli-interface Spec Delta

## ADDED Requirements

### Requirement: Custom Configuration File Path
The razd CLI MUST support specifying a custom configuration file path via command-line flags.

#### Scenario: Specify custom file with --taskfile flag
**Given** a user has a custom configuration file at `./config/custom.yml`  
**When** they run `razd list --taskfile ./config/custom.yml`  
**Then** the system should:
- Load configuration from `./config/custom.yml` instead of default `Razdfile.yml`
- List tasks from the custom configuration file
- Display appropriate error if custom file does not exist

#### Scenario: Specify custom file with --razdfile flag
**Given** a user has a custom configuration file at `./Taskfile.yml`  
**When** they run `razd list --razdfile ./Taskfile.yml`  
**Then** the system should:
- Load configuration from `./Taskfile.yml`
- List tasks from the specified file
- Behave identically to --taskfile flag

#### Scenario: Use short form -t flag
**Given** a user prefers concise commands  
**When** they run `razd list -t ./custom.yml`  
**Then** the system should load configuration from `./custom.yml` (same as --taskfile)

#### Scenario: Priority when both flags specified
**Given** a user specifies both `--taskfile` and `--razdfile` flags  
**When** they run `razd list --taskfile file1.yml --razdfile file2.yml`  
**Then** the system should:
- Use `file2.yml` (--razdfile takes priority)
- Ignore `file1.yml` (--taskfile)
- Document this priority in help text

#### Scenario: Custom path with run command
**Given** a user has tasks defined in `./config/tasks.yml`  
**When** they run `razd run build --taskfile ./config/tasks.yml`  
**Then** the system should:
- Load tasks from `./config/tasks.yml`
- Execute the `build` task from that file
- Show clear error if task not found in custom file

#### Scenario: Custom path with up command
**Given** a user wants to set up a project with custom config  
**When** they run `razd up --taskfile ./custom-setup.yml`  
**Then** the system should:
- Use `./custom-setup.yml` for project setup
- Run setup tasks defined in custom file
- Initialize tools based on custom configuration

#### Scenario: Absolute path support
**Given** a user provides an absolute path  
**When** they run `razd list --taskfile /home/user/configs/razd.yml` (Unix) or `razd list --taskfile C:\configs\razd.yml` (Windows)  
**Then** the system should:
- Load configuration from the absolute path
- Handle cross-platform path formats correctly
- Display appropriate error if file not found

#### Scenario: Relative path support
**Given** a user provides a relative path  
**When** they run `razd list --taskfile ../shared/config.yml`  
**Then** the system should:
- Resolve path relative to current working directory
- Load configuration from resolved path
- Handle parent directory references correctly

#### Scenario: Custom file not found error
**Given** a user specifies a non-existent file  
**When** they run `razd list --taskfile ./missing.yml`  
**Then** the system should:
- Display clear error message: "Specified configuration file not found: ./missing.yml"
- Exit with non-zero status code
- Not fall back to default Razdfile.yml

#### Scenario: Path with spaces and special characters
**Given** a user has a config file path with spaces  
**When** they run `razd list --taskfile "./my configs/razd file.yml"`  
**Then** the system should:
- Parse the path correctly (with quotes)
- Load configuration from the path with spaces
- Handle special characters in filenames

### Requirement: Global Flag Availability
The custom configuration flags MUST be available globally for all commands that use configuration.

#### Scenario: Global flag works with list command
**Given** a user wants to list tasks from a custom file  
**When** they run `razd list --taskfile custom.yml --list-all`  
**Then** the custom path flag should work alongside command-specific flags

#### Scenario: Global flag works with run command
**Given** a user wants to run a task from a custom file  
**When** they run `razd run build --taskfile custom.yml`  
**Then** the custom path flag should be recognized before the task name

#### Scenario: Global flag position flexibility
**Given** a user places the global flag in different positions  
**When** they run either:
- `razd --taskfile custom.yml list`
- `razd list --taskfile custom.yml`  
**Then** both command forms should work identically

## MODIFIED Requirements

### Requirement: Default Configuration Loading
The razd CLI configuration loading MUST maintain backward compatibility with default behavior.

#### Scenario: No custom path specified uses default
**Given** a user does not specify any custom path flags  
**When** they run `razd list`  
**Then** the system should:
- Load configuration from `./Razdfile.yml` (current directory)
- Behave exactly as before this feature was added
- Show appropriate error if Razdfile.yml not found

#### Scenario: Empty Razdfile.yml with custom path still works
**Given** a user has both `Razdfile.yml` and `custom.yml` in directory  
**When** they run `razd list --taskfile custom.yml`  
**Then** the system should:
- Load from `custom.yml` (not Razdfile.yml)
- Ignore the presence of default Razdfile.yml
- Only read the specified custom file

### Requirement: Error Messages and Help Text
Error messages and help text MUST clearly indicate which configuration file is being used.

#### Scenario: Help text shows custom path flags
**Given** a user wants to learn about custom path flags  
**When** they run `razd --help` or `razd list --help`  
**Then** the help text should:
- Show both `--taskfile` and `--razdfile` options
- Indicate short form `-t` is available
- Explain priority when both flags specified
- Include usage examples

#### Scenario: Error messages include file path
**Given** a task is not found in custom configuration  
**When** they run `razd run nonexistent --taskfile custom.yml`  
**Then** the error message should:
- Include the path to the file that was checked: "Task 'nonexistent' not found in custom.yml"
- Clearly indicate which file was searched
- Suggest using `razd list --taskfile custom.yml` to see available tasks

#### Scenario: Invalid YAML shows file path
**Given** a custom configuration file contains invalid YAML  
**When** they run `razd list --taskfile broken.yml`  
**Then** the error message should:
- Show the path to the broken file: "Failed to parse broken.yml"
- Include YAML parsing error details
- Indicate line number of YAML syntax error if available
