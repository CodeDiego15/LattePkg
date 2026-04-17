package cmd

import (
	"fmt"

	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/core/installer"
	"github.com/DiegoDev2/Fleet/internal/state"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall <tool>",
		Aliases: []string{"remove", "rm"},
		Short:   "Remove an installed tool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			store, err := state.Open(paths.State)
			if err != nil {
				return err
			}
			inst := installer.New(paths, store)
			if err := inst.Uninstall(args[0]); err != nil {
				return err
			}
			fmt.Printf("==> Removed %s\n", args[0])
			return nil
		},
	}
}
