package ast

// DevboxConfig represents the devbox configuration section in Razdfile
// Based on devbox.json schema from https://www.jetify.com/docs/devbox/configuration
type DevboxConfig struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Packages    DevboxPackages    `yaml:"packages,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Shell       *DevboxShell      `yaml:"shell,omitempty"`
	Include     []string          `yaml:"include,omitempty"`
	EnvFrom     string            `yaml:"env_from,omitempty"`
	Nixpkgs     *DevboxNixpkgs    `yaml:"nixpkgs,omitempty"`
}

// DevboxPackages represents packages in either array or object format
type DevboxPackages struct {
	// List of packages in "name@version" format
	List []string
	// Map of packages with detailed configuration
	Map map[string]*DevboxPackageConfig
}

// DevboxPackageConfig represents detailed package configuration
type DevboxPackageConfig struct {
	Version           string   `yaml:"version,omitempty"`
	Platforms         []string `yaml:"platforms,omitempty"`
	ExcludedPlatforms []string `yaml:"excluded_platforms,omitempty"`
	GlibcPatch        bool     `yaml:"glibc_patch,omitempty"`
}

// DevboxShell represents shell configuration
type DevboxShell struct {
	InitHook DevboxInitHook       `yaml:"init_hook,omitempty"`
	Scripts  map[string]DevboxScript `yaml:"scripts,omitempty"`
}

// DevboxInitHook can be string or array of strings
type DevboxInitHook struct {
	Commands []string
}

// DevboxScript can be string or array of strings
type DevboxScript struct {
	Commands []string
}

// DevboxNixpkgs represents nixpkgs configuration
type DevboxNixpkgs struct {
	Commit string `yaml:"commit,omitempty"`
}

// HasPackages returns true if devbox config has packages defined
func (d *DevboxConfig) HasPackages() bool {
	return d != nil && (len(d.Packages.List) > 0 || len(d.Packages.Map) > 0)
}

// HasShell returns true if devbox config has shell configuration
func (d *DevboxConfig) HasShell() bool {
	return d != nil && d.Shell != nil
}

// HasInclude returns true if devbox config has includes
func (d *DevboxConfig) HasInclude() bool {
	return d != nil && len(d.Include) > 0
}
