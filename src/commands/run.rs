use crate::config::get_workflow_config;
use crate::core::Result;
use crate::integrations::taskfile;
use colored::*;
use std::env;

/// Execute a custom task defined in Razdfile.yml
pub async fn execute(task_name: &str, _args: &[String]) -> Result<()> {
    println!(
        "{}",
        format!("🚀 Running task '{}'...", task_name).cyan().bold()
    );

    // Check and sync mise configuration before executing task
    let current_dir = env::current_dir()?;
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        eprintln!("Warning: Mise sync check failed: {}", e);
    }

    // Get workflow config with fallback chain
    if let Some(workflow_content) = get_workflow_config(task_name)? {
        // Execute via taskfile with the workflow content in interactive mode
        taskfile::execute_workflow_task_interactive(task_name, &workflow_content).await?;
    } else {
        return Err(crate::core::RazdError::command(&format!(
            "Task '{}' not found in Razdfile.yml. Try running 'task --list' to see available tasks",
            task_name
        )));
    }

    println!(
        "{}",
        format!("✅ Task '{}' completed successfully", task_name)
            .green()
            .bold()
    );
    Ok(())
}
