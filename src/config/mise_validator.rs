use crate::core::RazdError;
use regex::Regex;

/// Backend types for mise tools
#[derive(Debug, PartialEq)]
enum BackendType {
    /// Package managers: npm, pipx, cargo, gem - allow @, /, alphanumeric, hyphens, underscores, dots, brackets
    PackageManager,
    /// Repository backends: aqua, github, gitlab, ubi, asdf, vfox, vfox-backend, spm - allow /, alphanumeric, hyphens, underscores, dots
    Repository,
    /// Go module paths: allow full paths like github.com/owner/repo/cmd/tool
    Go,
    /// HTTP URLs: skip validation (complex URL patterns)
    Http,
    /// Core tools or standalone: strict alphanumeric, hyphens, underscores only
    Strict,
}

/// Known backend prefixes in mise
const PACKAGE_MANAGER_BACKENDS: &[&str] = &["npm:", "pipx:", "cargo:", "gem:", "dotnet:"];
const REPOSITORY_BACKENDS: &[&str] = &[
    "aqua:",
    "github:",
    "gitlab:",
    "ubi:",
    "asdf:",
    "vfox:",
    "vfox-backend:",
    "spm:",
];
const GO_BACKEND: &str = "go:";
const HTTP_BACKEND: &str = "http:";
const CORE_BACKEND: &str = "core:";

/// Validate a tool or plugin name according to mise naming rules
/// Supports backend-specific validation for different tool types:
/// - Package managers (npm:, pipx:, cargo:, gem:): Allow scoped packages like @scope/package
/// - Repository backends (aqua:, github:, etc.): Allow owner/repo format
/// - Go backend: Allow full module paths like github.com/owner/repo/cmd/tool
/// - HTTP backend: Skip validation (URLs are complex)
/// - Standalone tools: Strict alphanumeric, hyphens, underscores only
pub fn validate_tool_name(name: &str) -> Result<(), RazdError> {
    if name.is_empty() {
        return Err(RazdError::config("Tool name cannot be empty".to_string()));
    }

    // Determine backend type and extract name after prefix
    let (backend_type, name_to_validate) = determine_backend_type(name);

    // Validate based on backend type
    match backend_type {
        BackendType::Http => {
            // Skip validation for HTTP URLs - they have complex patterns
            Ok(())
        }
        BackendType::PackageManager => validate_package_manager_name(name, name_to_validate),
        BackendType::Repository => validate_repository_name(name, name_to_validate),
        BackendType::Go => validate_go_module_path(name, name_to_validate),
        BackendType::Strict => validate_strict_name(name, name_to_validate),
    }
}

/// Determine the backend type from the tool name
fn determine_backend_type(name: &str) -> (BackendType, &str) {
    // Check for package manager backends
    for prefix in PACKAGE_MANAGER_BACKENDS {
        if name.starts_with(prefix) {
            return (BackendType::PackageManager, &name[prefix.len()..]);
        }
    }

    // Check for repository backends
    for prefix in REPOSITORY_BACKENDS {
        if name.starts_with(prefix) {
            return (BackendType::Repository, &name[prefix.len()..]);
        }
    }

    // Check for go backend
    if name.starts_with(GO_BACKEND) {
        return (BackendType::Go, &name[GO_BACKEND.len()..]);
    }

    // Check for http backend
    if name.starts_with(HTTP_BACKEND) {
        return (BackendType::Http, &name[HTTP_BACKEND.len()..]);
    }

    // Check for core backend
    if name.starts_with(CORE_BACKEND) {
        return (BackendType::Strict, &name[CORE_BACKEND.len()..]);
    }

    // Check for any other prefix (legacy support)
    if let Some(idx) = name.find(':') {
        // Unknown prefix - treat as repository-style (permissive)
        return (BackendType::Repository, &name[idx + 1..]);
    }

    // No prefix - strict validation
    (BackendType::Strict, name)
}

/// Validate package manager tool names (npm, pipx, cargo, gem)
/// Allows: @scope/package, package[extras], package.name
fn validate_package_manager_name(full_name: &str, name: &str) -> Result<(), RazdError> {
    if name.is_empty() {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Package name after prefix cannot be empty.",
            full_name
        )));
    }

    // Allow: alphanumeric, @, /, -, _, ., [, ], ,
    let valid_pattern = Regex::new(r"^[@a-zA-Z0-9/_.\-\[\],]+$").unwrap();

    if !valid_pattern.is_match(name) {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Package manager tool names can contain alphanumeric characters, @, /, -, _, ., [, ], ,. Examples: 'npm:cowsay', 'npm:@scope/package', 'pipx:package[extra]'",
            full_name
        )));
    }

    Ok(())
}

