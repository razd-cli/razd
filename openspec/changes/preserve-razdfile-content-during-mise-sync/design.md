# Design: Preserve Razdfile Content During Mise Sync

## Current Architecture Issue

The current `sync_mise_to_razdfile()` method in `src/config/mise_sync.rs` follows this problematic flow:

```rust
// 1. Parse mise.toml
let mise_config = self.parse_mise_toml(&toml_content)?;

// 2. Load existing Razdfile OR create new one
let mut razdfile = RazdfileConfig::load_from_path(&razdfile_path)?

// 3. ❌ PROBLEM: Replace entire mise section
razdfile.mise = Some(mise_config);

// 4. ❌ PROBLEM: Reorder tasks alphabetically  
razdfile.tasks = Self::sort_tasks(razdfile.tasks);

// 5. ❌ PROBLEM: Serialize entire structure (loses formatting)
let yaml_content = serde_yaml::to_string(&razdfile)?;
```

**Issues:**
- Deserializing to `RazdfileConfig` loses original YAML structure, comments, and formatting
- Re-serializing creates new formatting that doesn't match original
- `sort_tasks()` unnecessarily reorders tasks
- Platform-specific command metadata gets lost in round-trip

## Proposed Architecture

### Strategy: YAML Document Manipulation

Instead of full deserialization → modification → serialization, use YAML document manipulation:

```rust
// 1. Parse YAML as document (preserves structure)
let mut doc = yaml_rust::YamlLoader::load_from_str(&yaml_content)?;

// 2. Parse mise.toml for new values
let mise_config = self.parse_mise_toml(&toml_content)?;

// 3. ✅ SOLUTION: Update only mise section in document
update_mise_section_in_yaml(&mut doc, mise_config)?;

// 4. ✅ SOLUTION: Write back preserving original structure
write_yaml_document(&doc, &razdfile_path)?;
```

### Implementation Approach

#### Option 1: yaml-rust + Manual Manipulation (Recommended)
- Use `yaml-rust` crate for document parsing that preserves structure
- Manually navigate to `mise:` section in YAML tree
- Replace only that subtree with new mise configuration
- Preserve everything else exactly as-is

#### Option 2: String-based Replacement
- Parse mise.toml for new values
- Use regex/string manipulation to replace only `mise:` section
- More fragile but potentially simpler for this specific case

#### Option 3: Enhanced serde with Custom Serializer
- Create custom YAML serializer that can preserve formatting
- More complex but most robust long-term solution

**Decision: Option 1** - yaml-rust provides good balance of precision and simplicity.

## Technical Implementation Details

### New YAML Manipulation Functions

```rust
// In src/config/yaml_updater.rs (new module)

/// Update only the mise section in a YAML document
fn update_mise_section_in_yaml(
    yaml_content: &str, 
    mise_config: &MiseConfig
) -> Result<String> {
    // 1. Parse as YAML document
    let mut docs = YamlLoader::load_from_str(yaml_content)?;
    let doc = docs.get_mut(0).ok_or("No YAML document found")?;
    
    // 2. Navigate to mise section
    if let Yaml::Hash(ref mut map) = doc {
        // 3. Convert MiseConfig to YAML value
        let mise_yaml = mise_config_to_yaml(mise_config)?;
        
        // 4. Replace mise section only
        map.insert(Yaml::String("mise".to_string()), mise_yaml);
    }
    
    // 5. Serialize back to string
    let mut out_str = String::new();
    YamlEmitter::new(&mut out_str).dump(doc)?;
    Ok(out_str)
}

/// Convert MiseConfig struct to YAML value
fn mise_config_to_yaml(config: &MiseConfig) -> Result<Yaml> {
    // Convert tools, plugins etc. to YAML representation
}
```

### Updated Sync Flow

```rust
// In src/config/mise_sync.rs - updated sync_mise_to_razdfile()

fn sync_mise_to_razdfile(&self) -> Result<SyncResult> {
    let razdfile_path = self.project_root.join("Razdfile.yml");
    let mise_toml_path = self.project_root.join("mise.toml");

    // Parse mise.toml (unchanged)
    let toml_content = fs::read_to_string(&mise_toml_path)?;
    let mise_config = self.parse_mise_toml(&toml_content)?;

    // ✅ NEW: Read Razdfile as raw YAML
    let yaml_content = fs::read_to_string(&razdfile_path)?;
    
    // ✅ NEW: Surgically update only mise section
    let updated_yaml = yaml_updater::update_mise_section_in_yaml(
        &yaml_content, 
        &mise_config
    )?;
    
    // ✅ NEW: Write back preserving structure
    fs::write(&razdfile_path, updated_yaml)?;
    
    // Update tracking (unchanged)
    file_tracker::update_tracking_state(&self.project_root)?;
    
    println!("✓ Synced mise.toml → Razdfile.yml (mise section only)");
    Ok(SyncResult::MiseToRazdfile)
}
```

## Benefits of This Approach

1. **Zero Data Loss**: Platform commands, formatting, comments all preserved
2. **Surgical Updates**: Only mise section changes, everything else untouched  
3. **Predictable Behavior**: Developers know exactly what gets modified
4. **Safer Syncing**: Reduced risk of unintended changes
5. **Better UX**: No surprise formatting changes or lost configuration

## Trade-offs and Considerations

### Pros
- Preserves original document structure and formatting
- Minimal, targeted changes reduce risk
- Platform-specific metadata stays intact
- Developer intent is respected

### Cons  
- More complex implementation than full re-serialization
- Need to handle YAML parsing edge cases
- Additional dependency on YAML manipulation library
- Need comprehensive testing for various YAML structures

### Edge Cases to Handle
- Empty or missing mise section in original Razdfile
- Invalid YAML syntax in original file
- Comments within mise section (may be lost)
- Different YAML formatting styles
- Very complex nested mise configurations

## Testing Strategy

### Test Categories
1. **Basic Preservation**: Simple platform commands preserved
2. **Complex Structure**: Nested configurations, multiple platforms
3. **Formatting Preservation**: Whitespace, indentation, field order
4. **Edge Cases**: Empty sections, missing files, invalid YAML
5. **Integration**: End-to-end sync workflows

### Example Test Case
```rust
#[test]
fn preserves_platform_commands_during_sync() {
    let original_razdfile = r#"
version: '3'
tasks:
  install:
    desc: Install dependencies  
    cmds:
    - cmd: scoop install gcc
      platform: windows
    - mise install
"#;
    
    let mise_toml = r#"
[tools]
node = "latest"
"#;
    
    // Perform sync
    let result = sync_mise_to_razdfile(original_razdfile, mise_toml)?;
    
    // Verify platform command preserved
    assert!(result.contains("platform: windows"));
    assert!(result.contains("scoop install gcc"));
    
    // Verify mise section updated
    assert!(result.contains("node: latest"));
}
```

This design ensures that the mise synchronization becomes a safe, predictable operation that respects the developer's existing Razdfile structure while enabling the intended tooling synchronization.