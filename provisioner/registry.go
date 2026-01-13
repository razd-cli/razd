package provisioner

import (
	"sort"
)

// registry holds all available provisioners.
// New provisioners can be added here to extend razd.
var registry = map[string]func(Config) Provisioner{
	"mise":   func(cfg Config) Provisioner { return NewMiseProvisioner(cfg) },
	"devbox": func(cfg Config) Provisioner { return NewDevboxProvisioner(cfg) },
}

// Get returns a provisioner by name.
// Returns ErrProvisionerNotFound if the provisioner is not registered.
func Get(name string, cfg Config) (Provisioner, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, &ErrProvisionerNotFound{
			Name:      name,
			Supported: Supported(),
		}
	}
	return factory(cfg), nil
}

// Supported returns a sorted list of supported provisioner names.
func Supported() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsSupported checks if a provisioner name is registered.
func IsSupported(name string) bool {
	_, ok := registry[name]
	return ok
}
