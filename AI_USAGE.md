# razd - AI Agent Context File
# Version: 0.4.0
# Last Updated: 2025-11-03

## What is razd?

**razd** (раздуплиться - "to get things sorted" in Russian) is a Rust CLI tool that simplifies project setup to **one command**.

**Problem**: Setting up projects requires multiple steps:
```bash
git clone https://github.com/user/repo
cd repo
mise install        # install language runtimes
task setup          # install dependencies
```

**Solution**: razd does it all:
```bash
razd up https://github.com/user/repo
```

## Core Concept

razd is a **smart orchestrator** that:
1. Clones git repositories (or works locally)
2. Installs development tools via **mise** (runtime manager)
3. Runs project setup via **taskfile** (task runner)
4. Manages configuration automatically

**Key insight**: razd doesn't reinvent wheels - it connects existing tools (git, mise, task) into a seamless workflow.

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
