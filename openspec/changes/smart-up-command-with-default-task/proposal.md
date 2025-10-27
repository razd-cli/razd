# Smart Up Command with Default Task

## Change ID
`smart-up-command-with-default-task`

## Type
Breaking Change / Enhancement

## Status
Proposed

## Why

Current command structure creates confusion and friction for users:

1. **Unclear Command Separation**: `razd init` vs `razd up` creates cognitive load - users must remember which command to use when
2. **Non-intuitive Task Naming**: `tasks.up` doesn't clearly communicate that this is the "main" project task
3. **Context Blindness**: `razd up` behavior doesn't adapt to project state (existing vs new projects)
4. **Multiple Steps for Simple Goals**: Users often want "just bring up my project" but need to know multiple commands

Users expect `razd up` to mean "bring up the project" regardless of context, similar to how `docker-compose up` or `vagrant up` work.

## What Changes

Transform `razd up` into a smart, context-aware command that handles all project "bring up" scenarios:

### 1. Smart Up Command Behavior
- **With Razdfile.yml**: Execute `tasks.default` workflow 
- **Without Razdfile.yml**: Offer interactive configuration creation
- **With URL argument**: Clone repository and setup project
- **No arguments anywhere**: Default to local project setup

### 2. Configuration Format Update
- Change `tasks.up` → `tasks.default` in Razdfile.yml
- More intuitive naming that aligns with common conventions
- Better semantic meaning for "primary project task"

### 3. Command Structure Simplification  
- Remove `razd init` command (functionality merged into `razd up`)
- Reduce command surface area
- Single command for all "bring up" scenarios

### 4. Migration Strategy
- Backward compatibility period supporting both `tasks.up` and `tasks.default`
- Deprecation warnings for old format
- Migration helper functionality

## Impact

### Breaking Changes
1. **Configuration Format**: `tasks.up` → `tasks.default` requires file updates
2. **Command Removal**: `razd init` will be removed entirely
3. **Behavior Change**: `razd up` without arguments changes from error to smart behavior

### Affected Components
- Commands: `up`, `init` (removal)
- Configuration: Razdfile.yml format
- Documentation: All examples and guides
- User workflows: Existing automation and scripts

### Migration Path
1. **Phase 1**: Add `tasks.default` support with backward compatibility
2. **Phase 2**: Deprecation warnings for `tasks.up` and `razd init`
3. **Phase 3**: Remove deprecated functionality

### User Benefits
- **Simplified Mental Model**: One command for all "bring up" scenarios
- **Intuitive Behavior**: Command adapts to context automatically  
- **Reduced Learning Curve**: Fewer commands to remember
- **Better UX**: Less typing, fewer decisions

### Risks
- **Breaking Changes**: Existing users need to update configurations
- **Migration Effort**: Projects and automation need updates
- **Learning Curve**: Existing users need to adapt to new patterns

## Alternatives Considered

1. **Keep Current Structure**: Rejected - doesn't address UX issues
2. **Add New Command**: Rejected - increases complexity instead of reducing it
3. **Only Change Configuration**: Rejected - doesn't solve command confusion
4. **Gradual Migration Without Breaking Changes**: Rejected - leads to permanent complexity

## Dependencies

- No external dependencies
- Builds on existing command infrastructure
- Compatible with current Taskfile and mise integrations

## Success Criteria

- [ ] `razd up` handles all three contexts seamlessly
- [ ] Migration path is smooth and well-documented
- [ ] No regression in existing URL-based cloning functionality
- [ ] Interactive configuration creation is user-friendly
- [ ] Performance impact is minimal

## Approval

- [ ] Reviewed by maintainer
- [ ] Breaking change impact assessed and accepted
- [ ] Migration strategy approved
- [ ] Ready for implementation