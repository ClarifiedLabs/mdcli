# Code blocks

Fenced code blocks are indented and left otherwise intact. When the fence
carries a language tag and color is on, the code is syntax highlighted. With
`-color never` — or piped to a file — the same block comes out as plain text,
byte for byte.

Highlighting only ever adds color. It never rewrites, reflows, or reindents
what you wrote, so code stays copy-pasteable straight out of the terminal.

## A tagged fence

```go
// Package greet says hello.
package greet

import "fmt"

// Greet writes a greeting for each name.
func Greet(names []string) {
	const limit = 0x1F
	for i, name := range names {
		/* Block comments span
		   as many lines as they need. */
		fmt.Printf("%d: hello, %s\n", i+1, name)
	}
}
```

Comments, strings, numbers, keywords, built-in types and call sites each get
their own color, drawn from the terminal's own palette so the result suits
whatever theme you use.

## Other languages

The tag is matched loosely: case is ignored, common aliases work (`js`, `py`,
`sh`, `c++`, `yml`), and trailing decoration such as `go:main.go` or
`{.python}` is discarded.

```python
# Sum the positive values.
def total(items: list[int]) -> int:
    """Docstrings, like block comments,
    keep their styling across lines."""
    return sum(x for x in items if x > 0)
```

```sh
#!/usr/bin/env bash
set -euo pipefail

for file in "$@"; do
  echo "processing ${file}"   # an apostrophe here doesn't break anything
done
```

```yaml
name: md
version: 1.2
args:
  - --color=always
env:
  URL: "http://example.com/#not-a-comment"   # this one is a comment
```

```diff
--- a/main.go
+++ b/main.go
@@ -1,4 +1,5 @@
 package main
-func old() {}
+func new() {}
```

Roughly two dozen languages are recognized, including Go, Rust, C, C++, Java,
Kotlin, Swift, C#, JavaScript, TypeScript, PHP, Python, Ruby, shell, SQL, Lua,
JSON, YAML, TOML, HTML, CSS, Dockerfile, Makefile, and unified diffs.

## Untagged fences

A fence with no language, or one naming a language that is not recognized,
is printed exactly as written:

```
this block is not highlighted
neither is *this* or `this`
```

Markdown inside a fence is never interpreted, tagged or not.

## Inline code

Inline spans are a separate thing and are styled as `one piece`, with no
attempt to read the language: `md -w 100 file.md`, `os.ReadFile`.
