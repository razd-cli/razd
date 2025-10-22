use crate::core::{output, Result, RazdError};
use crate::integrations::process;
use std::path::Path;

/// Check if Taskfile configuration exists in the directory
pub fn has_taskfile_config(dir: &Path) -> bool {
    dir.join("Taskfile.yml").exists() || dir.join("Taskfile.yaml").exists()
}

/// Run task setup to install project dependencies
pub async fn setup_project(working_dir: &Path) -> Result<()> {
    // Check if task is available
    if !process::check_command_available("task").await {
        return Err(RazdError::missing_tool(
            "task",
            "https://taskfile.dev/installation/"
        ));
    }

    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        output::warning("No Taskfile found (Taskfile.yml or Taskfile.yaml), skipping project setup");
        return Ok(());
    }

    output::step("Setting up project dependencies with task");
    
    process::execute_command("task", &["setup"], Some(working_dir)).await
        .map_err(|e| RazdError::task(format!("Failed to setup project: {}", e)))?;

    output::success("Successfully set up project dependencies");
    
    Ok(())
}

/// Execute a specific task
pub async fn execute_task(task_name: Option<&str>, args: &[String], working_dir: &Path) -> Result<()> {
    // Check if task is available
    if !process::check_command_available("task").await {
        return Err(RazdError::missing_tool(
            "task",
            "https://taskfile.dev/installation/"
        ));
    }

    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        return Err(RazdError::task(
            "No Taskfile found (Taskfile.yml or Taskfile.yaml). Please create one first."
        ));
    }

    let mut command_args = Vec::new();
    
    if let Some(name) = task_name {
        command_args.push(name);
    }
    
    // Add additional arguments
    for arg in args {
        command_args.push(arg);
    }

    let task_desc = if let Some(name) = task_name {
        format!("task {}", name)
    } else {
        "default task".to_string()
    };

    output::step(&format!("Executing {}", task_desc));
    
    let args_str: Vec<&str> = command_args.iter().map(|s| s.as_ref()).collect();
    process::execute_command("task", &args_str, Some(working_dir)).await
        .map_err(|e| RazdError::task(format!("Failed to execute task: {}", e)))?;

    output::success(&format!("Successfully executed {}", task_desc));
    
    Ok(())
}