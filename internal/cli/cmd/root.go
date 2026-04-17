// Package cmd defines the Armada command-line interface.
package cmd

import (
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X ...Version=...".
var Version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:   "armada",
	Short: "A reproducible cross-platform toolbox",
	Long: `Armada is a user-space package manager built for reproducible, ` +
		`cross-platform tool installs. Manifests describe a tool, Armada ` +
		`downloads the right asset for your platform, verifies its checksum ` +
		`and places binaries in ~/.armada/bin.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(
		newInstallCmd(),
		newUninstallCmd(),
		newListCmd(),
		newSearchCmd(),
		newRepoCmd(),
		newSimulateCmd(),
		newVersionCmd(),
	)
}
