# razd - AI Agent Context File
# Version: 0.4.1
# Last Updated: 2025-11-03

## What is razd?

**razd** (раздуплиться - "to get things sorted" in Russian) is a Rust CLI tool that simplifies project setup to **one command**.

**Problem**: Setting up projects requires multiple manual steps:
```bash
git clone https://github.com/user/repo
cd repo
mise install        # install language runtimes
task setup          # install dependencies
```

**Solution**: razd orchestrates everything:
```bash
razd up https://github.com/user/repo
```

## Core Architecture

razd is a **smart orchestrator** that integrates three external tools:
1. **git** - Repository management
2. **mise** - Runtime and tool version management
3. **taskfile.dev** - Task execution engine

**Key principle**: razd doesn't reinvent functionality - it connects existing tools into a seamless workflow.

---

## Commands

### `razd up [URL]`
Main command - sets up any project:

```bash
# Clone from GitHub and set up
razd up https://github.com/user/repo.git

# Set up current directory
cd my-project
razd up

# Initialize new Razdfile.yml
razd up --init
```

**Workflow**:
1. Clone repository (if URL provided)
2. Parse `Razdfile.yml`
3. Sync to `mise.toml` (if needed)
4. Install tools via `mise install`
5. Execute default task workflow

### `razd run <task>`
Execute custom tasks from Razdfile.yml:

```bash
razd run build      # Run build task
razd run test       # Run test task
razd run deploy     # Run custom deploy task
```

### Other Commands
```bash
razd install        # Install tools only (mise install)
razd setup          # Full setup without starting dev server
razd dev            # Start development environment
razd build          # Build project
```

---

## Configuration: Razdfile.yml

**Single source of truth** - defines tools and tasks. Auto-generates `mise.toml`.

### Minimal Example:
```yaml
mise:
  tools:
    node: "22"
    task: "latest"

tasks:
  default:
    desc: Set up and start project
    cmds:
      - task: install
      - task: dev
      
  install:
    desc: Install dependencies
    cmds:
      - mise install
      - npm install
      
  dev:
    desc: Start dev server
    cmds:
      - npm run dev
```

### Structure:

**`mise` section** - Development tools (synced to mise.toml):
```yaml
mise:
  tools:
    node: "22"           # Simple version
    python: "3.12"
    rust: "1.82.0"
    task: "latest"
  plugins:
    node: "https://github.com/asdf-vm/asdf-nodejs.git"
```

**`tasks` section** - Task definitions:
```yaml
tasks:
  default:
    desc: Main workflow
    cmds:
      - task: install    # Call another task
      - echo "Ready!"    # Shell command
      
  install:
    desc: Install dependencies
    cmds:
      - mise install
      - cargo fetch
      
  build:
    desc: Build project
    deps: [install]      # Run install first
    cmds:
      - cargo build
```

**Task properties**:
- `desc` - Description (shown in help)
- `cmds` - Array of commands (shell or `task:` references)
- `deps` - Array of prerequisite tasks
- `internal` - Hide from `razd run` list (default: false)

---

## Key Features

### 1. One-Command Setup
```bash
# Before
git clone URL
cd project
mise install
npm install
npm run dev

# After
razd up URL
```

### 2. Cross-Platform
Works identically on Windows (PowerShell) and Unix (bash/zsh).

### 3. Smart Config Sync
- **Razdfile.yml** → auto-generates **mise.toml**
- Semantic change detection prevents conflicts
- Creates backups before overwriting
- Warns on manual mise.toml edits

### 4. Flexible Task System
- Define tasks in Razdfile.yml
- Or use existing Taskfile.yml
- Or both (Razdfile tasks + Taskfile workflows)

### 5. Dogfooding Approach
razd itself is developed using razd (see `Razdfile.yml` in project root).

---

## Project Structure (For Code Changes)

