use colored::*;

pub fn success(msg: &str) {
    println!("{} {}", "✓".green().bold(), msg);
}

pub fn info(msg: &str) {
    println!("{} {}", "ℹ".blue().bold(), msg);
}

pub fn warning(msg: &str) {
    println!("{} {}", "⚠".yellow().bold(), msg.yellow());
}

pub fn error(msg: &str) {
    eprintln!("{} {}", "✗".red().bold(), msg.red());
}

pub fn step(step: &str) {
    println!("{} {}", "→".cyan().bold(), step.bold());
}