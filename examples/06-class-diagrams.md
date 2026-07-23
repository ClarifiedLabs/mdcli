# Class diagrams

Class diagrams start with `classDiagram`. Each class is drawn as a UML box
with compartments for its name, attributes and methods.

## Members

Members can be declared in a block or one at a time with `Class : member`.
Anything containing `()` is treated as a method, everything else as an
attribute:

```mermaid
classDiagram
  class Account {
    +String id
    +Money balance
    +deposit(Money) void
    +withdraw(Money) bool
  }

  class Ledger
  Ledger : +List~Entry~ entries
  Ledger : +append(Entry) void
```

Visibility markers are kept verbatim: `+` public, `-` private, `#` protected,
`~` package.

```mermaid
classDiagram
  class Connection {
    +String host
    -Socket socket
    #int timeout
    +open() void
    -reconnect() void
  }
```

## Generics and annotations

`Type~Param~` is displayed as `Type<Param>`. Annotations in `<< >>` sit above
the class name, and may contain spaces:

```mermaid
classDiagram
  class Repository~T~ {
    <<interface>>
    +find(String) T
    +save(T) void
  }

  class Money {
    <<value object>>
    +long amount
    +String currency
  }

  class Shape {
    <<abstract>>
    +area() double
  }
```

## Relations

Every relation type, with the marker at either end:

```mermaid
classDiagram
  Inheritance <|-- Derived
  Realization <|.. Implementor
  Composition *-- Part
  Aggregation o-- Member
```

```mermaid
classDiagram
  Association --> Target
  Dependency ..> Used
  Link -- Peer
  DashedLink .. Other
```

The markers may sit at either end, so `A <|-- B` and `B --|> A` describe the
same inheritance from opposite directions.

Cardinalities and a label describe the relation:

```mermaid
classDiagram
  Customer "1" --> "0..*" Order : places
  Order "1" *-- "1..*" LineItem : contains
  LineItem "*" --> "1" Product : refers to
```

## A worked model

Namespaces are flattened, so their classes are drawn alongside the rest:

```mermaid
classDiagram
  direction TB

  namespace billing {
    class Invoice {
      +String number
      +Money total
      +issue() void
    }
    class Payment {
      +Money amount
      +apply(Invoice) void
    }
  }

  class PaymentMethod {
    <<abstract>>
    #String token
    +charge(Money) Receipt
  }

  class Card
  class Transfer
  class Receipt

  PaymentMethod <|-- Card
  PaymentMethod <|-- Transfer
  Payment --> Invoice : settles
  Payment o-- PaymentMethod : via
  PaymentMethod ..> Receipt : issues
```

## Ignored statements

`style`, `cssClass`, `link`, `click`, `callback` and notes are parsed and
skipped, so diagrams written for the web renderer still work:

```mermaid
classDiagram
  class Node {
    +String id
  }
  class Edge {
    +Node from
    +Node to
  }
  Edge --> Node : connects

  note "Edges are directed"
  style Node fill:#eef
  cssClass "Node" highlight
  link Node "https://example.com" "docs"
```
