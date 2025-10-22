use crate::core::{output, Result};
use crate::integrations::taskfile;
use std::env;

/// Execute the `razd task` command: execute tasks from Taskfile.yml
pub async fn execute(task_name: Option<&str>, args: &[String]) -> Result<()> {
    let current_dir = env::current_dir()?;

    if task_name.is_none() && args.is_empty() {
        output::info("Running default development task...");
    }

    taskfile::execute_task(task_name, args, &current_dir).await?;

    Ok(())
}
