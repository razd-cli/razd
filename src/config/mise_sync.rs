use crate::config::file_tracker::{self, ChangeDetection};
use crate::config::mise_generator::generate_mise_toml;
use crate::config::razdfile::RazdfileConfig;
use crate::core::{RazdError, Result};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

/// Configuration for sync behavior
#[derive(Debug, Clone)]
pub struct SyncConfig {
    /// Skip all sync operations
    pub no_sync: bool,
    /// Auto-approve all sync operations without prompts
    pub auto_approve: bool,
    /// Create backups before overwriting files
    pub create_backups: bool,
}

impl Default for SyncConfig {
    fn default() -> Self {
        Self {
            no_sync: false,
            auto_approve: false,
            create_backups: true,
        }
    }
}

/// Result of a sync operation
#[derive(Debug, Clone, PartialEq)]
pub enum SyncResult {
    /// No sync was needed
    NoChangesNeeded,
    /// Razdfile was synced to mise.toml
    RazdfileToMise,
    /// mise.toml was synced to Razdfile
    MiseToRazdfile,
    /// Sync was skipped (no-sync flag or user declined)
    Skipped,
    /// Conflict detected, user needs to resolve manually
    Conflict,
}

/// Sync manager for coordinating file synchronization
pub struct MiseSyncManager {
    project_root: PathBuf,
    config: SyncConfig,
}

impl MiseSyncManager {
    pub fn new(project_root: PathBuf, config: SyncConfig) -> Self {
        Self {
            project_root,
            config,
        }
    }

    /// Main sync entry point - checks for changes and performs sync if needed
    pub fn check_and_sync_if_needed(&self) -> Result<SyncResult> {
        if self.config.no_sync {
            return Ok(SyncResult::Skipped);
        }

        let razdfile_path = self.project_root.join("Razdfile.yml");
        let mise_toml_path = self.project_root.join("mise.toml");

        // If neither file exists, nothing to sync
        if !razdfile_path.exists() && !mise_toml_path.exists() {
            return Ok(SyncResult::NoChangesNeeded);
        }

        let changes = file_tracker::check_file_changes(&self.project_root)?;

        match changes {
            ChangeDetection::NoChanges => Ok(SyncResult::NoChangesNeeded),
            ChangeDetection::RazdfileChanged => self.sync_razdfile_to_mise(),
            ChangeDetection::MiseTomlChanged => self.sync_mise_to_razdfile(),
            ChangeDetection::BothChanged => self.handle_conflict(),
        }
    }

    /// Sync Razdfile.yml mise config to mise.toml
    fn sync_razdfile_to_mise(&self) -> Result<SyncResult> {
        let razdfile_path = self.project_root.join("Razdfile.yml");
        let mise_toml_path = self.project_root.join("mise.toml");

        // Load Razdfile
        let razdfile = RazdfileConfig::load_from_path(&razdfile_path)?
            .ok_or_else(|| RazdError::config("Razdfile.yml not found"))?;

        // Check if there's mise config to sync
        let mise_config = match &razdfile.mise {
            Some(config) => config,
            None => {
                // No mise config in Razdfile, but mise.toml exists
                if mise_toml_path.exists() {
                    if !self.config.auto_approve {
                        println!("⚠️  Razdfile.yml has no mise config, but mise.toml exists.");
                        println!("   Sync mise.toml → Razdfile.yml? [Y/n]");
                        
                        if !self.prompt_user_approval()? {
                            return Ok(SyncResult::Skipped);
                        }
                    }
                    // Sync mise.toml to Razdfile instead
                    return self.sync_mise_to_razdfile();
                }
                // No mise config anywhere, nothing to sync
                return Ok(SyncResult::NoChangesNeeded);
            }
        };

        // Generate mise.toml content
        let toml_content = generate_mise_toml(mise_config)?;

        // Ask user about backup if file exists
        if mise_toml_path.exists() {
            if self.config.create_backups {
                if !self.config.auto_approve {
                    println!("⚠️  mise.toml will be overwritten. Create backup? [Y/n]");
                    if self.prompt_user_approval()? {
                        self.create_backup(&mise_toml_path)?;
                    }
                } else {
                    // Auto-approve: create backup without prompt
                    self.create_backup(&mise_toml_path)?;
                }
            }
        }

        // Write mise.toml
        let mut file = fs::File::create(&mise_toml_path)?;
        file.write_all(toml_content.as_bytes())?;

        // Update tracking state
        file_tracker::update_tracking_state(&self.project_root)?;

        println!("✓ Synced Razdfile.yml → mise.toml");
        Ok(SyncResult::RazdfileToMise)
    }

