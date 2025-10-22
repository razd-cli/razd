use crate::core::Result;
use crate::config::get_workflow_config;
use crate::integrations::taskfile;
use colored::*;

/// Execute development workflow
pub async fn execute() -> Result<()> {
    println!("{}", "🚀 Starting development workflow...".cyan().bold());

    // Get workflow config with fallback chain
    if let Some(workflow_content) = get_workflow_config("dev")? {
        // Execute via taskfile with the workflow content
        taskfile::execute_workflow_task("dev", &workflow_content).await?;
    } else {
        return Err(crate::core::RazdError::command(
            "No development workflow found. Try running 'razd init --config' to create a Razdfile.yml"
        ));
    }

    println!("{}", "✅ Development workflow completed".green().bold());
    Ok(())
}