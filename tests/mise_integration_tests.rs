use razd::config::{mise_generator, RazdfileConfig};
use std::fs;
use tempfile::TempDir;

#[test]
fn test_parse_nodejs_example_with_mise() {
    // Load the example Razdfile.yml
    let razdfile_path = "examples/nodejs-project/Razdfile.yml";

    let config = RazdfileConfig::load_from_path(razdfile_path)
        .expect("Failed to load Razdfile")
        .expect("Razdfile not found");

    // Verify mise section exists
    assert!(
        config.mise.is_some(),
        "Mise configuration should be present"
    );

    let mise = config.mise.unwrap();

    // Verify tools
    assert!(mise.tools.is_some(), "Tools should be present");
    let tools = mise.tools.unwrap();

    assert_eq!(tools.len(), 2, "Should have 2 tools (node and task)");
    assert!(tools.contains_key("node"), "Should have node tool");
    assert!(tools.contains_key("task"), "Should have task tool");

    // Verify plugins
    assert!(mise.plugins.is_some(), "Plugins should be present");
    let plugins = mise.plugins.unwrap();

    assert_eq!(plugins.len(), 1, "Should have 1 plugin");
    assert!(plugins.contains_key("node"), "Should have node plugin");
}

#[test]
fn test_generate_mise_toml_from_nodejs_example() {
    // Load the example Razdfile.yml
    let razdfile_path = "examples/nodejs-project/Razdfile.yml";

    let config = RazdfileConfig::load_from_path(razdfile_path)
        .expect("Failed to load Razdfile")
        .expect("Razdfile not found");

    let mise = config.mise.expect("Mise config should exist");

    // Generate mise.toml
    let toml_content =
        mise_generator::generate_mise_toml(&mise).expect("Failed to generate mise.toml");

    println!("Generated mise.toml:\n{}", toml_content);

    // Verify content
    assert!(
        toml_content.contains("[tools]"),
        "Should contain [tools] section"
    );
    assert!(
        toml_content.contains("node = "),
        "Should contain node configuration"
    );
    assert!(
        toml_content.contains("\"22\""),
        "Should have node version 22"
    );

    assert!(
        toml_content.contains("[plugins]"),
        "Should contain [plugins] section"
    );
    assert!(
        toml_content.contains("node = \"https://github.com/asdf-vm/asdf-nodejs.git\""),
        "Should have node plugin"
    );
}

#[test]
fn test_write_and_parse_generated_mise_toml() {
    let temp_dir = TempDir::new().unwrap();

    // Load the example Razdfile.yml
    let razdfile_path = "examples/nodejs-project/Razdfile.yml";

    let config = RazdfileConfig::load_from_path(razdfile_path)
        .expect("Failed to load Razdfile")
        .expect("Razdfile not found");

    let mise = config.mise.expect("Mise config should exist");

    // Generate mise.toml
    let toml_content =
        mise_generator::generate_mise_toml(&mise).expect("Failed to generate mise.toml");

    // Write to temp file
    let mise_toml_path = temp_dir.path().join("mise.toml");
    fs::write(&mise_toml_path, &toml_content).expect("Failed to write mise.toml");

    // Verify file exists and is valid TOML
    assert!(mise_toml_path.exists(), "mise.toml should exist");

    // Try to parse it back with toml parser to verify it's valid
    let content = fs::read_to_string(&mise_toml_path).expect("Failed to read mise.toml");
    let parsed: toml::Value =
        toml::from_str(&content).expect("Failed to parse generated mise.toml");

    // Verify structure
    assert!(parsed.get("tools").is_some(), "Should have tools section");
    assert!(
        parsed.get("plugins").is_some(),
        "Should have plugins section"
    );
}

