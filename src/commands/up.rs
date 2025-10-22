use crate::config::get_workflow_config;
use crate::core::{output, Result};
use crate::integrations::{git, mise, taskfile};
use std::env;

/// Execute the `razd up` command: clone repository + run up workflow
pub async fn execute(url: &str, name: Option<&str>) -> Result<()> {
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

    // Step 3: Execute up workflow (with fallback chain)
    if let Some(workflow_content) = get_workflow_config("up")? {
        output::step("Executing up workflow...");
        taskfile::execute_workflow_task("up", &workflow_content).await?;
    } else {
        // Fallback to legacy behavior if no workflow is found
        output::warning("No up workflow found, falling back to legacy setup");
        mise::install_tools(&absolute_repo_path).await?;
        taskfile::setup_project(&absolute_repo_path).await?;
    }

    // Step 4: Show success message and next steps
    output::success("Project setup completed successfully!");
    output::info("Next steps:");
    output::info("  razd dev            # Start development workflow");
    output::info("  razd build          # Build project");
    output::info("  razd task <name>    # Run specific task");

    Ok(())
}