    /// Sync mise.toml back to Razdfile.yml
    fn sync_mise_to_razdfile(&self) -> Result<SyncResult> {
        let razdfile_path = self.project_root.join("Razdfile.yml");
        let mise_toml_path = self.project_root.join("mise.toml");

        // Parse mise.toml
        let toml_content = fs::read_to_string(&mise_toml_path)
            .map_err(|e| RazdError::config(format!("Failed to read mise.toml: {}", e)))?;
        
        let mise_config = self.parse_mise_toml(&toml_content)?;

        // Load or create Razdfile
        let mut razdfile = if razdfile_path.exists() {
            RazdfileConfig::load_from_path(&razdfile_path)?
                .ok_or_else(|| RazdError::config("Failed to load Razdfile.yml"))?
        } else {
            // Prompt user before creating new Razdfile
            if !self.config.auto_approve {
                println!("⚠️  Razdfile.yml does not exist. Create it with mise config? [Y/n]");
                if !self.prompt_user_approval()? {
                    return Ok(SyncResult::Skipped);
                }
            }
            // Create minimal Razdfile
            RazdfileConfig {
                version: "3".to_string(),
                tasks: std::collections::HashMap::new(),
                mise: None,
            }
        };

        // Update mise config
        razdfile.mise = Some(mise_config);

        // Ask user about backup if file exists
        if razdfile_path.exists() {
            if self.config.create_backups {
                if !self.config.auto_approve {
                    println!("⚠️  Razdfile.yml will be modified. Create backup? [Y/n]");
                    if self.prompt_user_approval()? {
                        self.create_backup(&razdfile_path)?;
                    }
                } else {
                    // Auto-approve: create backup without prompt
                    self.create_backup(&razdfile_path)?;
                }
            }
        }

        // Write Razdfile
        let yaml_content = serde_yaml::to_string(&razdfile)
            .map_err(|e| RazdError::config(format!("Failed to serialize Razdfile: {}", e)))?;
        
        // Format YAML with better spacing
        let formatted_yaml = self.format_yaml(&yaml_content);
        
        let mut file = fs::File::create(&razdfile_path)?;
        file.write_all(formatted_yaml.as_bytes())?;

        // Update tracking state
        file_tracker::update_tracking_state(&self.project_root)?;

        println!("✓ Synced mise.toml → Razdfile.yml");
        Ok(SyncResult::MiseToRazdfile)
    }

    /// Handle conflict when both files changed
    fn handle_conflict(&self) -> Result<SyncResult> {
        println!("⚠️  Conflict detected: Both Razdfile.yml and mise.toml have been modified.");
        
        if self.config.auto_approve {
            println!("   Auto-approve enabled: preferring Razdfile.yml as source of truth.");
            return self.sync_razdfile_to_mise();
        }

        println!("\nOptions:");
        println!("  1) Use Razdfile.yml (overwrite mise.toml)");
        println!("  2) Use mise.toml (update Razdfile.yml)");
        println!("  3) Skip sync (resolve manually)");
        println!("\nYour choice [1-3]:");

        // Read user input
        let mut input = String::new();
        std::io::stdin()
            .read_line(&mut input)
            .map_err(|e| RazdError::config(format!("Failed to read input: {}", e)))?;

        match input.trim() {
            "1" => self.sync_razdfile_to_mise(),
            "2" => self.sync_mise_to_razdfile(),
            _ => {
                println!("Skipping sync. Please resolve manually.");
                Ok(SyncResult::Conflict)
            }
        }
    }

    /// Create backup of a file
    fn create_backup(&self, file_path: &Path) -> Result<()> {
        if !file_path.exists() {
            return Ok(());
        }

        let backup_path = file_path.with_extension(
            format!("{}.backup", file_path.extension().and_then(|s| s.to_str()).unwrap_or(""))
        );

        fs::copy(file_path, &backup_path)
            .map_err(|e| RazdError::config(format!("Failed to create backup: {}", e)))?;

        println!("  Created backup: {}", backup_path.display());
        Ok(())
    }

