use crate::core::{output, RazdError, Result};
use crate::defaults;
use crate::integrations::{mise, process};
use std::path::Path;
use std::time::Duration;

/// Execute task command, trying direct execution first, then mise exec as fallback
async fn execute_task_command(args: &[&str], working_dir: &Path) -> Result<()> {
    execute_task_command_with_mode(args, working_dir, false).await
}

/// Execute task command with option for interactive mode
async fn execute_task_command_with_mode(
    args: &[&str],
    working_dir: &Path,
    interactive: bool,
) -> Result<()> {
    // First try direct execution
    if process::check_command_available("task").await {
        if interactive {
            // Note: task doesn't have --interactive flag, but we use interactive execution
            // to properly handle stdin/stdout for commands that task runs
            return process::execute_command_interactive("task", args, Some(working_dir))
                .await
                .map_err(|e| RazdError::task(format!("Failed to execute task: {}", e)));
        } else {
            return process::execute_command("task", args, Some(working_dir))
                .await
                .map_err(|e| RazdError::task(format!("Failed to execute task: {}", e)));
        }
    }

    // Fallback: try through mise exec (task should be available via mise)
    output::step("Executing task via mise...");
    let mut mise_args = vec!["exec", "task", "--", "task"];
    mise_args.extend(args);

    if interactive {
        process::execute_command_interactive("mise", &mise_args, Some(working_dir))
            .await
            .map_err(|e| RazdError::task(format!("Failed to execute task via mise: {}", e)))
    } else {
        process::execute_command("mise", &mise_args, Some(working_dir))
            .await
            .map_err(|e| RazdError::task(format!("Failed to execute task via mise: {}", e)))
    }
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

/// Execute a workflow task using custom taskfile content
pub async fn execute_workflow_task(task_name: &str, workflow_content: &str) -> Result<()> {
    execute_workflow_task_with_mode(task_name, workflow_content, false).await
}

/// Execute a workflow task with option for interactive mode
pub async fn execute_workflow_task_interactive(
    task_name: &str,
    workflow_content: &str,
) -> Result<()> {
    execute_workflow_task_with_mode(task_name, workflow_content, true).await
}

/// Execute a workflow task using custom taskfile content with interactive option
async fn execute_workflow_task_with_mode(
    task_name: &str,
    workflow_content: &str,
    interactive: bool,
) -> Result<()> {
    use std::env;
    use std::fs;

    // Get current working directory
    let working_dir = env::current_dir()
        .map_err(|e| RazdError::task(format!("Failed to get current directory: {}", e)))?;

    // Ensure task tool is available
    mise::ensure_tool_available("task", "latest", &working_dir).await?;

    output::step(&format!("Executing workflow: {}", task_name));

    // Create temporary taskfile in system temp directory for task to load.
    let temp_taskfile = env::temp_dir().join(format!("razd-workflow-{}.yml", task_name));

    fs::write(&temp_taskfile, workflow_content)
        .map_err(|e| RazdError::task(format!("Failed to create temporary taskfile: {}", e)))?;

    // Use --dir/-d flag to ensure task executes in project directory, not temp directory
    let args = vec![
        "--taskfile",
        temp_taskfile.to_str().unwrap(),
        "--dir",
        working_dir.to_str().unwrap(),
        task_name,
    ];

    // Check if we can execute task directly (allows early cleanup) or need mise exec (keeps file)
    let result = if process::check_command_available("task").await {
        // Direct execution: spawn, wait briefly for file load, cleanup, then wait for completion
        if interactive {
            let child = process::spawn_command_interactive("task", &args, Some(&working_dir))
                .map_err(|e| {
                    let _ = fs::remove_file(&temp_taskfile);
                    e
                })?;

            // Wait briefly to ensure the process has loaded the file
            tokio::time::sleep(Duration::from_millis(defaults::DEFAULT_SPAWN_DELAY_MS)).await;

            // Clean up temporary file immediately after process has had time to load it
            let _ = fs::remove_file(&temp_taskfile);

            // Wait for the task process to complete
            process::wait_for_command_interactive(child, "task").await
        } else {
            let child = process::spawn_command("task", &args, Some(&working_dir))
                .await
                .map_err(|e| {
                    let _ = fs::remove_file(&temp_taskfile);
                    e
                })?;

            // Wait briefly to ensure the process has loaded the file
            tokio::time::sleep(Duration::from_millis(defaults::DEFAULT_SPAWN_DELAY_MS)).await;

            // Clean up temporary file immediately after process has had time to load it
            let _ = fs::remove_file(&temp_taskfile);

            // Wait for the task process to complete
            process::wait_for_command(child, "task").await
        }
    } else {
        // Fallback via mise exec: file must exist for duration of execution
        // because mise spawns a subshell that then runs task
        output::step("Executing task via mise...");
        let mut mise_args = vec!["exec", "task", "--", "task"];
        mise_args.extend(&args);

        let result = if interactive {
            process::execute_command_interactive("mise", &mise_args, Some(&working_dir)).await
        } else {
            process::execute_command("mise", &mise_args, Some(&working_dir)).await
        };

        // Clean up temporary file after mise exec completes
        let _ = fs::remove_file(&temp_taskfile);

        result
    };

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
