package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DiegoDev2/armada/internal/cli/ui"
	"github.com/DiegoDev2/armada/internal/config"
	"github.com/DiegoDev2/armada/internal/core/repository"
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
			out := cmd.OutOrStdout()
			ui.Info(out, fmt.Sprintf("syncing %d repositor%s",
				len(cfg.Repositories),
				map[bool]string{true: "y", false: "ies"}[len(cfg.Repositories) == 1]))
			if err := mgr.SyncAllRepositories(); err != nil {
				return fmt.Errorf("sync repositories: %w", err)
			}

			results, err := mgr.SearchManifests(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				ui.Warn(out, fmt.Sprintf("no manifests matching %q", args[0]))
				return nil
			}
			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})
			rows := make([][]string, 0, len(results))
			for _, m := range results {
				rows = append(rows, []string{
					ui.Bold(m.Name),
					ui.Accent(m.Version),
					truncateDesc(m.Description, 48),
					ui.Muted(strings.Join(m.Categories, ", ")),
				})
			}
			ui.Table(out, []string{"TOOL", "VERSION", "DESCRIPTION", "CATEGORIES"}, rows)
			fmt.Fprintf(out, "  %s\n",
				ui.Muted(fmt.Sprintf("%d result%s for %q", len(results), plural(len(results)), args[0])))
			return nil
		},
	}
}

func truncateDesc(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}
