package provisioner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/razd-cli/razd/razdfile/ast"
)

// MiseProvisioner implements Provisioner for mise (https://mise.jdx.dev).
type MiseProvisioner struct {
	BaseProvisioner
}

// NewMiseProvisioner creates a new MiseProvisioner with the given config.
func NewMiseProvisioner(cfg Config) *MiseProvisioner {
	return &MiseProvisioner{
		BaseProvisioner: BaseProvisioner{Config: cfg},
	}
}

func (m *MiseProvisioner) Name() string {
	return "mise"
}

func (m *MiseProvisioner) GenerateConfig(packages []ast.ParsedDependency, extra map[string]any) error {
	tools := make(map[string]any, len(packages))
	for _, pkg := range packages {
		tools[pkg.Tool] = pkg.Version
	}
	return m.writeConfig(tools, extra)
}

func (m *MiseProvisioner) WriteTools(tools map[string]string) error {
	// Merge into existing native tools; extra (env/settings/etc.) is untouched.
	return m.writeConfig(toStringAny(tools), nil)
}

// writeConfig merges the given tools and extra sections into the existing
// mise.toml, preserving any section or tool not declared in the inputs.
func (m *MiseProvisioner) writeConfig(tools map[string]any, extra map[string]any) error {
	misePath := filepath.Join(m.Config.Dir, "mise.toml")

	// Parse the existing file so unknown sections survive the merge.
	existing := make(map[string]any)
	if data, err := os.ReadFile(misePath); err == nil {
		if err := toml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("failed to parse mise.toml: %w", err)
		}
	}

	// Merge extra sections (e.g. env, settings) into the root if provided.
	if len(extra) > 0 {
		for k, v := range extra {
			existing[k] = v
		}
	}

	// Merge tools into the [tools] table without dropping existing tools.
	if len(tools) > 0 {
		toolsTable, _ := existing["tools"].(map[string]any)
		if toolsTable == nil {
			toolsTable = make(map[string]any)
		}
		for k, v := range tools {
			// mise rejects an empty version (uv = ''); a bare/versionless tool
			// means "latest" (same as devbox). Write "latest" so the package
			// is actually installable.
			if s, ok := v.(string); ok && s == "" {
				toolsTable[k] = "latest"
				continue
			}
			toolsTable[k] = v
		}
		existing["tools"] = toolsTable
	}

	content, err := toml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("failed to encode mise.toml: %w", err)
	}

	if existingFile, err := os.ReadFile(misePath); err == nil && string(existingFile) == string(content) {
		return nil
	}

	if m.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] Writing mise.toml to %s\n", misePath)
	}

	return os.WriteFile(misePath, content, 0644)
}

func (m *MiseProvisioner) ReadConfig() (map[string]string, error) {
	misePath := filepath.Join(m.Config.Dir, "mise.toml")

	data, err := os.ReadFile(misePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read mise.toml: %w", err)
	}

	var root map[string]any
	if err := toml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse mise.toml: %w", err)
	}

	toolsTable, _ := root["tools"].(map[string]any)
	if toolsTable == nil {
		return nil, nil
	}

	tools := make(map[string]string, len(toolsTable))
	for key, val := range toolsTable {
		tools[key] = versionOf(val)
	}

	return tools, nil
}

// versionOf extracts a scalar version string from a tool value, which may be
// a plain string, a nested table (complex tool), or an array (sequence tool).
// For complex/sequence forms it returns the primary version only.
func versionOf(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case map[string]any:
		if ver, ok := v["version"].(string); ok {
			return ver
		}
		return ""
	case []any:
		if len(v) > 0 {
			return versionOf(v[0])
		}
		return ""
	default:
		return ""
	}
}

// toStringAny converts a string map to an any map.
func toStringAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (m *MiseProvisioner) Install(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "mise", "install")
	cmd.Dir = m.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// AddTools delegates to `mise use`, which writes the tools into mise.toml and
// installs them. It is idempotent for already-present tools. Only called when
// mise.toml already exists.
func (m *MiseProvisioner) AddTools(ctx context.Context, tools map[string]string) error {
	args := []string{"use"}
	for name, version := range tools {
		args = append(args, toolArg(name, version))
	}
	if m.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] mise %s\n", strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, "mise", args...)
	cmd.Dir = m.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (m *MiseProvisioner) RunCommand(cmdArgs []string) []string {
	return append([]string{"mise", "exec", "--"}, cmdArgs...)
}

func (m *MiseProvisioner) Shell(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "mise", "shell")
	cmd.Dir = m.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (m *MiseProvisioner) Trust(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "mise", "trust")
	cmd.Dir = m.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *MiseProvisioner) Untrust(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "mise", "trust", "--untrust")
	cmd.Dir = m.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *MiseProvisioner) IsAvailable() bool {
	return checkBinary("mise")
}

// Ensure MiseProvisioner implements Provisioner
var _ Provisioner = (*MiseProvisioner)(nil)
