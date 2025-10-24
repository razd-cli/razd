use crate::core::{output, RazdError, Result};
use crate::integrations::{mise, process};
use std::path::Path;

/// Execute task command, trying direct execution first, then mise exec as fallback
async fn execute_task_command(args: &[&str], working_dir: &Path) -> Result<()> {
    // First try direct execution
    if process::check_command_available("task").await {
        return process::execute_command("task", args, Some(working_dir))
            .await
            .map_err(|e| RazdError::task(format!("Failed to execute task: {}", e)));
    }

    // Fallback: try through mise exec (task should be available via mise)
    output::step("Executing task via mise...");
    let mut mise_args = vec!["exec", "task", "--", "task"];
    mise_args.extend(args);
    
    process::execute_command("mise", &mise_args, Some(working_dir))
        .await
        .map_err(|e| RazdError::task(format!("Failed to execute task via mise: {}", e)))
}

/// Check if Taskfile configuration exists in the directory
pub fn has_taskfile_config(dir: &Path) -> bool {
    dir.join("Taskfile.yml").exists() || dir.join("Taskfile.yaml").exists()
}

/// Run task setup to install project dependencies
pub async fn setup_project(working_dir: &Path) -> Result<()> {
    // Ensure task tool is available
    mise::ensure_tool_available("task", "latest", working_dir).await?;

    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        output::warning(
            "No Taskfile found (Taskfile.yml or Taskfile.yaml), skipping project setup",
        );
        return Ok(());
    }

    output::step("Setting up project dependencies with task");

    execute_task_command(&["setup"], working_dir).await?;

    output::success("Successfully set up project dependencies");

    Ok(())
}

/// Execute a specific task
pub async fn execute_task(
    task_name: Option<&str>,
    args: &[String],
    working_dir: &Path,
) -> Result<()> {
    // Ensure task tool is available
    mise::ensure_tool_available("task", "latest", working_dir).await?;

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
    execute_task_command(&args_str, working_dir).await?;

    output::success(&format!("Successfully executed {}", task_desc));

    Ok(())
}

/// Execute a workflow task using custom taskfile content
pub async fn execute_workflow_task(task_name: &str, workflow_content: &str) -> Result<()> {
    use std::env;
    use std::fs;

    // Get current working directory
    let working_dir = env::current_dir()
        .map_err(|e| RazdError::task(format!("Failed to get current directory: {}", e)))?;

    // Ensure task tool is available
    mise::ensure_tool_available("task", "latest", &working_dir).await?;

    // Create a temporary taskfile in the working directory (not temp)
    let temp_taskfile = working_dir.join(format!(".razd-workflow-{}.yml", task_name));

    fs::write(&temp_taskfile, workflow_content)
        .map_err(|e| RazdError::task(format!("Failed to create temporary taskfile: {}", e)))?;

    output::step(&format!("Executing workflow: {}", task_name));

    // Execute task with custom taskfile in the working directory
    let args = vec!["--taskfile", temp_taskfile.to_str().unwrap(), task_name];

    let result = execute_task_command(&args, &working_dir).await;

    // Clean up temporary file
    let _ = fs::remove_file(&temp_taskfile);

    result?;

    output::success(&format!("Successfully executed workflow: {}", task_name));

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_has_taskfile_config_with_taskfile_yml() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();

        assert!(has_taskfile_config(temp_dir.path()));
    }

    #[test]
    fn test_has_taskfile_config_with_taskfile_yaml() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join("Taskfile.yaml"), "").unwrap();

        assert!(has_taskfile_config(temp_dir.path()));
    }

    #[test]
    fn test_has_taskfile_config_with_neither() {
        let temp_dir = TempDir::new().unwrap();

        assert!(!has_taskfile_config(temp_dir.path()));
    }

    #[test]
    fn test_has_taskfile_config_with_both() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join("Taskfile.yml"), "").unwrap();
        std::fs::write(temp_dir.path().join("Taskfile.yaml"), "").unwrap();

        assert!(has_taskfile_config(temp_dir.path()));
    }

    // Note: The async functions setup_project, execute_task, and execute_workflow_task
    // require external processes and are better tested as integration tests
    // rather than unit tests, since they depend on task being installed.
}