/// Validate repository-style tool names (aqua, github, gitlab, ubi, asdf, vfox)
/// Allows: owner/repo, owner/repo/subpath
fn validate_repository_name(full_name: &str, name: &str) -> Result<(), RazdError> {
    if name.is_empty() {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Repository path after prefix cannot be empty.",
            full_name
        )));
    }

    // Allow: alphanumeric, /, -, _, .
    let valid_pattern = Regex::new(r"^[a-zA-Z0-9/_.\-]+$").unwrap();

    if !valid_pattern.is_match(name) {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Repository tool names can contain alphanumeric characters, /, -, _, .. Examples: 'aqua:cli/cli', 'github:owner/repo'",
            full_name
        )));
    }

    Ok(())
}

/// Validate Go module paths
/// Allows: full paths like github.com/owner/repo/cmd/tool
fn validate_go_module_path(full_name: &str, name: &str) -> Result<(), RazdError> {
    if name.is_empty() {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Go module path after prefix cannot be empty.",
            full_name
        )));
    }

    // Allow: alphanumeric, /, -, _, .
    let valid_pattern = Regex::new(r"^[a-zA-Z0-9/_.\-]+$").unwrap();

    if !valid_pattern.is_match(name) {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Go module paths can contain alphanumeric characters, /, -, _, .. Examples: 'go:github.com/owner/repo/cmd/tool'",
            full_name
        )));
    }

    Ok(())
}

