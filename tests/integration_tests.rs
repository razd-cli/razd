use assert_cmd::prelude::*;
use predicates::prelude::*;
use std::fs;
use std::process::Command;

#[test]
fn test_help_command() {
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("--help");

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("razd"))
        .stdout(predicate::str::contains("Commands:"));
}

#[test]
fn test_up_command_help() {
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["up", "--help"]);

    cmd.assert()
        .success()
        .stdout(predicate::str::contains(
            "Clone repository and set up project, or set up local project",
        ))
        .stdout(predicate::str::contains("optional for local projects"));
}

#[test]
fn test_invalid_command() {
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("invalid");

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("error"));
}

#[test]
fn test_up_command_without_url_in_empty_directory() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("up");
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("No project configuration found"))
        .stderr(predicate::str::contains(
            "Create a Razdfile.yml manually or run 'razd up <url>'",
        ));
}

#[test]
fn test_up_command_without_url_with_taskfile() {
    use std::fs;
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create a Taskfile.yml in temp directory
    fs::write(
        temp_dir.path().join("Taskfile.yml"),
        "version: '3'\ntasks:\n  default:\n    cmds:\n      - echo 'test'",
    )
    .unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("up");
    cmd.current_dir(temp_dir.path());

    // Note: This will fail because there's no actual task setup, but it should pass validation
    // The error should NOT be about "No project detected"
    let output = cmd.output().unwrap();
    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        !stderr.contains("No project detected"),
        "Should detect Taskfile.yml as a valid project indicator"
    );
}

#[test]
fn test_task_auto_installation_behavior() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create a basic Taskfile.yml
    std::fs::write(
        temp_dir.path().join("Taskfile.yml"),
        r#"version: '3'
tasks:
  setup:
    desc: "Setup test project"
    cmds:
      - echo "Setup complete"
"#,
    )
    .unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("up");
    cmd.current_dir(temp_dir.path());

    // This test will exercise the task auto-installation code path
    // If task is not available, it should attempt to install it via mise
    // If mise is not available, it should provide a helpful error message
    let output = cmd.output().unwrap();
    let stderr = String::from_utf8_lossy(&output.stderr);
    let stdout = String::from_utf8_lossy(&output.stdout);

    // The command might succeed if task is already installed,
    // or fail with helpful error messages if tools are missing
    if !output.status.success() {
        // Should provide helpful error about missing tools, not generic failures
        assert!(
            stderr.contains("mise")
                || stderr.contains("task")
                || stderr.contains("Missing required tool"),
            "Should provide helpful error about missing tools. Stderr: {}",
            stderr
        );
    } else {
        // If successful, should show progress of task installation or execution
        assert!(
            stdout.contains("Installing")
                || stdout.contains("Setting up")
                || stdout.contains("Executing"),
            "Should show progress of task operations. Stdout: {}",
            stdout
        );
    }
}

#[test]
fn test_task_availability_check() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create a basic Taskfile.yml
    std::fs::write(
        temp_dir.path().join("Taskfile.yml"),
        r#"version: '3'
tasks:
  test:
    desc: "Test task"
    cmds:
      - echo "Test executed"
"#,
    )
    .unwrap();

    // Test the task command specifically
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["task", "test"]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let stderr = String::from_utf8_lossy(&output.stderr);

    if !output.status.success() {
        // Should attempt to install task or provide helpful error
        assert!(
            stderr.contains("Installing")
                || stderr.contains("Missing required tool")
                || stderr.contains("mise")
                || stderr.contains("task"),
            "Should handle missing task tool gracefully. Stderr: {}",
            stderr
        );
    }
}

#[test]
fn test_list_command_json_output() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    // Create a test Razdfile
    let razdfile_content = r#"
version: '3'
tasks:
  build:
    desc: Build project
    cmds:
      - echo "building"
  test:
    desc: Run tests
    cmds:
      - echo "testing"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("list").arg("--json");
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let stdout = String::from_utf8(output.stdout).unwrap();

    // Verify it's valid JSON
    let json: serde_json::Value = serde_json::from_str(&stdout).unwrap();
    assert!(json.get("tasks").is_some());
    assert!(json["tasks"].is_array());
    assert_eq!(json["tasks"].as_array().unwrap().len(), 2);
}

#[test]
fn test_list_command_list_all_flag() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    // Create a test Razdfile with internal task
    let razdfile_content = r#"
version: '3'
tasks:
  build:
    desc: Build project
    cmds:
      - echo "building"
  internal-setup:
    desc: Internal setup
    internal: true
    cmds:
      - echo "setup"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    // Test without --list-all (should not show internal)
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("list");
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let stdout = String::from_utf8(output.stdout).unwrap();
    assert!(!stdout.contains("internal-setup"));
    assert!(stdout.contains("build"));

    // Test with --list-all (should show internal)
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("list").arg("--list-all");
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let stdout = String::from_utf8(output.stdout).unwrap();
    assert!(stdout.contains("internal-setup"));
    assert!(stdout.contains("build"));
}

