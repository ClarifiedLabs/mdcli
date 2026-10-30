// Package highlight adds terminal ANSI syntax colors without changing source
// bytes. Callers must opt into color themselves; no TTY or environment detection
// is performed here. State and DiffState instances are not concurrency-safe.
package highlight

import internal "github.com/ClarifiedLabs/mdcli/internal/highlight"

// Theme selects a fixed syntax and diff palette. Its zero value is dark.
type Theme = internal.Theme

const (
	ThemeDark  = internal.ThemeDark
	ThemeLight = internal.ThemeLight
)

// State carries multiline syntax state. A nil *State leaves lines unchanged.
type State = internal.State

// New creates a dark-theme highlighter; unknown languages return nil.
func New(info string) *State { return internal.New(info) }

// NewWithTheme creates a highlighter for a fence language or common alias.
func NewWithTheme(info string, theme Theme) *State { return internal.NewWithTheme(info, theme) }

// DiffState tracks independent old/new lexers and unified diff headers.
// Its Line method accepts one line without its newline; a nil state is a no-op.
type DiffState = internal.DiffState

// NewDiffWithTheme creates a diff highlighter for a language. Unknown languages
// still receive diff header and addition/removal colors.
func NewDiffWithTheme(lang string, theme Theme) *DiffState {
	return internal.NewDiffWithTheme(lang, theme)
}

// ColorizeDiffWithTheme colors a unified diff, choosing syntax by path extension
// or extensionless basename and preserving the final newline state.
func ColorizeDiffWithTheme(path, text string, theme Theme) string {
	return internal.ColorizeDiffWithTheme(path, text, theme)
}
