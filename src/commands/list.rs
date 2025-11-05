use crate::config::razdfile::RazdfileConfig;
use crate::core::{output, Result};
use colored::*;

pub async fn execute() -> Result<()> {
    // Load Razdfile.yml
    let razdfile = match RazdfileConfig::load()? {
        Some(config) => config,
        None => {
            output::error("Razdfile.yml not found in current directory");
            return Err(crate::core::error::RazdError::config(
                "Razdfile.yml not found".to_string(),
            ));
        }
    };

    // Extract non-internal tasks
    let mut tasks: Vec<(String, String)> = razdfile
        .tasks
        .iter()
        .filter(|(_, config)| !config.internal)
        .map(|(name, config)| {
            let desc = config.desc.clone().unwrap_or_default();
            (name.clone(), desc)
        })
        .collect();

    // Check if there are any tasks
    if tasks.is_empty() {
        println!("No tasks found in Razdfile.yml");
        return Ok(());
    }

    // Sort alphabetically by task name
    tasks.sort_by(|a, b| a.0.cmp(&b.0));

    // Calculate maximum task name length for proper alignment
    let max_name_len = tasks.iter().map(|(name, _)| name.len()).max().unwrap_or(0);
    let column_width = max_name_len + 1; // +1 for the colon

    // Display header
    println!("{}", "task: Available tasks for this project:".bold());

    // Display each task with proper formatting
    for (name, desc) in tasks {
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
}