/// Validate strict tool names (standalone tools, core:)
/// Only allows: alphanumeric, hyphens, underscores
fn validate_strict_name(full_name: &str, name: &str) -> Result<(), RazdError> {
    if name.is_empty() {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Tool name cannot be empty.",
            full_name
        )));
    }

    let valid_pattern = Regex::new(r"^[a-zA-Z0-9_-]+$").unwrap();

    if !valid_pattern.is_match(name) {
        return Err(RazdError::config(format!(
            "Invalid tool name '{}'. Standalone tool names must contain only alphanumeric characters, hyphens, and underscores. Examples: 'node', 'python-3', 'my_tool'. For backend-prefixed tools, use formats like 'npm:@scope/package' or 'aqua:owner/repo'.",
            full_name
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
        return Err(RazdError::config("Plugin URL cannot be empty".to_string()));
    }

    // Allow URLs with protocols or git@ format
    let valid_url_pattern =
        Regex::new(r"^(https?://|git://|git@)[\w\-\.]+[:/][\w\-\./#]+$").unwrap();

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

    // ==================== Standalone (strict) tool name tests ====================

    #[test]
    fn test_valid_standalone_tool_names() {
        assert!(validate_tool_name("node").is_ok());
        assert!(validate_tool_name("python-3").is_ok());
        assert!(validate_tool_name("my_tool").is_ok());
        assert!(validate_tool_name("rust").is_ok());
        assert!(validate_tool_name("go-1-21").is_ok());
        assert!(validate_tool_name("my_custom_tool123").is_ok());
    }

    #[test]
    fn test_invalid_standalone_tool_names() {
        assert!(validate_tool_name("node js").is_err()); // space
        assert!(validate_tool_name("python@3").is_err()); // @ without prefix
        assert!(validate_tool_name("tool!name").is_err()); // !
        assert!(validate_tool_name("my/tool").is_err()); // / without prefix
        assert!(validate_tool_name("").is_err()); // empty
    }

    // ==================== Core backend tests ====================

    #[test]
    fn test_valid_core_backend_names() {
        assert!(validate_tool_name("core:node").is_ok());
        assert!(validate_tool_name("core:python").is_ok());
        assert!(validate_tool_name("core:go").is_ok());
        assert!(validate_tool_name("core:rust").is_ok());
    }

    #[test]
    fn test_invalid_core_backend_names() {
        assert!(validate_tool_name("core:").is_err()); // empty after prefix
        assert!(validate_tool_name("core:node/extra").is_err()); // / not allowed in core
    }

    // ==================== Package manager backend tests (npm, pipx, cargo, gem) ====================

    #[test]
    fn test_valid_npm_scoped_packages() {
        // Scoped packages - the main bug fix
        assert!(validate_tool_name("npm:@fission-ai/openspec").is_ok());
        assert!(validate_tool_name("npm:@babel/cli").is_ok());
        assert!(validate_tool_name("npm:@types/node").is_ok());
        assert!(validate_tool_name("npm:@scope/package-name").is_ok());
    }

    #[test]
    fn test_valid_npm_regular_packages() {
        assert!(validate_tool_name("npm:cowsay").is_ok());
        assert!(validate_tool_name("npm:typescript").is_ok());
        assert!(validate_tool_name("npm:eslint").is_ok());
    }

    #[test]
    fn test_valid_pipx_packages() {
        assert!(validate_tool_name("pipx:ansible").is_ok());
        assert!(validate_tool_name("pipx:black").is_ok());
        assert!(validate_tool_name("pipx:package[extra]").is_ok());
        assert!(validate_tool_name("pipx:package[extra1,extra2]").is_ok());
    }

    #[test]
    fn test_valid_cargo_packages() {
        assert!(validate_tool_name("cargo:ripgrep").is_ok());
        assert!(validate_tool_name("cargo:fd-find").is_ok());
        assert!(validate_tool_name("cargo:bat").is_ok());
    }

    #[test]
    fn test_valid_gem_packages() {
        assert!(validate_tool_name("gem:rails").is_ok());
        assert!(validate_tool_name("gem:bundler").is_ok());
    }

    #[test]
    fn test_invalid_package_manager_names() {
        assert!(validate_tool_name("npm:").is_err()); // empty after prefix
    }

    // ==================== Repository backend tests (aqua, github, gitlab, ubi, asdf, vfox) ====================

    #[test]
    fn test_valid_aqua_tools() {
        assert!(validate_tool_name("aqua:cli/cli").is_ok());
        assert!(validate_tool_name("aqua:sharkdp/fd").is_ok());
        assert!(validate_tool_name("aqua:BurntSushi/ripgrep").is_ok());
        assert!(validate_tool_name("aqua:junegunn/fzf").is_ok());
    }

    #[test]
    fn test_valid_github_tools() {
        assert!(validate_tool_name("github:jdx/mise").is_ok());
        assert!(validate_tool_name("github:owner/repo").is_ok());
        assert!(validate_tool_name("github:org-name/repo_name").is_ok());
    }

    #[test]
    fn test_valid_gitlab_tools() {
        assert!(validate_tool_name("gitlab:owner/repo").is_ok());
    }

    #[test]
    fn test_valid_ubi_tools() {
        assert!(validate_tool_name("ubi:sharkdp/fd").is_ok());
        assert!(validate_tool_name("ubi:BurntSushi/ripgrep").is_ok());
    }

    #[test]
    fn test_valid_asdf_vfox_tools() {
        assert!(validate_tool_name("asdf:nodejs").is_ok());
        assert!(validate_tool_name("asdf:mise-plugins/mise-python").is_ok());
        assert!(validate_tool_name("vfox:python").is_ok());
        assert!(validate_tool_name("vfox:mise-plugins/vfox-node").is_ok());
        assert!(validate_tool_name("vfox-backend:myplugin").is_ok());
    }

    #[test]
    fn test_invalid_repository_names() {
        assert!(validate_tool_name("aqua:").is_err()); // empty after prefix
        assert!(validate_tool_name("github:").is_err()); // empty after prefix
    }

    // ==================== Go backend tests ====================

    #[test]
    fn test_valid_go_module_paths() {
        assert!(
            validate_tool_name("go:github.com/golangci/golangci-lint/cmd/golangci-lint").is_ok()
        );
        assert!(validate_tool_name("go:golang.org/x/tools/gopls").is_ok());
        assert!(validate_tool_name("go:github.com/owner/repo").is_ok());
        assert!(validate_tool_name("go:mvdan.cc/sh/v3/cmd/shfmt").is_ok());
    }

    #[test]
    fn test_invalid_go_module_paths() {
        assert!(validate_tool_name("go:").is_err()); // empty after prefix
    }

    // ==================== HTTP backend tests ====================

    #[test]
    fn test_http_backend_skips_validation() {
        // HTTP backend should skip validation (URLs are complex)
        assert!(validate_tool_name("http:https://example.com/tool").is_ok());
        assert!(validate_tool_name("http:anything-goes-here").is_ok());
    }

    // ==================== Legacy prefix tests (backwards compatibility) ====================

    #[test]
    fn test_valid_tool_names_with_legacy_prefixes() {
        // These should still work for backwards compatibility
        assert!(validate_tool_name("asdf:nodejs").is_ok());
        assert!(validate_tool_name("vfox:python").is_ok());
        assert!(validate_tool_name("vfox-backend:myplugin").is_ok());
    }

    // ==================== Unknown prefix tests ====================

    #[test]
    fn test_unknown_prefix_treated_as_repository() {
        // Unknown prefixes should be treated as repository-style (permissive)
        assert!(validate_tool_name("unknown:owner/repo").is_ok());
        assert!(validate_tool_name("custom:tool-name").is_ok());
    }

    // ==================== Plugin URL tests ====================

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
