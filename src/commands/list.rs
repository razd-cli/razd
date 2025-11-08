use crate::config::razdfile::RazdfileConfig;
use crate::core::{output, Result};
use colored::*;
use serde::Serialize;
use std::path::PathBuf;

/// Helper function to skip serializing false boolean values for cleaner JSON
fn is_false(b: &bool) -> bool {
    !b
}

#[derive(Serialize)]
struct TaskLocation {
    taskfile: String,
    line: usize,
    column: usize,
}

#[derive(Serialize)]
struct TaskListOutput {
    tasks: Vec<TaskInfo>,
    #[serde(skip_serializing_if = "Option::is_none")]
    location: Option<String>,
}

#[derive(Serialize)]
struct TaskInfo {
    name: String,
    task: String,
    desc: String,
    summary: String,
    aliases: Vec<String>,
    location: TaskLocation,
    #[serde(skip_serializing_if = "is_false")]
    internal: bool,
}

/// Find the absolute path to Razdfile.yml or custom config file
fn find_razdfile_path(custom_path: Option<&PathBuf>) -> Result<std::path::PathBuf> {
    use std::env;

    if let Some(path) = custom_path {
        // Use custom path and make it absolute if relative
        let absolute = if path.is_absolute() {
            path.clone()
        } else {
            env::current_dir()
                .map_err(|e| {
                    crate::core::error::RazdError::config(format!(
                        "Failed to get current directory: {}",
                        e
                    ))
                })?
                .join(path)
        };

        if !absolute.exists() {
            return Err(crate::core::error::RazdError::config(format!(
                "Specified configuration file not found: {}",
                absolute.display()
            )));
        }

        return Ok(absolute);
    }

    // Default behavior: look for Razdfile.yml in current directory
    let current_dir = env::current_dir().map_err(|e| {
        crate::core::error::RazdError::config(format!("Failed to get current directory: {}", e))
    })?;

    let razdfile_path = current_dir.join("Razdfile.yml");

    if !razdfile_path.exists() {
        return Err(crate::core::error::RazdError::config(
            "Razdfile.yml not found in current directory".to_string(),
        ));
    }

    Ok(razdfile_path)
}

/// Estimate the line number where a task is defined in Razdfile.yml
fn estimate_task_line(task_name: &str, razdfile_path: &std::path::Path) -> Result<usize> {
    use std::fs;

    let content = fs::read_to_string(razdfile_path).map_err(|e| {
        crate::core::error::RazdError::config(format!("Failed to read file: {}", e))
    })?;

    // Find line containing "  task_name:"
    for (idx, line) in content.lines().enumerate() {
        let trimmed = line.trim_start();
        if trimmed.starts_with(&format!("{}:", task_name)) && !trimmed.starts_with('#') {
            return Ok(idx + 1); // 1-indexed
        }
    }

    // Fallback: return line 1 if not found
    Ok(1)
}

