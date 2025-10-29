use crate::config::get_workflow_config;
use crate::core::{output, RazdError, Result};
use crate::integrations::{git, mise, taskfile};
use std::env;
use std::path::Path;
use std::io::{self, Write};
use std::fs;

/// Execute the `razd up` command: clone repository + run up workflow, or set up local project
pub async fn execute(url: Option<&str>, name: Option<&str>, init: bool) -> Result<()> {
    if init {
        // Init mode: create new Razdfile.yml
        execute_init().await
    } else if let Some(url_str) = url {
        // Clone mode: existing behavior
        execute_with_clone(url_str, name).await
    } else {
        // Local mode: new behavior
        execute_local_project().await
    }
}

/// Initialize new Razdfile.yml with project template
async fn execute_init() -> Result<()> {
    let current_dir = env::current_dir()?;
    let razdfile_path = current_dir.join("Razdfile.yml");

    // Check if Razdfile already exists
    if razdfile_path.exists() {
        return Err(RazdError::config(
            "Razdfile.yml already exists. Remove it first if you want to reinitialize."
        ));
    }

    output::info("Initializing new Razdfile.yml...");
    output::info(&format!("Working in directory: {}", current_dir.display()));

    // Detect project type
    let project_type = detect_project_type(&current_dir);
    output::info(&format!("Detected project type: {}", project_type));

    // Generate Razdfile content
    let razdfile_content = get_razdfile_template(&project_type);

    // Write Razdfile.yml
    fs::write(&razdfile_path, razdfile_content)
        .map_err(|e| RazdError::config(format!("Failed to create Razdfile.yml: {}", e)))?;

    output::success("✓ Razdfile.yml created successfully!");
    output::info("\nNext steps:");
    output::info("  1. Review and customize Razdfile.yml");
    output::info("  2. Run 'razd up' or 'razd' to execute the setup");
    output::info("  3. Use 'razd dev', 'razd build', etc. for other tasks");

    Ok(())
}

/// Clone repository and set up project
async fn execute_with_clone(url: &str, name: Option<&str>) -> Result<()> {
    output::info(&format!("Setting up project from {}", url));

    // Step 1: Clone the repository
    let repo_path = git::clone_repository(url, name).await?;

    // Step 2: Change to the repository directory for subsequent operations
    let absolute_repo_path = env::current_dir()?.join(&repo_path);
    env::set_current_dir(&absolute_repo_path)?;
    output::info(&format!(
        "Working in directory: {}",
        absolute_repo_path.display()
    ));

    // Step 3: Execute up workflow
    execute_up_workflow().await?;

    // Step 4: Show success message
    show_success_message()?;

    Ok(())
}

/// Set up project in current directory
async fn execute_local_project() -> Result<()> {
    output::info("Setting up local project...");

    let current_dir = env::current_dir()?;
    output::info(&format!("Working in directory: {}", current_dir.display()));

    // Check if project has configuration
    if has_project_configuration(&current_dir) {
        // Step 1: Execute up workflow
        execute_up_workflow().await?;
        
        // Step 2: Show success message
        show_success_message()?;
    } else {
        // Step 1: Offer to create configuration interactively
        output::info("No project configuration found.");
        
        if prompt_yes_no("Would you like to create a Razdfile.yml?", false)? {
            create_interactive_razdfile(&current_dir).await?;
            output::info("Razdfile.yml created successfully!");
            
            // Run the workflow we just created
            execute_up_workflow().await?;
            show_success_message()?;
        } else {
            output::info("Hint: Run 'razd up <url>' to clone a repository, or manually create a Razdfile.yml");
            show_razdfile_example();
            
            return Err(RazdError::no_project_config(
                "Create a Razdfile.yml manually or run 'razd up <url>' to clone a repository with configuration."
            ));
        }
    }

    Ok(())
}

/// Check if directory has project configuration
fn has_project_configuration(dir: &Path) -> bool {
    let has_razdfile = dir.join("Razdfile.yml").exists();
    let has_taskfile = dir.join("Taskfile.yml").exists();
    let has_mise = dir.join("mise.toml").exists() || dir.join(".mise.toml").exists();
    
    has_razdfile || has_taskfile || has_mise
}

/// Execute up workflow (with fallback chain)
async fn execute_up_workflow() -> Result<()> {
    // Check and sync mise configuration before executing workflow
    let current_dir = env::current_dir()?;
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        output::warning(&format!("Mise sync check failed: {}", e));
    }

    if let Some(workflow_content) = get_workflow_config("default")? {
        output::step("Executing up workflow...");
        taskfile::execute_workflow_task_interactive("default", &workflow_content).await?;
    } else {
        // Fallback to legacy behavior if no workflow is found
        output::warning("No default task found, falling back to legacy setup");
        let current_dir = env::current_dir()?;
        mise::install_tools(&current_dir).await?;
        taskfile::setup_project(&current_dir).await?;
    }
    Ok(())
}

/// Show success message and next steps
fn show_success_message() -> Result<()> {
    output::success("Project setup completed successfully!");
    output::info("Next steps:");
    output::info("  razd dev            # Start development workflow");
    output::info("  razd build          # Build project");
    output::info("  razd task <name>    # Run specific task");
    Ok(())
}

