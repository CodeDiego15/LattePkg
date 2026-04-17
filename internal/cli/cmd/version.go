package cmd

import (
	"fmt"
	"runtime"

	"github.com/DiegoDev2/armada/internal/cli/ui"
	"github.com/spf13/cobra"
)

const banner = `
   █████╗ ██████╗ ███╗   ███╗ █████╗ ██████╗  █████╗
  ██╔══██╗██╔══██╗████╗ ████║██╔══██╗██╔══██╗██╔══██╗
  ███████║██████╔╝██╔████╔██║███████║██║  ██║███████║
  ██╔══██║██╔══██╗██║╚██╔╝██║██╔══██║██║  ██║██╔══██║
  ██║  ██║██║  ██║██║ ╚═╝ ██║██║  ██║██████╔╝██║  ██║
  ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝`

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Armada version and platform information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			if ui.ColorEnabled() {
				fmt.Fprintln(out, ui.Primary(banner))
			}
			fmt.Fprintf(out, "\n  %s %s  %s\n",
				ui.Bold("armada"),
				ui.Accent(Version),
				ui.Muted(fmt.Sprintf("%s/%s · go%s",
					runtime.GOOS, runtime.GOARCH, runtime.Version()[2:])))
			fmt.Fprintf(out, "  %s %s\n\n",
				ui.Muted("home:"),
				ui.Muted("https://github.com/DiegoDev2/armada"))
		},
	}
}
