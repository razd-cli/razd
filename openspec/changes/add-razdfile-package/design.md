# Design: Razdfile Package Architecture

## Overview

Пакет `razdfile` следует архитектурным паттернам из `go-task/task/taskfile`, адаптированным под Razdfile формат.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│                    (cmd/razd/main.go)                       │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      razdfile.Reader                         │
│  - NewReader(opts...)                                        │
│  - Read(ctx, node) → (*ast.Razdfile, error)                 │
│  - Functional options: WithDebugFunc, WithValidation        │
└─────────────────────────────┬───────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   FileNode      │ │   Validator     │ │   YAML Parser   │
│   - Location()  │ │   - Validate()  │ │   - Unmarshal   │
│   - Read()      │ │   - Schema      │ │   - Decode      │
└─────────────────┘ └─────────────────┘ └─────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       razdfile/ast                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  Razdfile    │  │  MiseConfig  │  │  Tasks       │       │
│  │  - Version   │  │  - Tools     │  │  (from       │       │
│  │  - Mise      │  │  - Env       │  │   go-task)   │       │
│  │  - Tasks     │  │  - Settings  │  │              │       │
│  │  - Vars      │  │  - Hooks     │  │              │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

## Package Dependencies

```
razdfile/
    ├── ast/           → go.yaml.in/yaml/v4
    │                  → github.com/go-task/task/v3/taskfile/ast (for Tasks, Vars)
    ├── reader.go      → razdfile/ast
    │                  → context, sync
    ├── node.go        → io
    ├── node_file.go   → os, path/filepath
    ├── validate.go    → razdfile/ast
    │                  → github.com/xeipuuv/gojsonschema (optional)
    └── errors.go      → fmt, errors
```

## AST Design

### Razdfile (root)

```go
package ast

type Razdfile struct {
    // Metadata
    Location string `yaml:"-"`         // file path, set by Reader
    
    // Required
    Version string `yaml:"version"`    // "1"
    
    // Optional: Mise configuration
    Mise *MiseConfig `yaml:"mise,omitempty"`
    
    // Taskfile-compatible fields (delegate to go-task types)
    Output   Output         `yaml:"output,omitempty"`
    Method   string         `yaml:"method,omitempty"`
    Includes *Includes      `yaml:"includes,omitempty"`
    Vars     *taskast.Vars  `yaml:"vars,omitempty"`
    Env      *taskast.Vars  `yaml:"env,omitempty"`
    Tasks    *taskast.Tasks `yaml:"tasks,omitempty"`
    Silent   bool           `yaml:"silent,omitempty"`
    Dotenv   []string       `yaml:"dotenv,omitempty"`
    Run      string         `yaml:"run,omitempty"`
    Interval string         `yaml:"interval,omitempty"`
    Set      []string       `yaml:"set,omitempty"`
    Shopt    []string       `yaml:"shopt,omitempty"`
}
```

### MiseConfig

```go
type MiseConfig struct {
    Tools      map[string]MiseTool   `yaml:"tools,omitempty"`
    Env        map[string]any        `yaml:"env,omitempty"`
    Settings   *MiseSettings         `yaml:"settings,omitempty"`
    Hooks      map[string]any        `yaml:"hooks,omitempty"`
    Plugins    map[string]string     `yaml:"plugins,omitempty"`
    MinVersion string                `yaml:"min_version,omitempty"`
}

type MiseTool struct {
    // Simple: "22" or complex object
    Version     string            `yaml:"version,omitempty"`
    OS          []string          `yaml:"os,omitempty"`
    InstallEnv  map[string]any    `yaml:"install_env,omitempty"`
    Postinstall string            `yaml:"postinstall,omitempty"`
}

type MiseSettings struct {
    AutoInstall  *bool `yaml:"auto_install,omitempty"`
    Experimental *bool `yaml:"experimental,omitempty"`
    Quiet        *bool `yaml:"quiet,omitempty"`
    Verbose      *bool `yaml:"verbose,omitempty"`
    Jobs         *int  `yaml:"jobs,omitempty"`
}
```

### Custom UnmarshalYAML

MiseTool поддерживает два формата:

```yaml
# Simple string
node: "22"

# Complex object
node:
  version: "22"
  os: [linux, darwin]
```

```go
func (t *MiseTool) UnmarshalYAML(node *yaml.Node) error {
    switch node.Kind {
    case yaml.ScalarNode:
        // Simple version string
        t.Version = node.Value
        return nil
    case yaml.MappingNode:
        // Complex object
        type toolAlias MiseTool
        return node.Decode((*toolAlias)(t))
    case yaml.SequenceNode:
        // Multiple versions
        var versions []string
        if err := node.Decode(&versions); err != nil {
            return err
        }
        t.Version = versions[0] // primary version
        return nil
    }
    return fmt.Errorf("invalid tool format at line %d", node.Line)
}
```

