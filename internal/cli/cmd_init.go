package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/huh"
	"go.yaml.in/yaml/v4"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/internal/trust"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
	taskast "github.com/go-task/task/v3/taskfile/ast"
)

// runInit implements the "razd init" command.
// It creates a new Razdfile.yml in the current directory.
func runInit(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	// Check if Razdfile already exists
	reader := razdfile.NewReader(razdfile.WithDir(dir))
	if reader.Exists() && !flags.Force {
		return fmt.Errorf("Razdfile already exists in %s. Use --force to overwrite", dir)
	}

	// Determine provisioner
	using, err := resolveInitProvider(dir, ctx.Log)
	if err != nil {
		return err
	}

	// Build Razdfile content
	rf := buildInitRazdfile(using)

	data, err := marshalInitRazdfile(rf)
	if err != nil {
		return fmt.Errorf("failed to marshal Razdfile: %w", err)
	}

	targetPath := filepath.Join(dir, "Razdfile.yml")
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Razdfile: %w", err)
	}

	ctx.Log.Successf("Created %s\n", targetPath)

	// Synchronize with an existing native config (e.g. a pre-existing mise.toml)
	// so its tools are imported into the new Razdfile's dependencies.
	if !flags.NoSync {
		if prov, ok := tryGetProvisioner(rf, dir, ctx.Log); ok {
			if err := syncRazdfile(rf, prov, dir, ctx.Log); err != nil {
				ctx.Log.Warnf("Failed to sync %s config: %v\n", prov.Name(), err)
			}
		}
	}

	return nil
}

// resolveInitProvider determines the provisioner to use for a new Razdfile.
// Priority:
//  1. --using flag (mise | devbox | none)
//  2. --yes flag → default "none" (skip the interactive prompt)
//  3. interactive TTY → promptInitProvider (huh select)
//  4. non-interactive → detectProvider (existing config files, fallback "none")
func resolveInitProvider(dir string, log *output.Logger) (string, error) {
	if flags.Using != "" {
		switch flags.Using {
		case "mise", "devbox", "none":
			log.Debugf("Using provisioner from --using flag: %s\n", flags.Using)
			return flags.Using, nil
		default:
			return "", fmt.Errorf("invalid provider %q: must be 'mise', 'devbox', or 'none'", flags.Using)
		}
	}

	if flags.Yes {
		log.Debugf("--yes set, skipping provisioner prompt, defaulting to none\n")
		return "none", nil
	}

	if trust.IsTerminal() {
		using, err := promptInitProvider(log, runtime.GOOS)
		if err != nil {
			return "", err
		}
		if using != "" {
			log.Debugf("Provisioner selected via prompt: %s\n", using)
			return using, nil
		}
		log.Debugf("Prompt returned empty, falling back to auto-detection\n")
	}

	using := detectProvider(dir)
	log.Debugf("Auto-detected provider: %s\n", using)
	return using, nil
}

// promptInitProvider shows an interactive select for the provisioner to use.
// Options: mise, devbox, none (default). On Windows the devbox option is
// hidden because devbox is only available on Unix systems. In non-interactive
// mode it returns an empty string without blocking.
func promptInitProvider(log *output.Logger, goos string) (string, error) {
	log.Debugf("Prompting for provisioner (goos=%s)\n", goos)

	options := initProviderOptions(goos)

	var choice string
	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which provisioner to use?").
				Description("mise and devbox manage project tools; choose none if you don't need a provisioner.").
				Options(options...).
				Value(&choice),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("Provisioner prompt aborted\n")
			return "", nil
		}
		return "", fmt.Errorf("provisioner prompt failed: %w", err)
	}

	log.Infof("Selected provisioner: %s\n", choice)
	return choice, nil
}

