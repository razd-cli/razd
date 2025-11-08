# cli-interface Spec Delta

## ADDED Requirements

### Requirement: CLI Argument Forwarding
The `razd run` command MUST support forwarding CLI arguments to tasks using the `--` separator, making them available via the `CLI_ARGS` variable.

#### Scenario: Forward arguments to task command
**Given** a user has a Razdfile.yml with a task that uses `{{.CLI_ARGS}}`  
```yaml
tasks:
  test:
    cmds:
      - go test {{.CLI_ARGS}}
```
**When** they run `razd run test -- -v -race`  
**Then** the system should:
- Capture arguments after `--` separator: `-v -race`
- Inject `CLI_ARGS` variable with value `-v -race` into the task execution
- Execute the command as `go test -v -race`

#### Scenario: Run task without CLI arguments
**Given** a user has a Razdfile.yml with a task that uses `{{.CLI_ARGS}}`  
```yaml
tasks:
  hello:
    cmds:
      - echo "Hello {{.CLI_ARGS}}"
```
**When** they run `razd run hello` without any arguments after `--`  
**Then** the system should:
- Set `CLI_ARGS` to an empty string
- Execute the command as `echo "Hello "`

#### Scenario: Pass arguments with spaces and special characters
**Given** a user has a Razdfile.yml with a task  
```yaml
tasks:
  docker-run:
    cmds:
      - docker run {{.CLI_ARGS}}
```
**When** they run `razd run docker-run -- -e "VAR=value with spaces" --name=myapp`  
**Then** the system should:
- Preserve argument formatting and spacing
- Execute the command with all arguments properly passed to docker

#### Scenario: CLI_ARGS variable is available in task vars
**Given** a user wants to reference CLI_ARGS in task variables  
```yaml
tasks:
  build:
    vars:
      FLAGS: "{{.CLI_ARGS}}"
    cmds:
      - echo "Building with flags: {{.FLAGS}}"
```
**When** they run `razd run build -- --release`  
**Then** the CLI_ARGS variable should be available for interpolation in vars and cmds
