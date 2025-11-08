## Tasks

### Phase 1: Analysis and Design
- [ ] **T1.1** Analyze current YAML processing in `sync_mise_to_razdfile()` method
- [ ] **T1.2** Research YAML preservation techniques for maintaining original formatting
- [ ] **T1.3** Design selective update strategy for `mise:` section only
- [ ] **T1.4** Create test cases covering platform-specific commands and formatting preservation

### Phase 2: Core Implementation  
- [ ] **T2.1** Implement YAML document parsing that preserves structure and comments
- [ ] **T2.2** Create selective mise section updater that leaves other sections untouched
- [ ] **T2.3** Update `sync_mise_to_razdfile()` to use surgical update approach
- [ ] **T2.4** Ensure backup creation still works with new approach

### Phase 3: Testing and Validation
- [ ] **T3.1** Add unit tests for platform command preservation during sync
- [ ] **T3.2** Add integration tests for formatting preservation 
- [ ] **T3.3** Test edge cases: empty mise section, missing sections, comments
- [ ] **T3.4** Validate that Razdfile-to-mise sync still works correctly

### Phase 4: Documentation and Rollout
- [ ] **T4.1** Update sync behavior documentation
- [ ] **T4.2** Add examples showing preserved platform commands
- [ ] **T4.3** Update error messages to reflect new selective behavior
- [ ] **T4.4** Test on real projects with complex Razdfiles

## Dependencies
- **T2.1** depends on **T1.2** (YAML preservation research)
- **T2.3** depends on **T2.1** and **T2.2** (core implementation pieces)
- **T3.x** depends on **T2.x** (implementation must be complete)
- **T4.x** can be parallelized after **T3.x** (documentation after validation)

## Validation Criteria
- Platform-specific commands (`platform: windows`) are preserved exactly
- YAML formatting, comments, and field order remain unchanged
- Only `mise:` section content is modified during mise.toml sync
- All existing tests continue to pass
- New tests verify preservation behavior