// initProviderOptions returns the selectable provisioner options for the given
// platform. On Windows the devbox option is hidden because devbox is only
// available on Unix systems. Order: mise, devbox (Unix only), none.
func initProviderOptions(goos string) []huh.Option[string] {
	options := []huh.Option[string]{
		huh.NewOption("mise", "mise"),
		huh.NewOption("none (no provisioner)", "none"),
	}
	if goos != "windows" {
		// Insert devbox before none so the order is mise, devbox, none.
		options = append(options[:1], append([]huh.Option[string]{huh.NewOption("devbox", "devbox")}, options[1:]...)...)
	}
	return options
}

// detectProvider detects the appropriate provider based on existing config files.
func detectProvider(dir string) string {
	miseConfigs := []string{
		"mise.toml",
		".mise.toml",
		".mise.local.toml",
		".tool-versions",
	}
	for _, f := range miseConfigs {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return "mise"
		}
	}

	devboxJSON := filepath.Join(dir, "devbox.json")
	if _, err := os.Stat(devboxJSON); err == nil {
		return "devbox"
	}

	return "none"
}

// marshalInitRazdfile serializes a freshly built init Razdfile to YAML.
// It builds the document as a node tree because taskast.Tasks has unexported
// fields and would otherwise marshal to an empty `tasks: {}` mapping.
func marshalInitRazdfile(rf *ast.Razdfile) ([]byte, error) {
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "version", Tag: "!!str"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: rf.Version, Tag: "!!str"},
	)

	if rf.Dependencies != nil {
		deps := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		deps.Content = append(deps.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "using", Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: rf.Dependencies.Using, Tag: "!!str"},
		)
		if len(rf.Dependencies.Ensure) > 0 {
			ensure := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
			for _, e := range rf.Dependencies.Ensure {
				ensure.Content = append(ensure.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: e, Tag: "!!str"})
			}
			deps.Content = append(deps.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "ensure", Tag: "!!str"},
				ensure,
			)
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "dependencies", Tag: "!!str"},
			deps,
		)
	}

	if rf.Tasks != nil && rf.Tasks.Len() > 0 {
		tasks := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for key := range rf.Tasks.Keys(nil) {
			task, ok := rf.Tasks.Get(key)
			if !ok || task == nil {
				continue
			}
			tasks.Content = append(tasks.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
				taskToInitYAMLNode(task),
			)
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "tasks", Tag: "!!str"},
			tasks,
		)
	}

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	return yaml.Marshal(doc)
}

// taskToInitYAMLNode serializes a single task into a YAML mapping node.
// A single command is written as `cmd: <command>`; multiple commands use
// `cmds: [...]`. Mirrors razdfile.taskToYAMLNode for the init path.
func taskToInitYAMLNode(task *taskast.Task) *yaml.Node {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	if len(task.Cmds) == 1 && task.Cmds[0] != nil && task.Cmds[0].Cmd != "" {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "cmd", Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: task.Cmds[0].Cmd, Tag: "!!str"},
		)
	} else if len(task.Cmds) > 0 {
		cmds := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, c := range task.Cmds {
			if c != nil && c.Cmd != "" {
				cmds.Content = append(cmds.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: c.Cmd, Tag: "!!str"})
			}
		}
		if len(cmds.Content) > 0 {
			mapping.Content = append(mapping.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "cmds", Tag: "!!str"},
				cmds,
			)
		}
	}

	return mapping
}

// buildInitRazdfile creates a default Razdfile structure.
// For "none" it omits the dependencies section (no provisioner) and adds a
// minimal default task so the file passes HasContent validation. For
// "mise"/"devbox" it writes the dependencies.using section as before.
func buildInitRazdfile(using string) *ast.Razdfile {
	rf := &ast.Razdfile{
		Version: "1",
	}

	if using == "none" {
		rf.Tasks = taskast.NewTasks(
			&taskast.TaskElement{
				Key: "default",
				Value: &taskast.Task{
					Cmds: []*taskast.Cmd{{Cmd: "echo \"razd project initialized\""}},
				},
			},
		)
		return rf
	}

	rf.Dependencies = &ast.DependenciesConfig{
		Using:  using,
		Ensure: []string{},
	}

	// Add a minimal default task
	return rf
}
