use crate::config::canonical::{compute_razdfile_semantic_hash, compute_mise_toml_semantic_hash};
use crate::core::{RazdError, Result};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::SystemTime;

/// File tracking state for mise configuration sync
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileTrackingState {
    /// Semantic hash of Razdfile.yml content
    pub razdfile_hash: Option<String>,
    /// Semantic hash of mise.toml content
    pub mise_toml_hash: Option<String>,
    /// Format version for tracking state (for future migrations)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format_version: Option<String>,
    /// Legacy fields for backwards compatibility
    #[serde(skip_serializing_if = "Option::is_none")]
    pub razdfile_modified: Option<SystemTime>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mise_toml_modified: Option<SystemTime>,
    pub last_sync_time: SystemTime,
}

/// Change detection result
#[derive(Debug, PartialEq)]
pub enum ChangeDetection {
    NoChanges,
    RazdfileChanged,
    MiseTomlChanged,
    BothChanged,
}

/// Get the tracking file path for a project
pub fn get_tracking_file_path(project_dir: &Path) -> Result<PathBuf> {
    // Get platform-specific data directory
    let data_dir = get_data_dir()?;

    // Hash the absolute project path
    let abs_path = project_dir
        .canonicalize()
        .map_err(|e| RazdError::config(format!("Failed to canonicalize project path: {}", e)))?;

    let hash = hash_path(&abs_path);

    // Create tracking path: <data_dir>/file_tracking/<hash>/tracking.json
    let tracking_path = data_dir
        .join("file_tracking")
        .join(&hash)
        .join("tracking.json");

    Ok(tracking_path)
}

/// Get platform-specific data directory
fn get_data_dir() -> Result<PathBuf> {
    #[cfg(target_os = "windows")]
    {
        if let Ok(local_app_data) = std::env::var("LOCALAPPDATA") {
            Ok(PathBuf::from(local_app_data).join("razd"))
        } else {
            Err(RazdError::config(
                "LOCALAPPDATA environment variable not set".to_string(),
            ))
        }
    }

    #[cfg(not(target_os = "windows"))]
    {
        if let Ok(home) = std::env::var("HOME") {
            Ok(PathBuf::from(home).join(".local/share/razd"))
        } else {
            Err(RazdError::config("HOME environment variable not set".to_string()))
        }
    }
}

/// Hash a path using SHA256
fn hash_path(path: &Path) -> String {
    let path_str = path.to_string_lossy();
    let mut hasher = Sha256::new();
    hasher.update(path_str.as_bytes());
    let result = hasher.finalize();
    format!("{:x}", result)
}

/// Load tracking state from file
pub fn load_tracking_state(project_dir: &Path) -> Result<Option<FileTrackingState>> {
    let tracking_path = get_tracking_file_path(project_dir)?;

    if !tracking_path.exists() {
        return Ok(None);
    }

    let content = fs::read_to_string(&tracking_path).map_err(|e| {
        RazdError::config(format!("Failed to read tracking state: {}", e))
    })?;

    let state: FileTrackingState = serde_json::from_str(&content).map_err(|e| {
        RazdError::config(format!("Failed to parse tracking state: {}", e))
    })?;

    Ok(Some(state))
}

/// Save tracking state to file
pub fn save_tracking_state(project_dir: &Path, state: &FileTrackingState) -> Result<()> {
    let tracking_path = get_tracking_file_path(project_dir)?;

    // Create parent directories if they don't exist
    if let Some(parent) = tracking_path.parent() {
        fs::create_dir_all(parent).map_err(|e| {
            RazdError::config(format!("Failed to create tracking directory: {}", e))
        })?;
    }

    // Serialize state to JSON
    let content = serde_json::to_string_pretty(&state).map_err(|e| {
        RazdError::config(format!("Failed to serialize tracking state: {}", e))
    })?;

    // Write atomically using temp file + rename
    atomic_write_file(&tracking_path, &content)?;

    Ok(())
}

