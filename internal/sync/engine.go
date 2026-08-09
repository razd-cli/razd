// Package sync implements bidirectional synchronization between a Razdfile
// and native provisioner config files (mise.toml / devbox.json).
//
// The core principle is non-destructive merging: neither file is ever fully
// overwritten. Tools are reconciled in both directions, and sections outside
// the managed tool list ([env], [settings], [tasks], etc.) are preserved.
package sync

// Tool represents a single tool dependency in either config.
type Tool struct {
	// Name is the canonical tool name (e.g. "node").
	Name string
	// Version is the version string (e.g. "22", "latest").
	Version string
}

// Resolution is the outcome of a version conflict between the two configs.
type Resolution int

const (
	// ResolutionUseRazdfile applies the Razdfile version to the native config.
	ResolutionUseRazdfile Resolution = iota
	// ResolutionUseNative applies the native config version to the Razdfile.
	ResolutionUseNative
	// ResolutionSkip leaves both sides unchanged.
	ResolutionSkip
)

// Resolver resolves a version conflict between the two configs.
// It receives the tool name and the conflicting versions and returns the
// resolution. It may return an error to abort the whole sync.
type Resolver func(tool string, razdVersion, nativeVersion string) (Resolution, error)

// Changes captures the outcome of a reconciliation pass. Both directions are
// independent; a tool may appear in either, both, or neither slice.
type Changes struct {
	// ToNative lists tools to add or update in the native config.
	ToNative []Tool
	// ToRazdfile lists tools to add to the Razdfile dependencies.ensure.
	ToRazdfile []Tool
}

// HasChanges reports whether any reconciliation is required.
func (c Changes) HasChanges() bool {
	return len(c.ToNative) > 0 || len(c.ToRazdfile) > 0
}

// Reconcile computes the merge between the tools declared in the Razdfile and
// the tools present in the native config.
//
// Merge rules:
//  1. Tool only in Razdfile          -> added to native (ToNative).
//  2. Tool only in native            -> added to Razdfile (ToRazdfile).
//  3. Tool in both, same version     -> no change.
//  4. Tool in both, different version -> ask the resolver; applies its answer.
//  5. Sections outside tools          -> preserved by the caller (not handled here).
//
// The resolver is invoked only for genuine version conflicts (rule 4).
func Reconcile(razdTools, nativeTools []Tool, resolve Resolver) (Changes, error) {
	razdMap := index(razdTools)
	nativeMap := index(nativeTools)

	var changes Changes

	for name, razd := range razdMap {
		native, exists := nativeMap[name]
		if !exists {
			// Rule 1: only in Razdfile.
			changes.ToNative = append(changes.ToNative, razd)
			continue
		}
		if native.Version == razd.Version {
			// Rule 3: identical, nothing to do.
			continue
		}
		// Rule 4: conflict.
		res, err := resolve(name, razd.Version, native.Version)
		if err != nil {
			return Changes{}, err
		}
		switch res {
		case ResolutionUseRazdfile:
			changes.ToNative = append(changes.ToNative, razd)
		case ResolutionUseNative:
			changes.ToRazdfile = append(changes.ToRazdfile, native)
		case ResolutionSkip:
			// No change.
		}
	}

	// Rule 2: only in native.
	for name, native := range nativeMap {
		if _, exists := razdMap[name]; !exists {
			changes.ToRazdfile = append(changes.ToRazdfile, native)
		}
	}

	return changes, nil
}

// index builds a name->tool map, keeping the first occurrence on duplicates.
func index(tools []Tool) map[string]Tool {
	m := make(map[string]Tool, len(tools))
	for _, t := range tools {
		if _, ok := m[t.Name]; !ok {
			m[t.Name] = t
		}
	}
	return m
}
