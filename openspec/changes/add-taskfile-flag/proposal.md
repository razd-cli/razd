# Proposal: Add --taskfile Flag Support

## Problem Statement

Currently, `razd` hardcodes the configuration file name as `Razdfile.yml`, which limits flexibility for users who:
- Want to use multiple configuration files for different environments or purposes
- Prefer to use custom naming conventions (e.g., `Taskfile.yml`, `razd.dev.yml`)
- Need to specify configuration files outside the current directory
- Want compatibility with existing taskfile.dev workflows that use `--taskfile` flag

The taskfile.dev CLI provides `--taskfile <file>` (short: `-t`) flag to specify custom Taskfile paths, and users expect similar functionality in `razd`.

## Proposed Solution

Add support for `--taskfile` and `--razdfile` flags (with short variant `-t`) to allow users to specify a custom configuration file path. Both flags will be synonyms, with `--razdfile` taking priority if both are specified.

### Key Features
1. **Global flag**: Available for all commands that use configuration (list, run, up, setup)
2. **Both long forms supported**: `--taskfile <file>` and `--razdfile <file>` as synonyms
3. **Short form**: `-t <file>` for convenience
4. **Priority**: `--razdfile` overrides `--taskfile` if both specified
5. **Path resolution**: Support both relative and absolute paths
6. **Backward compatibility**: Default to `Razdfile.yml` when flag not specified

### Example Usage
```bash
# Using --taskfile flag
razd list --taskfile ./custom/config.yml

# Using --razdfile flag
razd run build --razdfile ./Taskfile.yml

# Using short form
razd up -t ../shared/Razdfile.yml

# Default behavior (no flag)
razd list  # Uses ./Razdfile.yml
```

## Impact Assessment

### User Benefits
- **Flexibility**: Choose custom configuration file names and locations
- **Compatibility**: Align with taskfile.dev conventions
- **Multi-environment**: Use different configs for dev/staging/prod
- **Migration**: Easier transition from taskfile.dev to razd

### Technical Impact
- **Breaking changes**: None (default behavior unchanged)
- **New dependencies**: None
- **Performance**: Negligible (single path resolution per command)
- **Testing**: Requires new tests for custom paths

## Alternatives Considered

1. **Environment variable only** (`RAZD_TASKFILE`)
   - Rejected: Less discoverable, conflicts with global state
   
2. **Configuration file setting**
   - Rejected: Circular dependency (need config to find config)
   
3. **Only `--razdfile` flag**
   - Rejected: Less familiar to taskfile.dev users

## Open Questions

1. **Scope decision**: Should this be a global flag or command-specific?
   - Current proposal: Global flag (affects all commands)
   - Alternative: Only for commands that read config (list, run, setup)
   - **Decision needed**: User preference on scope

2. **Validation**: Should we validate file existence immediately or lazily?
   - Current proposal: Lazy (validate when command tries to read)
   - Alternative: Eager (validate in CLI parsing)

3. **Error messages**: How to indicate which flag was used in errors?
   - Current proposal: Show actual path used, not flag name
   - Alternative: Include flag name in error context