/// Atomically write file content
pub fn atomic_write_file(path: &Path, content: &str) -> Result<()> {
    // Create temp file in same directory
    let temp_path = path.with_extension("tmp");

    // Write to temp file
    fs::write(&temp_path, content).map_err(|e| {
        RazdError::config(format!("Failed to write temporary file: {}", e))
    })?;

    // Rename temp file to target (atomic on most filesystems)
    fs::rename(&temp_path, path).map_err(|e| {
        // Clean up temp file on error
        let _ = fs::remove_file(&temp_path);
        RazdError::config(format!("Failed to rename file atomically: {}", e))
    })?;

    Ok(())
}

/// Atomically write file content
pub fn compute_semantic_hash(path: &Path) -> Result<String> {
    let file_name = path
        .file_name()
        .and_then(|n| n.to_str())
        .ok_or_else(|| RazdError::config("Invalid file path"))?;

    match file_name {
        "Razdfile.yml" => compute_razdfile_semantic_hash(path),
        "mise.toml" => compute_mise_toml_semantic_hash(path),
        _ => {
            // Fallback to content hash for unknown files
            let content = fs::read_to_string(path)?;
            let mut hasher = Sha256::new();
            hasher.update(content.as_bytes());
            Ok(format!("{:x}", hasher.finalize()))
        }
    }
}

/// Check for file changes using semantic hashing
pub fn check_file_changes(project_dir: &Path) -> Result<ChangeDetection> {
    let razdfile_path = project_dir.join("Razdfile.yml");
    let mise_toml_path = project_dir.join("mise.toml");

    // Load tracking state (migrate if needed)
    let tracking_state = load_tracking_state(project_dir)?;

    // Compute current hashes
    let current_razdfile_hash = if razdfile_path.exists() {
        Some(compute_semantic_hash(&razdfile_path)?)
    } else {
        None
    };

    let current_mise_hash = if mise_toml_path.exists() {
        Some(compute_semantic_hash(&mise_toml_path)?)
    } else {
        None
    };

    match tracking_state {
        None => {
            // First run - no tracking state exists
            if current_razdfile_hash.is_some() || current_mise_hash.is_some() {
                Ok(ChangeDetection::RazdfileChanged) // Treat as Razdfile change to generate mise.toml
            } else {
                Ok(ChangeDetection::NoChanges)
            }
        }
        Some(state) => {
            let razdfile_changed = current_razdfile_hash != state.razdfile_hash;
            let mise_toml_changed = current_mise_hash != state.mise_toml_hash;

            if razdfile_changed && mise_toml_changed {
                Ok(ChangeDetection::BothChanged)
            } else if razdfile_changed {
                Ok(ChangeDetection::RazdfileChanged)
            } else if mise_toml_changed {
                Ok(ChangeDetection::MiseTomlChanged)
            } else {
                Ok(ChangeDetection::NoChanges)
            }
        }
    }
}

