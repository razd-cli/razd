# Semantic Change Detection for Mise Sync

## Problem Statement

Currently, the mise synchronization system tracks file changes using SHA-256 hashes of the **entire file content**, including all whitespace, indentation, and formatting. This causes unnecessary sync prompts when users make formatting-only changes to YAML files.

### Current Behavior (Problematic)
```yaml
# User adds a blank line on line 7
version: '3'
mise:
  tools:
    node: '22'
    pnpm: latest
    task: latest
                    # <- User adds blank line here
tasks:
  default:
    desc: Set up Node.js project
```

**Result**: System detects file change and prompts for sync, even though **no semantic content changed**.

### Issues
1. **False positives**: Formatting changes trigger sync warnings
2. **Poor UX**: Users interrupted by unnecessary prompts
3. **Workflow disruption**: Manual formatting breaks automation
4. **Trust issues**: Users lose confidence in sync reliability

## Proposed Solution

Implement **semantic change detection** that compares the **parsed structure** of configuration files rather than raw text.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    File Change Detection                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  Load File      │
                    │  Content        │
                    └────────┬────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
    ┌──────────────────┐      ┌──────────────────┐
    │  Parse to        │      │  Parse to        │
    │  Structure       │      │  Structure       │
    │  (YAML/TOML)     │      │  (YAML/TOML)     │
    └────────┬─────────┘      └────────┬─────────┘
             │                         │
             │      Current State      │
             │      from Tracking      │
             └────────────┬────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Serialize to         │
              │  Canonical Form       │
              │  (normalized)         │
              └───────────┬───────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Compare Hashes       │
              │  of Canonical Forms   │
              └───────────┬───────────┘
                          │
                ┌─────────┴──────────┐
                │                    │
                ▼                    ▼
          ┌──────────┐         ┌──────────┐
          │ Changed  │         │ No Change│
          └──────────┘         └──────────┘
```

### Key Changes

#### 1. Canonical Serialization

Instead of hashing raw file content, serialize to a **canonical form**:

**For Razdfile.yml:**
```rust
fn canonicalize_razdfile(config: &RazdfileConfig) -> String {
    // Serialize with consistent formatting
    let mut canonical = String::new();
    
    // Version (always present)
    canonical.push_str(&format!("version:{}\n", config.version));
    
    // Mise section (if present)
    if let Some(ref mise) = config.mise {
        canonical.push_str("mise:\n");
        
        // Tools (sorted by key)
        if let Some(ref tools) = mise.tools {
            canonical.push_str("  tools:\n");
            let mut sorted_tools: Vec<_> = tools.iter().collect();
            sorted_tools.sort_by_key(|(name, _)| *name);
            
            for (name, config) in sorted_tools {
                canonical.push_str(&format!("    {}:{}\n", name, tool_to_string(config)));
            }
        }
        
        // Plugins (sorted by key)
        if let Some(ref plugins) = mise.plugins {
            canonical.push_str("  plugins:\n");
            let mut sorted_plugins: Vec<_> = plugins.iter().collect();
            sorted_plugins.sort_by_key(|(name, _)| *name);
            
            for (name, url) in sorted_plugins {
                canonical.push_str(&format!("    {}:{}\n", name, url));
            }
        }
    }
    
    // Tasks (sorted by key)
    canonical.push_str("tasks:\n");
    let mut sorted_tasks: Vec<_> = config.tasks.iter().collect();
    sorted_tasks.sort_by_key(|(name, _)| *name);
    
    for (name, task) in sorted_tasks {
        canonical.push_str(&format!("  {}:{}\n", name, task_to_string(task)));
    }
    
    canonical
}
```

**For mise.toml:**
```rust
fn canonicalize_mise_toml(config: &MiseConfig) -> String {
    // Serialize TOML to canonical form
    let mut doc = toml_edit::DocumentMut::new();
    
    // Add sections in consistent order
    if let Some(ref tools) = config.tools {
        let mut tools_table = toml_edit::Table::new();
        let mut sorted_tools: Vec<_> = tools.iter().collect();
        sorted_tools.sort_by_key(|(name, _)| *name);
        
        for (name, config) in sorted_tools {
            tools_table.insert(name, serialize_tool_config(config));
        }
        doc.insert("tools", toml_edit::Item::Table(tools_table));
    }
    
    if let Some(ref plugins) = config.plugins {
        let mut plugins_table = toml_edit::Table::new();
        let mut sorted_plugins: Vec<_> = plugins.iter().collect();
        sorted_plugins.sort_by_key(|(name, _)| *name);
        
        for (name, url) in sorted_plugins {
            plugins_table.insert(name, toml_edit::value(url.as_str()));
        }
        doc.insert("plugins", toml_edit::Item::Table(plugins_table));
    }
    
    doc.to_string()
}
```

#### 2. Semantic Comparison

```rust
pub struct SemanticHasher {
    // Track semantic content, not formatting
}

