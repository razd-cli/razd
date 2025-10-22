# razd

[![CI](https://github.com/razd-cli/razd/workflows/CI/badge.svg)](https://github.com/razd-cli/razd/actions/workflows/ci.yml)
[![Release](https://github.com/razd-cli/razd/workflows/Release/badge.svg)](https://github.com/razd-cli/razd/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/razd-cli/razd/branch/main/graph/badge.svg)](https://codecov.io/gh/razd-cli/razd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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

### Pre-built Binaries (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/razd-cli/razd/releases):

**Windows:**
```powershell
# Download and extract to PATH
curl -L -o razd.zip https://github.com/razd-cli/razd/releases/latest/download/razd-v*-x86_64-pc-windows-msvc.zip
# Extract and add to PATH
```

**macOS:**
```sh
# Intel Macs
curl -L https://github.com/razd-cli/razd/releases/latest/download/razd-v*-x86_64-apple-darwin.tar.gz | tar -xz
# Apple Silicon Macs  
curl -L https://github.com/razd-cli/razd/releases/latest/download/razd-v*-aarch64-apple-darwin.tar.gz | tar -xz
# Move to PATH
sudo mv razd /usr/local/bin/
```

**Linux:**
```sh
curl -L https://github.com/razd-cli/razd/releases/latest/download/razd-v*-x86_64-unknown-linux-gnu.tar.gz | tar -xz
sudo mv razd /usr/local/bin/
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
