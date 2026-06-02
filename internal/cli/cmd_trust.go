package cli

import (
	"context"
	"fmt"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/trust"
	"github.com/razd-cli/razd/provisioner"
)

// runTrust implements the "razd trust" command.
// It manages project trust status for Razdfile execution.
func runTrust(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	// --untrust: remove trust for the project
	if flags.Untrust {
		return untrustProject(ctx, dir)
	}

	// --show: display trust status for the project
	if flags.Show {
		return showTrustStatus(ctx, dir)
	}

	// --all: list all trusted projects
	if flags.All {
		return listAllTrusted(ctx)
	}

	// --ignore: add project to ignore list
	if flags.Ignore {
		return ignoreProject(ctx, dir)
	}

	// Default: trust the current project
	return trustProject(ctx, dir)
}

// trustProject trusts the current project and syncs with the provisioner.
func trustProject(ctx *Context, dir string) error {
	var prov provisioner.Provisioner
	provResolved := false

	rf, err := readRazdfile(dir, ctx.Log)
	if err == nil && needsProvisioner(rf) {
		p, ok := tryGetProvisioner(rf, dir, ctx.Log)
		if ok {
			prov = p
			provResolved = true
		}
	}

	if err := trust.Trust(dir, prov, ctx.Log); err != nil {
		return &errors.TrustError{
			Path:    dir,
			Message: fmt.Sprintf("failed to trust project: %v", err),
		}
	}

	ctx.Log.Successf("Trusted: %s\n", dir)

	if provResolved && prov.Name() == "mise" {
		ctx.Log.Debugf("Running mise trust...\n")
		if err := prov.Trust(context.Background()); err != nil {
			ctx.Log.Warnf("mise trust failed: %v\n", err)
		}
	}

	return nil
}

// untrustProject removes trust for the current project.
func untrustProject(ctx *Context, dir string) error {
	var prov provisioner.Provisioner
	provResolved := false

	rf, err := readRazdfile(dir, ctx.Log)
	if err == nil && needsProvisioner(rf) {
		p, ok := tryGetProvisioner(rf, dir, ctx.Log)
		if ok {
			prov = p
			provResolved = true
		}
	}

	store, err := trust.Load()
	if err != nil {
		ctx.Log.Warnf("Failed to load trust store: %v\n", err)
	}

	store.Remove(dir)
	if err := store.Save(); err != nil {
		ctx.Log.Warnf("Failed to save trust store: %v\n", err)
	}

	if provResolved && prov.Name() == "mise" {
		ctx.Log.Debugf("Running mise trust --untrust...\n")
		if err := prov.Untrust(context.Background()); err != nil {
			ctx.Log.Warnf("mise trust --untrust failed: %v\n", err)
		}
	}

	ctx.Log.Successf("Untrusted: %s\n", dir)
	return nil
}

// showTrustStatus displays the trust status for the current project.
func showTrustStatus(ctx *Context, dir string) error {
	store, err := trust.Load()
	if err != nil {
		ctx.Log.Warnf("Failed to load trust store: %v\n", err)
		return err
	}

	status := store.GetStatus(dir)
	ctx.Log.Infof("Trust status for %s: %s\n", dir, status)
	return nil
}

// listAllTrusted lists all trusted projects.
func listAllTrusted(ctx *Context) error {
	store, err := trust.Load()
	if err != nil {
		ctx.Log.Warnf("Failed to load trust store: %v\n", err)
		return err
	}

	trusted := store.ListTrusted()
	ignored := store.ListIgnored()

	if len(trusted) == 0 && len(ignored) == 0 {
		fmt.Println("No trusted or ignored projects.")
		return nil
	}

	if len(trusted) > 0 {
		fmt.Println("Trusted projects:")
		for _, p := range trusted {
			fmt.Printf("  %s\n", p)
		}
	}

	if len(ignored) > 0 {
		fmt.Println("\nIgnored projects:")
		for _, p := range ignored {
			fmt.Printf("  %s\n", p)
		}
	}

	return nil
}

// ignoreProject adds the current project to the ignore list.
func ignoreProject(ctx *Context, dir string) error {
	if err := trust.Ignore(dir, ctx.Log); err != nil {
		return err
	}

	ctx.Log.Successf("Ignored: %s\n", dir)
	return nil
}