impl SemanticHasher {
    pub fn hash_razdfile(path: &Path) -> Result<String> {
        // 1. Load and parse file
        let config = RazdfileConfig::load_from_path(path)?
            .ok_or_else(|| RazdError::config("File not found"))?;
        
        // 2. Canonicalize structure
        let canonical = canonicalize_razdfile(&config);
        
        // 3. Hash canonical form
        Ok(hash_string(&canonical))
    }
    
    pub fn hash_mise_toml(path: &Path) -> Result<String> {
        // 1. Load and parse file
        let content = fs::read_to_string(path)?;
        let config = parse_mise_toml(&content)?;
        
        // 2. Canonicalize structure
        let canonical = canonicalize_mise_toml(&config);
        
        // 3. Hash canonical form
        Ok(hash_string(&canonical))
    }
}
```

#### 3. Update File Tracker

```rust
// In src/config/file_tracker.rs

pub fn compute_semantic_hash(path: &Path) -> Result<String> {
    let file_name = path.file_name()
        .and_then(|n| n.to_str())
        .ok_or_else(|| RazdError::config("Invalid file path"))?;
    
    match file_name {
        "Razdfile.yml" => SemanticHasher::hash_razdfile(path),
        "mise.toml" => SemanticHasher::hash_mise_toml(path),
        _ => {
            // Fallback to content hash for unknown files
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}

pub fn check_file_changes(project_root: &Path) -> Result<ChangeDetection> {
    let tracking_state = load_tracking_state(project_root)?;
    
    let razdfile_path = project_root.join("Razdfile.yml");
    let mise_toml_path = project_root.join("mise.toml");
    
    let razdfile_changed = if razdfile_path.exists() {
        let current_hash = compute_semantic_hash(&razdfile_path)?;
        tracking_state.razdfile_hash.as_ref() != Some(&current_hash)
    } else {
        false
    };
    
    let mise_toml_changed = if mise_toml_path.exists() {
        let current_hash = compute_semantic_hash(&mise_toml_path)?;
        tracking_state.mise_toml_hash.as_ref() != Some(&current_hash)
    } else {
        false
    };
    
    match (razdfile_changed, mise_toml_changed) {
        (false, false) => Ok(ChangeDetection::NoChanges),
        (true, false) => Ok(ChangeDetection::RazdfileChanged),
        (false, true) => Ok(ChangeDetection::MiseTomlChanged),
        (true, true) => Ok(ChangeDetection::BothChanged),
    }
}
```

## Benefits

### 1. **Formatting Independence**
Users can format YAML/TOML files as they prefer without triggering sync:
- Add/remove blank lines
- Adjust indentation
- Reorder comments
- Change quote styles (`'` vs `"`)

### 2. **Reduced False Positives**
Only **semantic changes** trigger sync:
- Tool version changes
- Added/removed tools or plugins
- Task modifications
- Configuration value updates

### 3. **Better User Experience**
- No interruptions for formatting changes
- Clearer sync prompts (only real changes)
- Increased trust in automation

### 4. **Editor Compatibility**
Works with formatters and linters:
- Prettier
- YAML formatter
- EditorConfig
- Auto-save with formatting

## Implementation Plan

### Phase 1: Canonical Serialization (2-3 hours)
- [ ] Create `src/config/canonical.rs` module
- [ ] Implement `canonicalize_razdfile()`
- [ ] Implement `canonicalize_mise_toml()`
- [ ] Add unit tests for canonical forms

### Phase 2: Semantic Hasher (1-2 hours)
- [ ] Create `SemanticHasher` struct
- [ ] Implement `hash_razdfile()`
- [ ] Implement `hash_mise_toml()`
- [ ] Add integration tests

### Phase 3: Update File Tracker (1 hour)
- [ ] Update `compute_semantic_hash()`
- [ ] Modify `check_file_changes()` to use semantic hashes
- [ ] Update tracking state storage

### Phase 4: Testing (2-3 hours)
- [ ] Test formatting changes (no sync triggered)
- [ ] Test semantic changes (sync triggered correctly)
- [ ] Test mixed scenarios
- [ ] Add regression tests

### Phase 5: Documentation (1 hour)
- [ ] Update CHANGELOG.md
- [ ] Document semantic change detection behavior
- [ ] Add troubleshooting guide

**Total Estimated Effort**: 7-10 hours

## Testing Strategy

### Test Cases

#### 1. Formatting Changes (Should NOT Trigger Sync)
```yaml
# Before
version: '3'
mise:
  tools:
    node: '22'
tasks:
  default:
    desc: Test

# After (added blank line)
version: '3'

mise:
  tools:
    node: '22'

tasks:
  default:
    desc: Test
```
**Expected**: No sync prompt

#### 2. Semantic Changes (Should Trigger Sync)
```yaml
# Before
mise:
  tools:
    node: '22'

# After (version changed)
mise:
  tools:
    node: '24'  # <- Changed
```
**Expected**: Sync prompt

#### 3. Key Order Changes (Should NOT Trigger Sync)
```yaml
# Before
mise:
  tools:
    node: '22'
    pnpm: 'latest'

# After (reordered)
mise:
  tools:
    pnpm: 'latest'
    node: '22'
```
**Expected**: No sync prompt (canonical form sorts keys)

#### 4. Comment Changes (Should NOT Trigger Sync)
```yaml
# Before
mise:
  tools:
    node: '22'

# After (added comment)
mise:
  tools:
    # Node.js LTS version
    node: '22'
```
**Expected**: No sync prompt

## Backwards Compatibility

### Migration Strategy
1. **Detect old tracking format**: Check if tracking state uses old content hashes
2. **Auto-migrate**: On first run, recompute semantic hashes
3. **Log migration**: Inform user of upgrade
4. **No data loss**: Preserve all configuration data

```rust
pub fn migrate_tracking_state(project_root: &Path) -> Result<()> {
    let state = load_tracking_state(project_root)?;
    
    // Check if migration needed (old format detection)
    if state.format_version.as_deref() != Some("semantic-v1") {
        println!("🔄 Upgrading tracking state to semantic change detection...");
        
        // Recompute semantic hashes
        let razdfile_path = project_root.join("Razdfile.yml");
        let mise_toml_path = project_root.join("mise.toml");
        
        let new_state = TrackingState {
            razdfile_hash: if razdfile_path.exists() {
                Some(compute_semantic_hash(&razdfile_path)?)
            } else {
                None
            },
            mise_toml_hash: if mise_toml_path.exists() {
                Some(compute_semantic_hash(&mise_toml_path)?)
            } else {
                None
            },
            format_version: Some("semantic-v1".to_string()),
        };
        
        save_tracking_state(project_root, &new_state)?;
        println!("✓ Tracking state upgraded successfully");
    }
    
    Ok(())
}
```

## Success Criteria

- [ ] Formatting-only changes do NOT trigger sync prompts
- [ ] Semantic changes DO trigger sync prompts correctly
- [ ] All existing tests pass
- [ ] New integration tests verify semantic detection
- [ ] Performance impact < 50ms per file check
- [ ] Zero data loss during migration
- [ ] Documentation updated

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Parsing errors on malformed YAML/TOML | High | Fallback to content hash on parse failure |
| Performance degradation | Medium | Cache canonical forms, optimize serialization |
| Edge cases with complex configs | Medium | Comprehensive test coverage |
| Migration issues | Low | Careful testing, backup tracking state |

## Alternative Approaches Considered

### 1. Whitespace-Normalized Hashing
**Pros**: Simple to implement
**Cons**: Still sensitive to comment changes, quote style changes
**Verdict**: ❌ Not robust enough

### 2. AST Comparison
**Pros**: Most accurate semantic comparison
**Cons**: More complex, requires deep tree traversal
**Verdict**: ⚠️ Overkill for current needs

### 3. Git-Style Diffing
**Pros**: Familiar to developers
**Cons**: Doesn't solve semantic vs formatting distinction
**Verdict**: ❌ Doesn't address core issue

## Future Enhancements

1. **User Configuration**: Allow users to specify what changes are "semantic"
2. **Diff Display**: Show semantic diff when prompting for sync
3. **Auto-Format**: Offer to auto-format files to canonical form
4. **Ignore Patterns**: Let users ignore specific field changes
5. **Merge Strategies**: Smart merging when both files changed semantically

## References

- [YAML Specification](https://yaml.org/spec/1.2/spec.html)
- [TOML Specification](https://toml.io/en/)
- [Semantic Versioning](https://semver.org/)
- [Git Hash-Object](https://git-scm.com/docs/git-hash-object)