## Reader Design

### Functional Options Pattern

```go
type Reader struct {
    debugFunc    DebugFunc
    validateFunc ValidateFunc
}

type ReaderOption interface {
    ApplyToReader(*Reader)
}

func NewReader(opts ...ReaderOption) *Reader {
    r := &Reader{}
    for _, opt := range opts {
        opt.ApplyToReader(r)
    }
    return r
}

func WithDebugFunc(fn DebugFunc) ReaderOption { ... }
func WithValidation(enabled bool) ReaderOption { ... }
```

### Read Flow

```go
func (r *Reader) Read(ctx context.Context, node Node) (*ast.Razdfile, error) {
    // 1. Read raw bytes
    data, err := node.Read()
    if err != nil {
        return nil, err
    }
    
    // 2. Parse YAML
    var rf ast.Razdfile
    if err := yaml.Unmarshal(data, &rf); err != nil {
        return nil, &RazdfileDecodeError{...}
    }
    
    // 3. Set location
    rf.Location = node.Location()
    
    // 4. Validate
    if err := r.validate(&rf); err != nil {
        return nil, err
    }
    
    return &rf, nil
}
```

## Node Interface

```go
type Node interface {
    // Location returns the path/URI of the node
    Location() string
    
    // Read returns the raw content
    Read() ([]byte, error)
}

type FileNode struct {
    entrypoint string
    dir        string
}

func NewFileNode(entrypoint, dir string) (*FileNode, error) {
    // Resolve path, check exists
    ...
}

func (n *FileNode) Location() string { return n.entrypoint }
func (n *FileNode) Read() ([]byte, error) { return os.ReadFile(n.entrypoint) }
```

## Validation Strategy

### Level 1: YAML Syntax
- Handled by `yaml.Unmarshal`
- Produces `RazdfileDecodeError` with line/column

### Level 2: Schema Validation
- Version must be "1"
- At least one of: `tasks`, `includes`, or `mise` required
- Tools format validation

### Level 3: Semantic Validation
- Task references exist
- Include paths resolvable
- Tool versions valid format

```go
func (r *Reader) validate(rf *ast.Razdfile) error {
    // Version check
    if rf.Version != "1" {
        return &VersionError{Got: rf.Version, Expected: "1"}
    }
    
    // Content check
    hasTasks := rf.Tasks != nil && rf.Tasks.Len() > 0
    hasIncludes := rf.Includes != nil && rf.Includes.Len() > 0
    hasMise := rf.Mise != nil
    
    if !hasTasks && !hasIncludes && !hasMise {
        return &EmptyRazdfileError{Location: rf.Location}
    }
    
    return nil
}
```

## Error Types

```go
// RazdfileDecodeError - YAML parsing error with location
type RazdfileDecodeError struct {
    Message  string
    Location string
    Line     int
    Column   int
    Err      error
}

// RazdfileVersionError - unsupported version
type RazdfileVersionError struct {
    Location string
    Got      string
    Expected string
}

// RazdfileNotFoundError - file not found
type RazdfileNotFoundError struct {
    Path string
}
```

## Integration with go-task

### Option A: Direct Reuse (Recommended)
Reuse `taskfile/ast` types directly:

```go
import taskast "github.com/go-task/task/v3/taskfile/ast"

type Razdfile struct {
    Tasks *taskast.Tasks
    Vars  *taskast.Vars
    ...
}
```

**Pros**: No duplication, compatible with Task executor
**Cons**: Coupled to go-task API changes

### Option B: Wrap Types
Create thin wrappers:

```go
type Tasks struct {
    inner *taskast.Tasks
}

func (t *Tasks) ToTaskAST() *taskast.Tasks {
    return t.inner
}
```

**Pros**: Isolation from go-task changes
**Cons**: More code, conversion overhead

### Decision: Start with Option A

We'll use direct reuse for v1. If go-task makes breaking changes, we can wrap later.

## File Detection

Similar to go-task's `DefaultTaskfiles`:

```go
var DefaultRazdfiles = []string{
    "Razdfile.yml",
    "Razdfile.yaml",
    "razdfile.yml",
    "razdfile.yaml",
}
```

Search order:
1. Exact path if specified
2. Search in current directory for DefaultRazdfiles
3. Walk up to find Razdfile (optional, for monorepos)

## Future Extensibility

### Remote Nodes
```go
type HTTPNode struct { ... }
type GitNode struct { ... }
```

### Devbox Support
```go
type Razdfile struct {
    Mise   *MiseConfig   `yaml:"mise,omitempty"`
    Devbox *DevboxConfig `yaml:"devbox,omitempty"` // future
}
```

### Includes Processing
Similar to go-task, build a graph of included files:
```go
type RazdfileGraph struct {
    graph.Graph[string, *RazdfileVertex]
}
```
