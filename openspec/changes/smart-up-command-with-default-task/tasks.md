# Tasks: Smart Up Command with Default Task

## Implementation Tasks

### Phase 1: Backward Compatibility Foundation
- [ ] 1.1 Add support for `tasks.default` in Razdfile.yml parsing
- [ ] 1.2 Implement priority logic: `tasks.default` > `tasks.up`
- [ ] 1.3 Add deprecation warning when only `tasks.up` is found
- [ ] 1.4 Update Razdfile.yml template generation to use `tasks.default`
- [ ] 1.5 Write unit tests for dual format support

### Phase 2: Smart Up Command Logic
- [ ] 2.1 Implement directory context detection in `up` command
- [ ] 2.2 Add interactive Razdfile.yml creation workflow
- [ ] 2.3 Create project type selection and template system
- [ ] 2.4 Implement smart configuration generation based on detected tools
- [ ] 2.5 Add user prompts for common project configurations
- [ ] 2.6 Write integration tests for context-aware behavior

### Phase 3: Migration Support
- [ ] 3.1 Create `razd migrate` subcommand for configuration updates
- [ ] 3.2 Add automatic detection of `tasks.up` in existing files
- [ ] 3.3 Implement safe configuration file updates with backup
- [ ] 3.4 Add validation for migration results
- [ ] 3.5 Create migration progress feedback and error handling

### Phase 4: Command Deprecation
- [ ] 4.1 Mark `razd init` as deprecated with helpful redirection message
- [ ] 4.2 Update help text to guide users toward `razd up`
- [ ] 4.3 Add clear migration guidance in deprecation warnings
- [ ] 4.4 Update CLI argument parsing to show deprecation notices
- [ ] 4.5 Test deprecation message clarity and helpfulness

### Phase 5: Enhanced User Experience
- [ ] 5.1 Improve error messages for missing project detection
- [ ] 5.2 Add progress indicators for long-running operations
- [ ] 5.3 Implement colorized output for better readability
- [ ] 5.4 Add confirmation prompts for destructive operations
- [ ] 5.5 Create helpful hints and suggestions in interactive mode

### Phase 6: Documentation and Examples
- [ ] 6.1 Update README.md with new command patterns and examples
- [ ] 6.2 Create migration guide for existing users
- [ ] 6.3 Update all example Razdfile.yml files to use `tasks.default`
- [ ] 6.4 Add troubleshooting section for common migration issues
- [ ] 6.5 Create video/gif demonstrations of new workflows

### Phase 7: Testing and Validation
- [ ] 7.1 Write comprehensive unit tests for smart up logic
- [ ] 7.2 Create integration tests for all context scenarios
- [ ] 7.3 Add tests for backward compatibility during transition
- [ ] 7.4 Test URL cloning with new configuration format
- [ ] 7.5 Validate cross-platform behavior (Windows, macOS, Linux)
- [ ] 7.6 Performance testing for directory scanning and context detection

### Phase 8: Breaking Changes (Final Phase)
- [ ] 8.1 Remove `razd init` command completely
- [ ] 8.2 Remove support for `tasks.up` (keep only `tasks.default`)
- [ ] 8.3 Update default templates to only use new format
- [ ] 8.4 Clean up deprecated code paths and unused functions
- [ ] 8.5 Final validation that all functionality works correctly

## Validation Criteria

Each task must meet these criteria before being marked complete:

- **Functionality**: Feature works as specified in all supported scenarios
- **Testing**: Comprehensive unit and integration tests with >80% coverage
- **Documentation**: Clear documentation with examples and troubleshooting
- **User Experience**: Intuitive behavior with helpful error messages
- **Performance**: No significant performance regression
- **Compatibility**: Smooth transition path for existing users

## Dependencies

- **Phase 2** depends on **Phase 1** (backward compatibility foundation)
- **Phase 3** depends on **Phase 1** (configuration format support)
- **Phase 4** can run parallel with **Phase 2-3**
- **Phase 5-7** depend on **Phase 2** (core functionality)
- **Phase 8** depends on **Phase 6** (documentation complete)

## Risk Mitigation

- **Breaking Changes**: Implemented through gradual migration with warnings
- **User Confusion**: Clear documentation and helpful error messages
- **Migration Failures**: Backup creation and validation before changes
- **Performance Impact**: Minimal directory scanning with caching where appropriate

## Estimated Timeline

- **Phase 1-2**: Core implementation (3-4 days)
- **Phase 3**: Migration support (2 days)
- **Phase 4**: Deprecation handling (1 day)  
- **Phase 5**: UX improvements (2 days)
- **Phase 6**: Documentation (2 days)
- **Phase 7**: Testing and validation (2 days)
- **Phase 8**: Breaking changes (1 day)

**Total estimated effort**: 13-15 days