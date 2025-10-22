# Project Context

## Purpose
`razd` (кир. разд, from "раздуплиться" - to get things sorted) is a Rust CLI tool that dramatically simplifies project setup across popular technology stacks. It provides one-command project initialization by integrating git, mise, and taskfile.dev to streamline development workflows on both Windows and Unix systems.

### Key Goals
- **One-command setup**: `razd up <url>` replaces multiple manual steps (git clone, mise install, task setup)
- **Cross-platform consistency**: Identical behavior on Windows and Unix systems  
- **Popular tech stack support**: Python, JavaScript, PHP, and other common development environments
- **Developer productivity**: Reduce project onboarding from hours to minutes

## Tech Stack
- **Rust**: Core implementation language for performance and cross-platform support
- **clap**: CLI argument parsing and command structure
- **tokio**: Async process execution and I/O operations
- **External tool integration**: git, mise, taskfile.dev

## Project Conventions

### Code Style
- **Rust formatting**: Use `rustfmt` with default settings
- **Naming conventions**: Follow Rust naming conventions (snake_case for functions/variables, PascalCase for types)
- **Documentation**: Comprehensive doc comments for all public APIs
- **Error handling**: Use Result types with custom error enums, no panics in user-facing code

### Architecture Patterns
- **Modular design**: Separate CLI, commands, integrations, and core functionality
- **External tool strategy**: Execute tools as child processes rather than embedding libraries
- **Cross-platform**: Single codebase using Rust std library and tokio
- **Configuration**: Optional configuration files with sensible defaults

### Testing Strategy
- **Unit tests**: Test each module in isolation with mocked dependencies
- **Integration tests**: End-to-end command execution with real tool integration
- **Cross-platform testing**: Automated testing on Windows, macOS, and Linux
- **Coverage target**: >80% test coverage across all modules

### Git Workflow
- **Branch strategy**: Feature branches with PR review process
- **Commit conventions**: Conventional commits format (feat:, fix:, docs:, etc.)
- **Release process**: Semantic versioning with automated releases

## Domain Context

### Development Tool Ecosystem
- **mise**: Modern runtime manager (replacement for asdf) - manages language versions and tools
- **taskfile.dev**: Task runner alternative to Makefiles - handles project-specific commands
- **git**: Version control system - primary source for project repositories

### Target Use Cases
- **Project onboarding**: New team members setting up existing projects
- **Multi-project workflows**: Developers working across different technology stacks
- **CI/CD integration**: Automated project setup in build environments
- **Educational environments**: Quick setup for tutorials and workshops

## Important Constraints

### Technical Constraints
- **No external dependencies**: Tool must work without internet for local operations
- **Minimal system requirements**: Should run on systems with basic development tools
- **Backward compatibility**: Must work with existing mise and taskfile configurations
- **Security**: Never store or cache user credentials, rely on existing git authentication

### Platform Constraints
- **Windows compatibility**: Must work with PowerShell and Windows file systems
- **Unix compatibility**: Must work with bash/zsh and Unix file systems
- **Permission handling**: Respect file permissions and executable requirements

## External Dependencies

### Required External Tools
- **git**: Must be available in system PATH for repository operations
- **mise**: Required for tool management operations (optional if not used in project)
- **task**: Required for taskfile operations (optional if not used in project)

### Optional Integrations
- **SSH agents**: For private repository access
- **Git credential managers**: For HTTPS authentication
- **Docker**: For containerized development environments (future)
