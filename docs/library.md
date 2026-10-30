# Go library reference

Import `github.com/ClarifiedLabs/mdcli/markdown` and
`github.com/ClarifiedLabs/mdcli/highlight`. Both use only the standard library;
the module requires Go 1.26+. Public packages are small facades over the same
internal implementations used by the `md` CLI. There is no separate fence parser
or diagram renderer in the CLI adapter.

## Markdown

```go
opts := markdown.Options{
    Enabled: true,
    ANSI: false,
    ColorTheme: highlight.ThemeLight,
    Width: 80,
    Prefix: "> ",
    MermaidMaxVirtualNodes: 1024,
}
text := markdown.Render(source, opts)
```

`Options` fields:

| Field | Behavior |
| --- | --- |
| `Enabled bool` | False leaves input byte-for-byte unchanged, including controls. Rendering is opt-in. |
| `ANSI bool` | Enables terminal emphasis, link styling, and syntax colors. False still normalizes supported Markdown. |
| `ColorTheme highlight.Theme` | Fixed syntax palette: `ThemeDark` (zero/default) or `ThemeLight`. Does not enable ANSI. |
| `Width int` | Positive values wrap prose, list bodies, and tables. Flowcharts use this width minus the prefix's terminal columns; nonpositive uses `DefaultWidth` (80) for flowcharts but leaves prose/tables unwrapped. Other diagrams and ordinary code retain natural width. |
| `Prefix string` | Prepended to nonempty output lines. Diagram space accounts for wide runes, combining marks, and ANSI styling in the prefix. |
| `MermaidMaxVirtualNodes int` | Optional per-diagram layout cap. Zero/negative uses 8192; positive uses the smaller of the value and 8192. Cannot raise the hard maximum. |

`DefaultWidth` and `HorizontalRule` are exported constants. `Render(string,
Options) string` preserves the input's final-newline state. This is a deliberately
small, terminal-friendly Markdown renderer, not a full CommonMark parser.

### Streaming

```go
stream := markdown.NewStream(opts)
// For every incoming text delta:
output := stream.Write(delta)
// At EOF (or a deliberate display boundary):
output += stream.Flush()
```

`NewStream(Options) *Stream` creates independent state. Methods:

- `Write(string) string`: consumes a delta and returns display-ready text. This
  is not `io.Writer`; it does not write to stdout. Complete lines normally render
  immediately, while tables and Mermaid fences wait for layout.
- `Flush() string`: emits pending text, tables, and unclosed Mermaid fences.
  Repeated flushes without new input return an empty string. A mid-fence flush
  renders what is available and treats later body lines as plain source until
  the closing delimiter, rather than starting a second diagram.
- `LineOpen() bool`: the last emitted physical line has no terminating newline.
- `AtLineBoundary() bool`: external scrolling output can be inserted without
  splitting a pending source line or an open display line. A complete buffered
  table/diagram source line can be at a safe boundary.
- `CloseLine()`: records that the caller wrote an external newline.
- `HasBufferedBlock() bool`: a table or Mermaid fence is waiting for layout.
  Do not flush just to repaint a transient status line: that splits layouts.

Streams are mutable and are **not safe for concurrent use**. Independent streams
and `Render` calls can run concurrently, including with different diagram budgets
and themes. No package-global settings are changed.

### Mermaid limits and fallback

Flowcharts adapt their labels and orientation, then use a readable node/connection
list if necessary. Sequence, state, and class diagrams retain their natural
layout. Labels are emitted literally, not reparsed as Markdown. Multiline and
wrapped list-item continuations are indented so they cannot impersonate sections
or other nodes.

All fixed limits in the README's Mermaid safety table apply. The optional virtual
node cap is passed to every adaptive layout candidate and to state/class graph
layout. A safety-limit error falls back to source; it is never hidden by retrying
a smaller layout or using a semantic list. The default remains 8192 virtual nodes;
consumers wanting a tighter bound can explicitly request 1024.

Public streams preserve the incremental renderer's source/newline accounting:
source lines expand tabs, trim trailing spaces, and count consumed line endings.
A Mermaid fence buffers at most 64 KiB of source before switching to plain-source
fallback; oversized incomplete lines stream in UTF-8-safe fragments instead of
remaining buffered indefinitely. Fallbacks preserve fence and final-newline
state. Supported Mermaid output, source fallbacks, labels, and titles are filtered
for terminal and bidi controls, including controls introduced by label decoding.
This is not a general-purpose sanitizer for arbitrary Markdown or code.

The CLI uses the same stream with an internal compatibility policy: nonpositive
width resolves to 80, nonempty documents always end in a newline, and resource
fallbacks include `[mermaid: rendering limit exceeded: ...; showing source]`.
CLI source budgeting retains trailing whitespace and excludes the final body
newline, preserving existing whole-document behavior. These CLI policies are
not extra public options; public library fallbacks have no explanatory notice.

## Syntax and unified diffs

```go
var theme highlight.Theme = highlight.ThemeLight
state := highlight.NewWithTheme("go", theme)
coloredLine := state.Line("func main() {}") // no trailing newline

diff := highlight.NewDiffWithTheme("go", theme)
coloredDiffLine := diff.Line("+func main() {}")
coloredPatch := highlight.ColorizeDiffWithTheme("main.go", patch, theme)
```

Public API:

- `Theme`, `ThemeDark`, `ThemeLight`: fixed palettes; zero/unknown values select dark.
- `State`, `New(info string) *State` (dark),
  `NewWithTheme(info string, theme Theme) *State`, and `(*State).Line(string) string`.
  Unknown languages return nil; a nil state's `Line` leaves input unchanged.
- `DiffState`, `NewDiffWithTheme(lang string, theme Theme) *DiffState`, and
  `(*DiffState).Line(string) string`. Unknown languages still receive diff
  header/addition/removal styling. A nil diff state is a no-op.
- `ColorizeDiffWithTheme(path, text string, theme Theme) string`: selects syntax by
  extension or extensionless basename and preserves newline state.

Syntax state persists across lines (including multiline strings/comments). Diff
state maintains independent old/new syntax lexers and resets them at file headers.
Colors are additive: stripping the inserted ANSI sequences recovers source bytes.
Diff tinting includes erase-to-end-of-line terminal controls for full-row
backgrounds. No TTY/environment detection is performed. Only call highlighters
when terminal styling is wanted; use the original bytes for logs/transcripts.
Instances are not concurrency-safe, but independent instances can run concurrently.
