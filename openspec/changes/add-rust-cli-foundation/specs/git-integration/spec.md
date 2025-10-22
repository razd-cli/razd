# Git Integration Specification

## ADDED Requirements

### Requirement: Repository cloning functionality
The system MUST be able to clone git repositories from various sources.

#### Scenario: Clone public HTTPS repository
**Given** a user provides a public HTTPS git URL  
**When** they run `razd up https://github.com/user/repo.git`  
**Then** the system should:
- Clone the repository to `./repo` directory
- Verify the clone was successful
- Change working directory to the cloned repository

#### Scenario: Clone public SSH repository
**Given** a user provides a public SSH git URL and has SSH keys configured  
**When** they run `razd up git@github.com:user/repo.git`  
**Then** the system should clone the repository successfully using SSH authentication

#### Scenario: Clone to existing directory
**Given** a directory with the repository name already exists  
**When** a user runs `razd up https://github.com/user/repo.git`  
**Then** the system should:
- Prompt the user for confirmation to overwrite or choose a different directory name
- Proceed based on user choice

#### Scenario: Clone private repository with authentication
**Given** a user provides a private repository URL  
**When** they run `razd up https://github.com/user/private-repo.git`  
**Then** the system should:
- Attempt to use existing git credentials (credential manager, SSH keys)
- Prompt for authentication if needed
- Clone successfully if credentials are valid

### Requirement: Git error handling
The system MUST handle git-related errors gracefully.

#### Scenario: Network connectivity issues
**Given** the user has no internet connection or git server is unreachable  
**When** they run `razd up https://github.com/user/repo.git`  
**Then** the system should display a clear error message about connectivity issues

#### Scenario: Invalid repository URL
**Given** a user provides a malformed or non-existent repository URL  
**When** they run `razd up https://invalid-url`  
**Then** the system should display an error message and suggest correct URL formats

#### Scenario: Permission denied
**Given** a user tries to clone a private repository without proper credentials  
**When** they run `razd up https://github.com/user/private-repo.git`  
**Then** the system should display an authentication error and suggest credential setup steps

### Requirement: Repository validation
The system MUST validate cloned repositories for razd compatibility.

#### Scenario: Repository has mise configuration
**Given** a cloned repository contains .mise.toml or .tool-versions file  
**When** the clone completes  
**Then** the system should detect mise configuration and proceed with tool installation

#### Scenario: Repository has Taskfile configuration  
**Given** a cloned repository contains Taskfile.yml or Taskfile.yaml  
**When** the clone completes  
**Then** the system should detect taskfile configuration and proceed with setup tasks

#### Scenario: Repository lacks configuration files
**Given** a cloned repository has no mise or taskfile configuration  
**When** the clone completes  
**Then** the system should:
- Display a warning about missing configuration
- Skip the corresponding setup steps
- Suggest adding configuration files