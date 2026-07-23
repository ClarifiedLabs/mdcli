# Sequence diagrams

Sequence diagrams start with `sequenceDiagram`. Participants appear in the
order they are first mentioned, with a header and footer box on each lifeline.

## Participants

Declare participants explicitly to fix their order, and use `as` to give a
lifeline a friendlier label than its identifier:

```mermaid
sequenceDiagram
  participant C as Client
  participant S as Server
  C->>S: GET /health
  S-->>C: 200 OK
```

`actor` is accepted as a synonym, and `box ... end` groups are parsed and
flattened:

```mermaid
sequenceDiagram
  box Frontend
    actor U as User
    participant W as Web app
  end
  box Backend
    participant A as API
  end
  U->>W: click "Save"
  W->>A: PUT /document
  A-->>W: 204
  W-->>U: saved
```

## Arrow forms

Every arrow Mermaid defines, solid and dashed:

```mermaid
sequenceDiagram
  participant A
  participant B
  A->B: solid, no head
  A->>B: solid arrow
  A-->B: dashed, no head
  A-->>B: dashed arrow
  A-xB: solid cross
  A--xB: dashed cross
  A-)B: solid async
  A--)B: dashed async
  A<<->>B: bidirectional
```

A message from a participant to itself is drawn as a loop off the lifeline:

```mermaid
sequenceDiagram
  participant Cache
  Cache->>Cache: evict expired keys
```

## Numbering

`autonumber` prefixes each message with a sequence number:

```mermaid
sequenceDiagram
  autonumber
  participant B as Browser
  participant S as Server
  B->>S: POST /login
  S-->>B: set-cookie
  B->>S: GET /profile
  S-->>B: 200 profile
```

## Blocks

`loop`, `alt`/`else`, `opt`, `par`/`and`, `critical`/`option`, `break` and
`rect` all draw a labelled frame around the messages they contain:

```mermaid
sequenceDiagram
  participant C as Client
  participant S as Server

  loop every 30s
    C->>S: heartbeat
    alt healthy
      S-->>C: pong
    else degraded
      S-->>C: 503
      opt retry budget left
        C->>S: heartbeat
      end
    end
  end
```

Frames nest, and `par` runs branches side by side:

```mermaid
sequenceDiagram
  participant O as Orchestrator
  participant I as Inventory
  participant P as Payments

  par reserve stock
    O->>I: hold items
    I-->>O: held
  and take payment
    O->>P: charge
    P-->>O: receipt
  end

  critical commit order
    O->>I: confirm
  option inventory gone
    O->>P: refund
    break order abandoned
      O->>O: emit alert
    end
  end
```

## Notes

Notes sit beside a single lifeline or span a range of them:

```mermaid
sequenceDiagram
  participant A as Producer
  participant B as Broker
  participant C as Consumer

  Note left of A: batches of 100
  A->>B: publish
  Note over B: durable log
  B->>C: deliver
  Note right of C: at-least-once
  Note over A,C: end-to-end acknowledged
```

## Activations and lifecycle

`activate`/`deactivate`, `create` and `destroy` are accepted so that diagrams
copied from elsewhere parse cleanly. Lifelines are drawn for the full height
of the diagram:

```mermaid
sequenceDiagram
  participant S as Supervisor
  S->>S: start
  activate S
  create participant W as Worker
  S->>W: spawn
  W-->>S: ready
  deactivate S
  destroy W
  S-xW: shut down
```

## Titles

A `title` line, or a YAML front matter title, is centered above the diagram:

```mermaid
sequenceDiagram
  title Token refresh
  participant A as App
  participant I as Identity
  A->>I: refresh_token
  I-->>A: access_token
```