```
razd/
├── src/
│   ├── main.rs              # CLI entry point (clap commands)
│   ├── commands/            # Command implementations
│   │   ├── up.rs           # Main setup orchestration
│   │   ├── run.rs          # Custom task execution
│   │   ├── install.rs      # Tool installation
│   │   ├── setup.rs        # Setup without dev server
│   │   ├── dev.rs          # Start dev environment
│   │   └── build.rs        # Build command
│   ├── config/              # Configuration management
│   │   ├── razdfile.rs     # Parse Razdfile.yml
│   │   ├── mise_sync.rs    # Auto-sync logic (Razdfile → mise.toml)
│   │   ├── mise_validator.rs # Validate mise configurations
│   │   ├── file_tracker.rs # Change detection (semantic hashing)
│   │   └── detection.rs    # Detect project configuration
│   ├── integrations/        # External tool wrappers
│   │   ├── git.rs          # Git clone/operations
│   │   ├── mise.rs         # Mise install/execution
│   │   ├── taskfile.rs     # Taskfile task execution
│   │   └── process.rs      # Cross-platform process spawning
│   └── core/
│       ├── error.rs        # RazdError types
│       └── output.rs       # Colored terminal output
├── tests/                   # Integration tests
│   ├── integration_tests.rs
│   ├── mise_integration_tests.rs
│   └── sync_integration_tests.rs
├── openspec/                # Spec-driven development
│   ├── AGENTS.md           # AI workflow instructions
│   ├── project.md          # Project context
│   ├── specs/              # Capability specifications
│   └── changes/            # Active proposals
├── Razdfile.yml            # Project tasks (dogfooding!)
├── Cargo.toml              # Rust manifest (MSRV: 1.82.0)
└── mise.toml               # Tool versions (generated from Razdfile.yml)
```

---

## Design Principles

### 1. External Tool Execution
**Never embed tools** - always execute as child processes:
```rust
// ✅ Good
Command::new("mise").args(["install"]).spawn()?;

// ❌ Bad
// Don't embed mise/git/task as libraries
```

### 2. Cross-Platform First
Single codebase using Rust std + tokio:
```rust
// ✅ Cross-platform
std::process::Command

// ❌ Platform-specific
#[cfg(windows)] ...
```

### 3. Error Handling
Always use `Result<T, RazdError>`:
```rust
// ✅ Good
pub fn parse_config(path: &Path) -> Result<Config, RazdError> {
    // ...
}

// ❌ Bad - never panic in user-facing code
pub fn parse_config(path: &Path) -> Config {
    fs::read_to_string(path).unwrap()  // ❌
}
```

### 4. Modular Design
Separation of concerns:
- **CLI layer** (`main.rs`, `commands/`) - User interface
- **Config layer** (`config/`) - Configuration parsing/sync
- **Integration layer** (`integrations/`) - External tools
- **Core layer** (`core/`) - Shared utilities

### 5. Semantic Change Detection
Use content hashing (not timestamps) to detect config changes:
```rust
// File tracker uses SHA-256 of normalized content
// Detects actual changes, not just file modification time
```

---

## OpenSpec Workflow (IMPORTANT)

razd uses **spec-driven development**. Before making changes:

### When to Create a Proposal

**Always create proposal for**:
- New features or commands
- Breaking changes (API, config format)
- Architecture changes
- Performance optimizations (changing behavior)
- Security pattern changes

**Skip proposal for**:
- Bug fixes (restoring intended behavior)
- Typos, comments, formatting
- Non-breaking dependency updates
- Configuration tweaks
- Tests for existing behavior

### How to Create a Proposal

1. **Review existing work**:
   ```bash
   openspec list              # Active changes
   openspec list --specs      # Existing specs
   rg "Requirement:|Scenario:" openspec/specs  # Search requirements
   ```

2. **Choose unique change-id**:
   - Format: `kebab-case`, verb-led
   - Examples: `add-docker-support`, `refactor-config-sync`, `remove-deprecated-flag`

3. **Scaffold proposal**:
   ```
   openspec/changes/<change-id>/
   ├── proposal.md           # Overview, motivation, goals
   ├── tasks.md              # Ordered implementation steps
   ├── design.md             # Architecture decisions (if complex)
   └── specs/                # Spec deltas per affected capability
       └── <capability>/
           └── spec.md       # ADDED/MODIFIED/REMOVED requirements
   ```

