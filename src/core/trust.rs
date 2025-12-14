//! Trust management for razd projects
//!
//! This module provides functionality to track which project directories
//! the user has explicitly trusted for execution.

use crate::core::{output, RazdError, Result};
use crate::integrations::{mise, process};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};

/// Status of a project's trust
#[derive(Debug, Clone, PartialEq)]
pub enum TrustStatus {
    /// Project is explicitly trusted
    Trusted,
    /// Project is explicitly ignored (never trust)
    Ignored,
    /// Project has no trust status set
    Unknown,
}

/// Entry for a trusted path
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrustedEntry {
    pub path: String,
    pub trusted_at: String,
}

/// Entry for an ignored path
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IgnoredEntry {
    pub path: String,
    pub ignored_at: String,
}

/// Trust store data structure
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct TrustStoreData {
    pub version: u32,
    #[serde(default)]
    pub trusted: Vec<TrustedEntry>,
    #[serde(default)]
    pub ignored: Vec<IgnoredEntry>,
}

/// Trust store manager
pub struct TrustStore {
    data: TrustStoreData,
    path: PathBuf,
}

impl TrustStore {
    /// Get the path to the trust store file
    pub fn get_store_path() -> PathBuf {
        if let Some(cache_dir) = dirs::cache_dir() {
            cache_dir.join("razd").join("trusted.json")
        } else if let Some(data_dir) = dirs::data_local_dir() {
            data_dir.join("razd").join("trusted.json")
        } else {
            // Fallback to home directory
            dirs::home_dir()
                .unwrap_or_else(|| PathBuf::from("."))
                .join(".razd")
                .join("trusted.json")
        }
    }

    /// Load trust store from disk
    pub fn load() -> Result<Self> {
        let path = Self::get_store_path();

        let data = if path.exists() {
            let content = fs::read_to_string(&path)
                .map_err(|e| RazdError::config(format!("Failed to read trust store: {}", e)))?;
            serde_json::from_str(&content).unwrap_or_default()
        } else {
            TrustStoreData {
                version: 1,
                trusted: Vec::new(),
                ignored: Vec::new(),
            }
        };

        Ok(Self { data, path })
    }

    /// Save trust store to disk
    pub fn save(&self) -> Result<()> {
        // Ensure directory exists
        if let Some(parent) = self.path.parent() {
            fs::create_dir_all(parent).map_err(|e| {
                RazdError::config(format!("Failed to create trust store directory: {}", e))
            })?;
        }

        let content = serde_json::to_string_pretty(&self.data)
            .map_err(|e| RazdError::config(format!("Failed to serialize trust store: {}", e)))?;

        fs::write(&self.path, content)
            .map_err(|e| RazdError::config(format!("Failed to write trust store: {}", e)))?;

        Ok(())
    }

    /// Normalize a path for consistent comparison
    fn normalize_path(path: &Path) -> String {
        let canonical = path.canonicalize().unwrap_or_else(|_| path.to_path_buf());
        let path_str = canonical.to_string_lossy().to_string();

        // On Windows, normalize to lowercase for case-insensitive comparison
        #[cfg(windows)]
        {
            path_str.to_lowercase().replace('\\', "/")
        }
        #[cfg(not(windows))]
        {
            path_str
        }
    }

    /// Check if a path is trusted
    pub fn is_trusted(&self, path: &Path) -> bool {
        let normalized = Self::normalize_path(path);
        self.data
            .trusted
            .iter()
            .any(|entry| entry.path == normalized)
    }

    /// Check if a path is ignored
    pub fn is_ignored(&self, path: &Path) -> bool {
        let normalized = Self::normalize_path(path);
        self.data
            .ignored
            .iter()
            .any(|entry| entry.path == normalized)
    }

    /// Get the trust status of a path
    pub fn get_status(&self, path: &Path) -> TrustStatus {
        if self.is_trusted(path) {
            TrustStatus::Trusted
        } else if self.is_ignored(path) {
            TrustStatus::Ignored
        } else {
            TrustStatus::Unknown
        }
    }

