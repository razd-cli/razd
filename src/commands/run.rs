use crate::core::Result;
use crate::integrations::taskfile;
use colored::*;
use std::env;
use std::path::PathBuf;

/// Execute a custom task defined in Razdfile.yml
pub async fn execute(
    task_name: &str,
    args: &[String],
    custom_path: Option<PathBuf>,
) -> Result<()> {
    println!(
        "{}",
        format!("🚀 Running task '{}'...", task_name).cyan().bold()
    );

    // Check and sync mise configuration before executing task
    let current_dir = env::current_dir()?;
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        eprintln!("Warning: Mise sync check failed: {}", e);
    }

    // Get workflow config with fallback chain (with custom path support)
    if let Some(workflow_content) =
        crate::config::get_workflow_config_with_path(task_name, custom_path)?
    {
        // Execute via taskfile with the workflow content and CLI arguments
        if args.is_empty() {
            taskfile::execute_workflow_task_interactive(task_name, &workflow_content).await?;
        } else {
            taskfile::execute_workflow_task_with_args(task_name, &workflow_content, args).await?;
        }
    } else {
        return Err(crate::core::RazdError::command(format!(
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