4. **Write spec deltas**:
   ```markdown
   ## ADDED Requirements
   
   ### Requirement: Docker container support
   razd SHALL support running projects in Docker containers.
   
   #### Scenario: Detect Dockerfile
   GIVEN a project with Dockerfile
   WHEN user runs `razd up`
   THEN razd SHALL offer to start Docker container
   
   #### Scenario: Build and run container
   GIVEN user accepts Docker option
   WHEN razd builds container
   THEN container SHALL have all mise tools installed
   ```

5. **Validate**:
   ```bash
   openspec validate <change-id> --strict
   # Fix all issues before proceeding
   ```

6. **Request approval** - Don't implement until proposal is approved

### Existing Specifications

Current capabilities (run `openspec list --specs`):
- **build-system** - Cargo build integration
- **cli-interface** - Command structure and arguments
- **continuous-integration** - GitHub Actions CI/CD
- **git-integration** - Repository cloning and management
- **release-automation** - Automated releases
- **task-auto-installation** - Auto-install missing tools
- **tool-integration** - Mise and taskfile integration

### Key Files

- `openspec/AGENTS.md` - Full workflow documentation
- `openspec/project.md` - Project context and constraints
- `openspec/specs/<capability>/spec.md` - Capability requirements
- `openspec/changes/<id>/` - Active change proposals

---

## Common Development Tasks

### Running Tests
```bash
razd run test              # All tests
cargo test                 # Same via cargo
cargo test --test integration_tests  # Specific test file
```

### Building
```bash
razd run build             # Debug build
cargo build --release      # Release build
```

### Code Quality
```bash
cargo fmt                  # Format code
cargo clippy -- -D warnings # Lint
cargo audit                # Security check
```

### Local Installation
```bash
cargo install --path .     # Install to ~/.cargo/bin/
razd --version             # Verify
```

---

## Troubleshooting

### "Razdfile.yml not found"
**Solution**: Create one:
```bash
razd up --init            # Interactive wizard
```

Or manually:
```yaml
mise:
  tools:
    node: "22"
tasks:
  default:
    cmds:
      - echo "Hello!"
```

### "mise not installed"
**Solution**: Install mise:
```bash
# Unix
curl https://mise.run | sh

# Windows
winget install jdx.mise

# Or via package manager
brew install mise  # macOS
```

### "Both Razdfile.yml and mise.toml changed"
**Cause**: Semantic conflict detection found both files modified.

**Solution**:
```bash
# Option 1: Let razd sync (creates backup)
razd up

# Option 2: Skip sync this time
razd up --no-sync

# Option 3: Resolve manually, commit mise.toml
git add mise.toml
git commit -m "Manual mise.toml changes"
razd up
```

### "Task not found"
**Solution**: Check available tasks:
```bash
razd run               # List tasks from Razdfile.yml
task --list            # List tasks from Taskfile.yml (if exists)
```

---

## AI Agent Quick Reference

### When User Asks About razd

**Setup/Usage questions** → Explain `razd up` workflow  
**Configuration** → Show Razdfile.yml structure  
**Errors** → Check Troubleshooting section  
**Features** → Check existing specs first (`openspec list --specs`)

### When Making Code Changes

1. ✅ **Check OpenSpec first**:
   ```bash
   openspec list --specs      # What specs exist?
   rg "keyword" openspec/     # Search requirements
   ```

2. ✅ **Create proposal if needed** (see OpenSpec Workflow above)

3. ✅ **Follow existing patterns**:
   - Look at similar command implementations
   - Match error handling style
   - Use existing integration wrappers

4. ✅ **Test cross-platform**:
   - Test on Windows (PowerShell) and Unix (bash)
   - Use `std::process::Command` (cross-platform)
   - Avoid platform-specific code

5. ✅ **Write tests**:
   - Unit tests for new functions
   - Integration tests for command flows
   - Add test cases to `tests/` directory

### Critical Files to Know

