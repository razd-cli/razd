# Semantic Change Detection - Implementation Tasks

## Overview

This document breaks down the implementation of semantic change detection into actionable tasks with estimates and dependencies.

## Task Breakdown

### Phase 1: Foundation (4-5 hours)

#### Task 1.1: Create Canonical Module
**Estimate**: 2 hours  
**Priority**: P0 (Blocker)  
**Dependencies**: None

**Subtasks**:
- [ ] Create `src/config/canonical.rs` file
- [ ] Add module declaration in `src/config/mod.rs`
- [ ] Define public API:
  ```rust
  pub fn canonicalize_razdfile(config: &RazdfileConfig) -> String
  pub fn canonicalize_mise_toml(config: &MiseConfig) -> String
  ```
- [ ] Add necessary imports (IndexMap, fmt::Write, etc.)

**Acceptance Criteria**:
- Module compiles without errors
- Public functions are accessible from other modules
- Basic structure in place

---

#### Task 1.2: Implement Razdfile Canonicalization
**Estimate**: 2 hours  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.1

**Subtasks**:
- [ ] Implement `canonicalize_razdfile()` function
  - [ ] Handle version field
  - [ ] Handle mise section (tools, plugins)
  - [ ] Handle tasks section (desc, cmds, internal)
- [ ] Handle ToolConfig enum variants (Simple, Detailed)
- [ ] Ensure deterministic ordering (use IndexMap iteration)
- [ ] Use consistent formatting (key=value pairs)

**Implementation Details**:
```rust
pub fn canonicalize_razdfile(config: &RazdfileConfig) -> String {
    let mut output = String::new();
    
    // Version
    writeln!(output, "v:{}", config.version).unwrap();
    
    // Mise section
    if let Some(ref mise) = config.mise {
        output.push_str("mise:\n");
        
        if let Some(ref tools) = mise.tools {
            output.push_str("  tools:\n");
            for (name, tool_config) in tools {
                let version = match tool_config {
                    ToolConfig::Simple(v) => v.clone(),
                    ToolConfig::Detailed { version, .. } => version.clone(),
                };
                writeln!(output, "    {}={}", name, version).unwrap();
            }
        }
        
        if let Some(ref plugins) = mise.plugins {
            output.push_str("  plugins:\n");
            for (name, url) in plugins {
                writeln!(output, "    {}={}", name, url).unwrap();
            }
        }
    }
    
    // Tasks section
    output.push_str("tasks:\n");
    for (name, task) in &config.tasks {
        writeln!(output, "  {}:", name).unwrap();
        
        if let Some(ref desc) = task.desc {
            writeln!(output, "    desc={}", desc).unwrap();
        }
        
        if let Some(ref cmds) = task.cmds {
            output.push_str("    cmds:\n");
            for cmd in cmds {
                writeln!(output, "      - {}", cmd).unwrap();
            }
        }
        
        if let Some(internal) = task.internal {
            writeln!(output, "    internal={}", internal).unwrap();
        }
    }
    
    output
}
```

**Acceptance Criteria**:
- Function produces deterministic output
- All RazdfileConfig fields are represented
- Output is independent of input formatting
- IndexMap ordering is preserved

---

#### Task 1.3: Implement Mise TOML Canonicalization
**Estimate**: 1 hour  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.1

**Subtasks**:
- [ ] Implement `canonicalize_mise_toml()` function
  - [ ] Handle [tools] section
  - [ ] Handle [plugins] section
- [ ] Use consistent formatting
- [ ] Ensure deterministic ordering

**Implementation Details**:
```rust
pub fn canonicalize_mise_toml(config: &MiseConfig) -> String {
    let mut output = String::new();
    
    if let Some(ref tools) = config.tools {
        output.push_str("[tools]\n");
        for (name, tool_config) in tools {
            let version = match tool_config {
                ToolConfig::Simple(v) => v.clone(),
                ToolConfig::Detailed { version, .. } => version.clone(),
            };
            writeln!(output, "{}={}", name, version).unwrap();
        }
    }
    
    if let Some(ref plugins) = config.plugins {
        output.push_str("[plugins]\n");
        for (name, url) in plugins {
            writeln!(output, "{}={}", name, url).unwrap();
        }
    }
    
    output
}
```

**Acceptance Criteria**:
- Function produces deterministic output
- All MiseConfig fields are represented
- Output is independent of TOML formatting

---

#### Task 1.4: Add Unit Tests for Canonical Module
**Estimate**: 1 hour  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.2, Task 1.3

**Subtasks**:
- [ ] Test: Same config produces same canonical form
- [ ] Test: Formatting differences ignored
- [ ] Test: Semantic differences detected
- [ ] Test: Empty sections handled correctly
- [ ] Test: All field types covered

