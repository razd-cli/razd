use crate::core::{output, Result};
use crate::integrations::mise;
use std::env;

/// Execute the `razd install` command: install development tools via mise
pub async fn execute() -> Result<()> {
    let current_dir = env::current_dir()?;
    output::info("Installing development tools...");
    
    mise::install_tools(&current_dir).await?;
    
    Ok(())
}