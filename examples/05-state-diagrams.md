# State diagrams

Both `stateDiagram` and `stateDiagram-v2` are supported. `[*]` marks the start
and end of a region, drawn as `(*)` and `(O)`.

## Transitions

```mermaid
stateDiagram-v2
  [*] --> Idle
  Idle --> Running: start
  Running --> Idle: stop
  Running --> [*]: exit
```

Direction works the same as in flowcharts:

```mermaid
stateDiagram-v2
  direction LR
  [*] --> Queued
  Queued --> Active: dequeue
  Active --> Done: finish
  Done --> [*]
```

## Naming and descriptions

Quote a name and bind it with `as` when the label is not a valid identifier.
A `State: text` line adds a description below the name:

```mermaid
stateDiagram-v2
  state "Waiting for input" as wait
  [*] --> wait
  wait --> Parsing: line read
  Parsing --> wait: incomplete
  Parsing --> Ready: complete
  Parsing: builds the AST
  Ready --> [*]
```

## Pseudo-states

`<<choice>>` renders as a diamond; `<<fork>>` and `<<join>>` render as bars:

```mermaid
stateDiagram-v2
  [*] --> Submitted
  state check <<choice>>
  Submitted --> check
  check --> Approved: score high
  check --> Rejected: score low
  Approved --> [*]
  Rejected --> [*]
```

Fork and join split and rejoin concurrent work:

```mermaid
stateDiagram-v2
  state split <<fork>>
  state merge <<join>>
  [*] --> split
  split --> Encoding
  split --> Thumbnails
  Encoding --> merge
  Thumbnails --> merge
  merge --> Published
```

## Composite states

Composite states are flattened — their children are drawn as ordinary states.
Each composite keeps its own start and end markers, so a nested `[*]` belongs
to its own region rather than the diagram as a whole:

```mermaid
stateDiagram-v2
  [*] --> Connection

  state Connection {
    [*] --> Handshake
    Handshake --> Open: accepted
    Handshake --> Closed: refused
    Open --> Closed: hang up
  }

  Connection --> [*]
```

Composites nest, and `--` separates concurrent regions within one composite —
each region gets its own start marker:

```mermaid
stateDiagram-v2
  state Session {
    [*] --> Transfer
    Transfer --> Transfer: chunk
    --
    [*] --> Keepalive
    Keepalive --> Keepalive: ping
  }
```

## Self-transitions and notes

A state may transition to itself. Notes are parsed and omitted from the
drawing, in both the one-line and block forms:

```mermaid
stateDiagram-v2
  [*] --> Polling
  Polling --> Polling: no work
  Polling --> Working: job found
  Working --> Polling: done

  note right of Polling: backs off exponentially
  note left of Working
    Runs at most one job
    at a time.
  end note
```

## Styling

`classDef`, `class` and `style` statements are parsed and ignored:

```mermaid
stateDiagram-v2
  [*] --> Healthy
  Healthy --> Degraded: probe failed
  Degraded --> Healthy: probe passed
  Degraded --> Down: 3 failures
  Down --> [*]

  classDef bad fill:#f99,stroke:#900
  class Down bad
```
