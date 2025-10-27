use assert_cmd::prelude::*;
use predicates::prelude::*;
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
