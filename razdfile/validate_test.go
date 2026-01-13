package razdfile

import (
	"strings"
	"testing"

	"github.com/razd-cli/razd/razdfile/ast"
)

func TestValidate_Dependencies(t *testing.T) {
	tests := []struct {
		name       string
		razdfile   *ast.Razdfile
		wantErr    bool
		errContain string
	}{
		{
			name: "valid dependencies with mise",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "mise",
					Ensure: []string{"node@22"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid dependencies with devbox",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "devbox",
					Ensure: []string{"node@22", "php@8.4"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid dependencies with extra",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "devbox",
					Ensure: []string{"go@1.21"},
					Extra: &ast.DependenciesExtra{
						Devbox: map[string]any{
							"shell": map[string]any{
								"init_hook": "echo hello",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid dependencies with empty ensure",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "mise",
					Ensure: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing using field",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Ensure: []string{"node@22"},
				},
			},
			wantErr:    true,
			errContain: "using is required",
		},
		{
			name: "invalid using value",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "npm",
					Ensure: []string{"node@22"},
				},
			},
			wantErr:    true,
			errContain: "must be 'mise' or 'devbox'",
		},
		{
			name: "invalid ensure format",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "mise",
					Ensure: []string{"node"},
				},
			},
			wantErr:    true,
			errContain: "invalid dependency format",
		},
		{
			name: "dependencies with mise section - mutual exclusion",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "mise",
					Ensure: []string{"node@22"},
				},
				Mise: &ast.MiseConfig{
					Tools: map[string]*ast.MiseTool{
						"node": {Version: "22"},
					},
				},
			},
			wantErr:    true,
			errContain: "cannot use 'dependencies' together with 'mise'",
		},
		{
			name: "dependencies with devbox section - mutual exclusion",
			razdfile: &ast.Razdfile{
				Version: "1",
				Dependencies: &ast.DependenciesConfig{
					Using:  "devbox",
					Ensure: []string{"node@22"},
				},
				Devbox: &ast.DevboxConfig{
					Packages: ast.DevboxPackages{
						List: []string{"nodejs@22"},
					},
				},
			},
			wantErr:    true,
			errContain: "cannot use 'dependencies' together with 'devbox'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.razdfile, "test.yml")
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContain != "" {
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errContain)
				}
			}
		})
	}
}

func TestValidate_HasContent_WithDependencies(t *testing.T) {
	// Razdfile with only dependencies should be valid
	rf := &ast.Razdfile{
		Version: "1",
		Dependencies: &ast.DependenciesConfig{
			Using:  "mise",
			Ensure: []string{"node@22"},
		},
	}

	err := Validate(rf, "test.yml")
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}
