use crate::core::{output, Result};
use crate::integrations::{git, mise, taskfile};
use std::env;

/// Execute the `razd up` command: clone repository + mise install + task setup
pub async fn execute(url: &str, name: Option<&str>) -> Result<()> {
    output::info(&format!("Setting up project from {}", url));
    
    // Step 1: Clone the repository
    let repo_path = git::clone_repository(url, name).await?;
    
    // Step 2: Change to the repository directory (for logging purposes)
    let absolute_repo_path = env::current_dir()?.join(&repo_path);
    output::info(&format!("Working in directory: {}", absolute_repo_path.display()));
    
    // Step 3: Install development tools with mise (if configuration exists)
    mise::install_tools(&repo_path).await?;
    
    // Step 4: Set up project dependencies with task (if Taskfile exists)
    taskfile::setup_project(&repo_path).await?;
    
    // Step 5: Show success message and next steps
    output::success("Project setup completed successfully!");
    output::info(&format!("Next steps:"));
    output::info(&format!("  cd {}", repo_path.display()));
    output::info(&format!("  razd task           # Start development server"));
    output::info(&format!("  razd task <name>    # Run specific task"));
    
    Ok(())
}