package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
	pkgMap := make(map[string]string, len(packages))
	for _, pkg := range packages {
		pkgMap[pkg.Tool] = pkg.Version
	}
	return d.writeConfig(pkgMap, extra)
}

func (d *DevboxProvisioner) WriteTools(tools map[string]string) error {
	return d.writeConfig(tools, nil)
}

// writeConfig merges the given packages and extra keys into the existing
// devbox.json, preserving any key not declared in the inputs.
func (d *DevboxProvisioner) writeConfig(pkgMap map[string]string, extra map[string]any) error {
	devboxPath := filepath.Join(d.Config.Dir, "devbox.json")

	// Parse the existing file so unknown keys survive the merge.
	existing := make(map[string]any)
	if data, err := os.ReadFile(devboxPath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("failed to parse devbox.json: %w", err)
		}
	}

	// Merge extra keys into the root if provided.
	if len(extra) > 0 {
		for k, v := range extra {
			existing[k] = v
		}
	}

	// Merge packages into the packages array without dropping existing entries.
	if len(pkgMap) > 0 {
		existingPackages, _ := existing["packages"].([]any)
		if existingPackages == nil {
			existingPackages = []any{}
		}

		// Build the final list: keep existing entries whose bare name is NOT in
		// pkgMap, then append exactly one entry per pkgMap tool. This removes
		// any pre-existing duplicate (e.g. nodejs@22 + nodejs@26) and replaces
		// it with a single entry for the new version.
		replaced := make(map[string]bool, len(pkgMap))
		for tool := range pkgMap {
			replaced[tool] = true
		}

		kept := make([]any, 0, len(existingPackages)+len(pkgMap))
		for _, p := range existingPackages {
			str, ok := p.(string)
			if !ok {
				kept = append(kept, p)
				continue
			}
			name := str
			if idx := strings.LastIndex(str, "@"); idx > 0 && idx < len(str)-1 {
				name = str[:idx]
			}
			if replaced[name] {
				// Dropped: this tool will be written once below.
				continue
			}
			kept = append(kept, str)
		}

		for tool, version := range pkgMap {
			// Emit "name@version" when a version is present, otherwise the bare
			// name (devbox treats a bare name as "latest"). Never emit a
			// trailing "@".
			entry := tool
			if version != "" {
				entry = tool + "@" + version
			}
			kept = append(kept, entry)
		}

		sort.Slice(kept, func(i, j int) bool {
			return kept[i].(string) < kept[j].(string)
		})
		existing["packages"] = kept
	}

	content, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal devbox.json: %w", err)
	}
	contentStr := string(content) + "\n"

	if existingFile, err := os.ReadFile(devboxPath); err == nil && string(existingFile) == contentStr {
		return nil
	}

	if d.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] Writing devbox.json to %s\n", devboxPath)
	}

	return os.WriteFile(devboxPath, []byte(contentStr), 0644)
}

func (d *DevboxProvisioner) ReadConfig() (map[string]string, error) {
	devboxPath := filepath.Join(d.Config.Dir, "devbox.json")

	data, err := os.ReadFile(devboxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read devbox.json: %w", err)
	}

	var config struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse devbox.json: %w", err)
	}

	tools := make(map[string]string)
	for _, pkg := range config.Packages {
		// A package may be "name@version" or a bare "name" (devbox defaults
		// the version to "latest" when omitted). Keep both: bare names get an
		// empty version so they participate in sync instead of being dropped.
		if idx := strings.LastIndex(pkg, "@"); idx > 0 && idx < len(pkg)-1 {
			tools[pkg[:idx]] = pkg[idx+1:]
		} else {
			tools[pkg] = ""
		}
	}

	return tools, nil
}

func (d *DevboxProvisioner) Install(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "devbox", "install")
	cmd.Dir = d.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// AddTools delegates to `devbox add`, which writes the packages into
// devbox.json and installs them. It is idempotent: re-adding an existing
// package is a no-op (devbox reports "already in devbox.json" and returns 0).
// Only called when devbox.json already exists.
func (d *DevboxProvisioner) AddTools(ctx context.Context, tools map[string]string) error {
	args := []string{"add"}
	for name, version := range tools {
		args = append(args, toolArg(name, version))
	}
	return d.runAddCommand(ctx, args)
}

func (d *DevboxProvisioner) runAddCommand(ctx context.Context, args []string) error {
	if d.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] devbox %s\n", strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, "devbox", args...)
	cmd.Dir = d.Config.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RemoveTools delegates to `devbox rm`, which removes the packages from
// devbox.json. It is idempotent: removing an absent package is a no-op (devbox
// reports "the following packages were not found" and returns 0). Only called
// when devbox.json already exists.
func (d *DevboxProvisioner) RemoveTools(ctx context.Context, tools map[string]string) error {
	args := []string{"rm"}
	for name := range tools {
		args = append(args, name)
	}
	if d.Config.Verbose {
		fmt.Fprintf(os.Stderr, "[FIX] devbox %s\n", strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, "devbox", args...)
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
