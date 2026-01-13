package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name       string
		want       string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "mise",
			want:    "mise",
			wantErr: false,
		},
		{
			name:    "devbox",
			want:    "devbox",
			wantErr: false,
		},
		{
			name:       "unknown",
			wantErr:    true,
			errMessage: "unknown provisioner: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Get(tt.name, Config{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, p)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, tt.want, p.Name())
			}
		})
	}
}

func TestRegistry_Supported(t *testing.T) {
	supported := Supported()
	assert.Contains(t, supported, "mise")
	assert.Contains(t, supported, "devbox")
	assert.Len(t, supported, 2)
}

func TestRegistry_IsSupported(t *testing.T) {
	assert.True(t, IsSupported("mise"))
	assert.True(t, IsSupported("devbox"))
	assert.False(t, IsSupported("unknown"))
	assert.False(t, IsSupported(""))
}

func TestMiseProvisioner_Name(t *testing.T) {
	p := NewMiseProvisioner(Config{})
	assert.Equal(t, "mise", p.Name())
}

func TestDevboxProvisioner_Name(t *testing.T) {
	p := NewDevboxProvisioner(Config{})
	assert.Equal(t, "devbox", p.Name())
}

func TestMiseProvisioner_RunCommand(t *testing.T) {
	p := NewMiseProvisioner(Config{})
	cmd := p.RunCommand([]string{"npm", "install"})
	assert.Equal(t, []string{"mise", "exec", "--", "npm", "install"}, cmd)
}

func TestDevboxProvisioner_RunCommand(t *testing.T) {
	p := NewDevboxProvisioner(Config{})
	cmd := p.RunCommand([]string{"npm", "install"})
	assert.Equal(t, []string{"devbox", "run", "--", "npm", "install"}, cmd)
}
