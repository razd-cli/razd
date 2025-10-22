use crate::core::{output, Result};
use crate::integrations::taskfile;
use std::env;

/// Execute the `razd setup` command: install project dependencies via task setup
pub async fn execute() -> Result<()> {
    let current_dir = env::current_dir()?;
    output::info("Setting up project dependencies...");

    taskfile::setup_project(&current_dir).await?;

    Ok(())
}
