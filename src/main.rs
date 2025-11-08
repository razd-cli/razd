mod commands;
mod config;
mod core;
mod defaults;
mod integrations;

use clap::{Parser, Subcommand};
use colored::*;

#[derive(Parser)]
#[command(
    name = "razd",
    version,
    about = "Streamlined project setup with git, mise, and taskfile integration",
    long_about = "razd (раздуплиться - to get things sorted) simplifies project setup by combining git clone, mise install, and task setup into single commands."
)]
struct Cli {
    /// Skip automatic synchronization between Razdfile.yml and mise.toml
    #[arg(long, global = true)]
    no_sync: bool,

    /// List all available tasks
    #[arg(long, global = true)]
    list: bool,

    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// Clone repository and set up project, or set up local project
    Up {
        /// Git repository URL to clone (optional for local projects)
        url: Option<String>,
        /// Directory name (defaults to repository name)
        #[arg(short, long)]
        name: Option<String>,
        /// Initialize new Razdfile.yml with project template
        #[arg(long)]
        init: bool,
    },
    /// List all available tasks from Razdfile.yml
    List {
        /// List all tasks, including internal ones
        #[arg(long)]
        list_all: bool,
        /// Output in JSON format
        #[arg(long)]
        json: bool,
    },
    /// Install development tools via mise
    Install,
    /// Install project dependencies via task setup
    Setup,
    /// Start development workflow
    Dev,
    /// Build project
    Build,
    /// Execute any custom task defined in Razdfile.yml
    Run {
        /// Task name to execute
        task_name: Option<String>,
        /// Arguments to pass to the task
        #[arg(trailing_var_arg = true)]
        args: Vec<String>,
        /// List all available tasks instead of running
        #[arg(long)]
        list: bool,
    },
}

#[tokio::main]
async fn main() {
    // Handle -v flag manually before clap parsing
    let args: Vec<String> = std::env::args().collect();
    if args.len() == 2 && (args[1] == "-v" || args[1] == "--version") {
        println!("{} {}", env!("CARGO_PKG_NAME"), env!("CARGO_PKG_VERSION"));
        return;
    }

    let cli = Cli::parse();

    if let Err(e) = run(cli).await {
        eprintln!("{} {}", "Error:".red().bold(), e);
        std::process::exit(1);
    }
}

async fn run(cli: Cli) -> core::Result<()> {
    // Store no_sync flag for use by commands
    std::env::set_var("RAZD_NO_SYNC", if cli.no_sync { "1" } else { "0" });

    // Handle global --list flag
    if cli.list {
        return commands::list::execute(false, false).await;
    }

    match cli.command {
        Some(Commands::Up { url, name, init }) => {
            commands::up::execute(url.as_deref(), name.as_deref(), init).await?;
        }
        Some(Commands::List { list_all, json }) => {
            commands::list::execute(list_all, json).await?;
        }
        Some(Commands::Install) => {
            commands::install::execute().await?;
        }
        Some(Commands::Setup) => {
            commands::setup::execute().await?;
        }
        Some(Commands::Dev) => {
            commands::dev::execute().await?;
        }
        Some(Commands::Build) => {
            commands::build::execute().await?;
        }
        Some(Commands::Run {
            task_name,
            args,
            list,
        }) => {
            if list {
                commands::list::execute(false, false).await?;
            } else {
                let task = task_name.ok_or_else(|| {
                    crate::core::error::RazdError::config(
                        "Task name required unless --list is specified".to_string(),
                    )
                })?;
                commands::run::execute(&task, &args).await?;
            }
        }
        None => {
            // If no subcommand provided, run 'razd up' (local project setup)
            commands::up::execute(None, None, false).await?;
        }
    }

    Ok(())
}
