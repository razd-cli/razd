# razd

A project setup tool powered by mise or devbox.

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
razd shell            # Start interactive shell with provisioned env
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
