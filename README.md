# razd

> Streamlined project setup with git, mise, and taskfile integration

`razd` (Russian: разд, from "раздуплиться" - to wake up and get yourself together) is a Rust CLI tool that dramatically simplifies project setup across popular technology stacks. It provides one-command project initialization by integrating git, mise, and taskfile.dev.


## Quick Start

Instead of running multiple commands:
```sh
git clone https://github.com/hello/world.git
cd world
mise install
task setup
```

Just run:
```sh
razd up https://github.com/hello/world.git
```

## Installation

```sh
# Build from source
git clone https://github.com/razd-cli/razd.git
cd razd
cargo install --path .
```

## Commands

### Primary Command
```sh
# Clone repository and set up project (git clone + mise install + task setup)
razd up https://github.com/hello/world.git      # Full URL
razd up https://github.com/hello/world.git -n my-project  # Custom directory name
```

### Individual Commands
```sh
# Install development tools via mise
razd install                        

# Install project dependencies via task setup
razd setup 

# Execute tasks from Taskfile.yml
razd task <name> [args...]          # Execute specific task
razd task                           # Start development server (default task)

# Initialize razd configuration
razd init                           
```

## Prerequisites

- **git**: Required for repository operations
- **mise**: Required for tool management (optional if project doesn't use mise)
- **task**: Required for task execution (optional if project doesn't use taskfile)

## How it Works

1. **Clone**: Uses git to clone the repository
2. **Tool Setup**: Detects `.mise.toml` or `.tool-versions` and runs `mise install`
3. **Project Setup**: Detects `Taskfile.yml` and runs `task setup`
4. **Ready**: Project is ready for development

## Features

- ✅ **Cross-platform**: Works on Windows, macOS, and Linux
- ✅ **Smart detection**: Automatically detects mise and taskfile configurations
- ✅ **Clear feedback**: Colored output with progress indicators
- ✅ **Error handling**: Helpful error messages with installation guidance
- ✅ **Non-intrusive**: Works alongside existing tools and workflows

## License

MIT
