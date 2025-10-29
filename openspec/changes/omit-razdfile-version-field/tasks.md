# Implementation Tasks

## 1. Update Razdfile Configuration Struct
- [x] 1.1 Modify `RazdfileConfig` struct in `src/config/razdfile.rs` to make `version` field optional with `#[serde(default = "default_version")]`
- [x] 1.2 Add `default_version()` function that returns `"3".to_string()`
- [x] 1.3 Ensure version defaults to "3" when not present in parsed YAML
- [x] 1.4 Verify parsing with unit test for Razdfile without version field
- [x] 1.5 Verify parsing with unit test for Razdfile with explicit version (backward compat)

## 2. Update Serialization Logic
- [x] 2.1 Ensure `version` field is always included when serializing `RazdfileConfig` to YAML (even if omitted in source)
- [x] 2.2 Test `serde_yaml::to_string()` output includes `version: '3'` at the top
- [x] 2.3 Verify `get_workflow_config()` generates valid Taskfile YAML with version
- [x] 2.4 Test that canonicalization in `canonical.rs` works correctly with optional version

## 3. Update Example Files
- [x] 3.1 Remove `version: '3'` from `examples/nodejs-project/Razdfile.yml`
- [x] 3.2 Keep version in `.razd-workflow-default.yml` for standalone compatibility
- [x] 3.3 Add comment in example explaining version field is optional

## 4. Update Init Command (if exists)
- [x] 4.1 Check if `razd init` command generates Razdfile.yml
- [x] 4.2 Update template to omit `version` field if init command exists
- [x] 4.3 Verify generated Razdfile.yml is clean and minimal

## 5. Integration Testing
- [x] 5.1 Test `razd task <name>` with Razdfile.yml without version field
- [x] 5.2 Test `razd up` workflow with Razdfile.yml without version
- [x] 5.3 Test `razd dev` workflow with Razdfile.yml without version
- [x] 5.4 Test `razd build` workflow with Razdfile.yml without version
- [x] 5.5 Test backward compatibility with existing Razdfile.yml containing explicit version
- [x] 5.6 Verify default workflows (DEFAULT_WORKFLOWS) continue to work

## 6. Update Tests
- [x] 6.1 Add unit test in `src/config/razdfile.rs` for optional version parsing
- [x] 6.2 Add unit test for version serialization behavior
- [x] 6.3 Update integration tests if they depend on version field presence
- [x] 6.4 Add regression test for backward compatibility

## 7. Validation
- [x] 7.1 Run `cargo test` and ensure all tests pass
- [x] 7.2 Run `cargo build --release` and verify successful compilation
- [x] 7.3 Manually test example project execution
- [x] 7.4 Verify `openspec validate` passes for the proposal
