package cli

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		goos     string
		want     string
	}{
		{name: "bash", shellEnv: "/bin/bash", goos: "linux", want: "bash"},
		{name: "zsh", shellEnv: "/bin/zsh", goos: "linux", want: "zsh"},
		{name: "fish", shellEnv: "/usr/bin/fish", goos: "linux", want: "fish"},
		{name: "pwsh", shellEnv: "/usr/bin/pwsh", goos: "linux", want: "pwsh"},
		{name: "powershell maps to pwsh", shellEnv: "/mnt/c/Program Files/PowerShell/pwsh.exe", goos: "windows", want: "pwsh"},
		// Empty $SHELL: fallback depends on the platform, not on which shells
		// happen to be installed on the runner. Unix falls back to bash even
		// when pwsh is present; Windows falls back to pwsh.
		{name: "empty unix falls back to bash", shellEnv: "", goos: "linux", want: "bash"},
		{name: "empty unix falls back to bash with pwsh present", shellEnv: "", goos: "darwin", want: "bash"},
		{name: "empty windows falls back to pwsh", shellEnv: "", goos: "windows", want: "pwsh"},
		{name: "unknown shell falls back to bash", shellEnv: "/bin/unknown", goos: "linux", want: "bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old, had := os.LookupEnv("SHELL")
			if tt.shellEnv == "" {
				os.Unsetenv("SHELL")
			} else {
				os.Setenv("SHELL", tt.shellEnv)
			}
			defer func() {
				if had {
					os.Setenv("SHELL", old)
				} else {
					os.Unsetenv("SHELL")
				}
			}()
			got := detectShellFor(tt.goos)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestActivationCommand_Mise(t *testing.T) {
	args, err := activationCommand("mise", "bash")
	require.NoError(t, err)
	assert.Equal(t, []string{"mise", "activate", "bash"}, args)
}

func TestActivationCommand_Devbox(t *testing.T) {
	args, err := activationCommand("devbox", "bash")
	require.NoError(t, err)
	assert.Equal(t, []string{"devbox", "shellenv", "--format", "bash"}, args)
}

func TestActivationCommand_Unsupported(t *testing.T) {
	_, err := activationCommand("unknown", "bash")
	assert.Error(t, err)
}

func TestActivationStartup_Bash(t *testing.T) {
	launch, env, cleanup, err := activationStartup("bash", "export MISE_SHELL=bash")
	require.NoError(t, err)
	defer cleanup()
	assert.Equal(t, "bash", launch[0])
	assert.Contains(t, launch, "--rcfile")
	assert.Contains(t, launch, "-i")
	assert.Nil(t, env)
}

// Regression: the activation script must be sourced verbatim, not wrapped in
// `eval "..."`. Wrapping re-parses the whole multi-line script as one string,
// so positional parameters ($1, $@) expand to empty in the outer eval context.
// mise's activation script uses `[[ $1 != "x" ]]` guards, which then produce
// `conditional binary operator expected` / `syntax error near "x"` and break
// the activation hooks. PATH entries containing spaces (e.g. WSL
// /mnt/c/Program Files/...) are likewise split into separate export arguments.
func TestActivationStartup_Bash_PathWithSpaces(t *testing.T) {
	// Mirrors the structure of `mise activate bash`: a PATH export with a
	// space-containing entry and a function using a positional-parameter
	// guard inside [[ ]].
	activation := "export PATH=\"/mnt/c/Program Files/dotnet:$PATH\"\n" +
		"_mise_hook_prompt_command() {\n" +
		"\tif [[ $1 != \"_mise_hook_prompt_command\" && $1 != \"_mise_hook\" ]]; then\n" +
		"\t\techo hook-ok\n" +
		"\tfi\n" +
		"}\n"
	launch, env, cleanup, err := activationStartup("bash", activation)
	require.NoError(t, err)
	defer cleanup()

	// Find the --rcfile argument.
	rcfile := ""
	for i, a := range launch {
		if a == "--rcfile" && i+1 < len(launch) {
			rcfile = launch[i+1]
		}
	}
	require.NotEmpty(t, rcfile, "expected --rcfile argument")

	// Run bash with the generated rcfile and confirm the PATH entry with
	// spaces survives and the function is defined.
	cmd := exec.Command("bash", "--rcfile", rcfile, "-i", "-c",
		`printf '%s' "$PATH"; printf '\n'; type _mise_hook_prompt_command`)
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "bash rcfile failed: %s", out)
	assert.Contains(t, string(out), "/mnt/c/Program Files/dotnet")
	assert.Contains(t, string(out), "_mise_hook_prompt_command is a function")
}

func TestActivationStartup_Zsh(t *testing.T) {
	launch, env, cleanup, err := activationStartup("zsh", "export MISE_SHELL=zsh")
	require.NoError(t, err)
	defer cleanup()
	assert.Equal(t, "zsh", launch[0])
	assert.Contains(t, launch, "-i")
	// ZDOTDIR must be set so the temp .zshrc is used.
	found := false
	for _, e := range env {
		if len(e) > len("ZDOTDIR=") && e[:len("ZDOTDIR=")] == "ZDOTDIR=" {
			found = true
		}
	}
	assert.True(t, found, "expected ZDOTDIR in env")
}

func TestActivationStartup_Fish(t *testing.T) {
	launch, env, cleanup, err := activationStartup("fish", "set -gx MISE_SHELL fish")
	require.NoError(t, err)
	assert.Nil(t, cleanup)
	assert.Equal(t, "fish", launch[0])
	assert.Contains(t, launch, "-C")
	assert.Nil(t, env)
}

func TestActivationStartup_Pwsh(t *testing.T) {
	launch, env, cleanup, err := activationStartup("pwsh", "export MISE_SHELL=pwsh")
	require.NoError(t, err)
	assert.Nil(t, cleanup)
	assert.Equal(t, "pwsh", launch[0])
	assert.Contains(t, launch, "-NoExit")
	// The activation command element must contain Invoke-Expression.
	found := false
	for _, a := range launch {
		if len(a) > len("Invoke-Expression") && a[:len("Invoke-Expression")] == "Invoke-Expression" {
			found = true
		}
	}
	assert.True(t, found, "expected an Invoke-Expression command element")
	assert.Nil(t, env)
}

func TestActivationStartup_Unsupported(t *testing.T) {
	_, _, _, err := activationStartup("cmd", "echo hi")
	assert.Error(t, err)
}
