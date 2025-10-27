# razd

[![CI](https://github.com/razd-cli/razd/workflows/CI/badge.svg)](https://github.com/razd-cli/razd/actions/workflows/ci.yml)
[![Release](https://github.com/razd-cli/razd/workflows/Release/badge.svg)]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> Streamlined project setup with git, mise, and taskfile integration

`razd` (Russian: разд, from "раздуплиться" - to wake up and get yourself together) is a Rust CLI tool that dramatically simplifies project setup across popular technology stacks. It provides one-command project initialization by integrating git, mise, and taskfile.dev.


## Quick Start

### Setting up a project

**From a repository:**
```sh
razd up https://github.com/hello/world.git
```
This clones the repository and runs the setup workflow.

**From an existing project:**
```sh
cd my-existing-project
razd up
```
This detects your project configuration and runs the setup workflow.

**No configuration? No problem:**
```sh
cd empty-project
razd up
```
If no configuration is found, razd will offer to create a `Razdfile.yml` with project-specific templates.

### Default behavior

With no arguments, `razd` defaults to `razd up` behavior:
```sh
razd  # Same as 'razd up'
```

## Requirements

- **[mise](https://mise.jdx.dev/getting-started.html)**: Required for razd installation and project tool management
- **[git](https://git-scm.com/)**: Required for repository operations

## Installation

### Using mise (Recommended)

> **Note:** Requires [mise](https://mise.jdx.dev/getting-started.html) to be installed first.

Install the plugin:

```bash
mise plugin install razd https://github.com/razd-cli/vfox-plugin-razd
```

Set global version:

```bash
mise use -g razd
```

Or install and use a specific version:

```bash
mise use -g razd@0.1.0
```

### Build from Source

```sh
git clone https://github.com/razd-cli/razd.git
cd razd
cargo install --path .
```

### Verify Installation

```sh
razd --version
```

## Commands

### Primary Command
```sh
# Smart project setup - automatically detects context
razd                                            # Default to 'razd up' behavior
razd up                                         # Setup current directory (local mode)
razd up https://github.com/hello/world.git      # Clone and setup from URL
razd up https://github.com/hello/world.git -n my-project  # Custom directory name
```

### Individual Commands
```sh
# Install development tools via mise
razd install       
                 
# Execute tasks from Taskfile.yml
razd task <name> [args...]          # Execute specific task
razd task                           # Run default task

# Development workflows
razd dev                           # Start development server
razd build                         # Build project
razd setup                         # Run project setup only
```

## Prerequisites

- **git**: Required for repository operations
- **mise**: Required for tool management (optional if project doesn't use mise)
- **task**: Required for task execution (optional if project doesn't use taskfile)

## How it Works

### Smart Context Detection

razd intelligently detects your project's context:

1. **Existing Configuration**: Looks for `Razdfile.yml`, `Taskfile.yml`, or `mise.toml`
2. **Project Type Detection**: Analyzes files like `package.json`, `Cargo.toml`, `requirements.txt`
3. **Interactive Setup**: Offers to create configuration if none found

### Workflow Execution

1. **Clone** (if URL provided): Uses git to clone the repository
2. **Tool Setup**: Detects mise configuration and runs `mise install`
3. **Project Setup**: Runs the `default` task from Razdfile.yml or Taskfile.yml
4. **Ready**: Project is ready for development

## Configuration

### Razdfile.yml

razd uses `Razdfile.yml` for project workflows. The key change from traditional taskfile usage is the `default` task:

```yaml
version: '3'

tasks:
  default:                    # ← This task runs with 'razd up'
    desc: "Set up and start project"
    cmds:
      - mise install
      - npm install
      - npm run dev

  dev:
    desc: "Start development server"
    cmds:
      - npm run dev

  build:
    desc: "Build project"
    cmds:
      - npm run build
```

### Project Type Templates

When creating a new `Razdfile.yml`, razd provides templates based on detected project type:

- **Node.js**: Detected by `package.json`
- **Rust**: Detected by `Cargo.toml`
- **Python**: Detected by `requirements.txt` or `pyproject.toml`
- **Go**: Detected by `go.mod`
- **Generic**: Fallback template

## Features

- ✅ **Cross-platform**: Works on Windows, macOS, and Linux
- ✅ **Smart detection**: Automatically detects mise and taskfile configurations
- ✅ **Clear feedback**: Colored output with progress indicators
- ✅ **Error handling**: Helpful error messages with installation guidance
- ✅ **Non-intrusive**: Works alongside existing tools and workflows

## Contributing

We welcome contributions! Please see our [Contributing Guide](.github/CONTRIBUTING.md) for details on:

- Development setup and workflow
- Code quality standards and CI pipeline
- Pull request process and testing requirements
- Release procedures and security measures

### Quick Development Setup
```sh
git clone https://github.com/razd-cli/razd.git
cd razd
cargo build
cargo test
```

For bug reports and feature requests, please use our [issue templates](.github/ISSUE_TEMPLATE/).

## License

MIT
