//! Trust command implementation
//!
//! Manages project trust status for razd.

use crate::core::trust::{run_mise_trust_if_needed, TrustStatus, TrustStore};
use crate::core::{output, Result};
use crate::integrations::mise;
use std::env;
use std::path::{Path, PathBuf};

/// Execute the trust command
pub async fn execute(
    path: Option<&str>,
    untrust: bool,
    show: bool,
    all: bool,
    ignore: bool,
) -> Result<()> {
    // Determine the target path
    let target_path = if let Some(p) = path {
        PathBuf::from(p)
    } else {
        env::current_dir()?
    };

    // Canonicalize the path
    let target_path = target_path.canonicalize().unwrap_or(target_path);

    if show {
        // Show trust status
        show_trust_status(&target_path)?;
    } else if untrust {
        // Remove from trust/ignore lists
        untrust_path(&target_path)?;
    } else if ignore {
        // Add to ignore list
        ignore_path(&target_path)?;
    } else if all {
        // Trust all parent directories with config
        trust_all(&target_path).await?;
    } else {
        // Trust the directory
        trust_path(&target_path).await?;
    }

    Ok(())
}

/// Show the trust status of a path
fn show_trust_status(path: &Path) -> Result<()> {
    let store = TrustStore::load()?;
    let status = store.get_status(path);

    output::info(&format!("Path: {}", path.display()));

    match status {
        TrustStatus::Trusted => {
            output::success("Status: Trusted ✓");
        }
        TrustStatus::Ignored => {
            output::warning("Status: Ignored (will not execute)");
        }
        TrustStatus::Unknown => {
            output::info("Status: Not trusted (will prompt on first run)");
        }
    }

    // Show mise trust status if mise config exists
    if mise::has_mise_config(path) {
        output::info("Mise config: Present");
    }

    Ok(())
}

/// Trust a path
async fn trust_path(path: &Path) -> Result<()> {
    let mut store = TrustStore::load()?;

    if store.is_trusted(path) {
        output::info(&format!("Already trusted: {}", path.display()));
        return Ok(());
    }

    store.add_trusted(path)?;
    output::success(&format!("Trusted: {}", path.display()));

    // Run mise trust if mise config exists
    run_mise_trust_if_needed(path).await?;

    Ok(())
}

/// Remove trust/ignore status from a path
fn untrust_path(path: &Path) -> Result<()> {
    let mut store = TrustStore::load()?;

    let status = store.get_status(path);
    store.remove_all(path)?;

    match status {
        TrustStatus::Trusted => {
            output::success(&format!("Removed trust: {}", path.display()));
        }
        TrustStatus::Ignored => {
            output::success(&format!("Removed ignore: {}", path.display()));
        }
        TrustStatus::Unknown => {
            output::info(&format!("Path was not in trust store: {}", path.display()));
        }
    }

    Ok(())
}

/// Add a path to the ignore list
fn ignore_path(path: &Path) -> Result<()> {
    let mut store = TrustStore::load()?;

    if store.is_ignored(path) {
        output::info(&format!("Already ignored: {}", path.display()));
        return Ok(());
    }

    store.add_ignored(path)?;
    output::success(&format!("Ignored: {}", path.display()));
    output::info("This project will not execute and you won't be prompted again.");

    Ok(())
}

/// Trust all parent directories with config
async fn trust_all(path: &Path) -> Result<()> {
    let mut current = path.to_path_buf();
    let mut trusted_count = 0;

    loop {
        // Check if this directory has config
        if has_config(&current) {
            let mut store = TrustStore::load()?;
            if !store.is_trusted(&current) {
                store.add_trusted(&current)?;
                output::success(&format!("Trusted: {}", current.display()));
                run_mise_trust_if_needed(&current).await?;
                trusted_count += 1;
            }
        }

        // Move to parent
        if let Some(parent) = current.parent() {
            if parent == current {
                break;
            }
            current = parent.to_path_buf();
        } else {
            break;
        }
    }

    if trusted_count == 0 {
        output::info("No new directories to trust");
    } else {
        output::success(&format!("Trusted {} directories", trusted_count));
    }

    Ok(())
}

/// Check if directory has any config file
fn has_config(dir: &Path) -> bool {
    dir.join("Razdfile.yml").exists()
        || dir.join("Taskfile.yml").exists()
        || dir.join("mise.toml").exists()
        || dir.join(".mise.toml").exists()
}
