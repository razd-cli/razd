<div align="center">

<img src="https://razd.dealenx.ru/docs/logo.png" alt="razd logo" width="200"/>

# razd

[![CI](https://github.com/razd-cli/razd/workflows/CI/badge.svg)](https://github.com/razd-cli/razd/actions/workflows/ci.yml)
![Release](https://github.com/razd-cli/razd/workflows/Release/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/razd-cli/razd/blob/main/LICENSE)
[![Context7](https://img.shields.io/badge/Context7-Enabled-blue)](https://context7.com/razd-cli/razd)

Razd is a modern project setup tool that helps you wake up your projects and get things sorted.

It orchestrates [mise](https://github.com/jdx/mise) and [taskfile](https://github.com/go-task/task) for one-command project initialization.

**Automation-friendly**: Use the `--yes` flag to skip all interactive prompts, perfect for CI/CD pipelines and unattended execution.

[Installation](https://razd.dealenx.ru/docs/installation) | [Documentation](https://razd.dealenx.ru/docs/guide/) | [Telegram](https://t.me/razd_cli)

</div>

## Commands

```bash
razd run <task>       # Run a task (default task if none specified)
razd list             # List tasks
razd list --json      # List tasks in JSON
razd up               # Install dependencies and run the default task
razd up --run         # Same as above (flag kept for compatibility)
razd init             # Create Razdfile.yml in current directory

razd init --using devbox   # Create with devbox provider
razd add node@22      # Add a dependency to Razdfile
razd add node         # Add a dependency without a version (defaults to "latest")
razd add task         # Add the 'task' package (bare form)
razd add task hello -- echo 'hi'   # Create a task with a command
razd shell            # Open interactive shell (provisioner activated)
razd shell --print    # Print the activation script (for eval/iex)
razd trust            # Trust current project
razd trust --show     # Show trust status
razd trust --untrust  # Remove trust
```

## Config Synchronization

`razd` keeps the `Razdfile.yml` and the native provisioner config in sync. When
you use the unified `dependencies` section, razd performs a **bidirectional,
non-destructive merge** between the two files on `razd up`, `razd add`, and
`razd init`:

- A tool added to `Razdfile.yml` (`dependencies.ensure`) is merged into the
  native config (`mise.toml` / `devbox.json`).
- A tool present in the native config is merged back into `Razdfile.yml`.
- Sections outside the managed tool list are **preserved as-is** — `[env]`,
  `[settings]`, `[hooks]`, `[plugins]` (mise) and `env`, `shell`, `nixpkgs`
  (devbox) are never dropped.

Versionless devbox packages (e.g. `php84Extensions.*`, `php84Packages.composer`)
are synced as bare entries in `dependencies.ensure` (no `@version` suffix), and
versionless entries in `ensure` are written back to `devbox.json` as bare names.

`razd add <tool>` accepts a bare package name without a version (the native
manager defaults it to "latest"). If the Razdfile has no `dependencies.ensure`
list yet (for example right after `razd init`), `razd add` creates it.

A package added via `razd add` is written to the native config without a
direction prompt when there is no version conflict — the direction is obvious
(Razdfile → native). When a tool exists in both files with **different
versions**, razd shows a single "Version conflict" prompt with **"Use Razdfile"**
as the first (recommended) option; after you choose, the change is applied
without a second prompt.

### Creating tasks

`razd add task <name> -- <cmd>` creates a task in the `tasks:` section of the
Razdfile. Everything after `--` is the command verbatim:

```bash
razd add task hello -- echo 'hi'
razd add task build -- go build ./...
```

A single command with no flags is written in the compact scalar form
(`build: go build ./...`); multiple commands or any flags use the mapping form.

Supported flags (before `--`):

```
--desc <text>        Description of the task
--dep <name>         Task dependency (repeatable)
--task-dir <path>    Run the task in a specific directory
--task-silent        Do not print the command or its output
--interactive        Mark the task as an interactive CLI application
```

Example:

```bash
razd add task test --desc "Run tests" --dep build -- go test ./...
```

When a provisioner (mise/devbox) is configured, `razd add task` also ensures
the `task` tool is available in the native config (e.g. `mise use task`),
defaulting to the `latest` version. An already-pinned version is left
untouched.

### Shell activation

`razd shell` opens an interactive subshell with the provisioner already
activated (mise/devbox tools and env vars are available). The activation is
temporary — `exit` returns to your original shell unchanged:

```bash
razd shell
# ... inside: node, python, etc. are available
exit
```

To activate the **current** shell instead (persists for the session), print the
activation script and evaluate it:

```bash
# Unix (bash/zsh/fish)
eval "$(razd shell --print)"

# PowerShell (Windows)
iex "$(razd shell --print)"
```

`razd shell` detects the current shell from `$SHELL` and uses the provisioner's
activation command (`mise activate <shell>` or `devbox shellenv --format
<shell>`).

### Version conflicts

If a tool exists in both files with **different versions**, razd asks which one
to use:

```
Version conflict for node
  [1] Use Razdfile (22)
  [2] Use mise.toml (23)
  [3] Skip (keep both unchanged)
```

In non-interactive mode (CI, pipes) the conflict is **skipped** — neither file
is modified, and a warning is printed.

### Backups

Before the native config is overwritten, razd prompts:

```
Make backup of mise.toml before overwrite? [y/N]
```

The default is **No**. Choose Yes to write a timestamped backup
(`mise.toml.bak.<timestamp>`). Pass `--backup` to force a backup without the
prompt.

### Flags

```
-d, --dir <path>      # Working directory (default: current)
-y, --yes             # Auto-trust project
-v, --verbose         # Verbose output
--json                # JSON output (list command)
--all                 # Show all tasks including internals
--no-sync             # Skip Razdfile <-> native config sync
--backup              # Force backup of native config before overwrite
```


## Dev

```bash
devbox run go build -o razd .
```


## Examples

```bash
# From project root
./razd run default -d examples/nodejs-devbox -y
./razd list -d examples/nodejs-project

# From example directory
cd examples/nodejs-devbox
../../razd run default -y
```
