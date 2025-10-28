use serde::{Deserialize, Serialize};
use indexmap::IndexMap;
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
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mise: Option<MiseConfig>,
    #[serde(default)]
    pub tasks: IndexMap<String, TaskConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<Command>,
    #[serde(default)]
    pub internal: bool,
}

/// Mise configuration section in Razdfile.yml
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MiseConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tools: Option<IndexMap<String, ToolConfig>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub plugins: Option<IndexMap<String, String>>,
}

/// Tool configuration supporting both simple versions and complex options
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum ToolConfig {
    /// Simple version string: "22"
    Simple(String),
    /// Complex configuration with options
    Complex {
        version: String,
        #[serde(skip_serializing_if = "Option::is_none")]
        postinstall: Option<String>,
        #[serde(skip_serializing_if = "Option::is_none")]
        os: Option<Vec<String>>,
        #[serde(skip_serializing_if = "Option::is_none")]
        install_env: Option<HashMap<String, String>>,
    },
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

        // Validate mise configuration if present
        if let Some(ref mise_config) = config.mise {
            config.validate_mise_config(mise_config)?;
        }

        Ok(Some(config))
    }

    /// Validate mise configuration
    fn validate_mise_config(&self, mise_config: &MiseConfig) -> Result<(), RazdError> {
        use crate::config::mise_validator;

        // Validate tool names
        if let Some(ref tools) = mise_config.tools {
            for name in tools.keys() {
                mise_validator::validate_tool_name(name)?;
            }
        }

        // Validate plugin names and URLs
        if let Some(ref plugins) = mise_config.plugins {
            for (name, url) in plugins.iter() {
                mise_validator::validate_tool_name(name)?;
                mise_validator::validate_plugin_url(url)?;
            }
        }

        Ok(())
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

    /// Get the primary task for "up" command, now only supports "default"
    pub fn get_primary_task(&self) -> Option<&str> {
        if self.has_task("default") {
            Some("default")
        } else {
            None
        }
    }
}

/// Get workflow configuration with fallback chain
/// Priority: Razdfile.yml → built-in defaults
/// For "default" task: uses get_primary_task which returns "default" task
pub fn get_workflow_config(command: &str) -> Result<Option<String>, RazdError> {
    // Try to load Razdfile.yml first
    if let Some(razdfile) = RazdfileConfig::load()? {
        let task_name = if command == "default" {
            // For "default" command, use get_primary_task
            razdfile.get_primary_task()
        } else {
            // For other commands, use exact task name
            if razdfile.has_task(command) {
                Some(command)
            } else {
                None
            }
        };

        if let Some(_task) = task_name {
            // Convert back to YAML for taskfile execution
            let yaml_content = serde_yaml::to_string(&razdfile).map_err(|e| {
                RazdError::config(format!("Failed to serialize Razdfile.yml: {}", e))
            })?;
            return Ok(Some(yaml_content));
        } else if command == "default" {
            // Special handling for "default" command when no default task found
            return Err(RazdError::no_default_task());
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

    #[test]
    fn test_get_primary_task() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        // Test with default task
        let content = r#"
version: '3'
tasks:
  default:
    desc: "Default task"
    cmds:
      - echo "default task"
  build:
    desc: "Build task"  
    cmds:
      - echo "build task"
"#;

        fs::write(&razdfile_path, content).unwrap();

        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        assert_eq!(config.get_primary_task(), Some("default"));

        // Test without default task
        let content_no_default = r#"
version: '3'
tasks:
  build:
    desc: "Build task"
    cmds:
      - echo "build task"
  test:
    desc: "Test task"
    cmds:
      - echo "test task"
"#;

        fs::write(&razdfile_path, content_no_default).unwrap();

        let config_no_default = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        assert_eq!(config_no_default.get_primary_task(), None);
    }

    #[test]
    fn test_workflow_config_default_priority() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        // Test up command with default task present
        let content = r#"
version: '3'
tasks:
  default:
    desc: "Default task"
    cmds:
      - echo "default task"
  up:
    desc: "Up task"  
    cmds:
      - echo "up task"
"#;

        fs::write(&razdfile_path, content).unwrap();

        // Load from specific path instead of changing current dir
        let razdfile = RazdfileConfig::load_from_path(&razdfile_path).unwrap().unwrap();
        
        // Test that it has default task and prefers it
        assert_eq!(razdfile.get_primary_task(), Some("default"));
        
        // Convert to YAML for workflow
        let yaml_content = serde_yaml::to_string(&razdfile).unwrap();
        
        // Should contain the config with default task prioritized
        assert!(yaml_content.contains("default:"));
        assert!(yaml_content.contains("Default task"));
    }

    #[test]
    fn test_parse_simple_mise_tools() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

mise:
  tools:
    node: "22"
    python: "3.11"

tasks:
  test:
    desc: "Test task"
    cmds:
      - echo "test"
"#;

        fs::write(&razdfile_path, content).unwrap();
        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        assert!(config.mise.is_some());
        let mise = config.mise.unwrap();
        assert!(mise.tools.is_some());
        
        let tools = mise.tools.unwrap();
        assert_eq!(tools.len(), 2);
        
        match tools.get("node").unwrap() {
            ToolConfig::Simple(v) => assert_eq!(v, "22"),
            _ => panic!("Expected simple tool config"),
        }
        
        match tools.get("python").unwrap() {
            ToolConfig::Simple(v) => assert_eq!(v, "3.11"),
            _ => panic!("Expected simple tool config"),
        }
    }

    #[test]
    fn test_parse_complex_mise_tools() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

