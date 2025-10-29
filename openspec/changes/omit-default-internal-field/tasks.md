# Omit Default Internal Field - Implementation Tasks

## Overview

Remove unnecessary `internal: false` from serialized task configurations by adding `skip_serializing_if` attribute.

## Task Breakdown

### Phase 1: Core Implementation (30 minutes)

#### Task 1.1: Add Helper Function
**Estimate**: 5 minutes  
**Priority**: P0 (Blocker)

**Steps**:
- [ ] Add `is_false` helper function to `src/config/razdfile.rs`
  ```rust
  /// Returns true if the value is false (for skip_serializing_if)
  fn is_false(value: &bool) -> bool {
      !*value
  }
  ```
- [ ] Place function near `TaskConfig` struct definition

**Acceptance Criteria**:
- Helper function compiles
- Function is accessible to serde attribute

---

#### Task 1.2: Update TaskConfig Attribute
**Estimate**: 5 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.1

**Steps**:
- [ ] Modify `internal` field in `TaskConfig` struct
- [ ] Add `skip_serializing_if = "is_false"` to serde attributes
  ```rust
  #[serde(default, skip_serializing_if = "is_false")]
  pub internal: bool,
  ```

**Acceptance Criteria**:
- Code compiles without errors
- Existing tests still pass

---

### Phase 2: Update Test Fixtures (20 minutes)

#### Task 2.1: Update Integration Test
**Estimate**: 10 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.2

**Steps**:
- [ ] Update `tests/order_integration_test.rs`
- [ ] Remove explicit `internal: false` assignments (lines 27, 32, 37, 42)
- [ ] Verify test creates tasks without internal field

**Acceptance Criteria**:
- Test compiles
- Test passes with updated fixture

---

#### Task 2.2: Update Canonical Test
**Estimate**: 10 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.2

**Steps**:
- [ ] Update `src/config/canonical.rs` test code
- [ ] Remove `internal: false` from test fixture (line 231)
- [ ] Verify canonical tests pass

**Acceptance Criteria**:
- Test compiles
- Canonical tests pass

---

### Phase 3: Add Serialization Tests (30 minutes)

#### Task 3.1: Test Omission of False Value
**Estimate**: 10 minutes  
**Priority**: P1 (High)  
**Dependencies**: Task 1.2

**Steps**:
- [ ] Add test `test_task_config_omits_default_internal` to `src/config/razdfile.rs`
- [ ] Create TaskConfig with `internal: false`
- [ ] Serialize to YAML
- [ ] Assert YAML does not contain "internal"

**Acceptance Criteria**:
- Test passes
- Verifies omission behavior

---

#### Task 3.2: Test Inclusion of True Value
**Estimate**: 10 minutes  
**Priority**: P1 (High)  
**Dependencies**: Task 1.2

**Steps**:
- [ ] Add test `test_task_config_includes_internal_true` to `src/config/razdfile.rs`
- [ ] Create TaskConfig with `internal: true`
- [ ] Serialize to YAML
- [ ] Assert YAML contains "internal: true"

**Acceptance Criteria**:
- Test passes
- Verifies true values are included

---

#### Task 3.3: Test Backwards Compatibility
**Estimate**: 10 minutes  
**Priority**: P1 (High)  
**Dependencies**: Task 1.2

**Steps**:
- [ ] Add test `test_task_config_parses_explicit_false` to verify explicit false works
- [ ] Add test `test_task_config_defaults_internal` to verify missing field defaults
- [ ] Both tests deserialize YAML and check internal field

**Acceptance Criteria**:
- Both tests pass
- Confirms backwards compatibility

---

### Phase 4: Integration Verification (20 minutes)

#### Task 4.1: Run Full Test Suite
**Estimate**: 10 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: All Phase 3 tasks

**Steps**:
- [ ] Run `cargo test --all-targets`
- [ ] Verify all 132+ tests pass
- [ ] Check for any serialization-related failures

**Acceptance Criteria**:
- All tests pass
- No regressions detected

---

#### Task 4.2: Manual Sync Test
**Estimate**: 10 minutes  
**Priority**: P1 (High)  
**Dependencies**: Task 4.1

**Steps**:
- [ ] Run `razd up` in `examples/nodejs-project`
- [ ] Trigger mise sync operation
- [ ] Verify `Razdfile.yml` does not contain `internal: false` lines
- [ ] Check that task execution still works

**Acceptance Criteria**:
- No `internal: false` in serialized YAML
- Sync works correctly
- Tasks execute properly

---

### Phase 5: Documentation (Optional, 10 minutes)

#### Task 5.1: Update CHANGELOG
**Estimate**: 5 minutes  
**Priority**: P2 (Nice to have)

**Steps**:
- [ ] Add entry to CHANGELOG.md under [Unreleased]
- [ ] Document as improvement/cleanup

**Acceptance Criteria**:
- CHANGELOG entry added

---

#### Task 5.2: Update Examples if Needed
**Estimate**: 5 minutes  
**Priority**: P2 (Nice to have)

**Steps**:
- [ ] Check if any example files have explicit `internal: false`
- [ ] Remove if found
- [ ] Test examples still work

**Acceptance Criteria**:
- Examples are clean

---

## Estimated Total Time

- **Phase 1**: 10 minutes
- **Phase 2**: 20 minutes
- **Phase 3**: 30 minutes
- **Phase 4**: 20 minutes
- **Phase 5**: 10 minutes (optional)

**Total**: 1-1.5 hours

## Dependencies

- No external dependencies
- All work is within existing codebase

## Risks

**Low Risk**: 
- Simple attribute change
- Backwards compatible
- Well-tested serialization library (serde)

## Success Metrics

- [ ] `internal: false` not present in serialized YAML
- [ ] `internal: true` present when set
- [ ] All 132+ tests pass
- [ ] Manual verification shows clean output
- [ ] No user-visible breaking changes
