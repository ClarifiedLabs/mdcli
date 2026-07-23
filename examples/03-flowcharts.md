# Flowcharts

Flowcharts are written as `flowchart` or `graph`, followed by a direction.

## Directions

`TD` (top-down, the default) and `BT` stack ranks vertically:

```mermaid
flowchart TD
  A[Parse] --> B[Analyse] --> C[Emit]
```

`LR` and `RL` lay them out horizontally:

```mermaid
flowchart LR
  A[Parse] --> B[Analyse] --> C[Emit]
```

Bottom-up, useful for dependency graphs where the leaves come first:

```mermaid
flowchart BT
  Binary --> Linker
  Linker --> Objects
  Objects --> Sources
```

## Node shapes

Every shape Mermaid's flowchart syntax offers:

```mermaid
flowchart TD
  rect[Rectangle] --> round(Rounded)
  round --> stadium([Stadium])
  stadium --> sub[[Subroutine]]
  sub --> cyl[(Database)]
  cyl --> circ((Circle))
```

```mermaid
flowchart TD
  diamond{Decision} --> hex{{Hexagon}}
  hex --> asym>Asymmetric]
  asym --> para[/Parallelogram/]
  para --> trap[/Trapezoid\]
```

The slanted shapes come in both leanings — `[/text/]` and `[\text\]` for
parallelograms, `[/text\]` and `[\text/]` for trapezoids:

```mermaid
flowchart LR
  a[/Input/] --> b[\Output\]
  b --> c[/Widen\] --> d[\Narrow/]
```

## Edge styles

Solid, dotted and thick lines, with and without arrowheads:

```mermaid
flowchart LR
  a --> b
  b --- c
  c -.-> d
  d -.- e
  e ==> f
  f === g
```

Alternative arrowheads, and arrows at both ends:

```mermaid
flowchart LR
  cross --x a
  circle --o b
  both <--> c
```

An invisible link (`~~~`) influences layout but draws nothing, which is handy
for nudging nodes into place:

```mermaid
flowchart LR
  visible --> target
  hidden ~~~ target
```

## Edge labels

Labels attach either in pipes after the arrow or inline within it:

```mermaid
flowchart TD
  Start --> Check{Valid?}
  Check -->|yes| Accept[Accept]
  Check -- no --> Reject[Reject]
  Reject -. retry .-> Start
```

## Chains and groups

Several nodes can be linked in one statement, and `&` applies a link to a
whole group at once:

```mermaid
flowchart TD
  Ingest --> Clean & Validate
  Clean & Validate --> Store[(Warehouse)]
  Store --> Report
```

Statements may also be separated with semicolons on a single line:

```mermaid
flowchart LR
  a --> b; b --> c; c --> a
```

## Labels

Quote a label to use characters that would otherwise end it, break lines with
`<br/>`, and write reserved characters as HTML entities:

```mermaid
flowchart TD
  quoted["Contains [brackets] and (parens)"]
  quoted --> multi["First line<br/>second line"]
  multi --> entity["a &amp; b, x &lt; y"]
  entity --> unicode[Ünïcodé ✓ 日本語 한글]
```

Double-width characters are measured in terminal columns, so boxes around CJK
text line up correctly.

## Subgraphs and styling

`subgraph` blocks are flattened — their nodes are drawn, the grouping box is
not. Styling statements (`style`, `classDef`, `class`, `linkStyle`, `click`)
are parsed and ignored, so diagrams copied from a web page still render:

```mermaid
flowchart TD
  Request --> Router

  subgraph Handlers
    Router --> Auth[Auth]
    Router --> Static[Static files]
  end

  Auth --> Response
  Static --> Response

  classDef fast fill:#9f9,stroke:#333
  class Static fast
  style Auth fill:#f99
  click Router href "https://example.com"
```

## Cycles and self-loops

Cycles are broken automatically so ranks can be assigned, and arrows still
point the right way. A node may also link to itself:

```mermaid
flowchart TD
  Poll[Poll queue] --> Work[Do work]
  Work --> Poll
  Work --> Work
  Work --> Done[Done]
```

## Titles

A YAML front matter block sets a title, which is centered above the diagram.
Init directives are stripped:

```mermaid
---
title: Release pipeline
---
%%{init: {'theme':'dark'}}%%
flowchart LR
  Tag --> Build --> Sign --> Publish
```
