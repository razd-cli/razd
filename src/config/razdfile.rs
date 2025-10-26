use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

use crate::core::RazdError;
use crate::defaults;

/// Command representation supporting both string commands and task references
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Command {
    /// Simple string command (e.g., "echo hello")
    String(String),
    /// Task reference with optional parameters
    TaskRef {
        task: String,
        #[serde(skip_serializing_if = "Option::is_none")]
        vars: Option<HashMap<String, String>>,
    },
}

/// Razdfile.yml configuration structure matching Taskfile v3 format
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RazdfileConfig {
    pub version: String,
    pub tasks: HashMap<String, TaskConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<Command>,
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

    #[test]
    fn test_command_string_parsing() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  test:
    desc: "Test with string commands"
    cmds:
      - echo "Installing..."
      - mise install
"#;

        fs::write(&razdfile_path, content).unwrap();

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_some());

        let config = result.unwrap();
        let task = config.get_task("test").unwrap();
        assert_eq!(task.cmds.len(), 2);

        // Verify commands are parsed as strings
        match &task.cmds[0] {
            Command::String(s) => assert_eq!(s, "echo \"Installing...\""),
            _ => panic!("Expected Command::String"),
        }
    }

    #[test]
    fn test_command_task_ref_parsing() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  up:
    desc: "Test with task references"
    cmds:
      - task: install
      - task: setup
  install:
    cmds:
      - mise install
"#;

        fs::write(&razdfile_path, content).unwrap();

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_some());

        let config = result.unwrap();
        let task = config.get_task("up").unwrap();
        assert_eq!(task.cmds.len(), 2);

        // Verify commands are parsed as task references
        match &task.cmds[0] {
            Command::TaskRef { task, vars } => {
                assert_eq!(task, "install");
                assert!(vars.is_none());
            }
            _ => panic!("Expected Command::TaskRef"),
        }
    }

    #[test]
    fn test_command_mixed_parsing() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  up:
    desc: "Test with mixed commands"
    cmds:
      - echo "Starting..."
      - task: install
      - echo "Done!"
"#;

        fs::write(&razdfile_path, content).unwrap();

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_some());

        let config = result.unwrap();
        let task = config.get_task("up").unwrap();
        assert_eq!(task.cmds.len(), 3);

        // Verify first command is string
        match &task.cmds[0] {
            Command::String(s) => assert_eq!(s, "echo \"Starting...\""),
            _ => panic!("Expected Command::String"),
        }

        // Verify second command is task reference
        match &task.cmds[1] {
            Command::TaskRef { task, vars } => {
                assert_eq!(task, "install");
                assert!(vars.is_none());
            }
            _ => panic!("Expected Command::TaskRef"),
        }

        // Verify third command is string
        match &task.cmds[2] {
            Command::String(s) => assert_eq!(s, "echo \"Done!\""),
            _ => panic!("Expected Command::String"),
        }
    }

    #[test]
    fn test_command_task_ref_with_vars() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  deploy:
    desc: "Test with task reference and vars"
    cmds:
      - task: build
        vars:
          ENV: production
          VERSION: v1.0.0
"#;

        fs::write(&razdfile_path, content).unwrap();

        let result = RazdfileConfig::load_from_path(&razdfile_path).unwrap();
        assert!(result.is_some());

        let config = result.unwrap();
        let task = config.get_task("deploy").unwrap();
        assert_eq!(task.cmds.len(), 1);

        // Verify command is task reference with vars
        match &task.cmds[0] {
            Command::TaskRef { task, vars } => {
                assert_eq!(task, "build");
                assert!(vars.is_some());
                let vars_map = vars.as_ref().unwrap();
                assert_eq!(vars_map.get("ENV").unwrap(), "production");
                assert_eq!(vars_map.get("VERSION").unwrap(), "v1.0.0");
            }
            _ => panic!("Expected Command::TaskRef with vars"),
        }
    }

    #[test]
    fn test_command_serialization() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

tasks:
  up:
    cmds:
      - echo "test"
      - task: install
"#;

        fs::write(&razdfile_path, content).unwrap();

        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        // Serialize back to YAML
        let yaml = serde_yaml::to_string(&config).unwrap();

        // Verify it contains both command types
        assert!(yaml.contains("echo \"test\"") || yaml.contains("echo"));
        assert!(yaml.contains("task: install"));
    }
}
