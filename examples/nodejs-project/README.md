# Node.js Project Example

This example demonstrates how to use `razd` with a Node.js project, including mise tool management configuration.

## Features Demonstrated

### Mise Tool Management in Razdfile.yml

The `Razdfile.yml` includes mise configuration for managing development tools:

```yaml
mise:
  tools:
    node: "22"

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
razd install
razd dev
razd build

# Or run custom tasks
razd run test 
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
node = "22"
```

### Automatic Synchronization

**razd** keeps `Razdfile.yml` and `mise.toml` in sync automatically:

- **Changes to Razdfile**: When you modify the `mise:` section, `mise.toml` is updated on next `razd` command
- **Changes to mise.toml**: If you edit `mise.toml` directly, changes are synced back to `Razdfile.yml`
- **Conflict Detection**: If both files are modified, razd shows a diff and prompts for resolution
- **Backups**: Automatic `.backup` files are created before any sync operation

**Skip Sync** if needed:
```bash
razd up --no-sync          # Skip synchronization for this command
RAZD_NO_SYNC=1 razd up    # Skip via environment variable
```

This ensures your team always has consistent tool versions without manual synchronization.
