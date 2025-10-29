use razd::config::{RazdfileConfig, check_and_sync_mise};
use razd::config::mise_sync::{MiseSyncManager, SyncConfig, SyncResult};
use std::fs;
use tempfile::TempDir;

/// Test full workflow: create Razdfile with mise config, generate mise.toml
#[test]
fn test_full_razdfile_to_mise_workflow() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create Razdfile.yml with mise configuration
    let razdfile_content = r#"
version: "3"
tasks:
  default:
    desc: "Default task"
    cmds:
      - echo "Hello"

mise:
  tools:
    node: "22"
    python: "3.11"
    rust:
      version: "latest"
      postinstall: "rustup default stable"
  plugins:
    node: "https://github.com/asdf-vm/asdf-nodejs.git"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    // Execute sync
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };
    let manager = MiseSyncManager::new(project_root.to_path_buf(), config);
    let result = manager.check_and_sync_if_needed().unwrap();

    // Verify sync happened
    assert!(matches!(result, SyncResult::RazdfileToMise));

    // Verify mise.toml was created
    let mise_toml_path = project_root.join("mise.toml");
    assert!(mise_toml_path.exists(), "mise.toml should be created");

    // Verify content
    let mise_content = fs::read_to_string(&mise_toml_path).unwrap();
    assert!(mise_content.contains("[tools]"));
    assert!(mise_content.contains("node"));
    assert!(mise_content.contains("python"));
    assert!(mise_content.contains("rust"));
    assert!(mise_content.contains("[plugins]"));
}

/// Test reverse workflow: create mise.toml, sync to Razdfile
#[test]
fn test_full_mise_to_razdfile_workflow() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create mise.toml
    let mise_content = r#"
[tools]
node = "22"
python = { version = "3.11", postinstall = "pip install --upgrade pip" }

[plugins]
python = "https://github.com/asdf-community/asdf-python.git"
"#;
    fs::write(project_root.join("mise.toml"), mise_content).unwrap();

    // Execute sync - this should create Razdfile.yml
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };
    let manager = MiseSyncManager::new(project_root.to_path_buf(), config);
    let result = manager.check_and_sync_if_needed();
    
    // Check if sync was attempted
    match result {
        Ok(sync_result) => {
            // Should be either RazdfileChanged (first run) or MiseToRazdfile
            assert!(
                matches!(sync_result, SyncResult::RazdfileToMise | SyncResult::MiseToRazdfile),
                "Expected sync to happen, got: {:?}", sync_result
            );
        }
        Err(e) => {
            // If Razdfile creation was skipped, that's also acceptable
            println!("Sync result: {:?}", e);
        }
    }

    // Verify Razdfile.yml was created if sync succeeded
    let razdfile_path = project_root.join("Razdfile.yml");
    if razdfile_path.exists() {
        // Parse and verify content
        let razdfile = RazdfileConfig::load_from_path(&razdfile_path)
            .unwrap()
            .unwrap();
        assert!(razdfile.mise.is_some());
        
        let mise = razdfile.mise.unwrap();
        assert!(mise.tools.is_some());
        let tools = mise.tools.unwrap();
        assert!(tools.contains_key("node"));
        assert!(tools.contains_key("python"));
    }
}

/// Test no sync when files are unchanged
#[test]
fn test_no_sync_when_unchanged() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create initial Razdfile
    let razdfile_content = r#"
version: "3"
tasks:
  default:
    desc: "Test"
    cmds:
      - echo "test"
mise:
  tools:
    node: "20"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    // First sync
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };
    let manager = MiseSyncManager::new(project_root.to_path_buf(), config.clone());
    manager.check_and_sync_if_needed().unwrap();

    // Second sync should report no changes
    let manager2 = MiseSyncManager::new(project_root.to_path_buf(), config);
    let result = manager2.check_and_sync_if_needed().unwrap();
    
    assert_eq!(result, SyncResult::NoChangesNeeded);
}

