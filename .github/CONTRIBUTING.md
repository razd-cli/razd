# Contributing to razd

Thank you for your interest in contributing to razd! This document provides guidelines for contributing to the project.

## Development Workflow

### Prerequisites

- Rust 1.71.0 or later (check `rust-version` in `Cargo.toml`)
- Git for version control
- GitHub account for pull requests

### Setting Up Development Environment

1. **Fork and clone the repository:**
   ```sh
   git clone https://github.com/your-username/razd.git
   cd razd
   ```

2. **Install Rust toolchain:**
   ```sh
   rustup update
   rustup component add rustfmt clippy
   ```

3. **Build and test:**
   ```sh
   cargo build
   cargo test
   ```

### Code Quality Standards

Our CI pipeline enforces the following quality standards:

#### Formatting
```sh
cargo fmt --check
```
All code must be formatted with `rustfmt` using default settings.

#### Linting
```sh
cargo clippy -- -D warnings
```
All clippy warnings must be resolved before merging.

#### Testing
```sh
cargo test
```
All tests must pass on all supported platforms (Windows, macOS, Linux).

#### Security Audit
```sh
cargo audit
```
Dependencies must be free of known security vulnerabilities.

### Pull Request Process

1. **Create a feature branch:**
   ```sh
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write clear, concise commit messages
   - Add tests for new functionality
   - Update documentation as needed

3. **Test locally:**
   ```sh
   cargo test
   cargo fmt --check
   cargo clippy -- -D warnings
   cargo audit
   ```

4. **Push and create PR:**
   ```sh
   git push origin feature/your-feature-name
   ```
   Then create a pull request through GitHub.

### CI/CD Pipeline

Our GitHub Actions workflows automatically:

#### Continuous Integration (`ci.yml`)
- **Matrix Testing**: Tests on Windows, macOS, and Linux
- **Rust Versions**: Tests on stable and MSRV (1.71.0)
- **Quality Checks**: Runs rustfmt, clippy, and security audit
- **Test Coverage**: Generates coverage reports via codecov

Triggers:
- Push to `main` branch
- Pull requests to `main` branch
- Manual workflow dispatch

#### Release Automation (`release.yml`)
- **Cross-platform Builds**: Creates binaries for all supported platforms
- **Asset Creation**: Generates archives with checksums
- **GitHub Releases**: Automatically publishes releases on version tags

Triggers:
- Git tags matching `v*` pattern (e.g., `v1.0.0`)

### Platform Support

razd supports the following platforms:

| Platform | Architecture | Status |
|----------|-------------|--------|
| Windows  | x86_64      | ✅ Supported |
| macOS    | x86_64      | ✅ Supported |
| macOS    | aarch64     | ✅ Supported |
| Linux    | x86_64      | ✅ Supported |

### Branch Protection

The `main` branch is protected with the following requirements:

- **Required Status Checks**: All CI jobs must pass
- **Up-to-date branches**: PRs must be current with main
- **Require review**: At least one approved review required
- **Dismiss stale reviews**: Reviews dismissed when new commits pushed

### Release Process

1. **Version Bump**: Update version in `Cargo.toml`
2. **Changelog**: Update `CHANGELOG.md` with release notes
3. **Commit**: Commit version changes
4. **Tag**: Create annotated git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. **Push**: Push tag to trigger release: `git push origin v1.0.0`
6. **Monitor**: Watch GitHub Actions for successful build and release

### Issue and PR Templates

#### Bug Reports
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Rust version)
- Relevant logs or error messages

#### Feature Requests
- Clear description of the feature
- Use case and motivation
- Proposed implementation approach
- Backward compatibility considerations

### Code Style

- Follow Rust standard conventions
- Use descriptive variable and function names
- Add doc comments for public APIs
- Keep functions focused and small
- Prefer explicit error handling over panics

### Testing Guidelines

- Write unit tests for new functionality
- Add integration tests for CLI interactions
- Test error conditions and edge cases
- Ensure tests are deterministic and fast
- Use `tempfile` for file system tests

### Documentation

- Update README.md for user-facing changes
- Add doc comments for public APIs  
- Update CLI help text when adding commands
- Include examples in documentation

## Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Security Issues**: Follow our [Security Policy](.github/SECURITY.md)

## Code of Conduct

Please be respectful and constructive in all interactions. We want razd to be a welcoming project for contributors of all backgrounds and experience levels.

Thank you for contributing to razd! 🦀