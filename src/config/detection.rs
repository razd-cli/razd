use std::path::Path;

/// Detect project type based on files in the current directory
#[allow(dead_code)]
pub fn detect_project_type<P: AsRef<Path>>(path: P) -> String {
    let path = path.as_ref();

    // Check for Node.js
    if path.join("package.json").exists() {
        return "node".to_string();
    }

    // Check for Python
    if path.join("requirements.txt").exists()
        || path.join("pyproject.toml").exists()
        || path.join("setup.py").exists()
    {
        return "python".to_string();
    }

    // Check for Rust
    if path.join("Cargo.toml").exists() {
        return "rust".to_string();
    }

    // Check for Go
    if path.join("go.mod").exists() {
        return "go".to_string();
    }

    // Check for other common project files
    if path.join("Dockerfile").exists() {
        return "docker".to_string();
    }

    "generic".to_string()
}

/// Get recommended tools for a project type
#[allow(dead_code)]
pub fn get_recommended_tools(project_type: &str) -> Vec<&'static str> {
    match project_type {
        "node" => vec!["node", "npm", "yarn"],
        "python" => vec!["python", "pip", "poetry"],
        "rust" => vec!["rust", "cargo"],
        "go" => vec!["go"],
        "docker" => vec!["docker", "docker-compose"],
        _ => vec![],
    }
}

/// Generate mise.toml content based on project type
#[allow(dead_code)]
pub fn generate_mise_config(project_type: &str) -> String {
    let tools = get_recommended_tools(project_type);

    if tools.is_empty() {
        return String::new();
    }

    let mut config = String::new();
    for tool in tools {
        match tool {
            "node" => config.push_str("node = \"lts\"\n"),
            "python" => config.push_str("python = \"3.11\"\n"),
            "rust" => config.push_str("rust = \"stable\"\n"),
            "go" => config.push_str("go = \"latest\"\n"),
            _ => {}
        }
    }

    config
}

/// Generate basic Taskfile.yml content based on project type
#[allow(dead_code)]
pub fn generate_taskfile_config(project_type: &str) -> String {
    match project_type {
        "node" => r#"version: '3'

tasks:
  default:
    desc: "Run development server"
    cmds:
      - npm run dev
      
  setup:
    desc: "Install dependencies"
    cmds:
      - npm install
      
  build:
    desc: "Build for production"
    cmds:
      - npm run build
      
  test:
    desc: "Run tests"
    cmds:
      - npm test
"#
        .to_string(),

        "python" => r#"version: '3'

tasks:
  default:
    desc: "Run application"
    cmds:
      - python -m src.main
      
  setup:
    desc: "Install dependencies"
    cmds:
      - pip install -r requirements.txt
      
  build:
    desc: "Build package"
    cmds:
      - python -m build
      
  test:
    desc: "Run tests"
    cmds:
      - pytest
"#
        .to_string(),

        "rust" => r#"version: '3'

tasks:
  default:
    desc: "Run application"
    cmds:
      - cargo run
      
  setup:
    desc: "Build dependencies"
    cmds:
      - cargo build
      
  build:
    desc: "Build for production"
    cmds:
      - cargo build --release
      
  test:
    desc: "Run tests"
    cmds:
      - cargo test
"#
        .to_string(),

        "go" => r#"version: '3'

tasks:
  default:
    desc: "Run application"
    cmds:
      - go run .
      
  setup:
    desc: "Download dependencies"
    cmds:
      - go mod download
      
  build:
    desc: "Build binary"
    cmds:
      - go build -o bin/app
      
  test:
    desc: "Run tests"
    cmds:
      - go test ./...
"#
        .to_string(),

        _ => r#"version: '3'

tasks:
  default:
    desc: "Default task"
    cmds:
      - echo "Please configure your tasks"
      
  setup:
    desc: "Setup project"
    cmds:
      - echo "Setup your project here"
      
  build:
    desc: "Build project"
    cmds:
      - echo "Build your project here"
      
  test:
    desc: "Run tests"
    cmds:
      - echo "Run your tests here"
"#
        .to_string(),
    }
}
