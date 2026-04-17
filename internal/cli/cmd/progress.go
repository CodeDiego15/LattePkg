package cmd

import (
	"fmt"
	"io"

	"github.com/DiegoDev2/Fleet/internal/core/manifest"
	"github.com/DiegoDev2/Fleet/internal/download"
)

// textProgress returns a ProgressFunc that prints a single-line progress
// update to w. It is safe to pass nil if w does not support line rewriting.
func textProgress(w io.Writer) download.ProgressFunc {
	return func(current, total int64) {
		if total > 0 {
			pct := float64(current) / float64(total) * 100
			fmt.Fprintf(w, "\r   %s / %s  (%.1f%%)    ",
				humanBytes(current), humanBytes(total), pct)
			if current >= total {
				fmt.Fprintln(w)
			}
		} else {
			fmt.Fprintf(w, "\r   %s    ", humanBytes(current))
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
