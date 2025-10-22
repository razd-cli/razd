use crate::core::{output, RazdError, Result};
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
            "https://taskfile.dev/installation/",
        ));
    }

    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        output::warning(
            "No Taskfile found (Taskfile.yml or Taskfile.yaml), skipping project setup",
        );
        return Ok(());
    }

    output::step("Setting up project dependencies with task");

    process::execute_command("task", &["setup"], Some(working_dir))
        .await
        .map_err(|e| RazdError::task(format!("Failed to setup project: {}", e)))?;

    output::success("Successfully set up project dependencies");

    Ok(())
}

/// Execute a specific task
pub async fn execute_task(
    task_name: Option<&str>,
    args: &[String],
    working_dir: &Path,
) -> Result<()> {
    // Check if task is available
    if !process::check_command_available("task").await {
        return Err(RazdError::missing_tool(
            "task",
            "https://taskfile.dev/installation/",
        ));
    }

    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        return Err(RazdError::task(
            "No Taskfile found (Taskfile.yml or Taskfile.yaml). Please create one first.",
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
    process::execute_command("task", &args_str, Some(working_dir))
        .await
        .map_err(|e| RazdError::task(format!("Failed to execute task: {}", e)))?;

    output::success(&format!("Successfully executed {}", task_desc));

    Ok(())
}

/// Execute a workflow task using custom taskfile content
pub async fn execute_workflow_task(task_name: &str, workflow_content: &str) -> Result<()> {
    use std::env;
    use std::fs;

    // Check if task is available
    if !process::check_command_available("task").await {
        return Err(RazdError::missing_tool(
            "task",
            "https://taskfile.dev/installation/",
        ));
    }

    // Get current working directory
    let working_dir = env::current_dir()
        .map_err(|e| RazdError::task(format!("Failed to get current directory: {}", e)))?;

    // Create a temporary taskfile in the working directory (not temp)
    let temp_taskfile = working_dir.join(format!(".razd-workflow-{}.yml", task_name));

    fs::write(&temp_taskfile, workflow_content)
        .map_err(|e| RazdError::task(format!("Failed to create temporary taskfile: {}", e)))?;

    output::step(&format!("Executing workflow: {}", task_name));

    // Execute task with custom taskfile in the working directory
    let args = vec!["--taskfile", temp_taskfile.to_str().unwrap(), task_name];

    let result = process::execute_command("task", &args, Some(&working_dir)).await;

    // Clean up temporary file
    let _ = fs::remove_file(&temp_taskfile);

    result.map_err(|e| RazdError::task(format!("Failed to execute workflow: {}", e)))?;

    output::success(&format!("Successfully executed workflow: {}", task_name));

    Ok(())
}
