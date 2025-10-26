# Implementation Tasks

## Overview
Implement support for Taskfile task references in Razdfile.yml by changing the command representation from simple strings to a flexible enum that handles both string commands and task references.

## Tasks

### 1. Core Implementation
- [x] Define `Command` enum in `src/config/razdfile.rs` with `String` and `TaskRef` variants
- [x] Add `#[serde(untagged)]` attribute for automatic variant detection
- [x] Update `TaskConfig.cmds` field from `Vec<String>` to `Vec<Command>`
- [x] Implement `TaskRef` struct with `task: String` and optional `vars: HashMap<String, String>`

### 2. Serialization
- [x] Test YAML deserialization of string commands (e.g., `- echo "test"`)
- [x] Test YAML deserialization of task references (e.g., `- task: install`)
- [x] Test YAML deserialization of task references with vars
- [x] Ensure serialization back to YAML maintains correct format for taskfile execution

### 3. Testing
- [x] Add unit test for parsing Razdfile with string commands only
- [x] Add unit test for parsing Razdfile with task references only  
- [x] Add unit test for parsing Razdfile with mixed commands (strings + task refs)
- [x] Add unit test for task references with variables
- [x] Add integration test that runs `razd up` with task references
- [x] Verify backward compatibility with existing test fixtures

### 4. Error Handling
- [x] Update error messages to provide clear guidance for invalid command formats
- [x] Add helpful error message when command is neither string nor valid task reference
- [x] Test error cases with malformed command structures

### 5. Documentation
- [x] Add inline code documentation for `Command` enum and its variants
- [x] Update module-level documentation in `razdfile.rs`
- [x] Add examples in code comments showing both command types

### 6. Validation
- [x] Run `cargo test` to ensure all tests pass
- [x] Run `cargo clippy` to check for warnings
- [x] Run `cargo fmt` to ensure consistent formatting
- [x] Test manually with the user's example Razdfile.yml
- [x] Verify `razd up` correctly executes task references

## Implementation Notes

### Command Enum Structure
```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Command {
    String(String),
    TaskRef {
        task: String,
        #[serde(skip_serializing_if = "Option::is_none")]
        vars: Option<HashMap<String, String>>,
    },
}
```

### Test Examples

**String commands:**
```yaml
cmds:
  - echo "Installing..."
  - mise install
```

**Task references:**
```yaml
cmds:
  - task: install
  - task: setup
```

**Mixed:**
```yaml
cmds:
  - echo "Starting..."
  - task: install
  - echo "Done!"
```

**With variables:**
```yaml
cmds:
  - task: deploy
    vars:
      ENV: production
      VERSION: v1.0.0
```

## Definition of Done
- [x] All unit tests pass
- [x] All integration tests pass
- [x] User's example Razdfile.yml parses without error
- [x] `razd up` correctly executes task references
- [x] No clippy warnings
- [x] Code formatted with rustfmt
- [x] Documentation complete
