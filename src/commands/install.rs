use crate::config::get_workflow_config;
use crate::core::trust::ensure_trusted;
use crate::core::{output, Result};
use crate::integrations::{mise, taskfile};
use std::env;

/// Execute the `razd install` command: run install workflow
pub async fn execute() -> Result<()> {
    output::info("Installing development tools...");

    // Check trust before executing
    let current_dir = env::current_dir()?;
    let auto_yes = env::var("RAZD_AUTO_YES").unwrap_or_default() == "1";
    ensure_trusted(&current_dir, auto_yes).await?;

    // Check and sync mise configuration before executing
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        output::warning(&format!("Mise sync check failed: {}", e));
    }

    // Execute install workflow (with fallback chain)
    if let Some(workflow_content) = get_workflow_config("install")? {
        taskfile::execute_workflow_task("install", &workflow_content).await?;
    } else {
        // Fallback to legacy behavior
        output::warning("No install workflow found, falling back to mise install");
        mise::install_tools(&current_dir).await?;
    }

    Ok(())
}
