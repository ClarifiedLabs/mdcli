# Document structure

Block-level Markdown: headings, lists, tables, rules and code blocks.

## Headings

All six levels are recognised. The `#` markers are kept so the outline stays
visible in plain text; level 1 is bold and underlined, the rest are bold.

## Level 2

### Level 3

#### Level 4

##### Level 5

###### Level 6

Up to three spaces of indentation are allowed before the `#`. A `#` with no
following space is not a heading, so #hashtag stays as written.

## Paragraphs and wrapping

Paragraphs are wrapped to the terminal width, or to an explicit `-w`. This
paragraph is deliberately long so that it demonstrates the behaviour: words
are moved to the next line rather than being split, and the wrap point is
recalculated for every line, which keeps prose readable at any width without
reflowing the source file itself.

Wrapping counts visible characters, so styled text wraps at the same place
whether or not color is enabled.

## Lists

Bullets accept `-`, `*` or `+` and are normalised to a single marker:

- a dash bullet
* a star bullet
+ a plus bullet

Ordered lists keep their own numbering and both delimiters:

1. first item
2. second item
3) third item, using a paren

Nesting follows indentation:

- top level
  - one level in
    - two levels in
  - back out again
- another top-level item

Long list items wrap with a hanging indent, so continuation lines line up
under the text rather than under the marker:

- This item is long enough to wrap at a typical terminal width, and its
  second line is indented to match the first, which keeps the list readable.
- A short item.

## Horizontal rules

Three or more `-`, `*` or `_` on their own line become a rule:

---

Spaces between the markers are allowed, so `- - -` and `***` work too:

***

## Tables

A header row, a separator row and any number of body rows. Column widths are
computed from the content, and cells may contain inline formatting:

| Flag | Type | Default | Notes |
| --- | --- | --- | --- |
| `-w`, `-width` | int | `0` | Wrap width; `0` means auto |
| `-color` | string | `auto` | One of `auto`, `always`, `never` |
| `-p`, `-pager` | string | `auto` | Page through **$PAGER** |
| `-h`, `-help` | bool | `false` | Show help and exit |

Alignment is taken from the separator row — `:---` left, `:---:` center and
`---:` right:

| Item | Qty | Unit price |
| :--- | :---: | ---: |
| Widget | 2 | 9.99 |
| Grommet | 14 | 0.45 |
| Flange assembly | 1 | 129.00 |

Each separator cell needs at least three dashes, so write `:---:` rather than
the shorter `:-:` that some renderers accept.

Rows with fewer cells than the header are padded out:

| A | B | C |
| --- | --- | --- |
| full | row | here |
| short | row |
| x |

A block of pipe-separated lines without a valid separator row is not a table,
and is rendered as ordinary paragraph text instead.

## Code blocks

Fenced blocks are passed through verbatim and indented, with no wrapping or
inline parsing, so the contents stay exactly as written:

```go
func Render(text string, opts Options) string {
	if !opts.Enabled || text == "" {
		return text // **not bold**, https://not-a-link.example
	}
	return NewStream(opts).Write(text)
}
```

A language tag is optional:

```
$ md -w 100 examples/02-document-structure.md
```

Tilde fences work the same way, which is handy when the block itself contains
backticks:

~~~
Use `md file.md` to render a document.
~~~

Tabs inside a block are expanded to four spaces so that indentation lines up
consistently in the terminal.
