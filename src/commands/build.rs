use crate::core::Result;
use crate::config::get_workflow_config;
use crate::integrations::taskfile;
use colored::*;

/// Execute build workflow
pub async fn execute() -> Result<()> {
    println!("{}", "🔨 Building project...".cyan().bold());

    // Get workflow config with fallback chain
    if let Some(workflow_content) = get_workflow_config("build")? {
        // Execute via taskfile with the workflow content
        taskfile::execute_workflow_task("build", &workflow_content).await?;
    } else {
        return Err(crate::core::RazdError::command(
            "No build workflow found. Try running 'razd init --config' to create a Razdfile.yml"
        ));
    }

    println!("{}", "✅ Build completed successfully".green().bold());
    Ok(())
}