package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
	misePath := filepath.Join(m.Config.Dir, "mise.toml")

	var sb strings.Builder
	sb.WriteString("[tools]\n")
	for _, pkg := range packages {
		sb.WriteString(fmt.Sprintf("%s = \"%s\"\n", pkg.Tool, pkg.Version))
	}

	if extra != nil {
		sb.WriteString("\n")
		writeTomlMap(&sb, extra, 0)
	}

	content := sb.String()

	if existing, err := os.ReadFile(misePath); err == nil && string(existing) == content {
		return nil
	}

	if m.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] Writing mise.toml to %s\n", misePath)
	}

	return os.WriteFile(misePath, []byte(content), 0644)
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

	tools := make(map[string]string)
	inToolsSection := false

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") {
			inToolsSection = line == "[tools]"
			continue
		}

		if !inToolsSection {
			continue
		}

		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = strings.Trim(val, "\"")
			tools[key] = val
		}
	}

	return tools, scanner.Err()
}

func writeTomlMap(sb *strings.Builder, m map[string]any, indent int) {
	prefix := strings.Repeat("  ", indent)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s%s = \"%s\"\n", prefix, k, val))
		case bool:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", prefix, k, val))
		case int, int64, float64:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", prefix, k, val))
		case map[string]any:
			sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
			writeTomlMap(sb, val, indent+1)
		case []any:
			if len(val) > 0 {
				if isAllStrings(val) {
					sb.WriteString(fmt.Sprintf("%s%s = [", prefix, k))
					for i, item := range val {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("\"%s\"", item.(string)))
					}
					sb.WriteString("]\n")
				} else {
					sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
					for _, item := range val {
						if m2, ok := item.(map[string]any); ok {
							writeTomlMap(sb, m2, indent+1)
						} else if s, ok := item.(string); ok {
							sb.WriteString(fmt.Sprintf("%s- \"%s\"\n", prefix+"  ", s))
						}
					}
				}
			}
		default:
			sb.WriteString(fmt.Sprintf("%s%s = %v\n", prefix, k, val))
		}
	}
}

func isAllStrings(arr []any) bool {
	for _, v := range arr {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}

func (m *MiseProvisioner) Install(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "mise", "install")
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