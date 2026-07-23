# Text formatting

Inline Markdown is rendered with terminal styling. With `-color never` the
markers are stripped and the text is left plain, so output stays useful in a
pipe or a file.

## Emphasis

Both marker styles work, and they nest:

- `*single asterisks*` → *italic*
- `_single underscores_` → _italic_
- `**double asterisks**` → **bold**
- `__double underscores__` → __bold__
- `***triple***` → ***bold italic***
- `___triple underscores___` → ___bold italic___

Emphasis can wrap other inline markup, so **bold text with `code` inside** and
*italic containing a [link](https://example.com)* both render as you would
expect.

Underscore markers only apply at a word boundary, so identifiers with interior
underscores survive intact: snake_case_name and a_b_c are left alone.

An identifier that *starts* and *ends* with underscores is a boundary case —
`__init__` written bare does become bold, because both markers sit against
whitespace. Wrap such names in backticks to keep them literal.

A marker with nothing between it, or with a space just inside it, is not
emphasis: `* not italic *` and `**` stay literal.

## Inline code

Backticks mark code spans: `md -w 100 file.md`, `os.ReadFile`, `--color`.

Code spans are not parsed further, so `**not bold**` and `https://x.example`
inside backticks stay literal.

## Links

A labelled link shows its text followed by the target, so the URL is still
visible and copyable in a terminal:

- [the project README](https://github.com/ClarifiedLabs/mdcli)
- [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110)

Bare URLs are picked up automatically — https://example.com/docs — including
`mailto:` addresses like mailto:someone@example.com.

Trailing punctuation is left out of the link, so a sentence ending in a URL
still reads correctly: see https://example.com/guide. The final full stop is
not part of the address.

When a link's label is the same as its target, or empty, only the target is
shown: [https://example.com](https://example.com).

## What is not styled

A few common extensions are not supported and pass through as written:

| Syntax | Renders as |
| --- | --- |
| `~~strikethrough~~` | literal `~~strikethrough~~` |
| `![alt](url)` | `!alt <url>` — images are not special |
| `> blockquote` | a plain paragraph, marker included |
| `[ref][1]` | literal text — reference links are not resolved |

Raw HTML is passed through untouched: <span class="x">like this</span>.
