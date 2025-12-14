# Design: Trust System Architecture

## Overview

The trust system provides a security layer that prevents execution of potentially dangerous project configurations without explicit user consent.

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ razd trust   │  │ razd up      │  │ razd run     │  ...  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │                 │                 │               │
│         v                 v                 v               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Trust Guard (middleware)                │    │
│  │  - check_trust() before command execution           │    │
│  │  - prompt_trust() if untrusted                      │    │
│  │  - auto_trust() if --yes flag                       │    │
│  └──────────────────────┬──────────────────────────────┘    │
└─────────────────────────┼───────────────────────────────────┘
                          │
                          v
┌─────────────────────────────────────────────────────────────┐
│                   Trust Storage Layer                        │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              TrustStore                              │    │
│  │  - load() / save() JSON file                        │    │
│  │  - is_trusted(path) -> bool                         │    │
│  │  - is_ignored(path) -> bool                         │    │
│  │  - add_trusted(path)                                │    │
│  │  - remove_trusted(path)                             │    │
│  │  - add_ignored(path)                                │    │
│  └──────────────────────┬──────────────────────────────┘    │
│                         │                                    │
│                         v                                    │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  ~/.cache/razd/trusted.json (or %LOCALAPPDATA%)     │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Trust Store File Format

```json
{
  "version": 1,
  "trusted": [
    {
      "path": "/home/user/projects/my-app",
      "trusted_at": "2025-12-14T10:30:00Z"
    }
  ],
  "ignored": [
    {
      "path": "/tmp/suspicious-repo",
      "ignored_at": "2025-12-14T11:00:00Z"
    }
  ]
}
```

### Path Normalization

- Paths are canonicalized (resolved symlinks, absolute paths)
- On Windows: lowercase drive letters, forward slashes normalized
- Trailing slashes removed

## Trust Check Flow

```
┌─────────────────┐
│ razd <command>  │
└────────┬────────┘
         │
         v
┌─────────────────────┐
│ Has Razdfile.yml?   │──No──> Execute without trust check
└────────┬────────────┘        (no config = no danger)
         │ Yes
         v
┌─────────────────────┐
│ Is path trusted?    │──Yes──> Execute command
└────────┬────────────┘
         │ No
         v
┌─────────────────────┐
│ Is path ignored?    │──Yes──> Exit with error
└────────┬────────────┘         "Project is ignored"
         │ No
         v
┌─────────────────────┐
│ Is --yes flag set?  │──Yes──> Auto-trust + Execute
└────────┬────────────┘
         │ No
         v
┌──────────────────────────────────────────────────────────────┐
│ razd config files in /path/to/project are not trusted.      │
│ Trust them?                                                  │
│                                                              │
│    [ Yes ]    [ No ]    [ Ignore ]                           │
│                                                              │
│ ←/→ toggle • y/n/i/enter submit                              │
└────────┬─────────────────────────────────────────────────────┘
         │
    ┌────┴────┬────────────┐
    v         v            v
  [Yes]     [No]      [Ignore]
    │         │            │
    v         v            v
 Trust     Exit        Add to
 + Run    (error)     ignored
                      + Exit
```

## Integration with mise trust

When `razd trust` is executed:

1. Add path to razd trust store
2. Check if mise config exists (mise.toml, .mise.toml, Razdfile.yml with mise section)
3. If yes, execute `mise trust` for that directory
4. Show combined success message

## Commands Requiring Trust

All commands that execute project configuration:

- `razd up` (except `--init`)
- `razd install`
- `razd setup`
- `razd dev`
- `razd build`
- `razd run`

Commands NOT requiring trust:

- `razd --help`
- `razd --version`
- `razd trust` (itself)
- `razd list` (read-only)
- `razd up --init` (creates config, doesn't execute)

## Error Messages

### Untrusted Project (user declined)

```
Error: Project is not trusted

Path: /home/user/suspicious-repo

To trust this project, run:
  razd trust

Or run with --yes to auto-trust:
  razd --yes up
```

### Ignored Project

```
Error: Project is ignored

Path: /tmp/malicious-repo

This project was previously marked as ignored.
To remove from ignore list, run:
  razd trust --untrust
  razd trust
```

## Security Considerations

1. **Path traversal**: Canonicalize all paths before storing/comparing
2. **Race conditions**: Lock file during read-modify-write operations
3. **Trust inheritance**: Only exact path match, no parent/child inheritance
4. **Symlink attacks**: Resolve symlinks before trust check
