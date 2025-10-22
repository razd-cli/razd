/// Built-in default workflows for razd
/// These workflows are used when no Razdfile.yml is present

pub const DEFAULT_WORKFLOWS: &str = r#"version: '3'

tasks:
  up:
    desc: "Clone repository and set up project"
    cmds:
      - echo "🚀 Setting up project..."
      - mise install
      - task setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      
  dev:
    desc: "Start development workflow"
    cmds:
      - echo "🚀 Starting development..."
      - task default --taskfile Taskfile.yml
      
  build:
    desc: "Build project"
    cmds:
      - echo "🔨 Building project..."
      - task build --taskfile Taskfile.yml
"#;

/// Get built-in workflow for a specific command
pub fn get_default_workflow(command: &str) -> Option<&'static str> {
    match command {
        "up" | "install" | "dev" | "build" => Some(DEFAULT_WORKFLOWS),
        _ => None,
    }
}

/// Check if a command has a built-in workflow
pub fn has_default_workflow(command: &str) -> bool {
    matches!(command, "up" | "install" | "dev" | "build")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_has_default_workflow() {
        assert!(has_default_workflow("up"));
        assert!(has_default_workflow("install"));
        assert!(has_default_workflow("dev"));
        assert!(has_default_workflow("build"));
        assert!(!has_default_workflow("unknown"));
    }

    #[test]
    fn test_default_workflows_valid_yaml() {
        // Test that the default workflows string is valid YAML
        let parsed: Result<serde_yaml::Value, _> = serde_yaml::from_str(DEFAULT_WORKFLOWS);
        assert!(parsed.is_ok(), "Default workflows should be valid YAML");
        
        let yaml = parsed.unwrap();
        assert_eq!(yaml["version"], "3");
        assert!(yaml["tasks"]["up"].is_mapping());
        assert!(yaml["tasks"]["install"].is_mapping());
        assert!(yaml["tasks"]["dev"].is_mapping());
        assert!(yaml["tasks"]["build"].is_mapping());
    }
}