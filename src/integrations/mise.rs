use crate::config::RazdfileConfig;
use crate::core::{output, RazdError, Result};
use crate::integrations::process;
use std::path::Path;

/// Check if mise configuration exists in the directory
pub fn has_mise_config(dir: &Path) -> bool {
    // Check for Razdfile.yml with mise section first
    if let Ok(Some(razdfile)) = RazdfileConfig::load_from_path(dir.join("Razdfile.yml")) {
        if razdfile.mise.is_some() {
            return true;
        }
    }

    // Fallback to traditional mise config files
    dir.join("mise.toml").exists()
        || dir.join(".mise.toml").exists()
        || dir.join(".tool-versions").exists()
}

/// Trust mise configuration files in the directory
pub async fn trust_config(working_dir: &Path) -> Result<()> {
    // Check if mise is available
    if !process::check_command_available("mise").await {
        return Err(RazdError::missing_tool(
            "mise",
            "https://mise.jdx.dev/getting-started.html",
        ));
    }

    // Check if mise configuration exists
    if !has_mise_config(working_dir) {
        return Ok(());
    }

    output::step("Trusting mise configuration...");

    // Run mise trust interactively so user can confirm
    process::execute_command_interactive("mise", &["trust"], Some(working_dir))
        .await
        .map_err(|e| RazdError::mise(format!("Failed to trust configuration: {}", e)))?;

    Ok(())
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
        output::warning("No mise configuration found (Razdfile.yml, mise.toml or .tool-versions), skipping tool installation");
        return Ok(());
    }

    // Trust configuration first
    trust_config(working_dir).await?;

    output::step("Installing development tools with mise");

    process::execute_command_interactive("mise", &["install"], Some(working_dir))
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

    // Trust configuration first (if exists)
    let _ = trust_config(working_dir).await; // Ignore error if no config to trust

    output::step(&format!("Installing {} via mise...", tool));

    // Install the tool
    let tool_spec = format!("{}@{}", tool, version);
    let install_args = vec!["install", &tool_spec];

    process::execute_command_interactive("mise", &install_args, Some(working_dir))
        .await
        .map_err(|e| {
            RazdError::mise(format!(
                "Failed to install {}: {}\n\
                 Please install {} manually: https://taskfile.dev/installation/",
                tool, e, tool
            ))
        })?;

    // Use the tool to make it available in current environment
    output::step(&format!(
        "Making {} available in current environment...",
        tool
    ));
    let use_args = vec!["use", &tool_spec];

    process::execute_command("mise", &use_args, Some(working_dir))
        .await
        .map_err(|e| {
            output::warning(&format!("Could not set {} as active version: {}", tool, e));
            e // Still propagate the error but with a warning
        })?;

    output::success(&format!("✓ {} installed and activated successfully", tool));
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

    // Note: After 'mise use', the tool should be available in the current directory
    // No need for complex verification - mise handles this

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
