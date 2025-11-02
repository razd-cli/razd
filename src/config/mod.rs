pub mod canonical;
pub mod defaults;
pub mod detection;
pub mod file_tracker;
pub mod mise_generator;
pub mod mise_sync;
pub mod mise_validator;
pub mod razdfile;

pub use razdfile::*;

use crate::core::Result;
use mise_sync::{MiseSyncManager, SyncConfig};
use std::env;
use std::path::Path;

/// Check and perform mise configuration sync if needed
/// Respects the RAZD_NO_SYNC environment variable
pub fn check_and_sync_mise(project_dir: &Path) -> Result<()> {
    // Check if sync is disabled
    let no_sync = env::var("RAZD_NO_SYNC").unwrap_or_default() == "1";

    let config = SyncConfig {
        no_sync,
        auto_approve: false, // Always prompt user for manual operations
        create_backups: true,
    };

    let manager = MiseSyncManager::new(project_dir.to_path_buf(), config);
    let result = manager.check_and_sync_if_needed()?;

    // Only print message if sync actually happened
    use mise_sync::SyncResult;
    match result {
        SyncResult::RazdfileToMise | SyncResult::MiseToRazdfile => {
            // Message already printed by sync manager
        }
        _ => {
            // No sync needed or skipped
        }
    }

    Ok(())
}