pub async fn execute(list_all: bool, json: bool, custom_path: Option<PathBuf>) -> Result<()> {
    // Load Razdfile.yml
    let razdfile = match RazdfileConfig::load_with_path(custom_path.clone())? {
        Some(config) => config,
        None => {
            if json {
                // Output error as JSON
                let error_json = serde_json::json!({
                    "error": "Razdfile.yml not found in current directory"
                });
                println!(
                    "{}",
                    serde_json::to_string_pretty(&error_json)
                        .unwrap_or_else(|_| r#"{"error":"Razdfile.yml not found"}"#.to_string())
                );
            } else {
                output::error("Razdfile.yml not found in current directory");
            }
            return Err(crate::core::error::RazdError::config(
                "Razdfile.yml not found".to_string(),
            ));
        }
    };

    // Extract tasks based on list_all flag
    let tasks: Vec<(String, String, bool)> = razdfile
        .tasks
        .iter()
        .filter(|(_, config)| list_all || !config.internal)
        .map(|(name, config)| {
            let desc = config.desc.clone().unwrap_or_default();
            (name.clone(), desc, config.internal)
        })
        .collect();

    // Check if there are any tasks
    if tasks.is_empty() {
        if json {
            let razdfile_path = find_razdfile_path(custom_path.as_ref()).ok();
            let location = razdfile_path.map(|p| p.to_string_lossy().to_string());

            let output = TaskListOutput {
                tasks: vec![],
                location,
            };
            println!(
                "{}",
                serde_json::to_string_pretty(&output)
                    .unwrap_or_else(|_| r#"{"tasks":[]}"#.to_string())
            );
        } else {
            println!("No tasks found in Razdfile.yml");
        }
        return Ok(());
    }

    // Tasks are already in the order they appear in Razdfile.yml (IndexMap preserves order)

    if json {
        // Get absolute path to Razdfile.yml for location metadata
        let razdfile_path = find_razdfile_path(custom_path.as_ref())?;
        let razdfile_path_str = razdfile_path.to_string_lossy().to_string();

        // Output as JSON with enhanced taskfile-compatible format
        let task_infos: Vec<TaskInfo> = tasks
            .iter()
            .map(|(name, desc, internal)| {
                let line = estimate_task_line(name, &razdfile_path).unwrap_or(1);

                TaskInfo {
                    name: name.clone(),
                    task: name.clone(), // Duplicate of name per taskfile convention
                    desc: desc.clone(),
                    summary: String::new(), // Placeholder for future feature
                    aliases: Vec::new(),    // Placeholder for future feature
                    location: TaskLocation {
                        taskfile: razdfile_path_str.clone(),
                        line,
                        column: 3, // Tasks are typically indented 2 spaces (column 3)
                    },
                    internal: *internal,
                }
            })
            .collect();

        let output = TaskListOutput {
            tasks: task_infos,
            location: Some(razdfile_path_str),
        };
        println!(
            "{}",
            serde_json::to_string_pretty(&output).unwrap_or_else(|_| r#"{"tasks":[]}"#.to_string())
        );
    } else {
        // Output as text
        // Calculate maximum task name length for proper alignment
        let max_name_len = tasks
            .iter()
            .map(|(name, _, _)| name.len())
            .max()
            .unwrap_or(0);
        let column_width = max_name_len + 1; // +1 for the colon

        // Display header
        println!("{}", "task: Available tasks for this project:".bold());

        // Display each task with proper formatting
        for (name, desc, _) in tasks {
            let formatted_name = format!("{}:", name);
            if desc.is_empty() {
                println!("* {}", formatted_name.cyan());
            } else {
                println!(
                    "* {:<width$} {}",
                    formatted_name.cyan(),
                    desc,
                    width = column_width
                );
            }
        }
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use crate::config::razdfile::{RazdfileConfig, TaskConfig};
    use indexmap::IndexMap;

    #[tokio::test]
    async fn test_list_extracts_non_internal_tasks() {
        let mut tasks = IndexMap::new();
        tasks.insert(
            "build".to_string(),
            TaskConfig {
                desc: Some("Build project".to_string()),
                cmds: vec![],
                internal: false,
                deps: None,
                env: None,
                vars: None,
                silent: None,
                platforms: None,
            },
        );
        tasks.insert(
            "internal-task".to_string(),
            TaskConfig {
                desc: Some("Internal task".to_string()),
                cmds: vec![],
                internal: true,
                deps: None,
                env: None,
                vars: None,
                silent: None,
                platforms: None,
            },
        );

        let razdfile = RazdfileConfig {
            version: "1.0.0".to_string(),
            mise: None,
            env: None,
            vars: None,
            tasks,
        };

        let visible_tasks: Vec<_> = razdfile
            .tasks
            .iter()
            .filter(|(_, config)| !config.internal)
            .map(|(name, _)| name.clone())
            .collect();

        assert_eq!(visible_tasks.len(), 1);
        assert_eq!(visible_tasks[0], "build");
    }

    #[test]
    fn test_task_sorting() {
        let task_names = vec!["zebra", "apple", "middle"];
        let mut sorted = task_names.clone();
        sorted.sort();

        assert_eq!(sorted, vec!["apple", "middle", "zebra"]);
    }

    #[test]
    fn test_list_all_includes_internal_tasks() {
        let mut tasks = IndexMap::new();
        tasks.insert(
            "public-task".to_string(),
            TaskConfig {
                desc: Some("Public task".to_string()),
                cmds: vec![],
                internal: false,
                deps: None,
                env: None,
                vars: None,
                silent: None,
                platforms: None,
            },
        );
        tasks.insert(
            "internal-task".to_string(),
            TaskConfig {
                desc: Some("Internal task".to_string()),
                cmds: vec![],
                internal: true,
                deps: None,
                env: None,
                vars: None,
                silent: None,
                platforms: None,
            },
        );

        // Test with list_all = false (default behavior)
        let list_all = false;
        let filtered_tasks: Vec<_> = tasks
            .iter()
            .filter(|(_, config)| list_all || !config.internal)
            .map(|(name, _)| name.clone())
            .collect();
        assert_eq!(filtered_tasks.len(), 1);
        assert_eq!(filtered_tasks[0], "public-task");

        // Test with list_all = true
        let list_all = true;
        let all_tasks: Vec<_> = tasks
            .iter()
            .filter(|(_, config)| list_all || !config.internal)
            .map(|(name, _)| name.clone())
            .collect();
        assert_eq!(all_tasks.len(), 2);
    }

    #[test]
    fn test_json_serialization() {
        use super::{TaskInfo, TaskListOutput, TaskLocation};

        let tasks = vec![
            TaskInfo {
                name: "build".to_string(),
                task: "build".to_string(),
                desc: "Build project".to_string(),
                summary: String::new(),
                aliases: Vec::new(),
                location: TaskLocation {
                    taskfile: "Razdfile.yml".to_string(),
                    line: 1,
                    column: 3,
                },
                internal: false,
            },
            TaskInfo {
                name: "test".to_string(),
                task: "test".to_string(),
                desc: "".to_string(),
                summary: String::new(),
                aliases: Vec::new(),
                location: TaskLocation {
                    taskfile: "Razdfile.yml".to_string(),
                    line: 5,
                    column: 3,
                },
                internal: false,
            },
        ];

        let output = TaskListOutput {
            tasks,
            location: Some("Razdfile.yml".to_string()),
        };
        let json = serde_json::to_string_pretty(&output).unwrap();

        // Verify it's valid JSON
        let parsed: serde_json::Value = serde_json::from_str(&json).unwrap();
        assert!(parsed.get("tasks").is_some());
        assert_eq!(parsed["tasks"].as_array().unwrap().len(), 2);
        assert_eq!(parsed["tasks"][0]["name"], "build");
        assert_eq!(parsed["tasks"][1]["desc"], "");
    }

    #[test]
    fn test_json_with_internal_task() {
        use super::{TaskInfo, TaskListOutput, TaskLocation};

        let tasks = vec![
            TaskInfo {
                name: "public".to_string(),
                task: "public".to_string(),
                desc: "Public task".to_string(),
                summary: String::new(),
                aliases: Vec::new(),
                location: TaskLocation {
                    taskfile: "Razdfile.yml".to_string(),
                    line: 1,
                    column: 3,
                },
                internal: false,
            },
            TaskInfo {
                name: "internal".to_string(),
                task: "internal".to_string(),
                desc: "Internal task".to_string(),
                summary: String::new(),
                aliases: Vec::new(),
                location: TaskLocation {
                    taskfile: "Razdfile.yml".to_string(),
                    line: 5,
                    column: 3,
                },
                internal: true,
            },
        ];

        let output = TaskListOutput {
            tasks,
            location: Some("Razdfile.yml".to_string()),
        };
        let json = serde_json::to_string_pretty(&output).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&json).unwrap();

        // When internal is false, it's omitted from JSON (for cleaner output)
        assert!(
            parsed["tasks"][0]["internal"].is_null() || parsed["tasks"][0]["internal"] == false
        );
        assert_eq!(parsed["tasks"][1]["internal"], true);
    }

    #[test]
    fn test_json_structure_with_enhanced_fields() {
        use super::{TaskInfo, TaskListOutput, TaskLocation};

        let tasks = vec![TaskInfo {
            name: "build".to_string(),
            task: "build".to_string(),
            desc: "Build project".to_string(),
            summary: String::new(),
            aliases: Vec::new(),
            location: TaskLocation {
                taskfile: "/path/to/Razdfile.yml".to_string(),
                line: 5,
                column: 3,
            },
            internal: false,
        }];

        let output = TaskListOutput {
            tasks,
            location: Some("/path/to/Razdfile.yml".to_string()),
        };

        let json = serde_json::to_string_pretty(&output).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&json).unwrap();

        // Verify all taskfile-compatible fields are present
        assert_eq!(parsed["tasks"][0]["name"], "build");
        assert_eq!(parsed["tasks"][0]["task"], "build"); // Duplicate of name
        assert_eq!(parsed["tasks"][0]["desc"], "Build project");
        assert_eq!(parsed["tasks"][0]["summary"], "");
        assert_eq!(parsed["tasks"][0]["aliases"], serde_json::json!([]));

        // Verify location object
        assert_eq!(
            parsed["tasks"][0]["location"]["taskfile"],
            "/path/to/Razdfile.yml"
        );
        assert_eq!(parsed["tasks"][0]["location"]["line"], 5);
        assert_eq!(parsed["tasks"][0]["location"]["column"], 3);

        // Verify root location
        assert_eq!(parsed["location"], "/path/to/Razdfile.yml");
    }

    #[test]
    fn test_task_field_equals_name() {
        use super::{TaskInfo, TaskLocation};

        let task = TaskInfo {
            name: "test-task".to_string(),
            task: "test-task".to_string(),
            desc: "".to_string(),
            summary: String::new(),
            aliases: Vec::new(),
            location: TaskLocation {
                taskfile: "Razdfile.yml".to_string(),
                line: 1,
                column: 3,
            },
            internal: false,
        };

        assert_eq!(task.name, task.task);
    }

    #[test]
    fn test_location_has_all_required_fields() {
        use super::TaskLocation;

        let location = TaskLocation {
            taskfile: "/absolute/path/Razdfile.yml".to_string(),
            line: 10,
            column: 3,
        };

        assert!(!location.taskfile.is_empty());
        assert!(location.line > 0);
        assert!(location.column > 0);
    }

    #[test]
    fn test_internal_field_skips_false() {
        use super::{TaskInfo, TaskListOutput, TaskLocation};

        let tasks = vec![TaskInfo {
            name: "public".to_string(),
            task: "public".to_string(),
            desc: "".to_string(),
            summary: String::new(),
            aliases: Vec::new(),
            location: TaskLocation {
                taskfile: "Razdfile.yml".to_string(),
                line: 1,
                column: 3,
            },
            internal: false, // Should be omitted from JSON
        }];

        let output = TaskListOutput {
            tasks,
            location: Some("Razdfile.yml".to_string()),
        };

        let json = serde_json::to_string(&output).unwrap();

        // When internal is false, it should not appear in JSON
        assert!(!json.contains("\"internal\""));
    }
}
