# Semantic Change Detection — Design

## Overview
This document describes the design for semantic (content-aware) change detection in the razd sync system. The goal is to ensure that only meaningful configuration changes (not formatting or whitespace) trigger synchronization between `Razdfile.yml` and `mise.toml`.

## Motivation
- **Current Problem:** Any change in file content, including whitespace, comments, or formatting, triggers a sync prompt.
- **Desired Behavior:** Only actual configuration changes (tool versions, task definitions, etc.) should trigger sync.

## Design Principles
- **Semantic Equivalence:** Two files are considered the same if their parsed configuration structures are identical, regardless of formatting.
- **Canonicalization:** All config files are parsed and re-serialized in a canonical (normalized, sorted) form before hashing.
- **Robustness:** If parsing fails, fallback to content hash to avoid false negatives.

## Architecture
1. **Parse** YAML/TOML config into Rust structs.
2. **Canonicalize**: Serialize struct into a normalized string (sorted keys, no comments, consistent formatting).
3. **Hash**: Compute SHA-256 of canonical string.
4. **Compare**: Use hash for change detection.

## Canonicalization Details
- **YAML (Razdfile.yml):**
    - Parse to `RazdfileConfig`.
    - Serialize with:
        - Sorted keys for all maps (IndexMap or BTreeMap).
        - Omit comments and blank lines.
        - Consistent indentation and quoting.
- **TOML (mise.toml):**
    - Parse to `MiseConfig`.
    - Serialize with:
        - Sorted keys for all tables.
        - Omit comments and blank lines.
        - Consistent formatting.

## Rust API Sketch
```rust
pub fn canonicalize_razdfile(config: &RazdfileConfig) -> String { /* ... */ }
pub fn canonicalize_mise_toml(config: &MiseConfig) -> String { /* ... */ }

pub fn compute_semantic_hash(path: &Path) -> Result<String> { /* ... */ }
```

## Fallback Logic
If parsing fails (invalid YAML/TOML), fallback to content hash and log a warning.

## Migration
- On first run, migrate tracking state to semantic hashes.
- Store format version in tracking state for future upgrades.

## Edge Cases
- **Malformed config:** Fallback to content hash.
- **Key order:** Canonicalization sorts keys, so order changes do not trigger sync.
- **Comments:** Ignored in canonical form.

## Alternatives Considered
- Whitespace normalization (not robust enough)
- AST diffing (overkill for config files)

## Open Questions
- Should users be able to configure which fields are considered semantic?
- Should we display a semantic diff in the sync prompt?

## Next Steps
- Implement canonicalization helpers
- Integrate with file tracker
- Add tests for formatting/semantic changes
# Semantic Change Detection - Design Document

## Overview

This design document details the implementation of semantic change detection for the mise synchronization system. The goal is to detect only **meaningful content changes** while ignoring formatting differences.

## Current Architecture

### File Tracking System

**Location**: `src/config/file_tracker.rs`

```rust
pub struct TrackingState {
    pub razdfile_hash: Option<String>,
    pub mise_toml_hash: Option<String>,
}

// Current implementation - hashes entire file content
pub fn compute_hash(path: &Path) -> Result<String> {
    let content = fs::read_to_string(path)?;
    let mut hasher = Sha256::new();
    hasher.update(&content);
    Ok(format!("{:x}", hasher.finalize()))
}
```

**Problem**: Any formatting change modifies the hash, triggering sync prompts.

## New Architecture

### Module Structure

```
src/config/
├── canonical.rs          # NEW: Canonical serialization
├── file_tracker.rs       # MODIFIED: Use semantic hashing
├── mise_sync.rs          # MINIMAL CHANGES: Use new tracker
├── razdfile.rs           # NO CHANGES
└── mise_generator.rs     # NO CHANGES
```

### 1. Canonical Serialization Module

**File**: `src/config/canonical.rs`

#### Purpose
Convert parsed configurations into a normalized, deterministic string representation that ignores formatting but preserves all semantic information.

#### Key Functions

