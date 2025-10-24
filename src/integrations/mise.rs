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

/// Install a specific tool using mise
pub async fn install_specific_tool(tool: &str, version: &str, working_dir: &Path) -> Result<()> {
    // Check if mise is available
    if !process::check_command_available("mise").await {
        return Err(RazdError::missing_tool(
            "mise",
            "https://mise.jdx.dev/getting-started.html",
        ));
    }

    output::step(&format!("Installing {} via mise...", tool));

    let tool_spec = format!("{}@{}", tool, version);
    let args = vec!["install", &tool_spec];

    process::execute_command("mise", &args, Some(working_dir))
        .await
        .map_err(|e| {
            RazdError::mise(format!(
                "Failed to install {}: {}\n\
                 Please install {} manually: https://taskfile.dev/installation/",
                tool, e, tool
            ))
        })?;

    output::success(&format!("✓ {} installed successfully", tool));
    Ok(())
}

/// Ensure a tool is available, installing it via mise if necessary
pub async fn ensure_tool_available(tool: &str, version: &str, working_dir: &Path) -> Result<()> {
    // Fast path: check if tool is already available
    if process::check_command_available(tool).await {
        return Ok(());
    }

    // Install tool via mise
    install_specific_tool(tool, version, working_dir).await?;

    // Verify installation
    if !process::check_command_available(tool).await {
        return Err(RazdError::config(format!(
            "Tool '{}' was installed but is not accessible.\n\
             This might be a PATH configuration issue.\n\
             Try running: mise reshim\n\
             Or install manually: https://taskfile.dev/installation/",
            tool
        )));
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_has_mise_config_with_mise_toml() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join(".mise.toml"), "").unwrap();

        assert!(has_mise_config(temp_dir.path()));
    }

    #[test]
    fn test_has_mise_config_with_tool_versions() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join(".tool-versions"), "").unwrap();

        assert!(has_mise_config(temp_dir.path()));
    }

    #[test]
    fn test_has_mise_config_with_neither() {
        let temp_dir = TempDir::new().unwrap();

        assert!(!has_mise_config(temp_dir.path()));
    }

    #[test]
    fn test_has_mise_config_with_both() {
        let temp_dir = TempDir::new().unwrap();
        std::fs::write(temp_dir.path().join(".mise.toml"), "").unwrap();
        std::fs::write(temp_dir.path().join(".tool-versions"), "").unwrap();

        assert!(has_mise_config(temp_dir.path()));
    }

    // Note: The async functions install_specific_tool and ensure_tool_available
    // require external processes and are better tested as integration tests
    // rather than unit tests, since they depend on mise being installed.
}
