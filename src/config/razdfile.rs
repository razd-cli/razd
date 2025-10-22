use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

use crate::core::RazdError;
use crate::defaults;

/// Razdfile.yml configuration structure matching Taskfile v3 format
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RazdfileConfig {
    pub version: String,
    pub tasks: HashMap<String, TaskConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<String>,
    #[serde(default)]
    pub internal: bool,
}

impl RazdfileConfig {
    /// Load Razdfile.yml from the current directory
    pub fn load() -> Result<Option<Self>, RazdError> {
        Self::load_from_path("Razdfile.yml")
    }

    /// Load Razdfile.yml from a specific path
    pub fn load_from_path<P: AsRef<Path>>(path: P) -> Result<Option<Self>, RazdError> {
        let path = path.as_ref();

        if !path.exists() {
            return Ok(None);
        }

        let content = fs::read_to_string(path)
            .map_err(|e| RazdError::config(format!("Failed to read Razdfile.yml: {}", e)))?;

        let config: RazdfileConfig = serde_yaml::from_str(&content)
            .map_err(|e| RazdError::config(format!("Failed to parse Razdfile.yml: {}", e)))?;

        Ok(Some(config))
    }

    /// Get a task configuration by name
    #[allow(dead_code)]
    pub fn get_task(&self, name: &str) -> Option<&TaskConfig> {
        self.tasks.get(name)
    }

    /// Check if a task exists in the configuration
    pub fn has_task(&self, name: &str) -> bool {
        self.tasks.contains_key(name)
    }
}

/// Get workflow configuration with fallback chain
/// Priority: Razdfile.yml → built-in defaults
pub fn get_workflow_config(command: &str) -> Result<Option<String>, RazdError> {
    // Try to load Razdfile.yml first
    if let Some(razdfile) = RazdfileConfig::load()? {
        if razdfile.has_task(command) {
            // Convert back to YAML for taskfile execution
            let yaml_content = serde_yaml::to_string(&razdfile).map_err(|e| {
                RazdError::config(format!("Failed to serialize Razdfile.yml: {}", e))
            })?;
            return Ok(Some(yaml_content));
        }
    }

    // Fallback to built-in defaults
    if defaults::has_default_workflow(command) {
        return Ok(Some(defaults::DEFAULT_WORKFLOWS.to_string()));
    }

    Ok(None)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::TempDir;

    #[test]
    fn test_razdfile_load_nonexistent() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_none());
    }

    #[test]
    fn test_razdfile_load_valid() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  test:
    desc: "Test task"
    cmds:
      - echo "test"
"#;

        fs::write(&razdfile_path, content).unwrap();

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_some());

        let config = result.unwrap();
        assert_eq!(config.version, "3");
        assert!(config.has_task("test"));
        assert!(!config.has_task("nonexistent"));
    }

    #[test]
    fn test_workflow_config_fallback() {
        // Test with no Razdfile.yml - should use built-in defaults
        let temp_dir = TempDir::new().unwrap();
        std::env::set_current_dir(temp_dir.path()).unwrap();

        let result = get_workflow_config("dev").unwrap();
        assert!(result.is_some());

        let workflow = result.unwrap();
        assert!(workflow.contains("version: '3'"));
        assert!(workflow.contains("dev:"));
    }
}
