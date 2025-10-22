use crate::config::get_workflow_config;
use crate::core::{output, Result};
use crate::integrations::{mise, taskfile};
use std::env;

/// Execute the `razd install` command: run install workflow
pub async fn execute() -> Result<()> {
    output::info("Installing development tools...");

    // Execute install workflow (with fallback chain)
    if let Some(workflow_content) = get_workflow_config("install")? {
        taskfile::execute_workflow_task("install", &workflow_content).await?;
    } else {
        // Fallback to legacy behavior
        output::warning("No install workflow found, falling back to mise install");
        let current_dir = env::current_dir()?;
        mise::install_tools(&current_dir).await?;
    }

    Ok(())
}
