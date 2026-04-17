// Package ui holds styled output helpers for the Armada CLI.
//
// Everything in this package honours the NO_COLOR and ARMADA_NO_COLOR
// environment variables and automatically degrades to plain text when
// stdout is not a TTY. Commands should call the helpers here instead of
// using fmt.Printf directly so output stays consistent across the binary.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
)

// Glyphs used across the CLI. The fancy versions are switched out for
// ASCII-only alternatives when the terminal does not support UTF-8.
var (
	GlyphSuccess  = "✓"
	GlyphFailure  = "✗"
	GlyphArrow    = "→"
	GlyphBullet   = "•"
	GlyphDot      = "·"
	GlyphDiamond  = "◆"
	GlyphDownload = "↓"
	GlyphRefresh  = "↻"
	GlyphInfo     = "ℹ"
	GlyphWarn     = "!"
)

const (
	colorPrimary   = "#7C3AED" // violet-600
	colorAccent    = "#22D3EE" // cyan-400
	colorSuccess   = "#10B981" // emerald-500
	colorWarn      = "#F59E0B" // amber-500
	colorError     = "#EF4444" // red-500
	colorMuted     = "#6B7280" // gray-500
	colorHeaderBg  = "#1F2937" // gray-800
	colorHeaderFg  = "#F9FAFB" // gray-50
	colorRuleLight = "#374151" // gray-700
)

var colorEnabled = detectColor()

// Palette exposes the colors used by the TUI so callers can build their own
// styled segments with the same theme.
var Palette = struct {
	Primary lipgloss.Color
	Accent  lipgloss.Color
	Success lipgloss.Color
	Warn    lipgloss.Color
	Error   lipgloss.Color
	Muted   lipgloss.Color
	Rule    lipgloss.Color
}{
	Primary: lipgloss.Color(colorPrimary),
	Accent:  lipgloss.Color(colorAccent),
	Success: lipgloss.Color(colorSuccess),
	Warn:    lipgloss.Color(colorWarn),
	Error:   lipgloss.Color(colorError),
	Muted:   lipgloss.Color(colorMuted),
	Rule:    lipgloss.Color(colorRuleLight),
}

// Shared styles.
var (
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleMuted   = lipgloss.NewStyle().Foreground(Palette.Muted)
	styleAccent  = lipgloss.NewStyle().Foreground(Palette.Accent)
	stylePrimary = lipgloss.NewStyle().Foreground(Palette.Primary)
	styleSuccess = lipgloss.NewStyle().Foreground(Palette.Success).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(Palette.Error).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(Palette.Warn).Bold(true)
	styleHeading = lipgloss.NewStyle().Foreground(Palette.Primary).Bold(true)
	styleChip    = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHeaderFg)).
			Background(Palette.Primary).
			Padding(0, 1).
			Bold(true)
)

// ColorEnabled reports whether styled output is enabled for the current
// process.
func ColorEnabled() bool { return colorEnabled }

// SetColorEnabled lets callers override detection, chiefly for tests or for
// forcing colored output into log files during recordings.
func SetColorEnabled(v bool) { colorEnabled = v }

