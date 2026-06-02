package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

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
	devboxPath := filepath.Join(d.Config.Dir, "devbox.json")

	config := make(map[string]any)

	pkgList := make([]string, 0, len(packages))
	for _, pkg := range packages {
		pkgList = append(pkgList, pkg.Raw)
	}
	sort.Strings(pkgList)

	if len(pkgList) > 0 {
		config["packages"] = pkgList
	}

	if extra != nil {
		for k, v := range extra {
			config[k] = v
		}
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal devbox.json: %w", err)
	}

	contentStr := string(content) + "\n"

	if existing, err := os.ReadFile(devboxPath); err == nil && string(existing) == contentStr {
		return nil
	}

	if d.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] Writing devbox.json to %s\n", devboxPath)
	}

	return os.WriteFile(devboxPath, []byte(contentStr), 0644)
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
	return nil
}

func (d *DevboxProvisioner) Untrust(ctx context.Context) error {
	return nil
}

func (d *DevboxProvisioner) IsAvailable() bool {
	return checkBinary("devbox")
}

// Ensure DevboxProvisioner implements Provisioner
var _ Provisioner = (*DevboxProvisioner)(nil)