package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DiegoDev2/Fleet/internal/cli/ui"
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
			out := cmd.OutOrStdout()
			if len(items) == 0 {
				ui.Info(out, "no tools installed yet")
				fmt.Fprintf(out, "  %s %s\n",
					ui.Muted("run:"),
					ui.Accent("armada install <tool>"))
				return nil
			}

			sort.Slice(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})

			rows := make([][]string, 0, len(items))
			for _, it := range items {
				rows = append(rows, []string{
					ui.Bold(it.Name),
					ui.Accent(it.Version),
					ui.Muted(it.Platform),
					strings.Join(it.Binaries, ", "),
				})
			}
			ui.Table(out, []string{"TOOL", "VERSION", "PLATFORM", "BINARIES"}, rows)
			fmt.Fprintf(out, "  %s\n",
				ui.Muted(fmt.Sprintf("%d tool%s installed in %s",
					len(items), plural(len(items)), paths.Bin)))
			return nil
		},
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
