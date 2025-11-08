use crate::config::razdfile::RazdfileConfig;
use crate::core::{output, Result};
use colored::*;
use serde::Serialize;

#[derive(Serialize)]
struct TaskListOutput {
    tasks: Vec<TaskInfo>,
}

#[derive(Serialize)]
struct TaskInfo {
    name: String,
    desc: String,
    internal: bool,
}

pub async fn execute(list_all: bool, json: bool) -> Result<()> {
    // Load Razdfile.yml
    let razdfile = match RazdfileConfig::load()? {
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
    let mut tasks: Vec<(String, String, bool)> = razdfile
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
            let output = TaskListOutput { tasks: vec![] };
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

    // Sort alphabetically by task name
    tasks.sort_by(|a, b| a.0.cmp(&b.0));

    if json {
        // Output as JSON
        let task_infos: Vec<TaskInfo> = tasks
            .iter()
            .map(|(name, desc, internal)| TaskInfo {
                name: name.clone(),
                desc: desc.clone(),
                internal: *internal,
            })
            .collect();

        let output = TaskListOutput { tasks: task_infos };
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
        use super::{TaskInfo, TaskListOutput};

        let tasks = vec![
            TaskInfo {
                name: "build".to_string(),
                desc: "Build project".to_string(),
                internal: false,
            },
            TaskInfo {
                name: "test".to_string(),
                desc: "".to_string(),
                internal: false,
            },
        ];

        let output = TaskListOutput { tasks };
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
        use super::{TaskInfo, TaskListOutput};

        let tasks = vec![
            TaskInfo {
                name: "public".to_string(),
                desc: "Public task".to_string(),
                internal: false,
            },
            TaskInfo {
                name: "internal".to_string(),
                desc: "Internal task".to_string(),
                internal: true,
            },
        ];

        let output = TaskListOutput { tasks };
        let json = serde_json::to_string_pretty(&output).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&json).unwrap();

        assert_eq!(parsed["tasks"][0]["internal"], false);
        assert_eq!(parsed["tasks"][1]["internal"], true);
    }
}