func detectColor() bool {
	if v, ok := os.LookupEnv("NO_COLOR"); ok && v != "" {
		return false
	}
	if v, ok := os.LookupEnv("ARMADA_NO_COLOR"); ok && v != "" {
		return false
	}
	if _, force := os.LookupEnv("ARMADA_FORCE_COLOR"); force {
		return true
	}
	if term := os.Getenv("TERM"); term == "dumb" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// render returns s styled, or the plain content when colors are disabled.
func render(style lipgloss.Style, s string) string {
	if !colorEnabled {
		return s
	}
	return style.Render(s)
}

// Success writes a green success line to w.
func Success(w io.Writer, msg string) {
	fmt.Fprintln(w, render(styleSuccess, GlyphSuccess+" ")+msg)
}

// Fail writes a red failure line.
func Fail(w io.Writer, msg string) {
	fmt.Fprintln(w, render(styleError, GlyphFailure+" ")+msg)
}

// Info writes a cyan info line.
func Info(w io.Writer, msg string) {
	fmt.Fprintln(w, render(styleAccent, GlyphArrow+" ")+msg)
}

// Warn writes an amber warning line.
func Warn(w io.Writer, msg string) {
	fmt.Fprintln(w, render(styleWarn, GlyphWarn+" ")+msg)
}

// Heading renders a bold violet heading.
func Heading(w io.Writer, text string) {
	fmt.Fprintln(w, render(styleHeading, GlyphDiamond+"  "+text))
}

// Muted renders secondary text in a dim color.
func Muted(s string) string { return render(styleMuted, s) }

// Bold renders text in bold.
func Bold(s string) string { return render(styleBold, s) }

// Accent renders text in the accent color.
func Accent(s string) string { return render(styleAccent, s) }

// Primary renders text in the primary brand color.
func Primary(s string) string { return render(stylePrimary, s) }

// Chip renders a small pill label (reverse video block).
func Chip(s string) string { return render(styleChip, s) }

// KV formats a key/value row with the key dimmed and aligned to the given
// width.
func KV(w io.Writer, key string, width int, value string) {
	padded := fmt.Sprintf("%-*s", width, key)
	fmt.Fprintf(w, "  %s  %s\n", Muted(padded), value)
}

// Section draws a framed heading with title text.
//
//	┌─ Installing ─────────────────────────────────
//	│  ffuf 2.1.0
//	│  linux/amd64 · tar.gz · 3.3 MB
//	└───────────────────────────────────────────────
func Section(w io.Writer, title string, lines []string) {
	if !colorEnabled {
		fmt.Fprintln(w, "== "+title+" ==")
		for _, l := range lines {
			fmt.Fprintln(w, "  "+l)
		}
		return
	}
	rule := lipgloss.NewStyle().Foreground(Palette.Rule)
	titleSt := lipgloss.NewStyle().Foreground(Palette.Primary).Bold(true)

	const width = 64
	tailLen := maxInt(3, width-4-len(title))
	top := rule.Render("┌─ ") + titleSt.Render(title) + rule.Render(" "+strings.Repeat("─", tailLen))
	fmt.Fprintln(w, top)
	for _, l := range lines {
		fmt.Fprintln(w, rule.Render("│")+"  "+l)
	}
	fmt.Fprintln(w, rule.Render("└"+strings.Repeat("─", width-1)))
}

// Table renders a styled table with the given headers and rows.
func Table(w io.Writer, headers []string, rows [][]string) {
	if !colorEnabled {
		writeASCIITable(w, headers, rows)
		return
	}
	header := lipgloss.NewStyle().
		Foreground(Palette.Accent).
		Bold(true)
	cell := lipgloss.NewStyle().Padding(0, 1)
	border := lipgloss.NewStyle().Foreground(Palette.Rule)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(border).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return header.Padding(0, 1)
			}
			return cell
		})
	for _, r := range rows {
		t.Row(r...)
	}
	fmt.Fprintln(w, t.Render())
}

func writeASCIITable(w io.Writer, headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if i >= len(widths) {
				continue
			}
			if l := len(c); l > widths[i] {
				widths[i] = l
			}
		}
	}

	line := func(cells []string) {
		parts := make([]string, len(cells))
		for i, c := range cells {
			parts[i] = fmt.Sprintf("%-*s", widths[i], c)
		}
		fmt.Fprintln(w, strings.Join(parts, "  "))
	}
	line(headers)
	sep := make([]string, len(headers))
	for i, wd := range widths {
		sep[i] = strings.Repeat("-", wd)
	}
	line(sep)
	for _, r := range rows {
		line(r)
	}
}

// Truncate shortens s to n runes, replacing the middle with an ellipsis.
// Useful for checksums and long URLs in compact output.
func Truncate(s string, n int) string {
	if n <= 3 || len([]rune(s)) <= n {
		return s
	}
	runes := []rune(s)
	half := (n - 1) / 2
	return string(runes[:half]) + "…" + string(runes[len(runes)-(n-1-half):])
}

// PlainTTY returns true when the current stdout is a TTY and colors are
// enabled. Useful for deciding between animated and static progress bars.
func PlainTTY() bool {
	return colorEnabled
}

// ProgressWriter returns stderr so progress bars don't pollute captured
// stdout pipes.
func ProgressWriter() io.Writer { return os.Stderr }

// maxInt returns the larger of a or b. Go 1.21 adds a builtin, but we keep
// a local helper so this package compiles on older toolchains that may use
// the ui internally.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