| File | Purpose |
|------|---------|
| `src/commands/up.rs` | Main setup orchestration logic |
| `src/config/razdfile.rs` | Razdfile.yml parsing |
| `src/config/mise_sync.rs` | Razdfile → mise.toml sync |
| `src/integrations/mise.rs` | Mise tool execution |
| `src/integrations/taskfile.rs` | Task execution |
| `openspec/AGENTS.md` | Full development workflow |
| `openspec/project.md` | Project constraints and context |

---

## Version History

**Current: 0.4.1** (2025-11-03)
- Removed `razd task` command (use `razd run` instead)
- Streamlined command structure
- Updated Rust MSRV to 1.82.0

**0.4.0**: Fixed version field injection  
**0.3.2**: Universal project template  
**0.3.1**: Optional version field  
**0.3.0**: Semantic change detection

See `CHANGELOG.md` for full history.

---

## Repository Info

- **GitHub**: https://github.com/razd-cli/razd
- **Docs**: https://razd-cli.github.io/docs/
- **Telegram**: https://t.me/razd_cli
- **License**: MIT
- **Language**: Rust 1.82.0+
- **Dependencies**: git, mise, task (external tools)

---

## Commands (The Essentials)

### `razd up [URL]`
**The main command** - sets up any project:

```bash
# Clone from GitHub and set up
razd up https://github.com/user/repo.git

# Set up local project (already cloned)
cd my-project
razd up

# Create new Razdfile.yml for current project
razd up --init
```

**What it does**:
1. Clones repo (if URL provided)
2. Reads `Razdfile.yml`
3. Installs tools (`mise install`)
4. Runs tasks

### `razd run <name>`
Run custom tasks from Razdfile.yml:

```bash
razd run deploy     # custom deployment task if exists
razd run backup     # custom backup task if exists
```

---

## Configuration: Razdfile.yml

**Single source of truth** for project setup. Auto-generates `mise.toml`.

### Minimal Example:
```yaml
mise:
  tools:
    node: "22"
    task: "latest"

tasks:
  default:
    desc: Set up and start project
    cmds:
      - mise install
      - npm install
      - npm run dev
```

### Structure:

**`mise` section** - Development tools:
```yaml
mise:
  tools:
    node: "22"           # simple version
    python: "3.12"
    task: "latest"
  plugins:
    node: "https://github.com/asdf-vm/asdf-nodejs.git"
```

**`tasks` section** - Custom commands:
```yaml
tasks:
  default:
    desc: Set up and start Node.js project
    cmds:
    - task: install
    - task: dev

  install:
    desc: Install dependencies
    cmds:
      - mise install
      - npm install
  dev:
    desc: Start dev server
    cmds:
      - npm run dev

```

**Task properties**:
- `desc` - Description (shown in help)
- `cmds` - Array of shell commands
- `deps` - Array of dependent tasks (run first)
- `internal` - Hide from task list (optional)

---

## Key Features

### 1. One-Command Setup
Replace 4+ commands with 1:
```bash
# Before
git clone URL && cd project && mise install && npm run dev

# After
razd up URL
```

### 2. Cross-Platform
Works identically on Windows (PowerShell) and Unix (bash/zsh).

### 3. Smart Config Sync
- Razdfile.yml → auto-generates mise.toml
- Detects conflicts with semantic hashing
- Prevents accidental overwrites

### 4. Flexible Tasks
- Use Taskfile.yml (standard task runner)
- Or define tasks in Razdfile.yml
- Or both!

### 5. No Installation Hassle
Warns about missing tools and provides install instructions.

---

## Architecture (For Code Changes)

### Project Structure:
```
src/
├── main.rs              # CLI entry point
├── commands/            # Command implementations
│   ├── up.rs           # Main setup command
│   ├── task.rs         # Task execution
│   ├── run.rs          # Custom task execution
│   └── install.rs      # Tool installation
├── config/              # Configuration management
│   ├── razdfile.rs     # Parse Razdfile.yml
│   ├── mise_sync.rs    # Auto-sync logic
│   └── file_tracker.rs # Change detection
├── integrations/        # External tools
│   ├── git.rs          # Git operations
│   ├── mise.rs         # Mise integration
│   └── taskfile.rs     # Taskfile integration
└── core/
    ├── error.rs        # Error types
    └── output.rs       # Colored output
```

