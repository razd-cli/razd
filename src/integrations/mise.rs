use crate::core::{output, RazdError, Result};
use crate::integrations::process;
use std::path::Path;

/// Check if mise configuration exists in the directory
pub fn has_mise_config(dir: &Path) -> bool {
    dir.join(".mise.toml").exists() || dir.join(".tool-versions").exists()
}

/// Install tools using mise
pub async fn install_tools(working_dir: &Path) -> Result<()> {
    // Check if mise is available
    if !process::check_command_available("mise").await {
        return Err(RazdError::missing_tool(
            "mise",
            "https://mise.jdx.dev/getting-started.html",
        ));
    }

    // Check if mise configuration exists
    if !has_mise_config(working_dir) {
        output::warning("No mise configuration found (.mise.toml or .tool-versions), skipping tool installation");
        return Ok(());
    }

    output::step("Installing development tools with mise");

    process::execute_command("mise", &["install"], Some(working_dir))
        .await
        .map_err(|e| RazdError::mise(format!("Failed to install tools: {}", e)))?;

    output::success("Successfully installed development tools");

    Ok(())
}
