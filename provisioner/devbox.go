package provisioner

import (
	"context"
	"os"
	"os/exec"

	"github.com/razd-cli/razd/razdfile/ast"
)

// DevboxProvisioner implements Provisioner for devbox (https://www.jetify.com/devbox).
type DevboxProvisioner struct {
	BaseProvisioner
}

// NewDevboxProvisioner creates a new DevboxProvisioner with the given config.
func NewDevboxProvisioner(cfg Config) *DevboxProvisioner {
	return &DevboxProvisioner{
		BaseProvisioner: BaseProvisioner{Config: cfg},
	}
}

func (d *DevboxProvisioner) Name() string {
	return "devbox"
}

func (d *DevboxProvisioner) GenerateConfig(packages []ast.ParsedDependency, extra map[string]any) error {
	// For devbox, we rely on the existing devbox.json or generate one.
	// In the current implementation, devbox reads packages from devbox.json
	// which can be synced from Razdfile.
	//
	// TODO: Implement config generation when sync feature is needed.
	// For now, devbox.json is expected to exist or be created manually.
	return nil
}

func (d *DevboxProvisioner) Install(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "devbox", "install")
	cmd.Dir = d.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (d *DevboxProvisioner) RunCommand(cmdArgs []string) []string {
	return append([]string{"devbox", "run", "--"}, cmdArgs...)
}

func (d *DevboxProvisioner) Shell(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "devbox", "shell")
	cmd.Dir = d.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (d *DevboxProvisioner) Trust(ctx context.Context) error {
	// devbox doesn't have a trust command like mise
	// Trust is handled at the razd level only
	return nil
}

func (d *DevboxProvisioner) Untrust(ctx context.Context) error {
	// devbox doesn't have an untrust command
	return nil
}

func (d *DevboxProvisioner) IsAvailable() bool {
	return checkBinary("devbox")
}

// Ensure DevboxProvisioner implements Provisioner
var _ Provisioner = (*DevboxProvisioner)(nil)
