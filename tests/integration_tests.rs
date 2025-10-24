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
        .stderr(predicate::str::contains("No project detected"))
        .stderr(predicate::str::contains(
            "Razdfile.yml, Taskfile.yml, or mise.toml",
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
