package cmd

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/DiegoDev2/Fleet/internal/cli/ui"
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
			out := cmd.OutOrStdout()
			m, err := coremanifest.ParseFile(args[0])
			if err != nil {
				return err
			}
			if verrs := coremanifest.Validate(m); len(verrs) > 0 {
				ui.Fail(out, "manifest validation errors:")
				for _, e := range verrs {
					fmt.Fprintf(out, "  %s %s\n", ui.Muted(ui.GlyphBullet), e.Error())
				}
				return fmt.Errorf("invalid manifest")
			}

			key := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

			rows := []string{
				fmt.Sprintf("%s  %s", ui.Bold(m.Name), ui.Accent(m.Version)),
			}
			if m.Description != "" {
				rows = append(rows, ui.Muted(m.Description))
			}
			ui.Section(out, "Install plan", rows)

			const w = 11
			ui.KV(out, "Platform", w, ui.Accent(key))

			if m.UsesAssets() {
				asset, ok := m.AssetFor(key)
				if !ok {
					ui.Fail(out, fmt.Sprintf("no asset available for %s", key))
					return fmt.Errorf("no asset available for %s", key)
				}
				ui.KV(out, "URL", w, asset.URL)
				ui.KV(out, "Checksum", w, ui.Muted(ui.Truncate(asset.Checksum, 32)))
				if asset.Type != "" {
					ui.KV(out, "Type", w, asset.Type)
				}
				if asset.StripComponents > 0 {
					ui.KV(out, "Strip", w, fmt.Sprintf("%d", asset.StripComponents))
				}
				ui.KV(out, "Binaries", w, strings.Join(m.ResolvedBinaries(), ", "))
				return nil
			}

			return simulateLegacy(out, m, key)
		},
	}
}

func simulateLegacy(w io.Writer, m *manifest.Manifest, key string) error {
	const kvWidth = 11
	for osName, pd := range m.Platforms {
		for arch, ad := range pd.Architecture {
			k := fmt.Sprintf("%s/%s", osName, arch)
			if k != key {
				continue
			}
			ui.KV(w, "URL", kvWidth, ad.URL)
			ui.KV(w, "Checksum", kvWidth, ui.Muted(ui.Truncate(ad.Checksum, 32)))
			ui.KV(w, "Steps", kvWidth, fmt.Sprintf("%d", len(ad.InstallSteps)))
			for i, step := range ad.InstallSteps {
				fmt.Fprintf(w, "    %s %d  %s\n",
					ui.Muted(ui.GlyphBullet), i+1, step)
			}
			return nil
		}
	}
	return fmt.Errorf("no legacy platform/arch entry for %s", key)
}
