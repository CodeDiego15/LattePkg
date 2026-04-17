package cmd

import (
	"fmt"
	"runtime"

	coremanifest "github.com/DiegoDev2/Fleet/internal/core/manifest"
	"github.com/DiegoDev2/Fleet/pkg/manifest"
	"github.com/spf13/cobra"
)

func newSimulateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "simulate <manifest-file>",
		Short: "Print the resolved install plan for a manifest without downloading anything",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := coremanifest.ParseFile(args[0])
			if err != nil {
				return err
			}
			if verrs := coremanifest.Validate(m); len(verrs) > 0 {
				fmt.Println("manifest validation errors:")
				for _, e := range verrs {
					fmt.Printf("  - %s\n", e.Error())
				}
				return fmt.Errorf("invalid manifest")
			}

			key := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("Tool:        %s\n", m.Name)
			fmt.Printf("Version:     %s\n", m.Version)
			if m.Description != "" {
				fmt.Printf("Description: %s\n", m.Description)
			}
			fmt.Printf("Platform:    %s\n", key)

			if m.UsesAssets() {
				asset, ok := m.AssetFor(key)
				if !ok {
					return fmt.Errorf("no asset available for %s", key)
				}
				fmt.Printf("URL:         %s\n", asset.URL)
				fmt.Printf("Checksum:    %s\n", asset.Checksum)
				if asset.Type != "" {
					fmt.Printf("Type:        %s\n", asset.Type)
				}
				fmt.Printf("Binaries:    %v\n", m.ResolvedBinaries())
				return nil
			}

			return simulateLegacy(m, key)
		},
	}
}

func simulateLegacy(m *manifest.Manifest, key string) error {
	for osName, pd := range m.Platforms {
		for arch, ad := range pd.Architecture {
			k := fmt.Sprintf("%s/%s", osName, arch)
			if k != key {
				continue
			}
			fmt.Printf("URL:         %s\n", ad.URL)
			fmt.Printf("Checksum:    %s\n", ad.Checksum)
			fmt.Printf("Steps:       %d step(s)\n", len(ad.InstallSteps))
			for i, step := range ad.InstallSteps {
				fmt.Printf("  [%d] %s\n", i+1, step)
			}
			return nil
		}
	}
	return fmt.Errorf("no legacy platform/arch entry for %s", key)
}
