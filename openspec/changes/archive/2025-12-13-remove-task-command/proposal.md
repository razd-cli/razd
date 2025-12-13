# Remove `razd task` Command

## Summary
Remove the `razd task` command from the CLI as it is redundant with `razd run`, which provides the same task execution functionality through Razdfile.yml. This simplification will reduce confusion for users and streamline the command interface.

## Motivation

### Current State
The CLI currently has two overlapping commands for executing tasks:
- **`razd task <name>`**: Executes tasks directly from Taskfile.yml
- **`razd run <name>`**: Executes custom tasks defined in Razdfile.yml, which internally uses Taskfile.yml

### Problems with Current Design
1. **User Confusion**: Having two similar commands (`task` vs `run`) creates ambiguity about which to use
2. **Documentation Overhead**: Must explain and maintain documentation for both commands
3. **Maintenance Burden**: Two code paths (src/commands/task.rs and src/commands/run.rs) for similar functionality
4. **Inconsistent Mental Model**: Users must understand the subtle difference between direct Taskfile.yml execution and Razdfile.yml task execution

### Why `razd run` is Preferred
- **More Intuitive**: "run" is a clearer verb for executing tasks
- **Razdfile-Centric**: Aligns with razd's philosophy of providing a unified configuration through Razdfile.yml
- **Better Abstraction**: Allows razd to add features and orchestration on top of raw task execution
- **Industry Patterns**: Matches conventions from tools like `npm run`, `cargo run`, `go run`

## Proposed Changes

### Remove `razd task` Command
- Remove `Commands::Task` variant from CLI enum in src/main.rs
- Delete src/commands/task.rs module
- Update documentation and examples to use `razd run` instead
- Update error messages that reference `razd task`

### Update Documentation
- Replace all references to `razd task` with `razd run` in:
  - README.md
  - examples/nodejs-project/README.md
  - Error messages in src/core/error.rs
  - Help text in src/commands/up.rs

## Impact Assessment

### User Impact
- **Breaking Change**: Users currently using `razd task` will need to switch to `razd run`
- **Migration Path**: Simple substitution: `razd task <name>` → `razd run <name>`
- **Documentation**: Clear migration guide in CHANGELOG.md for version 0.4.1

### Code Impact
- **Deletions**: ~20 lines of code removed from src/commands/task.rs
- **Modifications**: 4-5 files need updates for references
- **Tests**: No test files directly reference Commands::Task (verified via grep)

### Benefits
- Simpler CLI surface area
- Reduced cognitive load for new users
- Single clear path for task execution
- Easier to document and maintain

## Alternatives Considered

### Keep Both Commands
- **Rejected**: Maintains confusion and maintenance overhead
- Does not provide enough differentiated value

### Remove `razd run` Instead
- **Rejected**: "run" is more intuitive and aligns with industry conventions
- Razdfile.yml-centric approach is core to razd's value proposition

### Alias `task` to `run`
- **Rejected**: Still requires maintaining both command names in documentation
- Deprecation period adds complexity without sufficient value

## Release Plan

### Version 0.4.1
- Remove `razd task` command
- Update all documentation
- Add migration note to CHANGELOG.md

### Rollout Strategy
- **No deprecation period**: The tool is early stage (0.4.x), breaking changes are acceptable
- **Clear communication**: CHANGELOG and README will guide users through the one-line change

## Success Criteria
- [ ] `razd task` command removed from CLI
- [ ] All documentation updated to use `razd run`
- [ ] All error messages and help text updated
- [ ] Examples updated and tested
- [ ] CHANGELOG.md updated with migration notes
- [ ] Version bumped to 0.4.1 in Cargo.toml

## Related Specifications
- cli-interface: Primary CLI commands and interface design

## Timeline
- **Proposal**: Day 1
- **Implementation**: Day 1-2
- **Release**: Day 2

## Dependencies
None. This is an isolated change that does not depend on or block other changes.
