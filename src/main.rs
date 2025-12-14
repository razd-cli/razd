mod commands;
mod config;
mod core;
mod defaults;
mod integrations;

use clap::{Parser, Subcommand};
use colored::*;
use std::path::PathBuf;

#[derive(Parser)]
#[command(
    name = "razd",
    version,
    about = "Streamlined project setup with git, mise, and taskfile integration",
    long_about = "razd (раздуплиться - to get things sorted) simplifies project setup by combining git clone, mise install, and task setup into single commands."
)]
struct Cli {
    /// Specify custom taskfile/razdfile path
    #[arg(short = 't', long, global = true, value_name = "FILE")]
    taskfile: Option<String>,

    /// Specify custom razdfile path (overrides --taskfile)
    #[arg(long, global = true, value_name = "FILE")]
    razdfile: Option<String>,

    /// Skip automatic synchronization between Razdfile.yml and mise.toml
    #[arg(long, global = true)]
    no_sync: bool,

    /// Automatically answer "yes" to all prompts
    #[arg(short = 'y', long, global = true)]
    yes: bool,

    /// List all available tasks
    #[arg(long, global = true)]
    list: bool,

    #[command(subcommand)]
    command: Option<Commands>,
}

/// Resolve the configuration file path based on CLI flags
/// Priority: --razdfile > --taskfile > None (default)
fn resolve_config_path(cli: &Cli) -> Option<PathBuf> {
    cli.razdfile
        .as_ref()
        .or(cli.taskfile.as_ref())
        .map(PathBuf::from)
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
    /// Manage project trust status
    Trust {
        /// Path to trust (defaults to current directory)
        path: Option<String>,
        /// Remove trust status from the project
        #[arg(long)]
        untrust: bool,
        /// Show trust status without modifying
        #[arg(long)]
        show: bool,
        /// Trust all parent directories with config
        #[arg(short, long)]
        all: bool,
        /// Ignore this project (never trust, never prompt)
        #[arg(long)]
        ignore: bool,
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

    // Store yes flag for use by commands
    std::env::set_var("RAZD_AUTO_YES", if cli.yes { "1" } else { "0" });

    // Resolve custom config path from flags
    let custom_path = resolve_config_path(&cli);

    // Handle global --list flag
    if cli.list {
        return commands::list::execute(false, false, custom_path).await;
    }

    match cli.command {
        Some(Commands::Up { url, name, init }) => {
            commands::up::execute(url.as_deref(), name.as_deref(), init, custom_path).await?;
        }
        Some(Commands::List { list_all, json }) => {
            commands::list::execute(list_all, json, custom_path).await?;
        }
        Some(Commands::Install) => {
            commands::install::execute().await?;
        }
        Some(Commands::Setup) => {
            commands::setup::execute(custom_path).await?;
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
                commands::list::execute(false, false, custom_path).await?;
            } else {
                let task = task_name.ok_or_else(|| {
                    crate::core::error::RazdError::config(
                        "Task name required unless --list is specified".to_string(),
                    )
                })?;
                commands::run::execute(&task, &args, custom_path).await?;
            }
        }
        Some(Commands::Trust {
            path,
            untrust,
            show,
            all,
            ignore,
        }) => {
            commands::trust::execute(path.as_deref(), untrust, show, all, ignore).await?;
        }
        None => {
            // If no subcommand provided, run 'razd up' (local project setup)
            commands::up::execute(None, None, false, custom_path).await?;
        }
    }

    Ok(())
}