    /// Parse mise.toml into MiseConfig structure
    fn parse_mise_toml(&self, toml_content: &str) -> Result<crate::config::razdfile::MiseConfig> {
        use crate::config::razdfile::{MiseConfig, ToolConfig};
        use std::collections::HashMap;

        let doc: toml::Value = toml::from_str(toml_content)
            .map_err(|e| RazdError::config(format!("Invalid mise.toml: {}", e)))?;

        let mut tools = HashMap::new();
        let mut plugins = HashMap::new();

        // Parse [tools] section
        if let Some(tools_table) = doc.get("tools").and_then(|v| v.as_table()) {
            for (name, value) in tools_table {
                let tool_config = match value {
                    toml::Value::String(version) => ToolConfig::Simple(version.clone()),
                    toml::Value::Table(table) => {
                        let version = table
                            .get("version")
                            .and_then(|v| v.as_str())
                            .ok_or_else(|| {
                                RazdError::config(format!(
                                    "Tool '{}' is missing 'version' field",
                                    name
                                ))
                            })?
                            .to_string();

                        let postinstall = table.get("postinstall").and_then(|v| v.as_str()).map(String::from);
                        
                        let os = table
                            .get("os")
                            .and_then(|v| match v {
                                toml::Value::String(s) => Some(vec![s.clone()]),
                                toml::Value::Array(arr) => {
                                    let os_list: Vec<String> = arr
                                        .iter()
                                        .filter_map(|item| item.as_str().map(String::from))
                                        .collect();
                                    if os_list.is_empty() {
                                        None
                                    } else {
                                        Some(os_list)
                                    }
                                }
                                _ => None,
                            });
                        
                        let install_env = table
                            .get("install_env")
                            .and_then(|v| v.as_table())
                            .map(|env_table| {
                                env_table
                                    .iter()
                                    .map(|(k, v)| (k.clone(), v.as_str().unwrap_or("").to_string()))
                                    .collect()
                            });

                        ToolConfig::Complex {
                            version,
                            postinstall,
                            os,
                            install_env,
                        }
                    }
                    _ => {
                        return Err(RazdError::config(format!(
                            "Invalid tool config for '{}'",
                            name
                        )))
                    }
                };
                tools.insert(name.clone(), tool_config);
            }
        }

        // Parse [plugins] section
        if let Some(plugins_table) = doc.get("plugins").and_then(|v| v.as_table()) {
            for (name, value) in plugins_table {
                if let Some(url) = value.as_str() {
                    plugins.insert(name.clone(), url.to_string());
                }
            }
        }

        Ok(MiseConfig {
            tools: if tools.is_empty() { None } else { Some(tools) },
            plugins: if plugins.is_empty() { None } else { Some(plugins) },
        })
    }

    /// Prompt user for approval (returns true if approved)
    fn prompt_user_approval(&self) -> Result<bool> {
        let mut input = String::new();
        std::io::stdin()
            .read_line(&mut input)
            .map_err(|e| RazdError::config(format!("Failed to read input: {}", e)))?;

        let response = input.trim().to_lowercase();
        Ok(response.is_empty() || response == "y" || response == "yes")
    }

