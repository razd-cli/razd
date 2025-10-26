# Add Rust CLI Foundation

## Summary

Implement `razd` (кир. разд, from "раздуплиться" - to get things sorted) as a Rust CLI tool that dramatically simplifies project setup across popular technology stacks. The tool integrates git, mise, and taskfile.dev to provide one-command project initialization and streamlined development workflows on both Windows and Unix systems.

## Motivation

Setting up development projects typically requires multiple manual steps:
1. `git clone <url>` - Clone the repository  
2. `mise install` - Install required development tools
3. `task setup` - Install project dependencies
4. Navigate to various task runners for development

The `razd` CLI eliminates this friction by providing a single command (`razd up <url>`) that handles the entire project setup flow, plus additional commands for ongoing development tasks. This is especially valuable for:

- **Cross-platform consistency**: Works identically on Windows and Unix systems
- **Popular tech stacks**: Python, JavaScript, PHP, and other common development environments
- **Team onboarding**: New developers can get productive in minutes, not hours
- **Context switching**: Unified interface across different project types

## Proposed Changes

### Core CLI Commands

- `razd up <git-url>` - **Primary command**: Clone repository + mise install + task setup (full project initialization)
- `razd install` - Install development tools via mise (equivalent to `mise install`)
- `razd setup` - Install project dependencies (equivalent to `task setup`) 
- `razd task [name] [args...]` - Execute tasks from Taskfile.yml (with fallback to dev server if no task specified)
- `razd init` - Initialize razd configurations for a project

### Key Workflow
```bash
# Traditional approach (3+ commands):
git clone https://github.com/hello/world.git
cd world
mise install
task setup

# razd approach (1 command):
razd up https://github.com/hello/world.git
```

### Technical Implementation

- **Language**: Rust for performance and reliability
- **CLI Framework**: clap for argument parsing
- **Git Integration**: git2-rs or direct git command execution
- **Process Management**: tokio::process for running external commands
- **Configuration**: TOML or YAML for project configuration

### Integration Points

1. **Git**: Repository cloning and basic git operations
2. **Mise**: Tool installation and environment management
3. **Taskfile**: Task discovery and execution

## Impact Assessment

### Benefits
- Simplified developer onboarding
- Consistent project setup across teams
- Reduced context switching between tools
- Standardized development workflows

### Risks
- Additional dependency to maintain
- Learning curve for teams not familiar with razd
- Potential conflicts with existing tooling

### Migration Strategy
- Can be adopted incrementally alongside existing tools
- No breaking changes to existing workflows
- Optional tool that enhances rather than replaces

## Implementation Approach

1. Create Rust project structure with Cargo.toml
2. Implement core CLI argument parsing
3. Add git integration for repository cloning
4. Integrate with mise for tool management
5. Add taskfile support for task execution
6. Implement configuration management
7. Add comprehensive testing and documentation

## Success Criteria

- All specified CLI commands work correctly
- Successful integration with git, mise, and taskfile
- Comprehensive test coverage (>80%)
- Clear documentation and usage examples
- Performance benchmarks showing reasonable execution times