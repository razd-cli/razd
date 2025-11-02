# Contributing to razd

Thank you for your interest in contributing to razd! This document provides guidelines for contributing to the project.

## Development Workflow

### Prerequisites

- **Rust** 1.74.0 or later (check `rust-version` in `Cargo.toml`)
- **Git** for version control
- **GitHub account** for pull requests
- **razd** CLI tool (or use cargo directly)

### Setting Up Development Environment

1. **Fork and clone the repository:**
   ```sh
   git clone https://github.com/your-username/razd.git
   cd razd
   ```

2. **Option A: Using razd (dogfooding approach - recommended):**
   ```sh
   # If razd is not installed, install it first
   cargo install --path .
   
   # Setup project (installs Rust toolchain via mise, fetches dependencies)
   razd setup
   
   # Or run default workflow (build + test)
   razd
   ```

3. **Option B: Using cargo directly:**
   ```sh
   # Install Rust toolchain
   rustup update
   rustup component add rustfmt clippy
   
   # Build and test
   cargo build
   cargo test
   ```

### Available Commands

razd provides several useful commands for development (defined in `Razdfile.yml`):

```sh
razd                    # Default: build + test
razd run setup          # Setup dependencies
razd run build          # Build debug version
razd run build-release  # Build release version
razd run test           # Run all tests
razd run test-integration # Run integration tests only
razd run fmt            # Format code
razd run fmt-check      # Check formatting
razd run lint           # Run clippy
razd run audit          # Security audit
razd run ci             # Run all CI checks locally
razd run clean          # Clean build artifacts
razd run dev -- <args>  # Run razd in dev mode
razd run coverage       # Generate coverage report
razd run doc            # Generate and open docs
razd run version        # Show current version
razd run release-check  # Pre-release checks
```

### Code Quality Standards

Our CI pipeline enforces the following quality standards. You can run these checks locally using razd:

#### Quick CI Check
```sh
razd run ci  # Runs all checks: format, lint, test, and audit
```

#### Individual Checks

**Formatting:**
```sh
razd run fmt-check  # Check formatting
razd run fmt        # Auto-format code
```
All code must be formatted with `rustfmt` using default settings.

**Linting:**
```sh
razd run lint  # Run clippy
```
All clippy warnings must be resolved before merging.

**Testing:**
```sh
razd run test              # All tests
razd run test-integration  # Integration tests only
```
All tests must pass on all supported platforms (Windows, macOS, Linux).

**Security Audit:**
```sh
razd run audit  # Check dependencies
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
   razd run ci  # Run all checks
   ```
   
   Or manually:
   ```sh
   razd run test
   razd run fmt-check
   razd run lint
   razd run audit
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
- **Rust Versions**: Tests on stable Rust (MSRV: 1.74.0)
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
3. **Pre-release checks**:
   ```sh
   razd run release-check  # Runs CI + release build
   ```
4. **Commit**: Commit version changes
5. **Tag**: Create annotated git tag: 
   ```sh
   git tag -a v1.0.0 -m "Release v1.0.0"
   ```
6. **Push**: Push tag to trigger release: 
   ```sh
   git push origin v1.0.0
   ```
7. **Monitor**: Watch GitHub Actions for successful build and release

### Local Development Tips

**Building and Testing:**
```sh
razd run build          # Quick debug build
razd run build-release  # Optimized build
razd run test           # Run all tests
```

**Running razd in Development:**
```sh
razd run dev -- up https://github.com/user/repo  # Test 'razd up' command
razd run dev -- --help                            # Test help output
razd run dev -- run build                         # Test 'razd run' command
```

**Installing Locally:**
```sh
razd run install-local  # Install to ~/.cargo/bin/
```

**Generating Documentation:**
```sh
razd run doc  # Opens docs in browser
```

**Coverage Reports:**
```sh
razd run coverage  # Requires cargo-tarpaulin
```

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