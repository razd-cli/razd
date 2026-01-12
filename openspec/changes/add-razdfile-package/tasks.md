# Tasks: Add Razdfile Package

Implementation checklist for `razdfile` package.

## Phase 1: Core AST Types

- [x] **1.1** Create `razdfile/ast/version.go` with version constants and validation
- [x] **1.2** Create `razdfile/ast/mise.go` with MiseConfig, MiseTool, MiseSettings types
- [x] **1.3** Create `razdfile/ast/mise_test.go` with UnmarshalYAML tests for MiseTool
- [x] **1.4** Create `razdfile/ast/razdfile.go` with main Razdfile struct
- [x] **1.5** Create `razdfile/ast/razdfile_test.go` with parsing tests

## Phase 2: Node Abstraction

- [x] **2.1** Create `razdfile/node.go` with Node interface
- [x] **2.2** Create `razdfile/node_file.go` with FileNode implementation
- [x] **2.3** Create `razdfile/node_file_test.go` with file reading tests (covered in reader_test.go)
- [x] **2.4** Add `DefaultRazdfiles` constant for file detection

## Phase 3: Reader

- [x] **3.1** Create `razdfile/reader.go` with Reader struct and functional options
- [x] **3.2** Implement `Read(ctx, node)` method
- [x] **3.3** Add `WithDebugFunc` option
- [x] **3.4** Create `razdfile/reader_test.go` with integration tests

## Phase 4: Errors

- [x] **4.1** Create `razdfile/errors.go` with RazdfileDecodeError, RazdfileNotFoundError
- [x] **4.2** Add line/column information to decode errors
- [x] **4.3** Implement Error() and Code() methods for all error types

## Phase 5: Validation

- [x] **5.1** Create `razdfile/validate.go` with validation logic
- [x] **5.2** Implement version validation ("1" required)
- [x] **5.3** Implement content validation (tasks/includes/mise required)
- [x] **5.4** Create `razdfile/validate_test.go` with validation test cases (covered in reader_test.go)

## Phase 6: Integration

- [x] **6.1** Update `main.go` to use razdfile.Reader for parsing
- [x] **6.2** Parse Razdfile.yml from examples/ and extract tasks
- [x] **6.3** Pass parsed tasks to go-task Executor
- [x] **6.4** Verify examples/nodejs-project works end-to-end

## Verification

- [x] **V1** All tests pass: `go test ./razdfile/...`
- [x] **V2** Examples parse without errors
- [x] **V3** Mise section correctly parsed into structs
- [x] **V4** go-task execution works with parsed tasks

## Dependencies

- Phase 2 depends on Phase 1 (AST types needed for Node)
- Phase 3 depends on Phase 1, 2 (Reader uses AST and Node)
- Phase 4 can be done in parallel with Phase 2, 3
- Phase 5 depends on Phase 1, 3 (Validation uses AST and Reader)
- Phase 6 depends on all previous phases

## Parallelizable Work

Phases 1-4 can be developed in parallel by different developers:
- Dev A: Phase 1 (AST) + Phase 5 (Validation)
- Dev B: Phase 2 (Node) + Phase 3 (Reader)
- Dev C: Phase 4 (Errors)
