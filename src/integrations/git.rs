use crate::core::{output, Result, RazdError};
use crate::integrations::process;
use std::path::PathBuf;

/// Extract repository name from git URL
pub fn extract_repo_name(url: &str) -> Result<String> {
    let url = url.trim();
    
    // Handle common git URL formats
    let name = if url.ends_with(".git") {
        url.trim_end_matches(".git")
    } else {
        url
    };

    // Extract the last part of the path
    let name = name.split('/').last().unwrap_or(name);
    
    if name.is_empty() {
        return Err(RazdError::invalid_url("Could not extract repository name from URL"));
    }

    Ok(name.to_string())
}

/// Clone a git repository
pub async fn clone_repository(url: &str, target_dir: Option<&str>) -> Result<PathBuf> {
    // Check if git is available
    if !process::check_command_available("git").await {
        return Err(RazdError::missing_tool(
            "git",
            "https://git-scm.com/downloads"
        ));
    }

    let repo_name = if let Some(name) = target_dir {
        name.to_string()
    } else {
        extract_repo_name(url)?
    };

    let target_path = PathBuf::from(&repo_name);

    // Check if directory already exists
    if target_path.exists() {
        return Err(RazdError::git(format!(
            "Directory '{}' already exists. Please remove it or choose a different name.",
            repo_name
        )));
    }

    output::step(&format!("Cloning {} into {}", url, repo_name));
    
    process::execute_command("git", &["clone", url, &repo_name], None).await
        .map_err(|e| RazdError::git(format!("Failed to clone repository: {}", e)))?;

    output::success(&format!("Successfully cloned repository to {}", repo_name));
    
    Ok(target_path)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_repo_name() {
        assert_eq!(extract_repo_name("https://github.com/user/repo.git").unwrap(), "repo");
        assert_eq!(extract_repo_name("https://github.com/user/repo").unwrap(), "repo");
        assert_eq!(extract_repo_name("git@github.com:user/repo.git").unwrap(), "repo");
        assert_eq!(extract_repo_name("git@github.com:user/repo").unwrap(), "repo");
    }
}