    /// Format YAML with better spacing between sections
    fn format_yaml(&self, yaml: &str) -> String {
        let lines = yaml.lines().collect::<Vec<_>>();
        let mut formatted = Vec::new();
        
        for (i, line) in lines.iter().enumerate() {
            formatted.push(line.to_string());
            
            // Add blank line after top-level sections (version, tasks, mise)
            if i < lines.len() - 1 {
                let next_line = lines[i + 1];
                
                // Check if current line is a top-level key (no indentation, ends with :)
                let current_is_top_level = !line.starts_with(' ') && line.ends_with(':');
                let next_is_top_level = !next_line.starts_with(' ') && next_line.ends_with(':');
                
                // Add blank line between top-level sections
                if current_is_top_level && next_is_top_level {
                    formatted.push(String::new());
                }
                
                // Add blank line between task definitions (after "internal: false/true")
                if line.trim().starts_with("internal:") && next_line.starts_with("  ") && !next_line.trim().is_empty() {
                    formatted.push(String::new());
                }
                
                // Add blank line before "mise:" section (after last task's internal field)
                if line.trim().starts_with("internal:") && next_is_top_level {
                    formatted.push(String::new());
                }
            }
        }
        
        formatted.join("\n")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;
    use tempfile::TempDir;

    fn create_test_razdfile(dir: &Path) -> Result<()> {
        let mut tools = HashMap::new();
        tools.insert("node".to_string(), crate::config::razdfile::ToolConfig::Simple("22".to_string()));
        tools.insert("python".to_string(), crate::config::razdfile::ToolConfig::Simple("3.11".to_string()));

        let mut plugins = HashMap::new();
        plugins.insert("node".to_string(), "https://github.com/asdf-vm/asdf-nodejs.git".to_string());

        let mise_config = crate::config::razdfile::MiseConfig {
            tools: Some(tools),
            plugins: Some(plugins),
        };
        
        let razdfile = crate::config::razdfile::RazdfileConfig {
            version: "3".to_string(),
            tasks: HashMap::new(),
            mise: Some(mise_config),
        };

        let yaml = serde_yaml::to_string(&razdfile).unwrap();
        fs::write(dir.join("Razdfile.yml"), yaml)?;
        Ok(())
    }

    #[test]
    fn test_sync_razdfile_to_mise() {
        let temp_dir = TempDir::new().unwrap();
        let project_root = temp_dir.path().to_path_buf();

        create_test_razdfile(&project_root).unwrap();

        let config = SyncConfig {
            no_sync: false,
            auto_approve: true,
            create_backups: false,
        };

        let manager = MiseSyncManager::new(project_root.clone(), config);
        let result = manager.sync_razdfile_to_mise().unwrap();

        assert_eq!(result, SyncResult::RazdfileToMise);
        assert!(project_root.join("mise.toml").exists());

        let mise_content = fs::read_to_string(project_root.join("mise.toml")).unwrap();
        assert!(mise_content.contains("[tools]"));
        assert!(mise_content.contains("node"));
        assert!(mise_content.contains("python"));
    }

    #[test]
    fn test_sync_mise_to_razdfile() {
        let temp_dir = TempDir::new().unwrap();
        let project_root = temp_dir.path().to_path_buf();

        // Create mise.toml
        let mise_content = r#"
[tools]
node = "22"
python = "3.11"

[plugins]
node = "https://github.com/asdf-vm/asdf-nodejs.git"
"#;
        fs::write(project_root.join("mise.toml"), mise_content).unwrap();

        let config = SyncConfig {
            no_sync: false,
            auto_approve: true,
            create_backups: false,
        };

        let manager = MiseSyncManager::new(project_root.clone(), config);
        let result = manager.sync_mise_to_razdfile().unwrap();

        assert_eq!(result, SyncResult::MiseToRazdfile);
        assert!(project_root.join("Razdfile.yml").exists());

        let razdfile = RazdfileConfig::load_from_path(&project_root.join("Razdfile.yml"))
            .unwrap()
            .unwrap();
        assert!(razdfile.mise.is_some());
        let mise = razdfile.mise.unwrap();
        assert!(mise.tools.is_some());
        let tools = mise.tools.unwrap();
        assert!(tools.contains_key("node"));
        assert!(tools.contains_key("python"));
    }

    #[test]
    fn test_no_sync_flag() {
        let temp_dir = TempDir::new().unwrap();
        let project_root = temp_dir.path().to_path_buf();

        create_test_razdfile(&project_root).unwrap();

        let config = SyncConfig {
            no_sync: true,
            auto_approve: false,
            create_backups: false,
        };

        let manager = MiseSyncManager::new(project_root.clone(), config);
        let result = manager.check_and_sync_if_needed().unwrap();

        assert_eq!(result, SyncResult::Skipped);
        assert!(!project_root.join("mise.toml").exists());
    }

    #[test]
    fn test_parse_complex_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        let project_root = temp_dir.path().to_path_buf();

        let mise_content = r#"
[tools]
python = "3.11"
node = { version = "22", postinstall = "corepack enable" }

[plugins]
node = "https://github.com/asdf-vm/asdf-nodejs.git"
"#;

        let config = SyncConfig::default();
        let manager = MiseSyncManager::new(project_root, config);
        let parsed = manager.parse_mise_toml(mise_content).unwrap();

        assert!(parsed.tools.is_some());
        let tools = parsed.tools.unwrap();
        assert_eq!(tools.len(), 2);
        assert!(matches!(
            tools.get("python"),
            Some(crate::config::razdfile::ToolConfig::Simple(_))
        ));
        assert!(matches!(
            tools.get("node"),
            Some(crate::config::razdfile::ToolConfig::Complex { .. })
        ));
    }
}