    /// Add a path to the trusted list
    pub fn add_trusted(&mut self, path: &Path) -> Result<()> {
        let normalized = Self::normalize_path(path);

        // Remove from ignored if present
        self.data.ignored.retain(|entry| entry.path != normalized);

        // Add to trusted if not already present
        if !self
            .data
            .trusted
            .iter()
            .any(|entry| entry.path == normalized)
        {
            self.data.trusted.push(TrustedEntry {
                path: normalized,
                trusted_at: chrono_now(),
            });
        }

        self.save()
    }

    /// Remove a path from the trusted list
    pub fn remove_trusted(&mut self, path: &Path) -> Result<()> {
        let normalized = Self::normalize_path(path);
        self.data.trusted.retain(|entry| entry.path != normalized);
        self.save()
    }

    /// Add a path to the ignored list
    pub fn add_ignored(&mut self, path: &Path) -> Result<()> {
        let normalized = Self::normalize_path(path);

        // Remove from trusted if present
        self.data.trusted.retain(|entry| entry.path != normalized);

        // Add to ignored if not already present
        if !self
            .data
            .ignored
            .iter()
            .any(|entry| entry.path == normalized)
        {
            self.data.ignored.push(IgnoredEntry {
                path: normalized,
                ignored_at: chrono_now(),
            });
        }

        self.save()
    }

    /// Remove a path from the ignored list
    pub fn remove_ignored(&mut self, path: &Path) -> Result<()> {
        let normalized = Self::normalize_path(path);
        self.data.ignored.retain(|entry| entry.path != normalized);
        self.save()
    }

    /// Remove a path from both trusted and ignored lists
    pub fn remove_all(&mut self, path: &Path) -> Result<()> {
        let normalized = Self::normalize_path(path);
        self.data.trusted.retain(|entry| entry.path != normalized);
        self.data.ignored.retain(|entry| entry.path != normalized);
        self.save()
    }
}

/// Get current timestamp as ISO 8601 string
fn chrono_now() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};
    let duration = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default();
    let secs = duration.as_secs();
    // Simple ISO-like format without external crate
    format!("{}", secs)
}

/// Trust prompt response
#[derive(Debug, Clone, PartialEq)]
pub enum TrustResponse {
    Yes,
    No,
    Ignore,
}

/// Show interactive trust prompt using dialoguer
pub fn prompt_trust(path: &Path) -> Result<TrustResponse> {
    use dialoguer::{theme::ColorfulTheme, Select};

    let path_display = path.display();
    let prompt_message = format!(
        "razd config files in {} are not trusted. Trust them?",
        path_display
    );

    println!("{}", prompt_message);

    let items = vec!["Yes", "No", "Ignore"];
    let selection = Select::with_theme(&ColorfulTheme::default())
        .items(&items)
        .default(1) // Default to "No" for safety
        .interact()
        .map_err(|e| RazdError::config(format!("Failed to get user input: {}", e)))?;

    match selection {
        0 => Ok(TrustResponse::Yes),
        1 => Ok(TrustResponse::No),
        2 => Ok(TrustResponse::Ignore),
        _ => Ok(TrustResponse::No),
    }
}

