use crate::config::get_workflow_config;
use crate::core::{output, RazdError, Result};
use crate::integrations::{git, mise, taskfile};
use std::env;
use std::path::Path;

/// Execute the `razd up` command: clone repository + run up workflow, or set up local project
pub async fn execute(url: Option<&str>, name: Option<&str>) -> Result<()> {
    if let Some(url_str) = url {
        // Clone mode: existing behavior
        execute_with_clone(url_str, name).await
    } else {
        // Local mode: new behavior
        execute_local_project().await
    }
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

    // Step 1: Validate we're in a project directory
    let current_dir = env::current_dir()?;
    validate_project_directory(&current_dir)?;

    output::info(&format!("Working in directory: {}", current_dir.display()));

    // Step 2: Execute up workflow
    execute_up_workflow().await?;

    // Step 3: Show success message
    show_success_message()?;

    Ok(())
}

/// Execute up workflow (with fallback chain)
async fn execute_up_workflow() -> Result<()> {
    if let Some(workflow_content) = get_workflow_config("up")? {
        output::step("Executing up workflow...");
        taskfile::execute_workflow_task_interactive("up", &workflow_content).await?;
    } else {
        // Fallback to legacy behavior if no workflow is found
        output::warning("No up workflow found, falling back to legacy setup");
        let current_dir = env::current_dir()?;
        mise::install_tools(&current_dir).await?;
        taskfile::setup_project(&current_dir).await?;
    }
    Ok(())
}

/// Validate that the current directory contains a project
fn validate_project_directory(dir: &Path) -> Result<()> {
    // Check for at least one project indicator file
    let has_razdfile = dir.join("Razdfile.yml").exists();
    let has_taskfile = dir.join("Taskfile.yml").exists();
    let has_mise = dir.join("mise.toml").exists() || dir.join(".mise.toml").exists();

    if !has_razdfile && !has_taskfile && !has_mise {
        return Err(RazdError::command(
            "No project detected in current directory. Expected one of: Razdfile.yml, Taskfile.yml, or mise.toml\n\
             Hint: Run 'razd up <url>' to clone a repository, or 'razd init' to initialize a new project."
        ));
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

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::TempDir;

    #[test]
    fn test_validate_project_directory_with_razdfile() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Razdfile.yml"), "").unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_project_directory_with_taskfile() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_project_directory_with_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("mise.toml"), "").unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_project_directory_with_dot_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join(".mise.toml"), "").unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_project_directory_with_multiple_files() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();
        fs::write(temp_dir.path().join("mise.toml"), "").unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_project_directory_empty() {
        let temp_dir = TempDir::new().unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("No project detected"));
    }

    #[test]
    fn test_validate_project_directory_only_git() {
        let temp_dir = TempDir::new().unwrap();
        fs::create_dir(temp_dir.path().join(".git")).unwrap();

        let result = validate_project_directory(temp_dir.path());
        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("No project detected"));
    }
}
