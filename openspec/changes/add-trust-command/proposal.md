# Proposal: Add Trust Command

## Summary

Add a `razd trust` command that marks project directories as trusted before executing potentially dangerous operations. Similar to `mise trust`, this provides a security layer by requiring explicit user consent before running project configurations.

## Motivation

### Problem Statement

Currently, `razd` executes project configurations (Razdfile.yml, tasks, mise configurations) without asking for user consent. This poses security risks when:

- Cloning untrusted repositories
- Running `razd` in a directory with malicious configuration
- Executing tasks that could modify the system

### Solution

Implement a trust mechanism that:

1. Prompts user for trust confirmation on first `razd` command in a project
2. Stores trusted project paths in a temp/cache directory (outside the project)
3. Automatically triggers `mise trust` when `razd trust` is executed
4. Auto-approves trust when `--yes` flag is used

## Scope

### In Scope

- New `razd trust` command with `--untrust`, `--show`, and `--all` flags
- Trust state storage in temp/cache directory (`~/.cache/razd/trusted.json` or equivalent)
- Trust prompt on first execution of any razd command in untrusted project
- Automatic `mise trust` execution when `razd trust` is called
- `--yes` flag bypasses trust prompt (auto-trusts)

### Out of Scope

- Cryptographic verification of Razdfile.yml content
- Per-file trust (only per-directory trust)
- Trust expiration/TTL

## Design Decisions

### Trust Storage Location

Store in user cache directory:

- Unix: `~/.cache/razd/trusted.json`
- Windows: `%LOCALAPPDATA%\razd\trusted.json`

Format:

```json
{
  "trusted_paths": [
    "/home/user/projects/my-app",
    "C:\\Users\\user\\dev\\my-project"
  ],
  "ignored_paths": ["/tmp/untrusted-repo"]
}
```

### Trust Check Flow

1. Before any razd command execution
2. Check if current directory (or parent with Razdfile.yml) is in trusted list
3. If not trusted:
   - If `--yes` flag: auto-trust and continue
   - Otherwise: prompt user "Do you trust this project? [y/N]"
   - If user declines: exit with error

### Command Interface

```
razd trust [PATH]          # Trust current directory or specified path
razd trust --untrust       # Remove trust for current directory
razd trust --show          # Show trust status
razd trust --all           # Trust all parent configs too
razd trust --ignore        # Ignore this project (never prompt again, never trust)
```

## Success Criteria

- [ ] `razd trust` adds directory to trusted list and runs `mise trust`
- [ ] First razd command in untrusted project prompts for trust
- [ ] `--yes` flag auto-approves trust
- [ ] `razd trust --untrust` removes trust
- [ ] `razd trust --show` displays trust status
- [ ] Trust state persists across sessions

## References

- [mise trust documentation](https://mise.jdx.dev/cli/trust.html)
- [mise trust source code](https://github.com/jdx/mise/blob/main/src/cli/trust.rs)