/// Prompt user for yes/no answer
fn prompt_yes_no(message: &str, default: bool) -> Result<bool> {
    let default_str = if default { "Y/n" } else { "y/N" };
    output::info(&format!("{} [{}] ", message, default_str));
    
    io::stdout().flush().map_err(|e| RazdError::command(&format!("Failed to flush stdout: {}", e)))?;
    
    let mut input = String::new();
    io::stdin().read_line(&mut input)
        .map_err(|e| RazdError::command(&format!("Failed to read input: {}", e)))?;
    
    let input = input.trim().to_lowercase();
    
    match input.as_str() {
        "y" | "yes" => Ok(true),
        "n" | "no" => Ok(false),
        "" => Ok(default),
        _ => {
            output::info("Please answer 'y' or 'n'");
            prompt_yes_no(message, default)
        }
    }
}

/// Create interactive Razdfile.yml
async fn create_interactive_razdfile(dir: &Path) -> Result<()> {
    output::info("Creating Razdfile.yml...");
    
    // Detect project type
    let project_type = detect_project_type(dir);
    output::info(&format!("Detected project type: {}", project_type));
    
    // Get template based on project type
    let template = get_razdfile_template(&project_type);
    
    // Write file
    let razdfile_path = dir.join("Razdfile.yml");
    fs::write(&razdfile_path, template)
        .map_err(|e| RazdError::command(&format!("Failed to write Razdfile.yml: {}", e)))?;
    
    output::info(&format!("Created: {}", razdfile_path.display()));
    
    Ok(())
}

/// Detect project type based on files in directory
fn detect_project_type(dir: &Path) -> String {
    if dir.join("package.json").exists() {
        "Node.js".to_string()
    } else if dir.join("Cargo.toml").exists() {
        "Rust".to_string()
    } else if dir.join("requirements.txt").exists() || dir.join("pyproject.toml").exists() {
        "Python".to_string()
    } else if dir.join("go.mod").exists() {
        "Go".to_string()
    } else {
        "Generic".to_string()
    }
}

/// Get Razdfile.yml template for project type
fn get_razdfile_template(project_type: &str) -> String {
    match project_type {
        "Node.js" => {
            r#"tasks:
  default:
    desc: "Set up and start Node.js project"
    cmds:
      - mise install
      - npm install
      - npm run dev

  install:
    desc: "Install dependencies"
    cmds:
      - mise install
      - npm install

  dev:
    desc: "Start development server"
    cmds:
      - npm run dev

  build:
    desc: "Build project"
    cmds:
      - npm run build

  test:
    desc: "Run tests"
    cmds:
      - npm test
"#.to_string()
        }
        "Rust" => {
            r#"tasks:
  default:
    desc: "Set up and build Rust project"
    cmds:
      - mise install
      - cargo build

  install:
    desc: "Install tools"
    cmds:
      - mise install

  dev:
    desc: "Run in development mode"
    cmds:
      - cargo run

  build:
    desc: "Build project"
    cmds:
      - cargo build --release

  test:
    desc: "Run tests"
    cmds:
      - cargo test
"#.to_string()
        }
        "Python" => {
            r#"tasks:
  default:
    desc: "Set up and start Python project"
    cmds:
      - mise install
      - pip install -r requirements.txt
      - python main.py

  install:
    desc: "Install dependencies"
    cmds:
      - mise install
      - pip install -r requirements.txt

  dev:
    desc: "Start development server"
    cmds:
      - python main.py

  test:
    desc: "Run tests"
    cmds:
      - python -m pytest
"#.to_string()
        }
        "Go" => {
            r#"tasks:
  default:
    desc: "Set up and run Go project"
    cmds:
      - mise install
      - go mod download
      - go run .

  install:
    desc: "Install dependencies"
    cmds:
      - mise install
      - go mod download

  dev:
    desc: "Run in development mode"
    cmds:
      - go run .

  build:
    desc: "Build project"
    cmds:
      - go build

  test:
    desc: "Run tests"
    cmds:
      - go test ./...
"#.to_string()
        }
        _ => {
            r#"tasks:
  default:
    desc: "Set up project"
    cmds:
      - mise install
      - echo "Project setup completed!"

  install:
    desc: "Install tools and dependencies"
    cmds:
      - mise install

  dev:
    desc: "Start development"
    cmds:
      - echo "Starting development..."

  build:
    desc: "Build project"
    cmds:
      - echo "Building project..."

  test:
    desc: "Run tests"
    cmds:
      - echo "Running tests..."
"#.to_string()
        }
    }
}

/// Show example Razdfile.yml
fn show_razdfile_example() {
    output::info("Example Razdfile.yml:");
    output::info("  tasks:");
    output::info("    default:");
    output::info("      desc: \"Set up and start project\"");
    output::info("      cmds:");
    output::info("        - mise install");
    output::info("        - echo \"Project ready!\"");
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::TempDir;

    #[test]
    fn test_has_project_configuration_with_razdfile() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Razdfile.yml"), "").unwrap();

        assert!(has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_with_taskfile() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();

        assert!(has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_with_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("mise.toml"), "").unwrap();

        assert!(has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_with_dot_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join(".mise.toml"), "").unwrap();

        assert!(has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_with_multiple_files() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Razdfile.yml"), "").unwrap();
        fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();

        assert!(has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_empty() {
        let temp_dir = TempDir::new().unwrap();

        assert!(!has_project_configuration(temp_dir.path()));
    }

    #[test]
    fn test_has_project_configuration_only_git() {
        let temp_dir = TempDir::new().unwrap();
        fs::create_dir(temp_dir.path().join(".git")).unwrap();

        assert!(!has_project_configuration(temp_dir.path()));
    }
}
