# Capability: Project Trust

## Overview

Project trust capability provides a security layer that requires explicit user consent before executing project configurations.

---

## ADDED Requirements

### Requirement: Trust Storage

The system SHALL store trust state in a cache directory outside the project.

#### Scenario: Trust file location on Unix

- Given: razd is running on a Unix system
- When: Trust state is accessed
- Then: The trust file SHALL be at `~/.cache/razd/trusted.json`

#### Scenario: Trust file location on Windows

- Given: razd is running on Windows
- When: Trust state is accessed
- Then: The trust file SHALL be at `%LOCALAPPDATA%\razd\trusted.json`

#### Scenario: Trust file does not exist

- Given: The trust file does not exist
- When: Trust state is loaded
- Then: An empty trust store SHALL be returned without error

---

### Requirement: Trust Command

The system SHALL provide a `razd trust` command to manage project trust.

#### Scenario: Trust current directory

- Given: User is in a project directory
- When: User runs `razd trust`
- Then: The current directory SHALL be added to trusted list
- And: `mise trust` SHALL be executed if mise config exists

#### Scenario: Trust specific path

- Given: User provides a path argument
- When: User runs `razd trust /path/to/project`
- Then: The specified path SHALL be added to trusted list

#### Scenario: Untrust directory

- Given: A directory is in the trusted list
- When: User runs `razd trust --untrust`
- Then: The directory SHALL be removed from trusted and ignored lists

#### Scenario: Show trust status

- Given: User is in a project directory
- When: User runs `razd trust --show`
- Then: The trust status SHALL be displayed (trusted/untrusted/ignored)

#### Scenario: Ignore directory

- Given: User is in a project directory
- When: User runs `razd trust --ignore`
- Then: The directory SHALL be added to ignored list
- And: Future razd commands SHALL fail without prompting

---

### Requirement: Trust Check Before Execution

The system SHALL check trust status before executing project configurations.

#### Scenario: First run in untrusted project

- Given: A project is not in trusted or ignored list
- And: The project has a Razdfile.yml
- When: User runs any razd command (except trust, list, --help, --version)
- Then: The system SHALL prompt for trust confirmation

#### Scenario: Run in trusted project

- Given: A project is in the trusted list
- When: User runs a razd command
- Then: The command SHALL execute without prompting

#### Scenario: Run in ignored project

- Given: A project is in the ignored list
- When: User runs a razd command
- Then: The command SHALL fail with an error message
- And: No prompt SHALL be shown

#### Scenario: Auto-trust with --yes flag

- Given: A project is not trusted
- When: User runs `razd --yes up`
- Then: The project SHALL be automatically trusted
- And: The command SHALL execute without prompting

---

### Requirement: Trust Prompt

The system SHALL provide an interactive prompt for trust decisions.

#### Scenario: User accepts trust

- Given: Trust prompt is shown
- When: User enters "y" or "yes"
- Then: The project SHALL be added to trusted list
- And: Command execution SHALL continue

#### Scenario: User declines trust

- Given: Trust prompt is shown
- When: User enters "n" or "no"
- Then: The project SHALL NOT be added to any list
- And: Command execution SHALL abort with error

#### Scenario: User ignores project

- Given: Trust prompt is shown
- When: User enters "i" or "ignore"
- Then: The project SHALL be added to ignored list
- And: Command execution SHALL abort with error

---

### Requirement: Mise Trust Integration

When razd trusts a project, it SHALL also trigger mise trust.

#### Scenario: Mise config exists

- Given: Project has mise.toml or .mise.toml or Razdfile.yml with mise section
- When: User runs `razd trust`
- Then: `mise trust` SHALL be executed for the project directory

#### Scenario: No mise config

- Given: Project does not have mise configuration
- When: User runs `razd trust`
- Then: Only razd trust SHALL be added
- And: `mise trust` SHALL NOT be executed
