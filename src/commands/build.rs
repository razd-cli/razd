use crate::config::get_workflow_config;
use crate::core::Result;
use crate::integrations::taskfile;
use colored::*;
use std::env;

/// Execute build workflow
pub async fn execute() -> Result<()> {
    println!("{}", "🔨 Building project...".cyan().bold());

    // Check and sync mise configuration before executing workflow
    let current_dir = env::current_dir()?;
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        eprintln!("Warning: Mise sync check failed: {}", e);
    }

    // Get workflow config with fallback chain
    if let Some(workflow_content) = get_workflow_config("build")? {
        // Execute via taskfile with the workflow content in interactive mode
        taskfile::execute_workflow_task_interactive("build", &workflow_content).await?;
    } else {
        return Err(crate::core::RazdError::command(
            "No build workflow found. Try running 'razd init --config' to create a Razdfile.yml",
        ));
    }

    println!("{}", "✅ Build completed successfully".green().bold());
    Ok(())
}
