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
