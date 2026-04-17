package cmd

import (
	"fmt"
	"strconv"

	"github.com/DiegoDev2/armada/internal/cli/ui"
	"github.com/DiegoDev2/armada/internal/config"
	"github.com/DiegoDev2/armada/internal/core/repository"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "repo",
		Short: "Manage manifest repositories",
	}
	c.AddCommand(newRepoAddCmd(), newRepoRemoveCmd(), newRepoListCmd(), newRepoSyncCmd())
	return c
}

func newRepoAddCmd() *cobra.Command {
	var (
		repoType string
		priority int
	)
	c := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Register a new manifest repository",
		Args:  cobra.ExactArgs(2),
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
			added := cfg.AddRepository(config.Repository{
				Name:     args[0],
				URL:      args[1],
				Type:     repoType,
				Priority: priority,
			})
			if !added {
				return fmt.Errorf("repository %q already exists", args[0])
			}
			if err := cfg.Save(paths.ConfigFile); err != nil {
				return err
			}
			ui.Success(cmd.OutOrStdout(), fmt.Sprintf("added repository %s  %s",
				ui.Bold(args[0]), ui.Muted(args[1])))
			return nil
		},
	}
	c.Flags().StringVar(&repoType, "type", "git", "repository type: git or http")
	c.Flags().IntVar(&priority, "priority", 50, "priority (higher wins on conflicts)")
	return c
}

func newRepoRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a manifest repository",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			cfg, err := config.Load(paths.ConfigFile)
			if err != nil {
				return err
			}
			if !cfg.RemoveRepository(args[0]) {
				return fmt.Errorf("repository %q not found", args[0])
			}
			if err := cfg.Save(paths.ConfigFile); err != nil {
				return err
			}
			ui.Success(cmd.OutOrStdout(), fmt.Sprintf("removed repository %s", ui.Bold(args[0])))
			return nil
		},
	}
}

func newRepoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured repositories",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			cfg, err := config.Load(paths.ConfigFile)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(cfg.Repositories) == 0 {
				ui.Info(out, "no repositories configured")
				fmt.Fprintf(out, "  %s %s\n",
					ui.Muted("run:"),
					ui.Accent("armada repo add <name> <url>"))
				return nil
			}
			rows := make([][]string, 0, len(cfg.Repositories))
			for _, r := range cfg.Repositories {
				rows = append(rows, []string{
					ui.Bold(r.Name),
					ui.Muted(r.Type),
					ui.Accent(strconv.Itoa(r.Priority)),
					r.URL,
				})
			}
			ui.Table(out, []string{"NAME", "TYPE", "PRIORITY", "URL"}, rows)
			return nil
		},
	}
}

func newRepoSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync manifests from configured repositories",
		Args:  cobra.NoArgs,
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
			ui.Info(out, fmt.Sprintf("%s syncing %d repositor%s",
				ui.GlyphRefresh,
				len(cfg.Repositories),
				map[bool]string{true: "y", false: "ies"}[len(cfg.Repositories) == 1]))
			if err := mgr.SyncAllRepositories(); err != nil {
				return err
			}
			ui.Success(out, "repositories synced")
			return nil
		},
	}
}
