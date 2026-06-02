# Plan: `razd up <url>` — Clone and Provision from Git URL

**Branch:** 1.x (current)
**Created:** 2026-06-02
**Settings:** Testing: yes | Logging: verbose | Docs: no

## Problem

`razd up` currently only works in the current directory. Users want `razd up <url>` to clone a remote repository and provision it in one command, similar to `git clone <url>` followed by `cd <repo> && razd up`.

Behavior must mirror `git clone`:
- Clone into the current directory
- If target directory already exists → error from git
- Git errors (auth, network, etc.) are surfaced to the user as-is

## Architecture Decision

Add a `internal/git` package for git-related utilities (URL detection, git availability check, clone logic). Modify `runUp()` to check `ctx.Args` for a URL-like argument before proceeding with the normal flow.

URL detection patterns:
- `https://github.com/owner/repo`
- `https://github.com/owner/repo.git`
- `git@github.com:owner/repo.git`
- `ssh://git@github.com/owner/repo.git`
- `git://github.com/owner/repo.git`

## Tasks

- [x] T1: Create `internal/git/git.go` — git utility package
  - New file: `internal/git/git.go`
  - `IsGitAvailable() bool` — check if `git` binary is in PATH using `exec.LookPath("git")`
  - `IsURL(arg string) bool` — detect if arg looks like a git URL (https://, git@, ssh://, git://, or ends with .git)
  - `ExtractRepoName(url string) (string, error)` — extract repository directory name from URL (strip trailing `.git`, take last path segment)
  - `Clone(url string, dir string) error` — run `git clone <url>` in the given directory, streaming stdout/stderr to the caller
  - Logging: `DEBUG [git] IsGitAvailable: result=<true|false>`, `DEBUG [git] Clone: url=<url>, dir=<dir>`

- [x] T2: Add error types to `internal/errors/errors.go`
  - Add `CodeGitNotInstalled = 110` exit code constant
  - Add `CodeCloneFailed = 111` exit code constant
  - Add `GitNotInstalledError` struct implementing `RazdError` interface with `Code() int` returning `CodeGitNotInstalled`
  - Add `CloneError` struct with `URL string`, `Err error`, `Code() int` returning `CodeCloneFailed`
  - Add cases in `handleError()` in `internal/cli/cli.go` for both new error types with appropriate log messages

- [x] T3: Modify `internal/cli/cmd_up.go` — add URL argument handling
  - At the start of `runUp(ctx)`, check `ctx.Args` for a URL-like argument
  - If `len(ctx.Args) > 0` and `git.IsURL(ctx.Args[0])`:
    1. Check `git.IsGitAvailable()` — if not, return `&errors.GitNotInstalledError{}`
    2. Determine clone directory: `flags.Dir` if set, otherwise `os.Getwd()`
    3. Call `git.Clone(ctx.Args[0], cloneDir)` — if error, return `&errors.CloneError{URL: ctx.Args[0], Err: err}`
    4. Derive repo directory name: `git.ExtractRepoName(ctx.Args[0])`
    5. Set `dir = filepath.Join(cloneDir, repoName)` for the rest of the flow
    6. Log: `INFO Cloning <url>...`, `SUCCESS Cloned to <dir>`
    7. Continue with normal `runUp` flow using the cloned directory
  - If no URL argument, proceed with existing behavior (resolve dir, read Razdfile, etc.)
  - If `len(ctx.Args) > 1`, return error: "too many arguments for 'up' command"
  - If `len(ctx.Args) == 1` but it's not a URL, return error: "unexpected argument %q for 'up' command"
  - Logging: `DEBUG [up] args=%v, url_detected=<true|false>`

- [x] T4: Write tests for `internal/git/git.go`
  - New file: `internal/git/git_test.go`
  - `TestIsURL` — table-driven tests for all URL patterns:
    - `https://github.com/razd-cli/razd-nodejs-example` → true
    - `https://github.com/razd-cli/razd-nodejs-example.git` → true
    - `git@github.com:razd-cli/razd-nodejs-example.git` → true
    - `ssh://git@github.com/razd-cli/razd-nodejs-example.git` → true
    - `git://github.com/razd-cli/razd-nodejs-example.git` → true
    - `dev` → false (task name)
    - `build` → false
    - `./local/path` → false
    - empty string → false
  - `TestExtractRepoName` — table-driven tests:
    - `https://github.com/razd-cli/razd-nodejs-example` → `razd-nodejs-example`
    - `https://github.com/razd-cli/razd-nodejs-example.git` → `razd-nodejs-example`
    - `git@github.com:razd-cli/razd-nodejs-example.git` → `razd-nodejs-example`
  - `TestIsGitAvailable` — basic test (will pass on systems with git installed)

- [x] T5: Write tests for `internal/cli/cmd_up.go` URL handling
  - New file or extend: `internal/cli/cmd_up_test.go`
  - Test `runUp` with a URL argument when git is not available — should return `GitNotInstalledError`
  - Test `runUp` with too many arguments — should return error
  - Test `runUp` with a non-URL argument — should return error
  - Test `runUp` with no arguments — should follow existing behavior
  - Mock `git.Clone` and `git.IsGitAvailable` for unit testing without actual git

- [x] T6: Build and test
  - `go build ./...`
  - `go test ./...`
  - Verify all existing tests still pass

## Commit Plan

- Commit 1 (T1+T2): `feat(git): add git utility package and error types`
- Commit 2 (T3): `feat(up): support URL argument for cloning and provisioning`
- Commit 3 (T4+T5): `test(git,cli): add tests for URL detection and clone flow`
- Commit 4 (T6): build and verification

## Edge Cases

- **Git not installed:** Clear error message: "git is not installed. Please install git to clone repositories."
- **Private repository without access:** Git clone fails with auth error — surface git's error message directly to user
- **Target directory already exists:** Git clone will fail with `fatal: destination path already exists` — surface this as-is
- **Network errors:** Git clone will fail — surface the error from git
- **URL with trailing `.git`:** Handled by `ExtractRepoName` stripping `.git` suffix
- **SSH URL format (`git@host:user/repo.git`):** Recognized by `IsURL`
- **No Razdfile in cloned repo:** Normal error flow — `NoRazdfileError` after cloning succeeds