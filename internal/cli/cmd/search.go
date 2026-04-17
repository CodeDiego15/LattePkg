package cmd

import (
	"fmt"
	"strings"

	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/core/repository"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <pattern>",
		Short: "Search manifests across configured repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			if err := paths.EnsureDirs(); err != nil {
				return err
			}
			cfg, err := config.Load(paths.ConfigFile)
			if err != nil {
				return err
			}
			mgr, err := repository.NewManager(paths.Repos)
			if err != nil {
				return err
			}
			for _, r := range cfg.Repositories {
				_ = mgr.AddRepository(r.Name, r.URL, r.Type, r.Priority)
			}
			if err := mgr.SyncAllRepositories(); err != nil {
				return fmt.Errorf("sync repositories: %w", err)
			}

			results, err := mgr.SearchManifests(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Printf("no manifests matching %q\n", args[0])
				return nil
			}
			for _, m := range results {
				cats := strings.Join(m.Categories, ",")
				fmt.Printf("%-24s %-10s  %s  [%s]\n", m.Name, m.Version, m.Description, cats)
			}
			return nil
		},
	}
}