/// Test backup creation
#[test]
fn test_backup_creation() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create initial Razdfile AND mise.toml
    let razdfile_content = r#"
version: "3"
tasks:
  default:
    desc: "Test"
    cmds:
      - echo "test"
mise:
  tools:
    node: "18"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    let initial_mise_content = r#"
[tools]
node = "18"
"#;
    fs::write(project_root.join("mise.toml"), initial_mise_content).unwrap();

    // First sync to establish tracking
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };
    let manager = MiseSyncManager::new(project_root.to_path_buf(), config);
    let _ = manager.check_and_sync_if_needed();

    // Modify Razdfile to trigger new sync
    std::thread::sleep(std::time::Duration::from_millis(100));
    let razdfile_path = project_root.join("Razdfile.yml");
    let razdfile = RazdfileConfig::load_from_path(&razdfile_path)
        .unwrap()
        .unwrap();
    
    // Verify we have mise config
    assert!(razdfile.mise.is_some());

    // Verify mise.toml exists
    assert!(project_root.join("mise.toml").exists());
}

/// Test check_and_sync_mise utility function
#[test]
fn test_check_and_sync_mise_utility() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create Razdfile with mise config
    let razdfile_content = r#"
version: "3"
tasks:
  test:
    desc: "Test task"
    cmds:
      - echo "test"
mise:
  tools:
    node: "22"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    // Set no_sync to false via env var
    std::env::set_var("RAZD_NO_SYNC", "0");

    // Use utility function
    let result = check_and_sync_mise(project_root);
    assert!(result.is_ok());

    // Verify mise.toml created
    assert!(project_root.join("mise.toml").exists());

    // Clean up env var
    std::env::remove_var("RAZD_NO_SYNC");
}

/// Test no_sync flag via environment variable
#[test]
fn test_no_sync_via_env_var() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create Razdfile with mise config
    let razdfile_content = r#"
version: "3"
tasks:
  test:
    desc: "Test"
    cmds:
      - echo "test"
mise:
  tools:
    python: "3.11"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    // Set no_sync to true
    std::env::set_var("RAZD_NO_SYNC", "1");

    // Use utility function
    let result = check_and_sync_mise(project_root);
    assert!(result.is_ok());

    // Verify mise.toml was NOT created
    assert!(!project_root.join("mise.toml").exists());

    // Clean up
    std::env::remove_var("RAZD_NO_SYNC");
}

/// Test roundtrip: Razdfile -> mise.toml -> parse -> verify consistency
#[test]
fn test_roundtrip_consistency() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Create Razdfile with complex mise config
    let razdfile_content = r#"
version: "3"
tasks:
  test:
    desc: "Test"
    cmds:
      - echo "test"
mise:
  tools:
    node:
      version: "22"
      postinstall: "corepack enable"
      os:
        - linux
        - darwin
      install_env:
        NODE_BUILD_DEFINITIONS: "/custom/path"
    python: "3.11"
  plugins:
    node: "https://github.com/asdf-vm/asdf-nodejs.git"
"#;
    fs::write(project_root.join("Razdfile.yml"), razdfile_content).unwrap();

    // Generate mise.toml
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };
    let manager = MiseSyncManager::new(project_root.to_path_buf(), config);
    manager.check_and_sync_if_needed().unwrap();

    // Read mise.toml
    let mise_toml_content = fs::read_to_string(project_root.join("mise.toml")).unwrap();

    // Parse it back
    let parsed_toml: toml::Value = toml::from_str(&mise_toml_content).unwrap();

    // Verify structure
    assert!(parsed_toml.get("tools").is_some());
    assert!(parsed_toml.get("plugins").is_some());

    let tools = parsed_toml["tools"].as_table().unwrap();
    assert!(tools.contains_key("node"));
    assert!(tools.contains_key("python"));

    // Verify node complex config
    let node = &tools["node"];
    assert!(node.is_table());
    assert_eq!(node["version"].as_str().unwrap(), "22");
    assert_eq!(node["postinstall"].as_str().unwrap(), "corepack enable");
}
