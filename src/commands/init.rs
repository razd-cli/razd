use crate::core::{output, Result, RazdError};
use crate::config::{detection, defaults};
use std::fs;
use std::path::Path;
use std::env;

/// Execute the `razd init` command: initialize razd configuration for a project
pub async fn execute(create_config: bool, create_full: bool) -> Result<()> {
    let current_dir = env::current_dir()?;
    
    if !create_config && !create_full {
        // Default behavior - just show info about built-in workflows
        output::info("razd is ready to use with built-in workflows!");
        output::info("Available commands:");
        output::info("  razd up <url>       # Clone and set up project");
        output::info("  razd install        # Install development tools");
        output::info("  razd dev            # Start development workflow");
        output::info("  razd build          # Build project");
        output::info("  razd task <name>    # Run specific task");
        output::info("");
        output::info("To customize workflows:");
        output::info("  razd init --config  # Create Razdfile.yml");
        output::info("  razd init --full    # Create all config files");
        return Ok(());
    }
    
    output::info("Initializing project configuration...");
    
    // Detect project type
    let project_type = detection::detect_project_type(&current_dir);
    output::info(&format!("Detected project type: {}", project_type));
    
    if create_config || create_full {
        // Create Razdfile.yml
        create_razdfile(&current_dir, &project_type)?;
        output::success("Created Razdfile.yml");
    }
    
    if create_full {
        // Create Taskfile.yml if it doesn't exist
        if !Path::new("Taskfile.yml").exists() && !Path::new("Taskfile.yaml").exists() {
            create_taskfile(&current_dir, &project_type)?;
            output::success("Created Taskfile.yml");
        } else {
            output::info("Taskfile.yml already exists, skipping");
        }
        
        // Create mise.toml if it doesn't exist
        if !Path::new("mise.toml").exists() && !Path::new(".tool-versions").exists() {
            create_mise_config(&current_dir, &project_type)?;
            output::success("Created mise.toml");
        } else {
            output::info("mise configuration already exists, skipping");
        }
    }
    
    output::success("Project initialization completed!");
    output::info("Next steps:");
    output::info("  razd install        # Install tools");
    output::info("  razd dev            # Start development");
    
    Ok(())
}

fn create_razdfile(dir: &Path, project_type: &str) -> Result<()> {
    let content = defaults::generate_project_razdfile(project_type);
    let path = dir.join("Razdfile.yml");
    
    fs::write(path, content)
        .map_err(|e| RazdError::config(format!("Failed to create Razdfile.yml: {}", e)))?;
    
    Ok(())
}

fn create_taskfile(dir: &Path, project_type: &str) -> Result<()> {
    let content = detection::generate_taskfile_config(project_type);
    let path = dir.join("Taskfile.yml");
    
    fs::write(path, content)
        .map_err(|e| RazdError::config(format!("Failed to create Taskfile.yml: {}", e)))?;
    
    Ok(())
}

fn create_mise_config(dir: &Path, project_type: &str) -> Result<()> {
    let content = detection::generate_mise_config(project_type);
    if content.is_empty() {
        return Ok(()); // No tools to configure for this project type
    }
    
    let path = dir.join("mise.toml");
    
    fs::write(path, content)
        .map_err(|e| RazdError::config(format!("Failed to create mise.toml: {}", e)))?;
    
    Ok(())
}