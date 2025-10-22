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
    InvalidUrl(String),

    #[error("Missing required tool: {tool}. Please install it first.\nInstallation guide: {help}")]
    MissingTool { tool: String, help: String },

    #[error("Configuration error: {0}")]
    Config(String),
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

    pub fn invalid_url<S: Into<String>>(msg: S) -> Self {
        Self::InvalidUrl(msg.into())
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
}

pub type Result<T> = std::result::Result<T, RazdError>;