/// Ensure a project is trusted before executing commands
///
/// This function checks if the project is trusted, prompts the user if not,
/// and runs `mise trust` if the project becomes trusted.
pub async fn ensure_trusted(path: &Path, auto_yes: bool) -> Result<()> {
    // Check if project has configuration files
    if !has_razd_config(path) {
        // No config means no danger, allow execution
        return Ok(());
    }

    let mut store = TrustStore::load()?;

    match store.get_status(path) {
        TrustStatus::Trusted => {
            // Already trusted, proceed
            Ok(())
        }
        TrustStatus::Ignored => {
            // Explicitly ignored, block execution
            Err(RazdError::config(format!(
                "Project is ignored: {}\n\nThis project was previously marked as ignored.\nTo remove from ignore list, run:\n  razd trust --untrust\n  razd trust",
                path.display()
            )))
        }
        TrustStatus::Unknown => {
            if auto_yes {
                // Auto-trust with --yes flag
                output::step("Auto-trusting project (--yes flag)");
                store.add_trusted(path)?;
                run_mise_trust_if_needed(path).await?;
                Ok(())
            } else {
                // Show interactive prompt
                match prompt_trust(path)? {
                    TrustResponse::Yes => {
                        store.add_trusted(path)?;
                        output::success("Project trusted");
                        run_mise_trust_if_needed(path).await?;
                        Ok(())
                    }
                    TrustResponse::No => {
                        Err(RazdError::config(format!(
                            "Project is not trusted: {}\n\nTo trust this project, run:\n  razd trust\n\nOr run with --yes to auto-trust:\n  razd --yes up",
                            path.display()
                        )))
                    }
                    TrustResponse::Ignore => {
                        store.add_ignored(path)?;
                        Err(RazdError::config(format!(
                            "Project added to ignore list: {}",
                            path.display()
                        )))
                    }
                }
            }
        }
    }
}

/// Check if directory has razd configuration
fn has_razd_config(dir: &Path) -> bool {
    dir.join("Razdfile.yml").exists()
        || dir.join("Taskfile.yml").exists()
        || dir.join("mise.toml").exists()
        || dir.join(".mise.toml").exists()
}

/// Run mise trust if mise config exists
pub async fn run_mise_trust_if_needed(path: &Path) -> Result<()> {
    if mise::has_mise_config(path) {
        // Check if mise is available
        if process::check_command_available("mise").await {
            output::step("Running mise trust...");
            process::execute_command_interactive("mise", &["trust"], Some(path))
                .await
                .map_err(|e| RazdError::mise(format!("Failed to run mise trust: {}", e)))?;
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_trust_store_load_empty() {
        let store = TrustStore::load().unwrap();
        assert!(store.data.trusted.is_empty());
        assert!(store.data.ignored.is_empty());
    }

    #[test]
    fn test_trust_status_unknown_by_default() {
        let store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();
        assert_eq!(store.get_status(temp_dir.path()), TrustStatus::Unknown);
    }

    #[test]
    fn test_add_and_check_trusted() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_trusted(temp_dir.path()).unwrap();
        assert!(store.is_trusted(temp_dir.path()));
        assert!(!store.is_ignored(temp_dir.path()));
        assert_eq!(store.get_status(temp_dir.path()), TrustStatus::Trusted);
    }

    #[test]
    fn test_add_and_check_ignored() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_ignored(temp_dir.path()).unwrap();
        assert!(!store.is_trusted(temp_dir.path()));
        assert!(store.is_ignored(temp_dir.path()));
        assert_eq!(store.get_status(temp_dir.path()), TrustStatus::Ignored);
    }

    #[test]
    fn test_trust_replaces_ignore() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_ignored(temp_dir.path()).unwrap();
        assert!(store.is_ignored(temp_dir.path()));

        store.add_trusted(temp_dir.path()).unwrap();
        assert!(store.is_trusted(temp_dir.path()));
        assert!(!store.is_ignored(temp_dir.path()));
    }

    #[test]
    fn test_ignore_replaces_trust() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_trusted(temp_dir.path()).unwrap();
        assert!(store.is_trusted(temp_dir.path()));

        store.add_ignored(temp_dir.path()).unwrap();
        assert!(!store.is_trusted(temp_dir.path()));
        assert!(store.is_ignored(temp_dir.path()));
    }

    #[test]
    fn test_remove_trusted() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_trusted(temp_dir.path()).unwrap();
        assert!(store.is_trusted(temp_dir.path()));

        store.remove_trusted(temp_dir.path()).unwrap();
        assert!(!store.is_trusted(temp_dir.path()));
    }

    #[test]
    fn test_remove_all() {
        let mut store = TrustStore::load().unwrap();
        let temp_dir = TempDir::new().unwrap();

        store.add_trusted(temp_dir.path()).unwrap();
        store.remove_all(temp_dir.path()).unwrap();
        assert_eq!(store.get_status(temp_dir.path()), TrustStatus::Unknown);
    }
}
