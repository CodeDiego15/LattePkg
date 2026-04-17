package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/DiegoDev2/armada/internal/cli/ui"
	"github.com/DiegoDev2/armada/internal/config"
	"github.com/DiegoDev2/armada/internal/core/installer"
	coremanifest "github.com/DiegoDev2/armada/internal/core/manifest"
	"github.com/DiegoDev2/armada/internal/core/repository"
	"github.com/DiegoDev2/armada/internal/state"
	"github.com/DiegoDev2/armada/pkg/manifest"
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

			out := cmd.OutOrStdout()
			printInstallHeader(out, m, inst.Platform().Key())

			started := time.Now()
			entry, err := inst.Install(ctx, m, installer.Options{
				Force:    force,
				Progress: newDownloadProgress(ui.ProgressWriter(), filepath.Base(assetURL(m, inst.Platform().Key()))),
			})
			if errors.Is(err, installer.ErrAlreadyInstalled) {
				ui.Info(out, fmt.Sprintf("%s %s already installed  %s",
					ui.Bold(m.Name), m.Version,
					ui.Muted("(use --force to reinstall)")))
				return nil
			}
			if err != nil {
				ui.Fail(out, err.Error())
				return err
			}

			printInstallSuccess(out, entry, paths.Bin, time.Since(started))
			return nil
		},
	}

	c.Flags().StringVar(&fromFile, "from", "", "install from a local manifest YAML file")
	c.Flags().BoolVarP(&force, "force", "f", false, "reinstall even if the same version is already present")
	return c
}

func printInstallHeader(w io.Writer, m *manifest.Manifest, key string) {
	rows := []string{
		fmt.Sprintf("%s  %s", ui.Bold(m.Name), ui.Accent(m.Version)),
	}
	if m.Description != "" {
		rows = append(rows, ui.Muted(m.Description))
	}
	meta := []string{key}
	if asset, ok := m.AssetFor(key); ok {
		if asset.Type != "" {
			meta = append(meta, asset.Type)
		}
	}
	rows = append(rows, ui.Muted(strings.Join(meta, "  "+ui.GlyphDot+"  ")))
	ui.Section(w, "Installing", rows)
}

func printInstallSuccess(w io.Writer, entry state.Installed, binDir string, elapsed time.Duration) {
	for _, b := range entry.Binaries {
		ui.Success(w, fmt.Sprintf("linked %s %s %s",
			ui.Bold(b),
			ui.Muted(ui.GlyphArrow),
			filepath.Join(binDir, b)))
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s %s %s  %s\n",
		ui.Chip(" ready "),
		ui.Bold(entry.Name),
		ui.Accent(entry.Version),
		ui.Muted(fmt.Sprintf("in %s", elapsed.Round(100*time.Millisecond))))
	fmt.Fprintf(w, "  %s %s\n",
		ui.Muted("run:"),
		ui.Accent(entry.Name+" --help"))
	fmt.Fprintf(w, "  %s %s\n",
		ui.Muted("path:"),
		ui.Muted(binDir))
}

func assetURL(m *manifest.Manifest, key string) string {
	if a, ok := m.AssetFor(key); ok {
		return a.URL
	}
	return ""
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
