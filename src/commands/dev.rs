use crate::config::get_workflow_config;
use crate::core::trust::ensure_trusted;
use crate::core::Result;
use crate::integrations::taskfile;
use colored::*;
use std::env;

/// Execute development workflow
pub async fn execute() -> Result<()> {
    println!("{}", "🚀 Starting development workflow...".cyan().bold());

    // Check trust before executing
    let current_dir = env::current_dir()?;
    let auto_yes = env::var("RAZD_AUTO_YES").unwrap_or_default() == "1";
    ensure_trusted(&current_dir, auto_yes).await?;

    // Check and sync mise configuration before executing workflow
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        eprintln!("Warning: Mise sync check failed: {}", e);
    }

    // Get workflow config with fallback chain
    if let Some(workflow_content) = get_workflow_config("dev")? {
        // Execute via taskfile with the workflow content in interactive mode
        taskfile::execute_workflow_task_interactive("dev", &workflow_content).await?;
    } else {
        return Err(crate::core::RazdError::command(
            "No development workflow found. Try running 'razd init --config' to create a Razdfile.yml"
        ));
    }

    println!("{}", "✅ Development workflow completed".green().bold());
    Ok(())
}
