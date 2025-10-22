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
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Clone repository and set up project (git clone + mise install + task setup)
    Up {
        /// Git repository URL to clone
        url: String,
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
    /// Initialize razd configuration for a project
    Init {
        /// Create Razdfile.yml for workflow customization
        #[arg(long)]
        config: bool,
        /// Create all files (Razdfile.yml, Taskfile.yml, mise.toml)
        #[arg(long)]
        full: bool,
    },
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();

    if let Err(e) = run(cli).await {
        eprintln!("{} {}", "Error:".red().bold(), e);
        std::process::exit(1);
    }
}

async fn run(cli: Cli) -> core::Result<()> {
    match cli.command {
        Commands::Up { url, name } => {
            commands::up::execute(&url, name.as_deref()).await?;
        }
        Commands::Install => {
            commands::install::execute().await?;
        }
        Commands::Setup => {
            commands::setup::execute().await?;
        }
        Commands::Dev => {
            commands::dev::execute().await?;
        }
        Commands::Build => {
            commands::build::execute().await?;
        }
        Commands::Task { name, args } => {
            commands::task::execute(name.as_deref(), &args).await?;
        }
        Commands::Init { config, full } => {
            commands::init::execute(config, full).await?;
        }
    }

    Ok(())
}