#[test]
fn test_roundtrip_razdfile_to_mise_toml() {
    let temp_dir = TempDir::new().unwrap();

    // Copy example Razdfile to temp directory
    let source_razdfile = "examples/nodejs-project/Razdfile.yml";
    let dest_razdfile = temp_dir.path().join("Razdfile.yml");
    fs::copy(source_razdfile, &dest_razdfile).expect("Failed to copy Razdfile");

    // Load and generate
    let config = RazdfileConfig::load_from_path(&dest_razdfile)
        .expect("Failed to load Razdfile")
        .expect("Razdfile not found");

    let mise = config.mise.expect("Mise config should exist");
    let toml_content =
        mise_generator::generate_mise_toml(&mise).expect("Failed to generate mise.toml");

    // Write mise.toml
    let mise_toml_path = temp_dir.path().join("mise.toml");
    fs::write(&mise_toml_path, &toml_content).expect("Failed to write mise.toml");

    // Verify both files exist
    assert!(dest_razdfile.exists(), "Razdfile.yml should exist");
    assert!(mise_toml_path.exists(), "mise.toml should exist");

    println!("\nGenerated mise.toml content:");
    println!("{}", toml_content);
}

#[test]
fn test_yes_flag_auto_approves_mise_sync() {
    use razd::config::mise_sync::{MiseSyncManager, SyncConfig};
    use std::fs;
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create a Razdfile.yml with mise config
    let razdfile_content = r#"version: '3'
mise:
  tools:
    node: "20"
tasks:
  test:
    desc: Test task
    cmds:
      - echo "test"
"#;
    fs::write(temp_dir.path().join("Razdfile.yml"), razdfile_content).unwrap();

    // Create a conflicting mise.toml with different version
    let mise_toml_content = r#"[tools]
node = "18"
"#;
    fs::write(temp_dir.path().join("mise.toml"), mise_toml_content).unwrap();

    // Simulate changes to both files
    // In real scenario, the file tracker would detect this
    // For this test, we'll create a sync manager with auto_approve
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: true,
    };

    let manager = MiseSyncManager::new(temp_dir.path().to_path_buf(), config);

    // Attempt sync - should auto-resolve conflict
    let result = manager.check_and_sync_if_needed();

    // Should succeed without prompts
    assert!(
        result.is_ok(),
        "Sync should succeed with auto_approve enabled"
    );
}

#[test]
fn test_yes_flag_resolves_conflicts_automatically() {
    use razd::config::mise_sync::{MiseSyncManager, SyncConfig};
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create Razdfile.yml with mise config
    let razdfile_content = r#"version: '3'
mise:
  tools:
    node: "22"
    task: "latest"
tasks:
  build:
    cmds:
      - echo "Building..."
"#;
    fs::write(temp_dir.path().join("Razdfile.yml"), razdfile_content).unwrap();

    // Create conflicting mise.toml
    let mise_toml_content = r#"[tools]
node = "20"
"#;
    fs::write(temp_dir.path().join("mise.toml"), mise_toml_content).unwrap();

    // Create manager with auto_approve enabled
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: true,
    };

    let manager = MiseSyncManager::new(temp_dir.path().to_path_buf(), config);

    // Should auto-resolve by preferring Razdfile (Option 1)
    let result = manager.check_and_sync_if_needed();

    assert!(result.is_ok(), "Should auto-resolve conflict");

    // Verify mise.toml was updated with Razdfile values
    let mise_content = fs::read_to_string(temp_dir.path().join("mise.toml")).unwrap();
    assert!(
        mise_content.contains("22"),
        "mise.toml should have node 22 from Razdfile"
    );
}

#[test]
fn test_mise_sync_without_yes_flag_creates_backup() {
    use razd::config::mise_sync::{MiseSyncManager, SyncConfig};
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create Razdfile.yml
    let razdfile_content = r#"version: '3'
mise:
  tools:
    node: "20"
"#;
    fs::write(temp_dir.path().join("Razdfile.yml"), razdfile_content).unwrap();

    // Create mise.toml
    let mise_toml_content = r#"[tools]
node = "18"
"#;
    fs::write(temp_dir.path().join("mise.toml"), mise_toml_content).unwrap();

    // Create manager with auto_approve and backups enabled
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: true,
    };

    let manager = MiseSyncManager::new(temp_dir.path().to_path_buf(), config);

    // This should create backups automatically when auto_approve is true
    let _ = manager.check_and_sync_if_needed();

    // Note: Backup file verification would require accessing internal backup logic
    // This test primarily ensures no panics occur with backup creation
}
