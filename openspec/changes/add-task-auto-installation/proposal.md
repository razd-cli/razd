# Add Task Auto-Installation via Mise

## Summary

Enhance `razd up` to automatically install the `task` tool via mise (`mise install task@latest`) when needed, ensuring razd can reliably use task for workflow operations without requiring manual tool installation.

## Why

Currently, `razd up` requires the `task` tool to be manually installed by users before it can execute taskfile operations. This creates a dependency management burden and potential failure points:

1. **Manual dependency management**: Users must remember to install `task` separately
2. **Error-prone setup**: `razd up` fails with unhelpful errors if `task` is not available
3. **Inconsistent environments**: Different versions of `task` across team members
4. **Setup friction**: Additional step in project onboarding process

## Problem Statement

The current workflow forces users to manually manage the `task` tool dependency, breaking the "one-command setup" promise of `razd up` and creating friction in project onboarding.

## What Changes

Modify `razd up` to automatically ensure `task` is available by:

1. Check if `task` is already installed and accessible
2. If not available, and `mise` is present, automatically run `mise install task@latest`
3. Verify `task` installation before proceeding with taskfile operations
4. Provide clear feedback about tool installation progress

## Proposed Solution

The enhancement will integrate automatic tool installation into the existing `razd up` workflow seamlessly.

### Key Benefits

- **Zero-friction setup**: `razd up` becomes truly one-command project initialization
- **Consistent environments**: All team members get same version of `task` via mise
- **Reliable execution**: No more failures due to missing `task` dependency
- **Better UX**: Clear feedback when tools are being installed automatically

## Scope

### In Scope
- Auto-installation of `task` via mise during `razd up` execution
- Detection and validation of existing `task` installations
- Enhanced error handling for tool installation failures
- User feedback during tool installation process

### Out of Scope
- Auto-installation of other tools beyond `task`
- Alternative installation methods beyond mise
- Complex version management or tool conflicts resolution
- Installation without mise (manual tool installation remains user responsibility)

## Implementation Approach

1. **Enhanced tool checking**: Extend `taskfile.rs` to check for `task` availability and trigger installation
2. **Mise integration**: Add function to install specific tools via mise
3. **Workflow integration**: Integrate tool installation into the `razd up` workflow before taskfile operations
4. **Error handling**: Provide clear messages for installation success/failure

## Risk Assessment

### Low Risk
- **Backward compatibility**: No changes to existing successful workflows
- **Tool availability**: mise + task integration is well-established

### Medium Risk
- **Installation failures**: Network issues or mise configuration problems could cause failures
  - **Mitigation**: Clear error messages and fallback guidance
- **Permission issues**: Tool installation might require different permissions
  - **Mitigation**: Rely on mise's permission handling, provide clear error messages

### Dependencies
- Requires `mise` to be available and functional
- Depends on `task` package availability in mise registry
- No breaking changes to existing APIs

## Success Criteria

1. **Automated installation**: `razd up` successfully installs `task` when missing
2. **Tool verification**: Confirms `task` is functional after installation
3. **User experience**: Clear progress indication during tool installation
4. **Error handling**: Helpful error messages when installation fails
5. **Performance**: Minimal overhead when `task` is already installed
6. **Cross-platform**: Works on both Windows and Unix systems