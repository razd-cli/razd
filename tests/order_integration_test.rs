use razd::config::mise_sync::{MiseSyncManager, SyncConfig};
use razd::config::razdfile::{RazdfileConfig, TaskConfig, Command, MiseConfig, ToolConfig};
use indexmap::IndexMap;
use std::fs;
use tempfile::TempDir;

#[test]
fn test_sync_preserves_mise_before_tasks_order() {
    let temp_dir = TempDir::new().unwrap();
    let project_root = temp_dir.path();

    // Создаём mise.toml
    let mise_content = r#"[tools]
node = "24"
pnpm = "latest"

[plugins]
node = "https://github.com/asdf-vm/asdf-nodejs.git"
"#;
    fs::write(project_root.join("mise.toml"), mise_content).unwrap();

    // Создаём Razdfile с tasks в неправильном порядке
    let mut tasks = IndexMap::new();
    tasks.insert("build".to_string(), TaskConfig {
        desc: Some("Build".to_string()),
        cmds: vec![Command::String("echo build".to_string())],
        internal: false,
    });
    tasks.insert("custom".to_string(), TaskConfig {
        desc: Some("Custom".to_string()),
        cmds: vec![Command::String("echo custom".to_string())],
        internal: false,
    });
    tasks.insert("install".to_string(), TaskConfig {
        desc: Some("Install".to_string()),
        cmds: vec![Command::String("echo install".to_string())],
        internal: false,
    });
    tasks.insert("default".to_string(), TaskConfig {
        desc: Some("Default".to_string()),
        cmds: vec![Command::String("echo default".to_string())],
        internal: false,
    });

    let razdfile = RazdfileConfig {
        version: "3".to_string(),
        mise: None,
        tasks,
    };

    let yaml = serde_yaml::to_string(&razdfile).unwrap();
    fs::write(project_root.join("Razdfile.yml"), yaml).unwrap();

    // Инициализируем tracking state
    razd::config::file_tracker::update_tracking_state(project_root).unwrap();

    // Изменяем mise.toml
    let new_mise_content = r#"[tools]
node = "24"
pnpm = "latest"
python = "3.11"

[plugins]
node = "https://github.com/asdf-vm/asdf-nodejs.git"
"#;
    fs::write(project_root.join("mise.toml"), new_mise_content).unwrap();

    // Синхронизируем
    let config = SyncConfig {
        no_sync: false,
        auto_approve: true,
        create_backups: false,
    };

    let manager = MiseSyncManager::new(project_root.to_path_buf(), config);
    let result = manager.check_and_sync_if_needed().unwrap();

    println!("Sync result: {:?}", result);

    // Читаем обновлённый Razdfile
    let updated_yaml = fs::read_to_string(project_root.join("Razdfile.yml")).unwrap();
    println!("Updated Razdfile.yml:\n{}", updated_yaml);

    // Проверяем, что mise идёт перед tasks
    let mise_pos = updated_yaml.find("mise:").expect("mise section not found");
    let tasks_pos = updated_yaml.find("tasks:").expect("tasks section not found");
    assert!(mise_pos < tasks_pos, "mise section should come before tasks section");

    // Проверяем порядок tasks
    let config = RazdfileConfig::load_from_path(project_root.join("Razdfile.yml"))
        .unwrap()
        .unwrap();

    let task_names: Vec<_> = config.tasks.keys().map(|s| s.as_str()).collect();
    println!("Task order: {:?}", task_names);

    // Проверяем, что tasks в правильном порядке
    assert_eq!(task_names[0], "default");
    assert_eq!(task_names[1], "install");
    assert_eq!(task_names[2], "build");
    assert_eq!(task_names[3], "custom"); // По алфавиту после стандартных

    // Проверяем, что mise config обновился
    let mise = config.mise.unwrap();
    let tools = mise.tools.unwrap();
    assert!(tools.contains_key("python"), "Should have python tool");
}