### Design Principles:
- **External tool execution**: Run tools as child processes (not embedded libs)
- **Cross-platform**: Single codebase using Rust std + tokio
- **No panics**: Always use `Result<T, RazdError>`
- **Modular**: Separate concerns (CLI, config, integrations)

---

## Common Use Cases

### 1. Project Onboarding
New developer joins team:
```bash
razd up https://github.com/company/project
# ✓ Cloned, tools installed, dependencies ready, dev server running
```

### 2. Working on Multiple Projects
Switch between Python, Node.js, Rust projects:
```bash
cd ~/projects/python-api && razd up
cd ~/projects/react-app && razd up
cd ~/projects/rust-cli && razd up
```

### 3. CI/CD Pipelines
```bash
razd up && razd task test && razd task build
```

### 4. Teaching/Workshops
Share one-liner for students:
```bash
razd up https://github.com/workshop/tutorial
```

---

## Common Patterns

### Pattern 1: Fresh Start
```bash
razd up https://github.com/user/project
# Everything set up automatically
```

### Pattern 2: Resume Work
```bash
cd existing-project
razd up              # Re-sync tools and deps
razd dev        # Start working
```

### Pattern 3: Custom Workflows
```bash
# Morning routine
razd run test       # Run tests
razd run dev        # Start dev server

# Deployment
razd run build      # Build production
razd run deploy      # Custom deploy task
```

### Pattern 4: Avoid Auto-Sync
```bash
razd --no-sync up    # Skip mise.toml sync
```

---

## Troubleshooting

### Issue: "Razdfile.yml not found"
**Solution**: Create one:
```bash
razd up --init       # Interactive creation
```

### Issue: "mise not installed"
**Solution**: Install mise:
```bash
curl https://mise.run | sh    # Unix
# or
winget install jdx.mise    # Windows
```

### Issue: "Both Razdfile.yml and mise.toml changed"
**Solution**: 
- Resolve conflicts manually, OR
- Use `razd --no-sync up` to skip sync

### Issue: "Task not found"
**Solution**: Check task names:
- `razd task` (list Taskfile.yml tasks)
- Check `tasks:` section in Razdfile.yml

---

## OpenSpec Integration

razd uses **spec-driven development**:
- Specifications: `openspec/specs/`
- Active changes: `openspec/changes/`
- Workflow guide: `openspec/AGENTS.md`

**Before making changes**:
1. Read `openspec/AGENTS.md`
2. Check existing specs: `openspec list --specs`
3. Create proposal: `openspec/changes/[change-id]/`
4. Validate: `openspec validate --strict`

---

## AI Agent Quick Reference

### When User Asks About razd:

**Setup questions** → Explain `razd up`
**Configuration** → Show Razdfile.yml example
**Errors** → Check Troubleshooting section

### When Making Code Changes:

1. **Read specs first**: Check `openspec/specs/` for relevant capability
2. **Follow patterns**: Look at existing command implementations
3. **Cross-platform**: Test on Windows and Unix
4. **Error handling**: Use `Result<T, RazdError>`, no panics
5. **Create proposals**: For features, follow OpenSpec workflow

### Key Files to Know:

- `src/commands/up.rs` - Main setup logic
- `src/config/razdfile.rs` - Configuration parsing
- `src/integrations/*.rs` - Tool integrations
- `openspec/AGENTS.md` - Development workflow

---

## Repository Info

- **GitHub**: https://github.com/razd-cli/razd
- **Docs**: https://razd-cli.github.io/docs/
- **Telegram**: https://t.me/razd_cli
- **License**: MIT
- **Language**: Rust (1.74.0+)

---

## Version History

**Current: 0.4.0** (2025-11-03)
- Fixed version field injection in Razdfile.yml
- Improved workflow execution consistency

**0.3.2**: Universal template for all project types
**0.3.1**: Optional `version` field in Razdfile.yml
**0.3.0**: Semantic change detection for sync conflicts

See `CHANGELOG.md` for full history.