mise:
  tools:
    node:
      version: "22"
      postinstall: "corepack enable"
      os: ["linux", "macos"]
    go:
      version: "1.21"
      install_env:
        CGO_ENABLED: "1"

tasks:
  test:
    desc: "Test"
    cmds:
      - echo "test"
"#;

        fs::write(&razdfile_path, content).unwrap();
        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        let tools = config.mise.as_ref().unwrap().tools.as_ref().unwrap();
        
        match tools.get("node").unwrap() {
            ToolConfig::Complex { version, postinstall, os, .. } => {
                assert_eq!(version, "22");
                assert_eq!(postinstall.as_ref().unwrap(), "corepack enable");
                assert_eq!(os.as_ref().unwrap(), &vec!["linux".to_string(), "macos".to_string()]);
            },
            _ => panic!("Expected complex tool config"),
        }
        
        match tools.get("go").unwrap() {
            ToolConfig::Complex { version, install_env, .. } => {
                assert_eq!(version, "1.21");
                assert_eq!(install_env.as_ref().unwrap().get("CGO_ENABLED").unwrap(), "1");
            },
            _ => panic!("Expected complex tool config"),
        }
    }

    #[test]
    fn test_parse_mise_plugins() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'

mise:
  plugins:
    elixir: "https://github.com/my-org/mise-elixir.git"
    node: "https://github.com/my-org/mise-node.git#DEADBEEF"

tasks:
  test:
    desc: "Test"
    cmds:
      - echo "test"
"#;

        fs::write(&razdfile_path, content).unwrap();
        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        let plugins = config.mise.as_ref().unwrap().plugins.as_ref().unwrap();
        assert_eq!(plugins.len(), 2);
        assert_eq!(plugins.get("elixir").unwrap(), "https://github.com/my-org/mise-elixir.git");
        assert_eq!(plugins.get("node").unwrap(), "https://github.com/my-org/mise-node.git#DEADBEEF");
    }

    #[test]
    fn test_missing_mise_section() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        let content = r#"
version: '3'
tasks:
  test:
    desc: "Test"
    cmds:
      - echo "test"
"#;

        fs::write(&razdfile_path, content).unwrap();
        let config = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();

        assert!(config.mise.is_none());
    }
}
