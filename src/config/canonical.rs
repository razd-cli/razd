use crate::config::razdfile::{Command, MiseConfig, RazdfileConfig, TaskConfig, ToolConfig};
use crate::core::RazdError;
use indexmap::IndexMap;
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::fmt::Write;
use std::fs;
use std::path::Path;

/// Canonicalizes a RazdfileConfig into a deterministic string representation
/// This ignores formatting differences like whitespace, comments, and key order
pub fn canonicalize_razdfile(config: &RazdfileConfig) -> String {
    let mut output = String::new();
    
    // Version (always present)
    writeln!(output, "v:{}", config.version).unwrap();
    
    // Mise section (if present)
    if let Some(ref mise) = config.mise {
        writeln!(output, "mise:").unwrap();
        
        // Tools (sorted by key for determinism)
        if let Some(ref tools) = mise.tools {
            writeln!(output, "  tools:").unwrap();
            let mut sorted_tools: Vec<_> = tools.iter().collect();
            sorted_tools.sort_by_key(|(name, _)| *name);
            
            for (name, tool_config) in sorted_tools {
                let tool_str = match tool_config {
                    ToolConfig::Simple(version) => version.clone(),
                    ToolConfig::Complex { version, .. } => version.clone(),
                };
                writeln!(output, "    {}:{}", name, tool_str).unwrap();
            }
        }
        
        // Plugins (sorted by key)
        if let Some(ref plugins) = mise.plugins {
            writeln!(output, "  plugins:").unwrap();
            let mut sorted_plugins: Vec<_> = plugins.iter().collect();
            sorted_plugins.sort_by_key(|(name, _)| *name);
            
            for (name, url) in sorted_plugins {
                writeln!(output, "    {}:{}", name, url).unwrap();
            }
        }
    }
    
    // Tasks (sorted by key)
    writeln!(output, "tasks:").unwrap();
    let mut sorted_tasks: Vec<_> = config.tasks.iter().collect();
    sorted_tasks.sort_by_key(|(name, _)| *name);
    
    for (name, task) in sorted_tasks {
        writeln!(output, "  {}:", name).unwrap();
        
        if let Some(ref desc) = task.desc {
            writeln!(output, "    desc:{}", desc).unwrap();
        }
        
        writeln!(output, "    cmds:").unwrap();
        for cmd in &task.cmds {
            match cmd {
                Command::String(s) => {
                    writeln!(output, "      -{}", s).unwrap();
                }
                Command::TaskRef { task, vars } => {
                    write!(output, "      -task:{}", task).unwrap();
                    if let Some(ref vars) = vars {
                        // Sort vars for determinism
                        let sorted_vars: BTreeMap<_, _> = vars.iter().collect();
                        write!(output, ",vars:{{").unwrap();
                        for (i, (k, v)) in sorted_vars.iter().enumerate() {
                            if i > 0 {
                                write!(output, ",").unwrap();
                            }
                            write!(output, "{}:{}", k, v).unwrap();
                        }
                        write!(output, "}}").unwrap();
                    }
                    writeln!(output).unwrap();
                }
            }
        }
        
        if task.internal {
            writeln!(output, "    internal:true").unwrap();
        }
    }
    
    output
}

/// Canonicalizes a MiseConfig (from mise.toml) into a deterministic string
pub fn canonicalize_mise_toml(config: &MiseConfig) -> String {
    let mut output = String::new();
    
    // Tools section (sorted by key)
    if let Some(ref tools) = config.tools {
        writeln!(output, "[tools]").unwrap();
        let mut sorted_tools: Vec<_> = tools.iter().collect();
        sorted_tools.sort_by_key(|(name, _)| *name);
        
        for (name, tool_config) in sorted_tools {
            let tool_str = match tool_config {
                ToolConfig::Simple(version) => version.clone(),
                ToolConfig::Complex { version, .. } => version.clone(),
            };
            writeln!(output, "{}={}", name, tool_str).unwrap();
        }
    }
    
    // Plugins section (sorted by key)
    if let Some(ref plugins) = config.plugins {
        writeln!(output, "[plugins]").unwrap();
        let mut sorted_plugins: Vec<_> = plugins.iter().collect();
        sorted_plugins.sort_by_key(|(name, _)| *name);
        
        for (name, url) in sorted_plugins {
            writeln!(output, "{}={}", name, url).unwrap();
        }
    }
    
    output
}

/// Computes SHA-256 hash of a string
fn hash_string(content: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(content.as_bytes());
    format!("{:x}", hasher.finalize())
}