**Test Cases**:
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_canonical_is_deterministic() {
        let config = create_test_config();
        let canonical1 = canonicalize_razdfile(&config);
        let canonical2 = canonicalize_razdfile(&config);
        assert_eq!(canonical1, canonical2);
    }
    
    #[test]
    fn test_tool_version_change_detected() {
        let config1 = test_config_node_22();
        let config2 = test_config_node_24();
        
        let canonical1 = canonicalize_razdfile(&config1);
        let canonical2 = canonicalize_razdfile(&config2);
        
        assert_ne!(canonical1, canonical2);
    }
    
    #[test]
    fn test_empty_sections() {
        let config = RazdfileConfig {
            version: "3".to_string(),
            mise: None,
            tasks: IndexMap::new(),
        };
        
        let canonical = canonicalize_razdfile(&config);
        assert!(canonical.contains("v:3"));
        assert!(canonical.contains("tasks:\n"));
    }
}
```

**Acceptance Criteria**:
- All tests pass
- Code coverage > 90% for canonical.rs
- Edge cases covered

---

### Phase 2: File Tracker Integration (3-4 hours)

#### Task 2.1: Add Semantic Hash Functions
**Estimate**: 2 hours  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 1.4

**Subtasks**:
- [ ] Add `hash_string()` helper function
- [ ] Implement `compute_razdfile_semantic_hash()`
  - [ ] Load and parse Razdfile.yml
  - [ ] Canonicalize config
  - [ ] Hash canonical form
  - [ ] Add error handling (fallback to content hash)
- [ ] Implement `compute_mise_toml_semantic_hash()`
  - [ ] Parse mise.toml
  - [ ] Extract tools and plugins
  - [ ] Canonicalize config
  - [ ] Hash canonical form
  - [ ] Add error handling
- [ ] Update `compute_semantic_hash()` dispatcher

**Implementation Details**:
```rust
fn hash_string(content: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(content.as_bytes());
    format!("{:x}", hasher.finalize())
}

