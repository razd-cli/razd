# Enhance List JSON Output for Taskfile Compatibility

## Problem

The current `razd list --json` output provides minimal task information (name, desc, internal), which is insufficient for advanced tooling and IDE integrations that expect taskfile-compatible JSON format.

**Current output:**
```json
{
  "tasks": [
    {
      "name": "hello",
      "desc": "Test task from Razdfile",
      "internal": false
    }
  ]
}
```

**Expected taskfile-compatible output:**
```json
{
  "tasks": [
    {
      "name": "hello",
      "task": "hello",
      "desc": "Test task from Razdfile",
      "summary": "",
      "aliases": [],
      "up_to_date": false,
      "location": {
        "line": 7,
        "column": 3,
        "taskfile": "C:\\path\\to\\Razdfile.yml"
      }
    }
  ],
  "location": "C:\\path\\to\\Razdfile.yml"
}
```

### Missing Fields

1. **`task`** - Duplicate of name (taskfile convention)
2. **`summary`** - Extended description (currently not in Razdfile schema)
3. **`aliases`** - Alternative task names (not supported in Razdfile)
4. **`up_to_date`** - Runtime status (requires task execution state)
5. **`location`** - Source file position (line, column, file path)
6. **Root `location`** - Razdfile.yml path

### Impact

Without taskfile-compatible JSON:
- ❌ IDE extensions can't provide rich task information
- ❌ External tools can't parse razd output reliably
- ❌ Migration from taskfile to razd is harder
- ❌ Debugging task definitions is more difficult

## Solution

Enhance `razd list --json` output to include taskfile-compatible fields while maintaining backward compatibility through versioned output format.

### Phase 1: Essential Fields (This Proposal)
Add fields that can be computed from Razdfile.yml without external dependencies:

1. **`task`** - Mirror of `name` field
2. **`summary`** - Empty string (future: read from task config)
3. **`aliases`** - Empty array (future: add alias support)
4. **`location`** - File path, line number, column from YAML parser
5. **Root `location`** - Absolute path to Razdfile.yml

### Phase 2: Runtime Fields (Future)
Fields requiring task execution or state tracking:

1. **`up_to_date`** - Requires running taskfile's status check
2. **`namespaces`** - Requires multi-file Razdfile support

## Benefits

- ✅ **Tooling compatibility**: IDEs and extensions work seamlessly
- ✅ **Migration path**: Easier transition from taskfile to razd
- ✅ **Better debugging**: Source locations help identify task definitions
- ✅ **Standards compliance**: Matches established taskfile format
- ✅ **Future-proof**: Extensible for advanced features

## Scope

**In scope:**
- Add `task`, `summary`, `aliases` fields to JSON output
- Parse YAML with source location tracking (line/column)
- Include `location` object with file, line, column
- Add root `location` field with Razdfile.yml path
- Maintain existing `internal` field (razd-specific)

**Out of scope:**
- `up_to_date` status (requires task execution state)
- `namespaces` support (requires multi-file architecture)
- Adding `summary` to Razdfile schema (separate proposal)
- Alias functionality (separate proposal)

## Implementation Strategy

1. Extend JSON serialization structures in `src/commands/list.rs`
2. Use `serde_yaml::Value` to track source positions
3. Compute absolute Razdfile path from current directory
4. Populate placeholder values for unsupported fields (empty strings/arrays)
5. Add integration tests verifying JSON structure

## Non-Goals

- Changing text output format (stays as-is)
- Breaking existing JSON consumers (additive changes only)
- Implementing full taskfile feature parity
