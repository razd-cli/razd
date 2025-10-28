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
    assert!(config.mise.is_some(), "Mise configuration should be present");

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
    let toml_content = mise_generator::generate_mise_toml(&mise)
        .expect("Failed to generate mise.toml");

    println!("Generated mise.toml:\n{}", toml_content);

    // Verify content
    assert!(toml_content.contains("[tools]"), "Should contain [tools] section");
    assert!(toml_content.contains("node = "), "Should contain node configuration");
    assert!(toml_content.contains("\"22\""), "Should have node version 22");
    
    assert!(toml_content.contains("[plugins]"), "Should contain [plugins] section");
    assert!(toml_content.contains("node = \"https://github.com/asdf-vm/asdf-nodejs.git\""), "Should have node plugin");
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
    let toml_content = mise_generator::generate_mise_toml(&mise)
        .expect("Failed to generate mise.toml");

    // Write to temp file
    let mise_toml_path = temp_dir.path().join("mise.toml");
    fs::write(&mise_toml_path, &toml_content).expect("Failed to write mise.toml");

    // Verify file exists and is valid TOML
    assert!(mise_toml_path.exists(), "mise.toml should exist");

    // Try to parse it back with toml parser to verify it's valid
    let content = fs::read_to_string(&mise_toml_path).expect("Failed to read mise.toml");
    let parsed: toml::Value = toml::from_str(&content).expect("Failed to parse generated mise.toml");

    // Verify structure
    assert!(parsed.get("tools").is_some(), "Should have tools section");
    assert!(parsed.get("plugins").is_some(), "Should have plugins section");
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
    let toml_content = mise_generator::generate_mise_toml(&mise)
        .expect("Failed to generate mise.toml");

    // Write mise.toml
    let mise_toml_path = temp_dir.path().join("mise.toml");
    fs::write(&mise_toml_path, &toml_content).expect("Failed to write mise.toml");

    // Verify both files exist
    assert!(dest_razdfile.exists(), "Razdfile.yml should exist");
    assert!(mise_toml_path.exists(), "mise.toml should exist");

    println!("\nGenerated mise.toml content:");
    println!("{}", toml_content);
}