fn compute_razdfile_semantic_hash(path: &Path) -> Result<String> {
    match RazdfileConfig::load_from_path(path)? {
        Some(config) => {
            let canonical = canonicalize_razdfile(&config);
            Ok(hash_string(&canonical))
        }
        None => {
            // Fallback to content hash
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}

pub fn compute_semantic_hash(path: &Path) -> Result<String> {
    let file_name = path.file_name()
        .and_then(|n| n.to_str())
        .ok_or_else(|| RazdError::config("Invalid file path"))?;
    
    match file_name {
        "Razdfile.yml" => compute_razdfile_semantic_hash(path),
        "mise.toml" => compute_mise_toml_semantic_hash(path),
        _ => {
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}
```

**Acceptance Criteria**:
- Semantic hashing works for both file types
- Error handling provides graceful fallback
- Functions integrate with existing tracking system

---

#### Task 2.2: Add Tracking State Version Field
**Estimate**: 30 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: None

**Subtasks**:
- [ ] Add `format_version: Option<String>` to `TrackingState` struct
- [ ] Update serialization/deserialization
- [ ] Update existing code to handle new field

**Implementation Details**:
```rust
#[derive(Debug, Serialize, Deserialize)]
pub struct TrackingState {
    pub razdfile_hash: Option<String>,
    pub mise_toml_hash: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format_version: Option<String>,
}
```

**Acceptance Criteria**:
- Field added without breaking existing tracking files
- Serialization preserves backward compatibility

---

#### Task 2.3: Implement Migration Logic
**Estimate**: 1.5 hours  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 2.1, Task 2.2

**Subtasks**:
- [ ] Implement `migrate_tracking_state()` function
  - [ ] Check if migration needed
  - [ ] Recompute semantic hashes
  - [ ] Save new tracking state
  - [ ] Handle errors gracefully
- [ ] Add migration call to sync entry points
- [ ] Add logging for migration events

**Implementation Details**:
```rust
pub fn migrate_tracking_state(project_root: &Path) -> Result<()> {
    let tracking_file = project_root.join(".razd").join("tracking.json");
    
    if !tracking_file.exists() {
        return Ok(());
    }
    
    let state = load_tracking_state(project_root)?;
    
    // Already migrated?
    if state.format_version.as_deref() == Some("semantic-v1") {
        return Ok(());
    }
    
    // Recompute hashes
    let razdfile_path = project_root.join("Razdfile.yml");
    let mise_toml_path = project_root.join("mise.toml");
    
    let new_state = TrackingState {
        razdfile_hash: if razdfile_path.exists() {
            compute_semantic_hash(&razdfile_path).ok()
        } else {
            None
        },
        mise_toml_hash: if mise_toml_path.exists() {
            compute_semantic_hash(&mise_toml_path).ok()
        } else {
            None
        },
        format_version: Some("semantic-v1".to_string()),
    };
    
    save_tracking_state(project_root, &new_state)?;
    Ok(())
}
```

**Acceptance Criteria**:
- Migration runs automatically on first use
- No user intervention required
- Existing tracking data preserved when possible
- Errors logged but don't block operation

---

#### Task 2.4: Update check_file_changes()
**Estimate**: 30 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 2.1

**Subtasks**:
- [ ] Replace `compute_hash()` calls with `compute_semantic_hash()`
- [ ] Ensure comparison logic unchanged
- [ ] Verify error handling

**Acceptance Criteria**:
- Function uses semantic hashing
- Existing behavior preserved
- No regression in change detection

---

### Phase 3: Integration & Testing (3-4 hours)

#### Task 3.1: Update mise_sync.rs
**Estimate**: 30 minutes  
**Priority**: P1 (Important)  
**Dependencies**: Task 2.3

**Subtasks**:
- [ ] Add migration call at start of sync operations
- [ ] Update any direct hash computations
- [ ] Verify sync logic unchanged

**Implementation Details**:
```rust
pub fn sync_razdfile_to_mise(project_root: &Path) -> Result<SyncResult> {
    // Run migration if needed
    file_tracker::migrate_tracking_state(project_root)?;
    
    // Existing sync logic...
    let changes = file_tracker::check_file_changes(project_root)?;
    // ...
}
```

**Acceptance Criteria**:
- Migration runs before sync checks
- Sync behavior unchanged
- All existing tests pass

---

#### Task 3.2: Create Integration Tests
**Estimate**: 2 hours  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 2.4

**Subtasks**:
- [ ] Create `tests/semantic_detection_test.rs`
- [ ] Test: Formatting changes don't trigger sync
- [ ] Test: Semantic changes do trigger sync
- [ ] Test: Mixed scenarios
- [ ] Test: Migration works correctly

**Test Cases**:
```rust
#[test]
fn test_blank_line_no_sync() {
    // Setup temp directory with Razdfile.yml
    // Compute initial hash
    // Add blank lines
    // Compute new hash
    // Assert hashes match
}

#[test]
fn test_node_version_change_triggers_sync() {
    // Setup with node: '22'
    // Compute initial hash
    // Change to node: '24'
    // Compute new hash
    // Assert hashes differ
}

#[test]
fn test_comment_addition_no_sync() {
    // Add comments to YAML
    // Verify hash unchanged
}

#[test]
fn test_key_reorder_no_sync() {
    // Reorder tools or tasks
    // Verify hash unchanged (canonical form sorts)
}

#[test]
fn test_migration_from_old_format() {
    // Create old tracking state
    // Run migration
    // Verify new format used
}
```

**Acceptance Criteria**:
- All test scenarios pass
- Code coverage > 85% for semantic detection
- Tests run in CI pipeline

---

#### Task 3.3: Manual Testing
**Estimate**: 1 hour  
**Priority**: P1 (Important)  
**Dependencies**: Task 3.1

**Test Scenarios**:
1. **Formatting Test**:
   - Edit Razdfile.yml, add blank lines
   - Run `razd up`
   - Verify no sync prompt

2. **Version Change Test**:
   - Change tool version
   - Run `razd up`
   - Verify sync prompt appears

3. **Migration Test**:
   - Use old version, create tracking state
   - Upgrade to new version
   - Run any command
   - Verify migration happens silently

4. **Edge Cases**:
   - Malformed YAML (verify fallback)
   - Missing files (verify graceful handling)
   - Corrupted tracking state (verify recovery)

**Acceptance Criteria**:
- All scenarios work as expected
- User experience is smooth
- No unexpected errors or warnings

---

### Phase 4: Documentation & Release (2-3 hours)

#### Task 4.1: Update CHANGELOG.md
**Estimate**: 30 minutes  
**Priority**: P1 (Important)  
**Dependencies**: Task 3.3

**Subtasks**:
- [ ] Add section for new version
- [ ] Document semantic change detection feature
- [ ] Mention migration behavior
- [ ] Note any breaking changes (none expected)

**Content**:
```markdown
## [0.3.0] - 2025-10-XX

### Added
- **Semantic change detection**: Sync system now ignores formatting-only changes to YAML/TOML files
  - Adding/removing blank lines no longer triggers sync prompts
  - Comment changes are ignored
  - Indentation changes are ignored
  - Only semantic content changes (versions, tools, tasks) trigger sync

### Changed
- File tracking now uses semantic hashing instead of content hashing
- Tracking state format upgraded to "semantic-v1" (automatic migration)

### Technical
- Added `src/config/canonical.rs` for normalized serialization
- Updated `src/config/file_tracker.rs` with semantic hash computation
- Automatic migration of existing `.razd/tracking.json` files
```

**Acceptance Criteria**:
- CHANGELOG accurately reflects changes
- User-facing improvements clearly described
- Migration behavior documented

---

#### Task 4.2: Update Documentation
**Estimate**: 1 hour  
**Priority**: P2 (Nice to have)  
**Dependencies**: Task 3.3

**Subtasks**:
- [ ] Update README.md if needed
- [ ] Add developer documentation for canonical module
- [ ] Document fallback behavior for parse errors
- [ ] Add troubleshooting section

**Acceptance Criteria**:
- Documentation is clear and complete
- Developers can understand semantic detection
- Users understand behavior changes

---

#### Task 4.3: Performance Testing
**Estimate**: 1 hour  
**Priority**: P2 (Nice to have)  
**Dependencies**: Task 3.2

**Subtasks**:
- [ ] Benchmark semantic hash vs content hash
- [ ] Test with various file sizes
- [ ] Measure impact on `razd up` command
- [ ] Verify < 10ms overhead per file

**Acceptance Criteria**:
- Performance impact documented
- No significant slowdown (< 10ms)
- Benchmarks added to documentation

---

#### Task 4.4: Create Release
**Estimate**: 30 minutes  
**Priority**: P0 (Blocker)  
**Dependencies**: Task 4.1, All P0 tasks

**Subtasks**:
- [ ] Bump version in Cargo.toml (0.3.0)
- [ ] Run full test suite
- [ ] Build release binary
- [ ] Create git tag
- [ ] Push to GitHub
- [ ] Create GitHub release with notes

**Acceptance Criteria**:
- Version bumped correctly
- All tests pass
- Release published
- GitHub release created

---

## Summary

### Total Estimated Time: 12-16 hours

### Priority Breakdown:
- **P0 (Blocker)**: 10-12 hours
- **P1 (Important)**: 2-3 hours
- **P2 (Nice to have)**: 2 hours

### Dependencies Graph:
```
Phase 1: Foundation
├─ 1.1: Create Canonical Module (2h)
│  ├─ 1.2: Implement Razdfile Canonicalization (2h)
│  └─ 1.3: Implement Mise TOML Canonicalization (1h)
└─ 1.4: Add Unit Tests (1h) [depends on 1.2, 1.3]

Phase 2: File Tracker Integration
├─ 2.1: Add Semantic Hash Functions (2h) [depends on 1.4]
├─ 2.2: Add Tracking State Version Field (0.5h)
├─ 2.3: Implement Migration Logic (1.5h) [depends on 2.1, 2.2]
└─ 2.4: Update check_file_changes() (0.5h) [depends on 2.1]

Phase 3: Integration & Testing
├─ 3.1: Update mise_sync.rs (0.5h) [depends on 2.3]
├─ 3.2: Create Integration Tests (2h) [depends on 2.4]
└─ 3.3: Manual Testing (1h) [depends on 3.1]

Phase 4: Documentation & Release
├─ 4.1: Update CHANGELOG.md (0.5h) [depends on 3.3]
├─ 4.2: Update Documentation (1h) [depends on 3.3]
├─ 4.3: Performance Testing (1h) [depends on 3.2]
└─ 4.4: Create Release (0.5h) [depends on 4.1, all P0]
```

### Critical Path:
1.1 → 1.2 → 1.4 → 2.1 → 2.3 → 3.1 → 3.3 → 4.1 → 4.4  
**Total Critical Path Time**: ~10 hours

### Parallel Work Opportunities:
- Task 1.3 (Mise TOML) can run parallel to 1.2 (Razdfile)
- Task 2.2 (Version field) can run parallel to 2.1 (Semantic hash)
- Task 3.2 (Integration tests) can run parallel to 3.1 (mise_sync update)
- Task 4.2 (Docs) and 4.3 (Performance) can run parallel to each other

### Risk Mitigation:
- **Risk**: Parse errors in production files
  - **Mitigation**: Fallback to content hash
  
- **Risk**: Migration fails for some users
  - **Mitigation**: Detailed error logging, recovery mechanism
  
- **Risk**: Performance impact too high
  - **Mitigation**: Benchmarking in Phase 4, optimization if needed

### Success Metrics:
- [ ] All 124+ existing tests pass
- [ ] 10+ new tests for semantic detection pass
- [ ] Zero reported false positives for formatting changes
- [ ] Zero reported false negatives for semantic changes
- [ ] Performance overhead < 10ms per file check
- [ ] Successful migration for all existing users
