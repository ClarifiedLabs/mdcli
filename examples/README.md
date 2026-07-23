# md examples

Markdown documents that exercise everything `md` renders — text formatting,
document structure, syntax highlighted code, and the four Mermaid diagram
types. Each file is a readable document in its own right, so view one with:

```sh
md examples/01-text-formatting.md
```

Or page through all of them at once:

```sh
cat examples/*.md | md
```

## Markdown

| File | Covers |
| --- | --- |
| `01-text-formatting.md` | Emphasis, inline code, links, bare URLs, what is not styled |
| `02-document-structure.md` | Headings, lists, tables, rules, code blocks, wrapping |
| `09-code-blocks.md` | Syntax highlighting, language tags, untagged fences |

## Mermaid diagrams

| File | Covers |
| --- | --- |
| `03-flowcharts.md` | Directions, node shapes, edge styles, labels, subgraphs |
| `04-sequence-diagrams.md` | Participants, arrows, nested blocks, notes |
| `05-state-diagrams.md` | Composite states, concurrent regions, fork/join |
| `06-class-diagrams.md` | Members, generics, annotations, relation types |

## Putting it together

| File | Covers |
| --- | --- |
| `07-architecture-review.md` | Prose, tables and diagrams in one document |
| `08-fallbacks.md` | Unsupported and malformed diagrams |

## Rendering notes

Everything is laid out for a terminal around 100 columns wide. If a line
wraps, widen the terminal or pass an explicit width:

```sh
md -w 120 examples/03-flowcharts.md
```

To capture the output as plain text, turn off color and paging. Markdown
markers are stripped rather than styled, and diagrams are unaffected:

```sh
md -color never -p never examples/02-document-structure.md > out.txt
```