/// Update tracking state after a sync operation
pub fn update_tracking_state(project_dir: &Path) -> Result<()> {
    let razdfile_path = project_dir.join("Razdfile.yml");
    let mise_toml_path = project_dir.join("mise.toml");

    let razdfile_hash = if razdfile_path.exists() {
        Some(compute_semantic_hash(&razdfile_path)?)
    } else {
        None
    };

    let mise_toml_hash = if mise_toml_path.exists() {
        Some(compute_semantic_hash(&mise_toml_path)?)
    } else {
        None
    };

    let state = FileTrackingState {
        razdfile_hash,
        mise_toml_hash,
        format_version: Some("semantic-v1".to_string()),
        razdfile_modified: None,
        mise_toml_modified: None,
        last_sync_time: SystemTime::now(),
    };

    save_tracking_state(project_dir, &state)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::TempDir;

    #[test]
    fn test_hash_path_deterministic() {
        let path = PathBuf::from("/home/user/project");
        let hash1 = hash_path(&path);
        let hash2 = hash_path(&path);
        assert_eq!(hash1, hash2);
    }

    #[test]
    fn test_hash_path_unique() {
        let path1 = PathBuf::from("/home/user/project1");
        let path2 = PathBuf::from("/home/user/project2");
        let hash1 = hash_path(&path1);
        let hash2 = hash_path(&path2);
        assert_ne!(hash1, hash2);
    }

    #[test]
    fn test_save_and_load_tracking_state() {
        let temp_dir = TempDir::new().unwrap();
        let now = SystemTime::now();

        let state = FileTrackingState {
            razdfile_hash: Some("abc123".to_string()),
            mise_toml_hash: Some("def456".to_string()),
            format_version: Some("semantic-v1".to_string()),
            razdfile_modified: None,
            mise_toml_modified: None,
            last_sync_time: now,
        };

        save_tracking_state(temp_dir.path(), &state).unwrap();
        let loaded = load_tracking_state(temp_dir.path()).unwrap().unwrap();

        assert_eq!(loaded.razdfile_hash, Some("abc123".to_string()));
        assert_eq!(loaded.mise_toml_hash, Some("def456".to_string()));
        assert_eq!(loaded.format_version, Some("semantic-v1".to_string()));
    }

    #[test]
    fn test_load_nonexistent_tracking_state() {
        let temp_dir = TempDir::new().unwrap();
        let result = load_tracking_state(temp_dir.path()).unwrap();
        assert!(result.is_none());
    }

    #[test]
    fn test_check_file_changes_first_run() {
        let temp_dir = TempDir::new().unwrap();
        fs::write(temp_dir.path().join("Razdfile.yml"), "version: '3'\ntasks: {}").unwrap();

        let detection = check_file_changes(temp_dir.path()).unwrap();
        assert_eq!(detection, ChangeDetection::RazdfileChanged);
    }

    #[test]
    fn test_check_file_changes_no_changes() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");
        let mise_path = temp_dir.path().join("mise.toml");

        fs::write(&razdfile_path, "version: '3'\ntasks: {}").unwrap();
        fs::write(&mise_path, "[tools]\nnode = \"22\"").unwrap();

        // Establish initial state with semantic hashes
        update_tracking_state(temp_dir.path()).unwrap();

        // Check - should be no changes (same semantic content)
        let detection = check_file_changes(temp_dir.path()).unwrap();
        assert_eq!(detection, ChangeDetection::NoChanges);
    }

    #[test]
    fn test_check_file_changes_razdfile_modified() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");
        let mise_path = temp_dir.path().join("mise.toml");

        fs::write(&razdfile_path, "version: '3'\ntasks: {}").unwrap();
        fs::write(&mise_path, "[tools]\nnode = \"22\"").unwrap();

        // Establish initial state
        update_tracking_state(temp_dir.path()).unwrap();

        // Modify Razdfile semantically (not just formatting)
        fs::write(&razdfile_path, "version: '3'\ntasks:\n  test:\n    desc: New task\n    cmds: [\"echo test\"]").unwrap();

        // Check - should detect Razdfile change
        let detection = check_file_changes(temp_dir.path()).unwrap();
        assert_eq!(detection, ChangeDetection::RazdfileChanged);
    }

    #[test]
    fn test_semantic_change_detection_formatting_only() {
        let temp_dir = TempDir::new().unwrap();
        let razdfile_path = temp_dir.path().join("Razdfile.yml");

        // Write initial version
        fs::write(&razdfile_path, "version: '3'\ntasks: {}").unwrap();
        update_tracking_state(temp_dir.path()).unwrap();

        // Modify only formatting (add blank lines, change spacing)
        fs::write(&razdfile_path, "version: '3'\n\n\ntasks: {}").unwrap();

        // Should NOT detect change (semantic content unchanged)
        let detection = check_file_changes(temp_dir.path()).unwrap();
        assert_eq!(detection, ChangeDetection::NoChanges);
    }

    #[test]
    fn test_atomic_write_file() {
        let temp_dir = TempDir::new().unwrap();
        let test_file = temp_dir.path().join("test.txt");

        atomic_write_file(&test_file, "test content").unwrap();

        let content = fs::read_to_string(&test_file).unwrap();
        assert_eq!(content, "test content");

        // Verify temp file was cleaned up
        let temp_file = test_file.with_extension("tmp");
        assert!(!temp_file.exists());
    }
}