#[test]
fn test_list_command_json_with_list_all() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    // Create a test Razdfile with internal task
    let razdfile_content = r#"
version: '3'
tasks:
  public:
    desc: Public task
    cmds:
      - echo "public"
  internal:
    desc: Internal task
    internal: true
    cmds:
      - echo "internal"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("list").arg("--list-all").arg("--json");
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let stdout = String::from_utf8(output.stdout).unwrap();

    // Verify JSON contains both tasks with correct internal flags
    let json: serde_json::Value = serde_json::from_str(&stdout).unwrap();
    let tasks = json["tasks"].as_array().unwrap();
    assert_eq!(tasks.len(), 2);

    // Find the internal task
    let internal_task = tasks
        .iter()
        .find(|t| t["name"] == "internal")
        .expect("Should find internal task");
    assert_eq!(internal_task["internal"], true);

    let public_task = tasks
        .iter()
        .find(|t| t["name"] == "public")
        .expect("Should find public task");
    // When internal is false, it may be omitted from JSON
    assert!(public_task["internal"].is_null() || public_task["internal"] == false);
}

#[test]
fn test_list_backward_compatibility() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    // Create a test Razdfile
    let razdfile_content = r#"
version: '3'
tasks:
  build:
    desc: Build project
    cmds:
      - echo "building"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    // Test that basic list still works as before
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("list");
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("Available tasks"))
        .stdout(predicate::str::contains("build"));
}

#[test]
fn test_list_json_enhanced_output() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    let razdfile_content = r#"version: '3'
tasks:
  hello:
    desc: Test task
    cmds:
      - echo "hello"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--json"]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    assert!(output.status.success());

    let json_str = String::from_utf8(output.stdout).unwrap();
    let json: serde_json::Value = serde_json::from_str(&json_str).unwrap();

    // Verify structure
    assert!(json["tasks"].is_array());
    assert_eq!(json["tasks"].as_array().unwrap().len(), 1);

    // Verify taskfile-compatible fields
    let task = &json["tasks"][0];
    assert_eq!(task["name"], "hello");
    assert_eq!(task["task"], "hello"); // Duplicate of name
    assert_eq!(task["desc"], "Test task");
    assert_eq!(task["summary"], "");
    assert_eq!(task["aliases"], serde_json::json!([]));

    // Verify location object
    assert!(task["location"]["taskfile"].is_string());
    assert!(task["location"]["taskfile"]
        .as_str()
        .unwrap()
        .ends_with("Razdfile.yml"));
    assert!(task["location"]["line"].is_number());
    assert_eq!(task["location"]["column"], 3);

    // Verify root location
    assert!(json["location"].is_string());
    assert!(json["location"].as_str().unwrap().ends_with("Razdfile.yml"));
}

#[test]
fn test_list_json_with_internal_tasks() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    let razdfile_content = r#"version: '3'
tasks:
  public:
    desc: Public task
    cmds:
      - echo "public"
  _internal:
    desc: Internal task
    internal: true
    cmds:
      - echo "internal"
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    // Test without --list-all (should exclude internal)
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--json"]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let json_str = String::from_utf8(output.stdout).unwrap();
    let json: serde_json::Value = serde_json::from_str(&json_str).unwrap();

    assert_eq!(json["tasks"].as_array().unwrap().len(), 1);
    assert_eq!(json["tasks"][0]["name"], "public");

    // Test with --list-all (should include internal)
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--list-all", "--json"]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let json_str = String::from_utf8(output.stdout).unwrap();
    let json: serde_json::Value = serde_json::from_str(&json_str).unwrap();

    assert_eq!(json["tasks"].as_array().unwrap().len(), 2);

    // Find internal task
    let internal_task = json["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|t| t["name"] == "_internal")
        .unwrap();

    assert_eq!(internal_task["internal"], true);
}

#[test]
fn test_list_json_empty_tasks() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let razdfile_path = temp_dir.path().join("Razdfile.yml");

    let razdfile_content = r#"version: '3'
tasks: {}
"#;
    fs::write(&razdfile_path, razdfile_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--json"]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();
    let json_str = String::from_utf8(output.stdout).unwrap();
    let json: serde_json::Value = serde_json::from_str(&json_str).unwrap();

    assert!(json["tasks"].is_array());
    assert_eq!(json["tasks"].as_array().unwrap().len(), 0);
    assert!(json["location"].is_string());
}

#[test]
fn test_list_with_custom_taskfile_flag() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let custom_file = temp_dir.path().join("custom.yml");

    let custom_content = r#"version: '3'
tasks:
  build:
    desc: Build the project
    cmds:
      - echo "Building..."
  test:
    desc: Run tests
    cmds:
      - echo "Testing..."