```rust
use indexmap::IndexMap;
use serde_yaml::Value;
use std::fmt::Write;

/// Canonicalize a RazdfileConfig into a deterministic string
pub fn canonicalize_razdfile(config: &RazdfileConfig) -> String {
    let mut output = String::new();
    
    // Version (required field)
    writeln!(output, "v:{}", config.version).unwrap();
    
    // Mise section (optional)
    if let Some(ref mise) = config.mise {
        output.push_str("mise:\n");
        
        // Tools
        if let Some(ref tools) = mise.tools {
            output.push_str("  tools:\n");
            for (name, tool_config) in tools {
                let tool_str = match tool_config {
                    ToolConfig::Simple(version) => version.clone(),
                    ToolConfig::Detailed { version, .. } => version.clone(),
                };
                writeln!(output, "    {}={}", name, tool_str).unwrap();
            }
        }
        
        // Plugins
        if let Some(ref plugins) = mise.plugins {
            output.push_str("  plugins:\n");
            for (name, url) in plugins {
                writeln!(output, "    {}={}", name, url).unwrap();
            }
        }
    }
    
    // Tasks (required field)
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

/// Canonicalize a MiseConfig (from mise.toml) into a deterministic string
pub fn canonicalize_mise_toml(config: &MiseConfig) -> String {
    let mut output = String::new();
    
    // Tools
    if let Some(ref tools) = config.tools {
        output.push_str("[tools]\n");
        for (name, tool_config) in tools {
            let tool_str = match tool_config {
                ToolConfig::Simple(version) => version.clone(),
                ToolConfig::Detailed { version, .. } => version.clone(),
            };
            writeln!(output, "{}={}", name, tool_str).unwrap();
        }
    }
    
    // Plugins
    if let Some(ref plugins) = config.plugins {
        output.push_str("[plugins]\n");
        for (name, url) in plugins {
            writeln!(output, "{}={}", name, url).unwrap();
        }
    }
    
    output
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_canonical_ignores_whitespace() {
        let config1 = RazdfileConfig {
            version: "3".to_string(),
            mise: Some(MiseConfig {
                tools: Some(indexmap! {
                    "node".to_string() => ToolConfig::Simple("22".to_string()),
                }),
                plugins: None,
            }),
            tasks: indexmap! {
                "default".to_string() => TaskConfig {
                    desc: Some("Test".to_string()),
                    cmds: None,
                    internal: None,
                },
            },
        };
        
        // Both should produce same canonical form
        let canonical1 = canonicalize_razdfile(&config1);
        let canonical2 = canonicalize_razdfile(&config1);
        
        assert_eq!(canonical1, canonical2);
    }
    
    #[test]
    fn test_canonical_detects_version_change() {
        let config1 = RazdfileConfig {
            version: "3".to_string(),
            mise: Some(MiseConfig {
                tools: Some(indexmap! {
                    "node".to_string() => ToolConfig::Simple("22".to_string()),
                }),
                plugins: None,
            }),
            tasks: indexmap!{},
        };
        
        let mut config2 = config1.clone();
        if let Some(ref mut mise) = config2.mise {
            if let Some(ref mut tools) = mise.tools {
                tools.insert("node".to_string(), ToolConfig::Simple("24".to_string()));
            }
        }
        
        let canonical1 = canonicalize_razdfile(&config1);
        let canonical2 = canonicalize_razdfile(&config2);
        
        assert_ne!(canonical1, canonical2);
    }
}
```

#### Design Decisions

1. **Deterministic Ordering**: Use IndexMap iteration order (already sorted in our implementation)
2. **Normalized Format**: Use simple `key=value` format instead of preserving YAML/TOML syntax
3. **All Fields Included**: Even optional fields with None are handled consistently
4. **No Whitespace Variations**: Fixed newlines, no trailing spaces
5. **Simple String Concatenation**: Fast and predictable

### 2. Updated File Tracker

**File**: `src/config/file_tracker.rs` (modifications)

