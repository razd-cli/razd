# Omit Default Internal Field in Task Serialization

## Problem Statement

Currently, when tasks are serialized to YAML in `Razdfile.yml`, the `internal: false` field is always written, even though `false` is the default value. This creates unnecessary noise in configuration files, especially during mise synchronization when tasks are re-serialized.

### Current Behavior (Problematic)

```yaml
tasks:
  default:
    desc: Set up and start Node.js project
    cmds:
    - mise install
    - npm install
    - npm run dev
    internal: false  # ← Unnecessary, this is the default

  install:
    desc: Install dependencies
    cmds:
    - mise install
    - npm install
    internal: false  # ← Unnecessary
```

### Expected Behavior

```yaml
tasks:
  default:
    desc: Set up and start Node.js project
    cmds:
    - mise install
    - npm install
    - npm run dev
  # No internal field when it's false

  install:
    desc: Install dependencies
    cmds:
    - mise install
    - npm install
  # No internal field when it's false
  
  _helper:
    desc: Internal helper task
    cmds:
    - echo "helper"
    internal: true  # ← Only present when true
```

## Root Cause

The `TaskConfig` struct in `src/config/razdfile.rs` defines the `internal` field as:

```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<Command>,
    #[serde(default)]
    pub internal: bool,  // ← Missing skip_serializing_if attribute
}
```

The field has `#[serde(default)]` which correctly sets default value to `false` on deserialization, but lacks `#[serde(skip_serializing_if = "...")]` to omit it during serialization when it's the default value.

## Proposed Solution

Add the `skip_serializing_if` attribute to skip serialization when `internal` is `false`:

```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<Command>,
    #[serde(default, skip_serializing_if = "is_false")]
    pub internal: bool,
}

fn is_false(value: &bool) -> bool {
    !*value
}
```

Alternatively, use a more generic approach with a reusable helper:

```rust
fn is_default<T: Default + PartialEq>(value: &T) -> bool {
    value == &T::default()
}

#[serde(default, skip_serializing_if = "is_default")]
pub internal: bool,
```

## Impact Analysis

### Benefits

1. **Cleaner YAML files**: Reduces visual clutter in configuration files
2. **Consistency**: Matches behavior of other optional fields (`desc`, `mise.tools`, etc.) which use `skip_serializing_if`
3. **Better UX**: Users don't see confusing `internal: false` lines they didn't write
4. **Semantic correctness**: Only serialize non-default values, following YAML best practices

### Compatibility

**Backwards Compatible**: ✅ Yes

- Existing `Razdfile.yml` files with explicit `internal: false` will still parse correctly (serde will read it and set the field)
- Files without the field will continue to work (serde default kicks in)
- Only affects serialization output, not deserialization input

### Affected Components

1. **src/config/razdfile.rs**: Update `TaskConfig` struct definition
2. **Tests**: Update test fixtures that currently include `internal: false`
   - `tests/order_integration_test.rs`
   - `src/config/canonical.rs` (test code)
3. **Examples**: Update example files if they contain explicit `internal: false`

## Testing Strategy

### Unit Tests

1. **Serialization test**: Verify `internal: false` is omitted
   ```rust
   #[test]
   fn test_task_config_omits_default_internal() {
       let task = TaskConfig {
           desc: Some("Test".to_string()),
           cmds: vec![Command::String("echo test".to_string())],
           internal: false,
       };
       let yaml = serde_yaml::to_string(&task).unwrap();
       assert!(!yaml.contains("internal"));
   }
   ```

2. **Serialization test**: Verify `internal: true` is included
   ```rust
   #[test]
   fn test_task_config_includes_internal_true() {
       let task = TaskConfig {
           desc: Some("Helper".to_string()),
           cmds: vec![Command::String("echo helper".to_string())],
           internal: true,
       };
       let yaml = serde_yaml::to_string(&task).unwrap();
       assert!(yaml.contains("internal: true"));
   }
   ```

3. **Deserialization test**: Verify explicit `false` still works
   ```rust
   #[test]
   fn test_task_config_parses_explicit_false() {
       let yaml = r#"
       desc: Test
       cmds:
         - echo test
       internal: false
       "#;
       let task: TaskConfig = serde_yaml::from_str(yaml).unwrap();
       assert_eq!(task.internal, false);
   }
   ```

4. **Deserialization test**: Verify missing field defaults to `false`
   ```rust
   #[test]
   fn test_task_config_defaults_internal() {
       let yaml = r#"
       desc: Test
       cmds:
         - echo test
       "#;
       let task: TaskConfig = serde_yaml::from_str(yaml).unwrap();
       assert_eq!(task.internal, false);
   }
   ```

### Integration Tests

1. **Sync roundtrip test**: Verify mise sync doesn't add `internal: false`
2. **Example project test**: Verify example `Razdfile.yml` remains clean after operations

## Implementation Steps

1. Add helper function `is_false` or `is_default` to `src/config/razdfile.rs`
2. Update `TaskConfig.internal` field with `skip_serializing_if` attribute
3. Update test fixtures to remove explicit `internal: false` assignments
4. Add new unit tests for serialization behavior
5. Run full test suite to verify backwards compatibility
6. Update any example files if needed

## Success Criteria

- [ ] `internal: false` is not present in serialized YAML
- [ ] `internal: true` is present when set
- [ ] All existing tests pass
- [ ] New tests verify serialization behavior
- [ ] Example files remain clean after sync operations
- [ ] Backwards compatibility maintained (existing files with explicit `false` still work)

## Related Issues

This is a minor quality-of-life improvement that addresses user confusion during mise synchronization when unnecessary `internal: false` lines appear in their configuration files.
