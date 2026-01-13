package provisioner

import (
	"context"
	"os"
	"os/exec"

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
	// For mise, we rely on the existing mise.toml or generate one.
	// In the current implementation, mise reads tools from mise.toml
	// which can be synced from Razdfile.
	//
	// TODO: Implement config generation when sync feature is needed.
	// For now, mise.toml is expected to exist or be created manually.
	return nil
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
