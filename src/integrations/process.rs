use crate::core::{output, RazdError, Result};
use std::path::Path;
use tokio::process::Command;

/// Spawn a command and return the child process handle
pub async fn spawn_command(
    program: &str,
    args: &[&str],
    working_dir: Option<&Path>,
) -> Result<tokio::process::Child> {
    let mut cmd = Command::new(program);
    cmd.args(args);

    // Inherit current environment to ensure tools are found
    cmd.env_clear().envs(std::env::vars());

    if let Some(dir) = working_dir {
        cmd.current_dir(dir);
    }

    cmd.spawn()
        .map_err(|e| RazdError::config(format!("Failed to spawn {}: {}", program, e)))
}

/// Spawn an interactive command and return the child process handle
pub fn spawn_command_interactive(
    program: &str,
    args: &[&str],
    working_dir: Option<&Path>,
) -> Result<std::process::Child> {
    let mut cmd = std::process::Command::new(program);
    cmd.args(args);

    // Inherit current environment and stdio for interactive execution
    cmd.env_clear().envs(std::env::vars());
    cmd.stdin(std::process::Stdio::inherit());
    cmd.stdout(std::process::Stdio::inherit());
    cmd.stderr(std::process::Stdio::inherit());

    if let Some(dir) = working_dir {
        cmd.current_dir(dir);
    }

    cmd.spawn()
        .map_err(|e| RazdError::config(format!("Failed to spawn {}: {}", program, e)))
}

/// Wait for a spawned process to complete
#[allow(unused_mut)]
pub async fn wait_for_command(mut child: tokio::process::Child, program: &str) -> Result<()> {
    let output = child
        .wait_with_output()
        .await
        .map_err(|e| RazdError::config(format!("Failed to wait for {}: {}", program, e)))?;

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

/// Wait for an interactive spawned process to complete
pub async fn wait_for_command_interactive(mut child: std::process::Child, program: &str) -> Result<()> {
    let status = tokio::task::spawn_blocking(move || child.wait())
        .await
        .map_err(|e| RazdError::config(format!("Failed to wait for task: {}", e)))?
        .map_err(|e| RazdError::config(format!("Failed to wait for {}: {}", program, e)))?;

    if !status.success() {
        return Err(RazdError::config(format!(
            "Command '{}' failed with exit code {:?}",
            program,
            status.code()
        )));
    }

    Ok(())
}

/// Execute a command and return success/failure
pub async fn execute_command(
    program: &str,
    args: &[&str],
    working_dir: Option<&Path>,
) -> Result<()> {
    output::step(&format!("Running: {} {}", program, args.join(" ")));

    let child = spawn_command(program, args, working_dir).await?;
    wait_for_command(child, program).await
}

/// Execute a command interactively, showing real-time output
pub async fn execute_command_interactive(
    program: &str,
    args: &[&str],
    working_dir: Option<&Path>,
) -> Result<()> {
    output::step(&format!("Running: {} {}", program, args.join(" ")));

    let child = spawn_command_interactive(program, args, working_dir)?;
    wait_for_command_interactive(child, program).await
}

/// Check if a command is available in PATH
pub async fn check_command_available(program: &str) -> bool {
    // On Windows, also try the .exe extension
    let exe_name = format!("{}.exe", program);
    let programs_to_try = if cfg!(windows) {
        vec![program, exe_name.as_str()]
    } else {
        vec![program]
    };

    for prog in programs_to_try {
        // Try with --version flag first
        if let Ok(output) = Command::new(prog)
            .arg("--version")
            .env_clear()
            .envs(std::env::vars())
            .output()
            .await
        {
            if output.status.success() {
                return true;
            }
        }

        // Fallback: try with -v flag (some tools use this instead)
        if let Ok(output) = Command::new(prog)
            .arg("-v")
            .env_clear()
            .envs(std::env::vars())
            .output()
            .await
        {
            if output.status.success() {
                return true;
            }
        }
    }

    false
}
