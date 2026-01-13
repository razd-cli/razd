package ast

import (
	"testing"
)

func TestParseDependencyString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ParsedDependency
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "node@22",
			want:  ParsedDependency{Tool: "node", Version: "22", Raw: "node@22"},
		},
		{
			name:  "dotted version",
			input: "php@8.4",
			want:  ParsedDependency{Tool: "php", Version: "8.4", Raw: "php@8.4"},
		},
		{
			name:  "latest version",
			input: "pnpm@latest",
			want:  ParsedDependency{Tool: "pnpm", Version: "latest", Raw: "pnpm@latest"},
		},
		{
			name:  "full semver",
			input: "python@3.11.5",
			want:  ParsedDependency{Tool: "python", Version: "3.11.5", Raw: "python@3.11.5"},
		},
		{
			name:  "tool with hyphen",
			input: "node-lts@22",
			want:  ParsedDependency{Tool: "node-lts", Version: "22", Raw: "node-lts@22"},
		},
		{
			name:  "tool with underscore",
			input: "my_tool@1.0",
			want:  ParsedDependency{Tool: "my_tool", Version: "1.0", Raw: "my_tool@1.0"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing version",
			input:   "node",
			wantErr: true,
		},
		{
			name:    "missing tool",
			input:   "@22",
			wantErr: true,
		},
		{
			name:    "special characters in version",
			input:   "node@22!",
			wantErr: true,
		},
		{
			name:    "tool starting with number",
			input:   "123node@22",
			wantErr: true,
		},
		{
			name:    "uppercase tool",
			input:   "Node@22",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDependencyString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDependencyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseDependencyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDependenciesConfig_ParseEnsure(t *testing.T) {
	tests := []struct {
		name    string
		config  *DependenciesConfig
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantLen: 0,
		},
		{
			name: "empty ensure",
			config: &DependenciesConfig{
				Using:  "mise",
				Ensure: []string{},
			},
			wantLen: 0,
		},
		{
			name: "valid ensure list",
			config: &DependenciesConfig{
				Using:  "mise",
				Ensure: []string{"node@22", "php@8.4", "pnpm@latest"},
			},
			wantLen: 3,
		},
		{
			name: "invalid ensure string",
			config: &DependenciesConfig{
				Using:  "mise",
				Ensure: []string{"node@22", "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ParseEnsure()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnsure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("ParseEnsure() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestDependenciesConfig_HasEnsure(t *testing.T) {
	tests := []struct {
		name   string
		config *DependenciesConfig
		want   bool
	}{
		{
			name:   "nil config",
			config: nil,
			want:   false,
		},
		{
			name:   "empty ensure",
			config: &DependenciesConfig{Ensure: []string{}},
			want:   false,
		},
		{
			name:   "with ensure",
			config: &DependenciesConfig{Ensure: []string{"node@22"}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.HasEnsure(); got != tt.want {
				t.Errorf("HasEnsure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDependenciesConfig_GetProviderExtra(t *testing.T) {
	miseExtra := map[string]any{"env": map[string]string{"NODE_ENV": "dev"}}
	devboxExtra := map[string]any{"shell": map[string]any{"init_hook": "echo hi"}}

	tests := []struct {
		name   string
		config *DependenciesConfig
		want   map[string]any
	}{
		{
			name:   "nil config",
			config: nil,
			want:   nil,
		},
		{
			name: "mise provider",
			config: &DependenciesConfig{
				Using: "mise",
				Extra: &DependenciesExtra{Mise: miseExtra, Devbox: devboxExtra},
			},
			want: miseExtra,
		},
		{
			name: "devbox provider",
			config: &DependenciesConfig{
				Using: "devbox",
				Extra: &DependenciesExtra{Mise: miseExtra, Devbox: devboxExtra},
			},
			want: devboxExtra,
		},
		{
			name: "unknown provider",
			config: &DependenciesConfig{
				Using: "unknown",
				Extra: &DependenciesExtra{Mise: miseExtra},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetProviderExtra()
			if tt.want == nil && got != nil {
				t.Errorf("GetProviderExtra() = %v, want nil", got)
			}
			if tt.want != nil && got == nil {
				t.Errorf("GetProviderExtra() = nil, want %v", tt.want)
			}
		})
	}
}

func TestInvalidDependencyFormatError(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty value",
			value: "",
			want:  "invalid dependency format: empty string, expected 'tool@version'",
		},
		{
			name:  "invalid value",
			value: "node",
			want:  "invalid dependency format 'node', expected 'tool@version' (e.g., 'node@22')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &InvalidDependencyFormatError{Value: tt.value}
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