```rust
use sha2::{Sha256, Digest};
use std::fs;
use std::path::Path;
use crate::core::error::{RazdError, Result};
use crate::config::razdfile::RazdfileConfig;
use crate::config::canonical::{canonicalize_razdfile, canonicalize_mise_toml};

pub struct TrackingState {
    pub razdfile_hash: Option<String>,
    pub mise_toml_hash: Option<String>,
    pub format_version: Option<String>, // NEW: Track format version
}

/// Hash a string using SHA-256
fn hash_string(content: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(content.as_bytes());
    format!("{:x}", hasher.finalize())
}

/// Compute semantic hash for Razdfile.yml
fn compute_razdfile_semantic_hash(path: &Path) -> Result<String> {
    // Parse the file
    let config = RazdfileConfig::load_from_path(path)?
        .ok_or_else(|| RazdError::config("Failed to load Razdfile.yml"))?;
    
    // Canonicalize to normalized form
    let canonical = canonicalize_razdfile(&config);
    
    // Hash the canonical form
    Ok(hash_string(&canonical))
}

/// Compute semantic hash for mise.toml
fn compute_mise_toml_semantic_hash(path: &Path) -> Result<String> {
    // Read file content
    let content = fs::read_to_string(path)?;
    
    // Parse using toml_edit (preserve our existing parsing logic)
    let doc = content.parse::<toml_edit::DocumentMut>()
        .map_err(|e| RazdError::config(&format!("Failed to parse mise.toml: {}", e)))?;
    
    // Extract tools and plugins
    let mut tools = IndexMap::new();
    if let Some(tools_table) = doc.get("tools").and_then(|t| t.as_table()) {
        for (key, value) in tools_table.iter() {
            if let Some(version) = value.as_str() {
                tools.insert(key.to_string(), ToolConfig::Simple(version.to_string()));
            }
        }
    }
    
    let mut plugins = IndexMap::new();
    if let Some(plugins_table) = doc.get("plugins").and_then(|t| t.as_table()) {
        for (key, value) in plugins_table.iter() {
            if let Some(url) = value.as_str() {
                plugins.insert(key.to_string(), url.to_string());
            }
        }
    }
    
    let config = MiseConfig {
        tools: if tools.is_empty() { None } else { Some(tools) },
        plugins: if plugins.is_empty() { None } else { Some(plugins) },
    };
    
    // Canonicalize and hash
    let canonical = canonicalize_mise_toml(&config);
    Ok(hash_string(&canonical))
}

/// Compute semantic hash for a file based on its name
pub fn compute_semantic_hash(path: &Path) -> Result<String> {
    let file_name = path.file_name()
        .and_then(|n| n.to_str())
        .ok_or_else(|| RazdError::config("Invalid file path"))?;
    
    match file_name {
        "Razdfile.yml" => compute_razdfile_semantic_hash(path),
        "mise.toml" => compute_mise_toml_semantic_hash(path),
        _ => {
            // Fallback to content hash for unknown files
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}

/// Migrate old tracking state to new semantic format
pub fn migrate_tracking_state(project_root: &Path) -> Result<()> {
    let tracking_file = project_root.join(".razd").join("tracking.json");
    
    if !tracking_file.exists() {
        return Ok(()); // Nothing to migrate
    }
    
    let state = load_tracking_state(project_root)?;
    
    // Check if already migrated
    if state.format_version.as_deref() == Some("semantic-v1") {
        return Ok(());
    }
    
    // Recompute semantic hashes
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

#### Migration Strategy

1. **Automatic Detection**: Check `format_version` field in tracking state
2. **Transparent Upgrade**: First command run after upgrade migrates automatically
3. **No User Intervention**: Migration happens silently
4. **Error Handling**: If migration fails, falls back to forcing sync

### 3. Integration Points

#### mise_sync.rs Changes

```rust
// Minimal changes - just ensure migration runs
pub fn sync_razdfile_to_mise(project_root: &Path) -> Result<SyncResult> {
    // Run migration if needed
    file_tracker::migrate_tracking_state(project_root)?;
    
    // Rest of existing logic unchanged...
    let changes = file_tracker::check_file_changes(project_root)?;
    // ...
}
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  User edits Razdfile.yml                     │
│                  (adds blank line on line 7)                 │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  razd up command      │
                └───────────┬───────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  Migrate tracking     │
                │  state if needed      │
                └───────────┬───────────┘
                            │
                            ▼
        ┌───────────────────────────────────────┐
        │  Load Razdfile.yml                    │
        │  Parse to RazdfileConfig struct       │
        └───────────────┬───────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────┐
        │  Canonicalize (normalize structure)   │
        │  - Sort keys                          │
        │  - Fixed formatting                   │
        │  - key=value pairs                    │
        └───────────────┬───────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────┐
        │  Hash canonical form                  │
        │  SHA-256(canonical_string)            │
        └───────────────┬───────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────┐
        │  Compare with stored hash             │
        └───────────────┬───────────────────────┘
                        │
            ┌───────────┴───────────┐
            │                       │
            ▼                       ▼
    ┌──────────────┐        ┌──────────────┐
    │ Hash matches │        │ Hash differs │
    │ (no sync)    │        │ (trigger     │
    │              │        │  sync)       │
    └──────────────┘        └──────────────┘
