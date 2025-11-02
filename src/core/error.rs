use thiserror::Error;

#[derive(Error, Debug)]
pub enum RazdError {
    #[error("Git operation failed: {0}")]
    Git(String),

    #[error("Mise operation failed: {0}")]
    Mise(String),

    #[error("Task operation failed: {0}")]
    Task(String),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Invalid URL: {0}")]
    #[allow(dead_code)]
    InvalidUrl(String),

    #[error("Missing required tool: {tool}. Please install it first.\nInstallation guide: {help}")]
    MissingTool { tool: String, help: String },

    #[error("Configuration error: {0}")]
    Config(String),

    #[error("Command error: {0}")]
    Command(String),

    #[error("No project configuration found in current directory.\n{suggestion}")]
    NoProjectConfig { suggestion: String },

    #[error("No default task found in Razdfile.yml.\nPlease add a 'default' task or specify a task name: razd run <name>")]
    NoDefaultTask,

    #[error("Interactive setup cancelled by user")]
    #[allow(dead_code)]
    SetupCancelled,

    #[error("Project type '{project_type}' not recognized. Using generic template.")]
    #[allow(dead_code)]
    UnknownProjectType { project_type: String },
}

impl RazdError {
    pub fn git<S: Into<String>>(msg: S) -> Self {
        Self::Git(msg.into())
    }

    pub fn mise<S: Into<String>>(msg: S) -> Self {
        Self::Mise(msg.into())
    }

    pub fn task<S: Into<String>>(msg: S) -> Self {
        Self::Task(msg.into())
    }

    #[allow(dead_code)]
    pub fn invalid_url<S: Into<String>>(msg: S) -> Self {
        RazdError::InvalidUrl(msg.into())
    }

    pub fn missing_tool<S: Into<String>>(tool: S, help: S) -> Self {
        Self::MissingTool {
            tool: tool.into(),
            help: help.into(),
        }
    }

    pub fn config<S: Into<String>>(msg: S) -> Self {
        Self::Config(msg.into())
    }

    pub fn command<S: Into<String>>(msg: S) -> Self {
        Self::Command(msg.into())
    }

    pub fn no_project_config<S: Into<String>>(suggestion: S) -> Self {
        Self::NoProjectConfig {
            suggestion: suggestion.into(),
        }
    }

    pub fn no_default_task() -> Self {
        Self::NoDefaultTask
    }

    #[allow(dead_code)]
    pub fn setup_cancelled() -> Self {
        Self::SetupCancelled
    }

    #[allow(dead_code)]
    pub fn unknown_project_type<S: Into<String>>(project_type: S) -> Self {
        Self::UnknownProjectType {
            project_type: project_type.into(),
        }
    }
}

pub type Result<T> = std::result::Result<T, RazdError>;
