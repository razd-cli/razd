/// Default workflow definitions and utilities
use crate::defaults::DEFAULT_WORKFLOWS;

/// Generate default Razdfile.yml content
pub fn generate_default_razdfile() -> String {
    DEFAULT_WORKFLOWS.to_string()
}

/// Generate default Razdfile.yml with project-specific customizations
pub fn generate_project_razdfile(project_type: &str) -> String {
    match project_type {
        "node" | "nodejs" => generate_node_razdfile(),
        "python" => generate_python_razdfile(),
        "rust" => generate_rust_razdfile(),
        "go" => generate_go_razdfile(),
        _ => generate_default_razdfile(),
    }
}

fn generate_node_razdfile() -> String {
    r#"version: '3'

tasks:
  up:
    desc: "Set up Node.js project"
    cmds:
      - echo "🚀 Setting up Node.js project..."
      - mise install
      - npm install
      - task setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      - npm install
      
  dev:
    desc: "Start development server"
    cmds:
      - echo "🚀 Starting development..."
      - npm run dev
      
  build:
    desc: "Build project"
    cmds:
      - echo "🔨 Building project..."
      - npm run build
"#.to_string()
}

fn generate_python_razdfile() -> String {
    r#"version: '3'

tasks:
  up:
    desc: "Set up Python project"
    cmds:
      - echo "🚀 Setting up Python project..."
      - mise install
      - pip install -r requirements.txt
      - task setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      - pip install -r requirements.txt
      
  dev:
    desc: "Start development workflow"
    cmds:
      - echo "🚀 Starting development..."
      - python -m src.main
      
  build:
    desc: "Build project"
    cmds:
      - echo "🔨 Building project..."
      - python -m build
"#.to_string()
}

fn generate_rust_razdfile() -> String {
    r#"version: '3'

tasks:
  up:
    desc: "Set up Rust project"
    cmds:
      - echo "🚀 Setting up Rust project..."
      - mise install
      - cargo build
      - task setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      
  dev:
    desc: "Start development workflow"
    cmds:
      - echo "🚀 Starting development..."
      - cargo run
      
  build:
    desc: "Build project"
    cmds:
      - echo "🔨 Building project..."
      - cargo build --release
"#.to_string()
}

fn generate_go_razdfile() -> String {
    r#"version: '3'

tasks:
  up:
    desc: "Set up Go project"
    cmds:
      - echo "🚀 Setting up Go project..."
      - mise install
      - go mod download
      - task setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      - go mod download
      
  dev:
    desc: "Start development workflow"
    cmds:
      - echo "🚀 Starting development..."
      - go run .
      
  build:
    desc: "Build project"
    cmds:
      - echo "🔨 Building project..."
      - go build -o bin/app
"#.to_string()
}