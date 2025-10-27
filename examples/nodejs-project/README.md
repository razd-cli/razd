# Node.js Project Example

This example demonstrates how to use `razd` with a Node.js project, including mise tool management configuration.

## Features Demonstrated

### Mise Tool Management in Razdfile.yml

The `Razdfile.yml` includes mise configuration for managing development tools:

```yaml
mise:
  tools:
    node:
      version: "22"
      postinstall: "corepack enable"  # Enables corepack after Node.js installation
    python: "3.11"
    rust: "latest"
  plugins:
    node: "https://github.com/asdf-vm/asdf-nodejs.git"
```

**Benefits:**
- Single configuration file for both tasks and tools
- Automatic mise.toml generation
- Team-wide tool version consistency

### Task Definitions

The example includes common Node.js project tasks:

- **default**: Complete project setup and dev server start
- **install**: Install dependencies (mise + npm)
- **dev**: Start development server
- **build**: Build production bundle
- **test**: Run test suite

## Usage

From this directory:

```bash
# Complete setup and start dev server
razd up

# Or run specific tasks
razd task install
razd task dev
razd task build
razd task test
```

## How It Works

1. **razd** reads `Razdfile.yml`
2. Detects `mise` section
3. Automatically generates `mise.toml` (if needed)
4. Runs `mise install` to install tools
5. Executes the specified task

## Generated mise.toml

When razd processes this Razdfile, it generates:

```toml
[tools]
python = "3.11"
node = { version = "22", postinstall = "corepack enable" }
rust = "latest"

[plugins]
node = "https://github.com/asdf-vm/asdf-nodejs.git"
```

This file is automatically kept in sync with your Razdfile.yml.
