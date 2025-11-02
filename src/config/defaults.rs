/// Default workflow definitions and utilities
use crate::defaults::DEFAULT_WORKFLOWS;

/// Generate default Razdfile.yml content
pub fn generate_default_razdfile() -> String {
    DEFAULT_WORKFLOWS.to_string()
}

/// Generate default Razdfile.yml with project-specific customizations
pub fn generate_project_razdfile(_project_type: &str) -> String {
    generate_default_razdfile()
}
