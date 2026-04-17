package cmd

import (
	"fmt"
	"io"

	"github.com/DiegoDev2/Fleet/internal/cli/ui"
	"github.com/DiegoDev2/Fleet/internal/core/manifest"
	"github.com/DiegoDev2/Fleet/internal/download"
	"github.com/schollz/progressbar/v3"
)

// newDownloadProgress returns a ProgressFunc that drives a styled progress
// bar on the given writer. When the terminal is not interactive, a plain
// line is emitted once the transfer completes so logs stay readable.
func newDownloadProgress(w io.Writer, label string) download.ProgressFunc {
	if !ui.PlainTTY() {
		var total int64
		return func(current, t int64) {
			if t > 0 {
				total = t
			}
			if total > 0 && current >= total {
				fmt.Fprintf(w, "    %s  %s\n",
					ui.Muted("downloaded"),
					humanBytes(total))
			}
		}
	}

	var bar *progressbar.ProgressBar
	return func(current, total int64) {
		if bar == nil {
			bar = progressbar.NewOptions64(total,
				progressbar.OptionSetWriter(w),
				progressbar.OptionSetDescription("  "+ui.Accent(ui.GlyphDownload+" downloading")+"  "+ui.Muted(label)),
				progressbar.OptionSetWidth(24),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetPredictTime(false),
				progressbar.OptionThrottle(80_000_000),
				progressbar.OptionSetRenderBlankState(true),
				progressbar.OptionShowCount(),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "█",
					SaucerPadding: "░",
					BarStart:      "",
					BarEnd:        "",
				}),
				progressbar.OptionOnCompletion(func() { fmt.Fprintln(w) }),
			)
		}
		if total > 0 && bar.GetMax64() != total {
			bar.ChangeMax64(total)
		}
		_ = bar.Set64(current)
		if total > 0 && current >= total {
			_ = bar.Finish()
		}
	}
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// joinErrors formats a slice of ValidationError as a newline-joined string.
func joinErrors(errs []manifest.ValidationError) string {
	if len(errs) == 0 {
		return ""
	}
	s := errs[0].Error()
	for _, e := range errs[1:] {
		s += "\n  - " + e.Error()
	}
	return s
}
