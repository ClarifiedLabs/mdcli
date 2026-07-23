// Command md is a standalone terminal Markdown viewer. It renders Markdown
// with terminal styling and draws ```mermaid code fences as ASCII diagrams.
// It uses only the Go standard library.
//
// Usage:
//
//	md [flags] [file...]
//
// With no file arguments it reads Markdown from standard input.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ClarifiedLabs/mdcli/internal/buildinfo"
	"github.com/ClarifiedLabs/mdcli/internal/markdown"
	"github.com/ClarifiedLabs/mdcli/internal/pager"
	"github.com/ClarifiedLabs/mdcli/internal/term"
	"github.com/ClarifiedLabs/mdcli/internal/viewer"
)

func main() {
	var (
		width       int
		colorMode   string
		pagerMode   string
		showVersion bool
	)
	fs := flag.NewFlagSet("md", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.IntVar(&width, "w", 0, "wrap width in columns (0 = auto)")
	fs.IntVar(&width, "width", 0, "wrap width in columns (0 = auto)")
	fs.StringVar(&colorMode, "color", "auto", "color output: auto, always, or never")
	fs.StringVar(&colorMode, "colour", "auto", "alias for -color")
	fs.StringVar(&pagerMode, "p", "auto", "pager: auto, always, or never")
	fs.StringVar(&pagerMode, "pager", "auto", "pager: auto, always, or never")
	fs.BoolVar(&showVersion, "version", false, "print version information and exit")
	fs.Usage = usage
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if showVersion {
		fmt.Println(buildinfo.Line("md"))
		return
	}

	ansi, err := resolveColor(colorMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "md:", err)
		os.Exit(2)
	}
	usePager, err := resolvePager(pagerMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "md:", err)
		os.Exit(2)
	}
	w := resolveWidth(width)
	opts := viewer.Options{ANSI: ansi, Width: w}

	// Render everything into one buffer, then emit it once: through a pager
	// when interactive, or straight to stdout otherwise.
	var buf bytes.Buffer

	args := fs.Args()
	if len(args) == 0 {
		if err := renderReader(&buf, os.Stdin, opts); err != nil {
			fmt.Fprintln(os.Stderr, "md:", err)
			os.Exit(1)
		}
		emit(buf.Bytes(), usePager)
		return
	}

	exit := 0
	for _, path := range args {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "md:", err)
			exit = 1
			continue
		}
		err = renderReader(&buf, f, opts)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "md: %s: %v\n", path, err)
			exit = 1
		}
	}
	emit(buf.Bytes(), usePager)
	os.Exit(exit)
}

// emit writes the rendered output, paging it when interactive and enabled.
// A pager failure is reported on stderr but never changes the exit status.
func emit(content []byte, usePager bool) {
	if usePager {
		if err := pager.Page(content, os.Getenv("PAGER"), os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "md:", err)
		}
		return
	}
	os.Stdout.Write(content)
}

func renderReader(w io.Writer, r io.Reader, opts viewer.Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	io.WriteString(w, viewer.Render(string(data), opts))
	return nil
}

// resolveColor turns the -color flag value into an ANSI on/off decision.
func resolveColor(mode string) (bool, error) {
	switch strings.ToLower(mode) {
	case "always", "true", "on", "yes", "1":
		return true, nil
	case "never", "false", "off", "no", "0":
		return false, nil
	case "auto", "":
		return defaultColor(), nil
	default:
		return false, fmt.Errorf("invalid -color value %q (want auto, always, or never)", mode)
	}
}

// resolvePager decides whether to route output through a pager. Explicit
// always/never win; auto pages only when stdout is a terminal.
func resolvePager(mode string) (bool, error) {
	switch strings.ToLower(mode) {
	case "always", "true", "on", "yes", "1":
		return true, nil
	case "never", "false", "off", "no", "0":
		return false, nil
	case "auto", "":
		return term.IsTerminal(os.Stdout), nil
	default:
		return false, fmt.Errorf("invalid -pager value %q (want auto, always, or never)", mode)
	}
}

// defaultColor enables color only when stdout is a terminal and the
// environment does not disable it.
func defaultColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return term.IsTerminal(os.Stdout)
}

// resolveWidth picks the wrap width: an explicit flag value wins, then the
// COLUMNS environment variable, then the terminal size, then a fallback.
func resolveWidth(explicit int) int {
	if explicit > 0 {
		return explicit
	}
	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	if _, cols, ok := term.Size(); ok {
		return cols
	}
	return markdown.DefaultWidth
}

func usage() {
	fmt.Fprint(os.Stderr, `md — a terminal Markdown viewer with ASCII Mermaid diagrams

Usage:
  md [flags] [file...]

With no file arguments, Markdown is read from standard input.

Flags:
  -w, -width int    wrap width in columns (0 = auto)
  -color string     color output: auto, always, or never (default "auto")
  -p, -pager string page output through a pager: auto, always, or never (default "auto")
  -version          print version information and exit
  -h, -help         show this help

When stdout is a terminal, output is piped through a pager resolved from
$PAGER, then less, then more. A less pager is invoked as less -FRX unless the
LESS environment variable is set.

Examples:
  md README.md
  md -w 100 notes.md
  cat doc.md | md -color never
  md -p never long-doc.md
`)
}
