package cmd

import (
	"fmt"
	"strings"

	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/state"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed tools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := config.ResolvePaths()
			if err != nil {
				return err
			}
			store, err := state.Open(paths.State)
			if err != nil {
				return err
			}
			items := store.List()
			if len(items) == 0 {
				fmt.Println("no tools installed")
				return nil
			}
			for _, it := range items {
				bins := strings.Join(it.Binaries, ", ")
				fmt.Printf("%-24s %-10s  %s  [%s]\n", it.Name, it.Version, it.Platform, bins)
			}
			return nil
		},
	}
}
