package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/core/installer"
	coremanifest "github.com/DiegoDev2/Fleet/internal/core/manifest"
	"github.com/DiegoDev2/Fleet/internal/core/repository"
	"github.com/DiegoDev2/Fleet/internal/state"
	"github.com/DiegoDev2/Fleet/pkg/manifest"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		fromFile string
		force    bool
	)

	c := &cobra.Command{
		Use:   "install [tool]",
		Short: "Install a tool by name or from a local manifest",
		Long: "Install resolves a tool from the configured repositories, " +
			"downloads its asset for the current platform, verifies the " +
			"checksum and symlinks the binaries into ~/.armada/bin.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			if err := paths.EnsureDirs(); err != nil {
				return err
			}

			var m *manifest.Manifest
			switch {
			case fromFile != "":
				m, err = coremanifest.ParseFile(fromFile)
				if err != nil {
					return err
				}
			case len(args) == 1:
				m, err = resolveManifest(paths, args[0])
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("provide a tool name or use --from")
			}

			if verrs := coremanifest.Validate(m); len(verrs) > 0 {
				return fmt.Errorf("invalid manifest:\n  - %s", joinErrors(verrs))
			}

			store, err := state.Open(paths.State)
			if err != nil {
				return err
			}
			inst := installer.New(paths, store)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			fmt.Printf("==> Installing %s %s for %s\n", m.Name, m.Version, inst.Platform())
			entry, err := inst.Install(ctx, m, installer.Options{
				Force:    force,
				Progress: textProgress(cmd.OutOrStdout()),
			})
			if errors.Is(err, installer.ErrAlreadyInstalled) {
				fmt.Printf("==> %s %s already installed (use --force to reinstall)\n", m.Name, m.Version)
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Printf("==> Linked binaries: %s\n", strings.Join(entry.Binaries, ", "))
			fmt.Printf("==> Done. Make sure %s is on your PATH.\n", paths.Bin)
			return nil
		},
	}

	c.Flags().StringVar(&fromFile, "from", "", "install from a local manifest YAML file")
	c.Flags().BoolVarP(&force, "force", "f", false, "reinstall even if the same version is already present")
	return c
}

func resolveManifest(paths config.Paths, name string) (*manifest.Manifest, error) {
	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		return nil, err
	}

	mgr, err := repository.NewManager(paths.Repos)
	if err != nil {
		return nil, err
	}
	for _, r := range cfg.Repositories {
		_ = mgr.AddRepository(r.Name, r.URL, r.Type, r.Priority)
	}
	if err := mgr.SyncAllRepositories(); err != nil {
		return nil, fmt.Errorf("sync repositories: %w", err)
	}

	m, err := mgr.GetManifest(name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return m, nil
}
