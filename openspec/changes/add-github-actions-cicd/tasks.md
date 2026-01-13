# Tasks: GitHub Actions CI/CD Implementation

## Phase 1: Lint Workflow

- [x] **1.1 Create lint workflow**
  - Create `.github/workflows/lint.yml`
  - Configure trigger on push (main, tags) and pull_request
  - Set up Go with matrix (1.24.x, 1.25.x)
  - Run golangci-lint with latest version
  - Test workflow locally with `act` or push

## Phase 2: Test Workflow

- [x] **2.1 Create test workflow**
  - Create `.github/workflows/test.yml`
  - Configure trigger on push (main, tags) and pull_request
  - Set up matrix: go-version × platform
  - Download Go modules
  - Build binary
  - Run tests

## Phase 3: GoReleaser Setup

- [x] **3.1 Create GoReleaser config**
  - Create `.goreleaser.yml` in root
  - Configure build targets (linux, darwin, windows × amd64, arm64)
  - Set ldflags for version injection
  - Configure archive formats (tar.gz for unix, zip for windows)
  - Configure changelog generation

- [x] **3.2 Verify version injection**
  - Ensure `internal/version/version.go` supports ldflags
  - Test local build with ldflags

## Phase 4: Release Workflow

- [x] **4.1 Create release workflow**
  - Create `.github/workflows/release.yml`
  - Configure trigger on tag `v*`
  - Set up Go environment
  - Run GoReleaser with `release --clean`
  - Use `GITHUB_TOKEN` for publishing

- [ ] **4.2 Test release workflow**
  - Create test tag (e.g., v0.0.1-test)
  - Verify artifacts on GitHub Releases
  - Delete test release if needed

## Phase 5: Documentation

- [ ] **5.1 Update README**
  - Add badges for CI status
  - Add installation instructions (download from releases)
  - Document release process

## Validation Criteria

- [ ] Lint workflow passes on clean code
- [ ] Test workflow passes on all platforms (ubuntu, macos, windows)
- [ ] GoReleaser produces binaries for all 6 targets
- [ ] Release creates GitHub Release with all artifacts
- [ ] Version info correctly embedded in binary
