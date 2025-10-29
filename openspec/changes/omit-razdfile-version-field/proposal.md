# Omit Razdfile Version Field

## Why
The `version: '3'` field in Razdfile.yml is required by the Taskfile tool but adds unnecessary boilerplate for users. Since razd internally manages the interaction with Taskfile, the version field can be automatically injected during task execution, simplifying the user-facing configuration format.

This change improves user experience by:
- Reducing boilerplate in Razdfile.yml configuration
- Maintaining a cleaner, more focused configuration format
- Preserving full compatibility with Taskfile.dev under the hood

## What Changes
- Make `version` field optional in Razdfile.yml parsing
- Automatically inject `version: '3'` when serializing Razdfile to YAML for taskfile execution
- Update example files and documentation to omit version field
- Maintain backward compatibility: existing Razdfile.yml files with explicit version will continue to work

## Impact
**Affected specs:**
- `cli-interface` - Razdfile.yml format and initialization behavior
- `tool-integration` - Taskfile execution and YAML serialization

**Affected code:**
- `src/config/razdfile.rs` - Make version field optional, add default value
- `src/integrations/taskfile.rs` - Ensure version is injected when serializing
- `examples/nodejs-project/Razdfile.yml` - Remove explicit version field
- `src/defaults.rs` - Update default workflow templates (keep version for backward compatibility)

**User impact:**
- Existing Razdfile.yml with `version: '3'` continues to work (backward compatible)
- New users can omit the version field for cleaner configuration
- No breaking changes
