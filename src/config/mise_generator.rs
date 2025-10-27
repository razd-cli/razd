use crate::config::razdfile::{MiseConfig, ToolConfig};
use crate::core::Result;
use toml_edit::{value, Array, DocumentMut, InlineTable, Item, Table, Value};

/// Generate mise.toml content from MiseConfig
pub fn generate_mise_toml(mise_config: &MiseConfig) -> Result<String> {
    let mut toml_doc = DocumentMut::new();

    // Add tools section if present
    if let Some(ref tools) = mise_config.tools {
        if !tools.is_empty() {
            let mut tools_table = Table::new();

            for (name, config) in tools {
                match config {
                    ToolConfig::Simple(version) => {
                        tools_table.insert(name, value(version.as_str()));
                    }
                    ToolConfig::Complex {
                        version,
                        postinstall,
                        os,
                        install_env,
                    } => {
                        let mut tool_table = InlineTable::new();
                        tool_table.insert("version", Value::from(version.as_str()));

                        if let Some(cmd) = postinstall {
                            tool_table.insert("postinstall", Value::from(cmd.as_str()));
                        }

                        if let Some(os_list) = os {
                            let mut os_array = Array::new();
                            for os_item in os_list {
                                os_array.push(os_item.as_str());
                            }
                            tool_table.insert("os", Value::from(os_array));
                        }

                        if let Some(env) = install_env {
                            let mut env_table = InlineTable::new();
                            for (k, v) in env {
                                env_table.insert(k, Value::from(v.as_str()));
                            }
                            tool_table.insert("install_env", Value::from(env_table));
                        }

                        tools_table.insert(name, value(tool_table));
                    }
                }
            }

            toml_doc.insert("tools", Item::Table(tools_table));
        }
    }

    // Add plugins section if present
    if let Some(ref plugins) = mise_config.plugins {
        if !plugins.is_empty() {
            let mut plugins_table = Table::new();

            for (name, url) in plugins {
                plugins_table.insert(name, value(url.as_str()));
            }

            toml_doc.insert("plugins", Item::Table(plugins_table));
        }
    }

    Ok(toml_doc.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    #[test]
    fn test_generate_simple_tools() {
        let mut tools = HashMap::new();
        tools.insert("node".to_string(), ToolConfig::Simple("22".to_string()));
        tools.insert("python".to_string(), ToolConfig::Simple("3.11".to_string()));

        let mise_config = MiseConfig {
            tools: Some(tools),
            plugins: None,
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        assert!(toml.contains("[tools]"));
        assert!(toml.contains("node = \"22\""));
        assert!(toml.contains("python = \"3.11\""));
    }

    #[test]
    fn test_generate_complex_tools() {
        let mut tools = HashMap::new();
        tools.insert(
            "node".to_string(),
            ToolConfig::Complex {
                version: "22".to_string(),
                postinstall: Some("corepack enable".to_string()),
                os: Some(vec!["linux".to_string(), "macos".to_string()]),
                install_env: None,
            },
        );

        let mise_config = MiseConfig {
            tools: Some(tools),
            plugins: None,
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        assert!(toml.contains("[tools]"));
        assert!(toml.contains("node = "));
        assert!(toml.contains("version = \"22\""));
        assert!(toml.contains("postinstall = \"corepack enable\""));
        assert!(toml.contains("os = [\"linux\", \"macos\"]"));
    }

    #[test]
    fn test_generate_tools_with_install_env() {
        let mut install_env = HashMap::new();
        install_env.insert("CGO_ENABLED".to_string(), "1".to_string());
        install_env.insert("GOARCH".to_string(), "amd64".to_string());

        let mut tools = HashMap::new();
        tools.insert(
            "go".to_string(),
            ToolConfig::Complex {
                version: "1.21".to_string(),
                postinstall: None,
                os: None,
                install_env: Some(install_env),
            },
        );

        let mise_config = MiseConfig {
            tools: Some(tools),
            plugins: None,
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        assert!(toml.contains("[tools]"));
        assert!(toml.contains("go = "));
        assert!(toml.contains("version = \"1.21\""));
        assert!(toml.contains("install_env"));
        assert!(toml.contains("CGO_ENABLED = \"1\""));
        assert!(toml.contains("GOARCH = \"amd64\""));
    }

    #[test]
    fn test_generate_plugins() {
        let mut plugins = HashMap::new();
        plugins.insert(
            "elixir".to_string(),
            "https://github.com/my-org/mise-elixir.git".to_string(),
        );
        plugins.insert(
            "node".to_string(),
            "https://github.com/my-org/mise-node.git#DEADBEEF".to_string(),
        );

        let mise_config = MiseConfig {
            tools: None,
            plugins: Some(plugins),
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        assert!(toml.contains("[plugins]"));
        assert!(toml.contains("elixir = \"https://github.com/my-org/mise-elixir.git\""));
        assert!(toml.contains("node = \"https://github.com/my-org/mise-node.git#DEADBEEF\""));
    }

    #[test]
    fn test_generate_both_tools_and_plugins() {
        let mut tools = HashMap::new();
        tools.insert("node".to_string(), ToolConfig::Simple("22".to_string()));

        let mut plugins = HashMap::new();
        plugins.insert(
            "elixir".to_string(),
            "https://github.com/my-org/mise-elixir.git".to_string(),
        );

        let mise_config = MiseConfig {
            tools: Some(tools),
            plugins: Some(plugins),
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        assert!(toml.contains("[tools]"));
        assert!(toml.contains("node = \"22\""));
        assert!(toml.contains("[plugins]"));
        assert!(toml.contains("elixir = \"https://github.com/my-org/mise-elixir.git\""));
    }

    #[test]
    fn test_generate_empty_config() {
        let mise_config = MiseConfig {
            tools: None,
            plugins: None,
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        // Should generate empty document
        assert!(toml.trim().is_empty());
    }

    #[test]
    fn test_generate_empty_maps() {
        let mise_config = MiseConfig {
            tools: Some(HashMap::new()),
            plugins: Some(HashMap::new()),
        };

        let toml = generate_mise_toml(&mise_config).unwrap();

        // Should generate empty document for empty maps
        assert!(toml.trim().is_empty());
    }
}
