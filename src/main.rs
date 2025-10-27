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
    },
    /// Install development tools via mise
    Install,
    /// Install project dependencies via task setup
    Setup,
    /// Start development workflow
    Dev,
    /// Build project
    Build,
    /// Execute tasks from Taskfile.yml
    Task {
        /// Task name to execute (if empty, runs default dev task)
        name: Option<String>,
        /// Arguments to pass to the task
        #[arg(trailing_var_arg = true)]
        args: Vec<String>,
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
    match cli.command {
        Some(Commands::Up { url, name }) => {
            commands::up::execute(url.as_deref(), name.as_deref()).await?;
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
        Some(Commands::Task { name, args }) => {
            commands::task::execute(name.as_deref(), &args).await?;
        }
        None => {
            // If no subcommand provided, run 'razd up' (local project setup)
            commands::up::execute(None, None).await?;
        }
    }

    Ok(())
}
