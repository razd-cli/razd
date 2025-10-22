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
            "Clone repository and set up project",
        ))
        .stdout(predicate::str::contains("Git repository URL to clone"));
}

#[test]
fn test_invalid_command() {
    let mut cmd = Command::cargo_bin("razd").unwrap();
    cmd.arg("invalid");

    cmd.assert()
        .failure()
        .stderr(predicate::str::contains("error"));
}