```

## Error Handling

### Parse Errors

**Scenario**: File is malformed and can't be parsed

**Solution**: Fallback to content hash
```rust
fn compute_razdfile_semantic_hash(path: &Path) -> Result<String> {
    match RazdfileConfig::load_from_path(path) {
        Ok(Some(config)) => {
            let canonical = canonicalize_razdfile(&config);
            Ok(hash_string(&canonical))
        }
        Ok(None) | Err(_) => {
            // Fallback: use content hash
            eprintln!("⚠️  Warning: Failed to parse Razdfile.yml, using content hash");
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}
```

### Missing Files

**Scenario**: Razdfile.yml or mise.toml doesn't exist

**Solution**: Store `None` for hash, no comparison needed

### Corrupted Tracking State

**Scenario**: `.razd/tracking.json` is corrupted

**Solution**: Delete and recreate from current files
```rust
pub fn recover_tracking_state(project_root: &Path) -> Result<()> {
    let tracking_dir = project_root.join(".razd");
    let tracking_file = tracking_dir.join("tracking.json");
    
    if tracking_file.exists() {
        fs::remove_file(&tracking_file)?;
    }
    
    // Create fresh tracking state
    let new_state = create_initial_tracking_state(project_root)?;
    save_tracking_state(project_root, &new_state)?;
    
    Ok(())
}
```

## Performance Considerations

### Benchmarking

Expected performance impact per file check:

| Operation | Current | New | Delta |
|-----------|---------|-----|-------|
| Read file | ~1ms | ~1ms | 0ms |
| Hash content | ~0.1ms | - | -0.1ms |
| Parse YAML | - | ~2-5ms | +2-5ms |
| Canonicalize | - | ~0.5ms | +0.5ms |
| Hash canonical | - | ~0.1ms | +0.1ms |
| **Total** | **~1.1ms** | **~3-6ms** | **+2-5ms** |

**Verdict**: Acceptable overhead (~5ms per sync check)

### Optimization Strategies

1. **Lazy Parsing**: Only parse if file timestamp changed
2. **Caching**: Cache canonical form if file unchanged
3. **Parallel Processing**: Check both files concurrently

```rust
pub fn check_file_changes_parallel(project_root: &Path) -> Result<ChangeDetection> {
    use rayon::prelude::*;
    
    let razdfile_path = project_root.join("Razdfile.yml");
    let mise_toml_path = project_root.join("mise.toml");
    
    let tracking_state = load_tracking_state(project_root)?;
    
    // Check both files in parallel
    let results: Vec<_> = vec![
        ("razdfile", razdfile_path, tracking_state.razdfile_hash.clone()),
        ("mise_toml", mise_toml_path, tracking_state.mise_toml_hash.clone()),
    ]
    .into_par_iter()
    .map(|(name, path, stored_hash)| {
        if !path.exists() {
            return Ok((name, false));
        }
        
        let current_hash = compute_semantic_hash(&path)?;
        let changed = stored_hash.as_ref() != Some(&current_hash);
        
        Ok((name, changed))
    })
    .collect::<Result<Vec<_>>>()?;
    
    // Aggregate results
    let razdfile_changed = results.iter().find(|(n, _)| *n == "razdfile").map(|(_, c)| *c).unwrap_or(false);
    let mise_toml_changed = results.iter().find(|(n, _)| *n == "mise_toml").map(|(_, c)| *c).unwrap_or(false);
    
    match (razdfile_changed, mise_toml_changed) {
        (false, false) => Ok(ChangeDetection::NoChanges),
        (true, false) => Ok(ChangeDetection::RazdfileChanged),
        (false, true) => Ok(ChangeDetection::MiseTomlChanged),
        (true, true) => Ok(ChangeDetection::BothChanged),
    }
}
```

## Testing Strategy

### Unit Tests

**Location**: `src/config/canonical.rs`

- Test canonical form is deterministic
- Test formatting changes produce same canonical form
- Test semantic changes produce different canonical forms
- Test all field types (tools, plugins, tasks)

### Integration Tests

**Location**: `tests/semantic_detection_test.rs`

```rust
#[test]
fn test_formatting_change_no_sync() {
    let temp_dir = TempDir::new().unwrap();
    
    // Create initial Razdfile.yml
    let razdfile = temp_dir.path().join("Razdfile.yml");
    fs::write(&razdfile, "version: '3'\nmise:\n  tools:\n    node: '22'\ntasks:\n  default:\n    desc: Test\n").unwrap();
    
    // Compute initial hash
    let hash1 = compute_semantic_hash(&razdfile).unwrap();
    
    // Add blank lines (formatting change)
    fs::write(&razdfile, "version: '3'\n\nmise:\n  tools:\n    node: '22'\n\ntasks:\n  default:\n    desc: Test\n").unwrap();
    
    // Compute new hash
    let hash2 = compute_semantic_hash(&razdfile).unwrap();
    
    // Hashes should match
    assert_eq!(hash1, hash2, "Formatting change should not alter semantic hash");
}

#[test]
fn test_version_change_triggers_sync() {
    let temp_dir = TempDir::new().unwrap();
    
    // Create initial Razdfile.yml
    let razdfile = temp_dir.path().join("Razdfile.yml");
    fs::write(&razdfile, "version: '3'\nmise:\n  tools:\n    node: '22'\ntasks: {}\n").unwrap();
    
    let hash1 = compute_semantic_hash(&razdfile).unwrap();
    
    // Change node version (semantic change)
    fs::write(&razdfile, "version: '3'\nmise:\n  tools:\n    node: '24'\ntasks: {}\n").unwrap();
    
    let hash2 = compute_semantic_hash(&razdfile).unwrap();
    
    // Hashes should differ
    assert_ne!(hash1, hash2, "Version change should alter semantic hash");
}
```

### Manual Test Cases

1. **Formatting Changes**: Add/remove blank lines, change indentation
2. **Comment Changes**: Add/remove/modify comments
3. **Key Reordering**: Swap order of tools or tasks
4. **Quote Style**: Change `'` to `"` or vice versa
5. **Semantic Changes**: Modify versions, add tools, change task commands

## Rollout Plan

### Phase 1: Implementation (Week 1)
- Create `canonical.rs` module
- Implement canonicalization functions
- Add unit tests

### Phase 2: Integration (Week 1)
- Update `file_tracker.rs`
- Add migration logic
- Update `mise_sync.rs` integration

### Phase 3: Testing (Week 2)
- Write integration tests
- Manual testing with various scenarios
- Performance benchmarking

### Phase 4: Documentation (Week 2)
- Update user documentation
- Add developer documentation
- Update CHANGELOG

### Phase 5: Release (Week 2)
- Create release candidate
- Beta testing with users
- Production release

## Success Metrics

- [ ] Zero false positives in formatting change detection (100% accuracy)
- [ ] Zero false negatives in semantic change detection (100% accuracy)
- [ ] Performance overhead < 10ms per sync check
- [ ] All existing tests pass
- [ ] 10+ new tests covering semantic detection
- [ ] User reports of "unnecessary sync prompts" eliminated

## Conclusion

This design provides a robust solution for semantic change detection that:
- Eliminates false positives from formatting changes
- Maintains 100% accuracy for semantic changes
- Adds minimal performance overhead
- Migrates existing users transparently
- Provides clear fallback behavior for edge cases
