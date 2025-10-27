use crate::core::RazdError;
use regex::Regex;

/// Validate a tool or plugin name according to mise naming rules
/// Valid names contain alphanumeric characters, hyphens, underscores
/// and optional type prefixes like "asdf:", "vfox:", "vfox-backend:"
pub fn validate_tool_name(name: &str) -> Result<(), RazdError> {
    // Allow type prefixes
    let name_without_prefix = if let Some(idx) = name.find(':') {
        &name[idx + 1..]
    } else {
        name
    };

    // Check for valid characters: alphanumeric, hyphens, underscores
    let valid_pattern = Regex::new(r"^[a-zA-Z0-9_-]+$").unwrap();
    
    if !valid_pattern.is_match(name_without_prefix) {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Tool names must contain only alphanumeric characters, hyphens, and underscores. Examples: 'node', 'python-3', 'my_tool'",
            name
        )));
    }

    Ok(())
}

/// Validate a plugin URL
/// Must be a valid URL format, optionally with git refs
pub fn validate_plugin_url(url: &str) -> Result<(), RazdError> {
    // Check for basic URL patterns (http://, https://, git://)
    // or GitHub shorthand (github.com/...)
    if url.is_empty() {
        return Err(RazdError::config(
            "Plugin URL cannot be empty".to_string()
        ));
    }

    // Allow URLs with protocols or git@ format
    let valid_url_pattern = Regex::new(
        r"^(https?://|git://|git@)[\w\-\.]+[:/][\w\-\./#]+$"
    ).unwrap();
    
    if !valid_url_pattern.is_match(url) {
        return Err(RazdError::config(format!(
            "Invalid plugin URL '{}'. Plugin URLs must be valid git repository URLs. Examples: 'https://github.com/org/repo.git', 'https://github.com/org/repo.git#ref'",
            url
        )));
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_tool_names() {
        assert!(validate_tool_name("node").is_ok());
        assert!(validate_tool_name("python-3").is_ok());
        assert!(validate_tool_name("my_tool").is_ok());
        assert!(validate_tool_name("rust").is_ok());
        assert!(validate_tool_name("go-1-21").is_ok());
        assert!(validate_tool_name("my_custom_tool123").is_ok());
    }

    #[test]
    fn test_valid_tool_names_with_prefixes() {
        assert!(validate_tool_name("asdf:nodejs").is_ok());
        assert!(validate_tool_name("vfox:python").is_ok());
        assert!(validate_tool_name("vfox-backend:myplugin").is_ok());
    }

    #[test]
    fn test_invalid_tool_names() {
        assert!(validate_tool_name("node js").is_err());
        assert!(validate_tool_name("python@3").is_err());
        assert!(validate_tool_name("tool!name").is_err());
        assert!(validate_tool_name("my/tool").is_err());
        assert!(validate_tool_name("").is_err());
    }

    #[test]
    fn test_valid_plugin_urls() {
        assert!(validate_plugin_url("https://github.com/org/repo.git").is_ok());
        assert!(validate_plugin_url("https://github.com/org/repo.git#DEADBEEF").is_ok());
        assert!(validate_plugin_url("https://gitlab.com/user/project.git").is_ok());
        assert!(validate_plugin_url("git://github.com/org/repo.git").is_ok());
        assert!(validate_plugin_url("git@github.com:org/repo.git").is_ok());
    }

    #[test]
    fn test_invalid_plugin_urls() {
        assert!(validate_plugin_url("").is_err());
        assert!(validate_plugin_url("not-a-url").is_err());
        assert!(validate_plugin_url("ftp://invalid.com").is_err());
    }
}
