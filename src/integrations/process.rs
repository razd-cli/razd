use crate::core::{output, Result, RazdError};
use std::path::Path;
use tokio::process::Command;

/// Execute a command and return success/failure
pub async fn execute_command(program: &str, args: &[&str], working_dir: Option<&Path>) -> Result<()> {
    output::step(&format!("Running: {} {}", program, args.join(" ")));

    let mut cmd = Command::new(program);
    cmd.args(args);

    if let Some(dir) = working_dir {
        cmd.current_dir(dir);
    }

    let output = cmd
        .output()
        .await
        .map_err(|e| RazdError::config(format!("Failed to execute {}: {}", program, e)))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(RazdError::config(format!(
            "Command '{}' failed with exit code {:?}:\n{}",
            program,
            output.status.code(),
            stderr
        )));
    }

    // Print stdout if there's output
    let stdout = String::from_utf8_lossy(&output.stdout);
    if !stdout.trim().is_empty() {
        println!("{}", stdout);
    }

    Ok(())
}

/// Check if a command is available in PATH
pub async fn check_command_available(program: &str) -> bool {
    Command::new(program)
        .arg("--version")
        .output()
        .await
        .map(|output| output.status.success())
        .unwrap_or(false)
}