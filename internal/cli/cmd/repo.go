package cmd

import (
	"fmt"

	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/core/repository"
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
			fmt.Printf("==> Added %s (%s)\n", args[0], args[1])
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
			fmt.Printf("==> Removed %s\n", args[0])
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
			if len(cfg.Repositories) == 0 {
				fmt.Println("no repositories configured")
				return nil
			}
			for _, r := range cfg.Repositories {
				fmt.Printf("%-16s %-6s prio=%d  %s\n", r.Name, r.Type, r.Priority, r.URL)
			}
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
			if err := mgr.SyncAllRepositories(); err != nil {
				return err
			}
			fmt.Println("==> Repositories synced")
			return nil
		},
	}
}
