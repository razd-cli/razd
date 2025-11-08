use crate::core::{output, Result};
use crate::integrations::taskfile;
use std::env;
use std::path::PathBuf;

/// Execute the `razd setup` command: install project dependencies via task setup
pub async fn execute(custom_path: Option<PathBuf>) -> Result<()> {
    let current_dir = env::current_dir()?;

    // Check and sync mise configuration before executing
    if let Err(e) = crate::config::check_and_sync_mise(&current_dir) {
        output::warning(&format!("Mise sync check failed: {}", e));
    }

    output::info("Setting up project dependencies...");

    taskfile::setup_project_with_path(&current_dir, custom_path).await?;

    Ok(())
}