/// Computes semantic hash for Razdfile.yml
/// Falls back to content hash if parsing fails
pub fn compute_razdfile_semantic_hash(path: &Path) -> crate::core::Result<String> {
    match RazdfileConfig::load_from_path(path)? {
        Some(config) => {
            let canonical = canonicalize_razdfile(&config);
            Ok(hash_string(&canonical))
        }
        None => {
            // Fallback to content hash
            eprintln!("Warning: Failed to parse Razdfile.yml, using content hash");
            let content = fs::read_to_string(path)?;
            Ok(hash_string(&content))
        }
    }
}

/// Computes semantic hash for mise.toml
/// Falls back to content hash if parsing fails
pub fn compute_mise_toml_semantic_hash(path: &Path) -> crate::core::Result<String> {
    // Parse mise.toml and extract tools/plugins
    let content = fs::read_to_string(path)?;
    
    match parse_mise_toml(&content) {
        Ok(config) => {
            let canonical = canonicalize_mise_toml(&config);
            Ok(hash_string(&canonical))
        }
        Err(_) => {
            // Fallback to content hash
            eprintln!("Warning: Failed to parse mise.toml, using content hash");
            Ok(hash_string(&content))
        }
    }
}

/// Parses mise.toml content into MiseConfig
fn parse_mise_toml(content: &str) -> crate::core::Result<MiseConfig> {
    let doc: toml_edit::DocumentMut = content.parse()
        .map_err(|e| RazdError::config(&format!("Failed to parse mise.toml: {}", e)))?;
    
    let mut tools = None;
    let mut plugins = None;
    
    // Extract [tools] section
    if let Some(tools_table) = doc.get("tools").and_then(|t| t.as_table()) {
        let mut tools_map = IndexMap::new();
        for (name, value) in tools_table.iter() {
            if let Some(version) = value.as_str() {
                tools_map.insert(name.to_string(), ToolConfig::Simple(version.to_string()));
            }
        }
        if !tools_map.is_empty() {
            tools = Some(tools_map);
        }
    }
    
    // Extract [plugins] section
    if let Some(plugins_table) = doc.get("plugins").and_then(|t| t.as_table()) {
        let mut plugins_map = IndexMap::new();
        for (name, value) in plugins_table.iter() {
            if let Some(url) = value.as_str() {
                plugins_map.insert(name.to_string(), url.to_string());
            }
        }
        if !plugins_map.is_empty() {
            plugins = Some(plugins_map);
        }
    }
    
    Ok(MiseConfig { tools, plugins })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_canonicalize_razdfile_basic() {
        let config = RazdfileConfig {
            version: "3".to_string(),
            mise: Some(MiseConfig {
                tools: Some({
                    let mut map = IndexMap::new();
                    map.insert("node".to_string(), ToolConfig::Simple("22".to_string()));
                    map.insert("pnpm".to_string(), ToolConfig::Simple("latest".to_string()));
                    map
                }),
                plugins: None,
            }),
            tasks: {
                let mut map = IndexMap::new();
                map.insert(
                    "default".to_string(),
                    TaskConfig {
                        desc: Some("Test task".to_string()),
                        cmds: vec![Command::String("echo test".to_string())],
                        internal: false, // Keep explicit false for testing
                    },
                );
                map
            },
        };

        let canonical = canonicalize_razdfile(&config);
        
        assert!(canonical.contains("v:3"));
        assert!(canonical.contains("mise:"));
        assert!(canonical.contains("node:22"));
        assert!(canonical.contains("pnpm:latest"));
        assert!(canonical.contains("tasks:"));
        assert!(canonical.contains("default:"));
    }

    #[test]
    fn test_canonicalize_sorts_keys() {
        let mut tools1 = IndexMap::new();
        tools1.insert("zsh".to_string(), ToolConfig::Simple("1.0".to_string()));
        tools1.insert("node".to_string(), ToolConfig::Simple("22".to_string()));
        
        let mut tools2 = IndexMap::new();
        tools2.insert("node".to_string(), ToolConfig::Simple("22".to_string()));
        tools2.insert("zsh".to_string(), ToolConfig::Simple("1.0".to_string()));
        
        let config1 = RazdfileConfig {
            version: "3".to_string(),
            mise: Some(MiseConfig {
                tools: Some(tools1),
                plugins: None,
            }),
            tasks: IndexMap::new(),
        };
        
        let config2 = RazdfileConfig {
            version: "3".to_string(),
            mise: Some(MiseConfig {
                tools: Some(tools2),
                plugins: None,
            }),
            tasks: IndexMap::new(),
        };
        
        // Canonical forms should be identical despite different insertion order
        assert_eq!(canonicalize_razdfile(&config1), canonicalize_razdfile(&config2));
    }

    #[test]
    fn test_hash_string() {
        let hash1 = hash_string("test");
        let hash2 = hash_string("test");
        let hash3 = hash_string("different");
        
        assert_eq!(hash1, hash2);
        assert_ne!(hash1, hash3);
        assert_eq!(hash1.len(), 64); // SHA-256 produces 64 hex characters
    }
}
