package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{name: "bash", shellEnv: "/bin/bash", want: "bash"},
		{name: "zsh", shellEnv: "/bin/zsh", want: "zsh"},
		{name: "fish", shellEnv: "/usr/bin/fish", want: "fish"},
		{name: "pwsh", shellEnv: "/usr/bin/pwsh", want: "pwsh"},
		{name: "powershell maps to pwsh", shellEnv: "/mnt/c/Program Files/PowerShell/pwsh.exe", want: "pwsh"},
		{name: "empty falls back to bash", shellEnv: "", want: "bash"},
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
			got := detectShell()
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