"#;
    fs::write(&custom_file, custom_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--taskfile", custom_file.to_str().unwrap()]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("build"))
        .stdout(predicate::str::contains("Build the project"))
        .stdout(predicate::str::contains("test"))
        .stdout(predicate::str::contains("Run tests"));
}

#[test]
fn test_list_with_custom_razdfile_flag() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let custom_file = temp_dir.path().join("my-config.yml");

    let custom_content = r#"version: '3'
tasks:
  deploy:
    desc: Deploy application
    cmds:
      - echo "Deploying..."
"#;
    fs::write(&custom_file, custom_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--razdfile", custom_file.to_str().unwrap()]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("deploy"))
        .stdout(predicate::str::contains("Deploy application"));
}

#[test]
fn test_short_form_taskfile_flag() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let custom_file = temp_dir.path().join("tasks.yml");

    let custom_content = r#"version: '3'
tasks:
  lint:
    desc: Lint code
    cmds:
      - echo "Linting..."
"#;
    fs::write(&custom_file, custom_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "-t", custom_file.to_str().unwrap()]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("lint"))
        .stdout(predicate::str::contains("Lint code"));
}

#[test]
fn test_custom_file_not_found_error() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let nonexistent = temp_dir.path().join("nonexistent.yml");

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["list", "--taskfile", nonexistent.to_str().unwrap()]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("not found"));
}

#[test]
fn test_razdfile_priority_over_taskfile() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let file1 = temp_dir.path().join("file1.yml");
    let file2 = temp_dir.path().join("file2.yml");

    let content1 = r#"version: '3'
tasks:
  task1:
    desc: From file1
    cmds:
      - echo "File 1"
"#;
    let content2 = r#"version: '3'
tasks:
  task2:
    desc: From file2
    cmds:
      - echo "File 2"
"#;
    fs::write(&file1, content1).unwrap();
    fs::write(&file2, content2).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args([
        "list",
        "--taskfile",
        file1.to_str().unwrap(),
        "--razdfile",
        file2.to_str().unwrap(),
    ]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("task2")) // Should use file2
        .stdout(predicate::str::contains("From file2"))
        .stdout(predicate::str::contains("task1").not()); // Should NOT have task1 from file1
}

#[test]
fn test_list_json_with_custom_path() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();
    let custom_file = temp_dir.path().join("custom.yml");

    let custom_content = r#"version: '3'
tasks:
  build:
    desc: Build the project
    cmds:
      - echo "Building..."
"#;
    fs::write(&custom_file, custom_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args([
        "list",
        "--taskfile",
        custom_file.to_str().unwrap(),
        "--json",
    ]);
    cmd.current_dir(temp_dir.path());

    let output = cmd.output().unwrap();

    // Check if command succeeded
    if !output.status.success() {
        eprintln!("Command failed!");
        eprintln!("stdout: {}", String::from_utf8_lossy(&output.stdout));
        eprintln!("stderr: {}", String::from_utf8_lossy(&output.stderr));
        panic!("Command failed with status: {}", output.status);
    }

    let json_str = String::from_utf8(output.stdout).unwrap();
    if json_str.is_empty() {
        panic!("Empty JSON output");
    }

    let json: serde_json::Value = serde_json::from_str(&json_str).unwrap();

    assert!(json["tasks"].is_array());
    let tasks = json["tasks"].as_array().unwrap();
    assert_eq!(tasks.len(), 1);
    assert_eq!(tasks[0]["name"], "build");
    assert_eq!(tasks[0]["desc"], "Build the project");
}

#[test]
fn test_yes_flag_in_help() {
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("--help");

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("-y, --yes"))
        .stdout(predicate::str::contains(
            "Automatically answer \"yes\" to all prompts",
        ));
}

#[test]
fn test_short_yes_flag_works() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    // Create a basic Razdfile.yml for list command
    let razdfile_content = r#"version: '3'
tasks:
  test:
    desc: Test task
    cmds:
      - echo "test"
"#;
    fs::write(temp_dir.path().join("Razdfile.yml"), razdfile_content).unwrap();

    // Test with short form -y
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["-y", "list"]);
    cmd.current_dir(temp_dir.path());

    cmd.assert().success();

    // Test with long form --yes
    let mut cmd2 = Command::cargo_bin("razd").unwrap();
    cmd2.args(["--yes", "list"]);
    cmd2.current_dir(temp_dir.path());

    cmd2.assert().success();
}

#[test]
fn test_yes_flag_with_list_command() {
    use tempfile::TempDir;

    let temp_dir = TempDir::new().unwrap();

    let razdfile_content = r#"version: '3'
tasks:
  build:
    desc: Build project
    cmds:
      - echo "Building..."
"#;
    fs::write(temp_dir.path().join("Razdfile.yml"), razdfile_content).unwrap();

    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.args(["--yes", "list"]);
    cmd.current_dir(temp_dir.path());

    cmd.assert()
        .success()
        .stdout(predicate::str::contains("build"))
        .stdout(predicate::str::contains("Build project"));
}
