# tool-integration Specification Deltas

## MODIFIED Requirements

### Requirement: Taskfile workflow execution with immediate temporary file cleanup
The system MUST create temporary workflow files only for the duration needed to spawn the task process, then immediately clean them up.

#### Scenario: Temporary file deleted immediately after process spawn
**Given** a project has Razdfile.yml with workflow tasks  
**When** user runs `razd dev` or any workflow command  
**Then** the system should:
- Create temporary workflow file in system temp directory with format `razd-workflow-{task_name}.yml`
- Spawn the `task` process with the temporary file as `--taskfile` argument
- Wait a brief delay (e.g., 100ms) to ensure the process has loaded the file
- Delete the temporary file from disk
- Continue monitoring the task process until completion
- Display task output to the user in real-time

#### Scenario: Long-running workflow does not keep temporary file
**Given** a project has a long-running dev server workflow  
**When** user runs `razd dev` which takes 30 minutes or more  
**Then** the system should:
- Create and delete the temporary file within the first second of execution
- Keep the task process running for the full duration
- Not leave temporary files on disk for the entire workflow duration

#### Scenario: Multiple simultaneous workflows don't accumulate temp files
**Given** user runs multiple workflow commands in different terminals  
**When** workflows for dev, build, and test are running simultaneously  
**Then** the system should:
- Create unique temporary files for each workflow (using task name in filename)
- Delete each temporary file immediately after its process spawns
- Not accumulate multiple temporary files in the temp directory

#### Scenario: Temporary file cleanup on workflow failure
**Given** a workflow task fails during execution  
**When** the task process exits with an error  
**Then** the system should:
- Have already deleted the temporary file early in execution
- Not leave orphaned temporary files even on error
- Display the error from the task process

#### Scenario: Process spawn delay ensures file is loaded
**Given** system is under heavy load and process spawn is slow  
**When** razd creates temporary file and spawns task process  
**Then** the system should:
- Wait a configurable delay (default 100ms) after process spawn
- Ensure the task process has time to open and read the file
- Delete the file only after the delay completes
- Avoid race conditions where file is deleted before task reads it
