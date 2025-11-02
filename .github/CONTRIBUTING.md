# Contributing to razd

Thank you for your interest in contributing to razd!

## Quick Start

### Prerequisites

- Rust 1.82.0+
- Git
- mise installed (https://mise.jdx.dev)

**Install razd:**
```sh
# Install the Razd plugin
mise plugin install razd https://github.com/razd-cli/vfox-plugin-razd

# Install and use the latest version globally
mise use -g razd@latest

# Verify installation
razd --version
```

### Setup

```sh
git clone https://github.com/your-username/razd.git
cd razd

razd  # Installs deps + builds
```

## Development Commands

```sh
# Main commands (from Razdfile.yml)
razd                # Default: install + build
razd run install    # Install dependencies
razd run build      # Build debug version
razd run test       # Run all tests

# Cargo alternatives
cargo build         # Build
cargo test          # Test
cargo run -- <args> # Run CLI
```

## Code Quality

Before submitting PR:

```sh
cargo fmt                   # Format code
cargo clippy -- -D warnings # Lint
cargo test                  # Test
cargo audit                 # Security check
```

**Requirements:**
- ✅ Code formatted with `rustfmt`
- ✅ No clippy warnings
- ✅ All tests pass (Windows, macOS, Linux)
- ✅ No security vulnerabilities

## Pull Request Workflow

1. Create branch: `git checkout -b feature/name`
2. Make changes + add tests
3. Run quality checks (see above)
4. Push: `git push origin feature/name`
5. Create PR on GitHub

## Release Process

1. Update version in `Cargo.toml`
2. Update `CHANGELOG.md`
3. Test: `cargo test && cargo build --release`
4. Commit changes
5. Tag: `git tag -a v0.4.2 -m "Release v0.4.2"`
6. Push: `git push origin v0.4.2`
7. GitHub Actions will build and publish release

## Project Structure

```
razd/
├── src/
│   ├── commands/      # CLI commands
│   ├── config/        # Configuration
│   ├── core/          # Core utilities
│   └── integrations/  # Tool integrations
├── tests/             # Integration tests
├── Razdfile.yml       # Task definitions
└── Cargo.toml         # Rust manifest
```

## Common Tasks

```sh
# Local testing
cargo run -- up https://github.com/user/repo
cargo run -- run build
cargo run -- --help

# Install locally
cargo install --path .

# Clean build
cargo clean && cargo build
```

## Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Windows  | x86_64       | ✅     |
| macOS    | x86_64       | ✅     |
| macOS    | aarch64      | ✅     |
| Linux    | x86_64       | ✅     |

## Getting Help

- GitHub Issues: Bug reports and features
- GitHub Discussions: Questions

Thank you for contributing! 🤩