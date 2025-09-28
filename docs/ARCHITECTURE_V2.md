# Architecture Design: Independent YAML Library v2.0

## Overview

This document describes the architecture for the next-generation golang-yaml-advanced library, removing all external dependencies while maintaining YAML 1.2.2 compliance.

## Design Principles

### SOLID Principles Applied

#### Single Responsibility Principle
Each component has one clear responsibility:
- **Scanner**: Reads bytes and produces runes
- **Tokenizer**: Converts runes to tokens
- **Parser**: Converts tokens to events
- **Builder**: Converts events to nodes
- **Emitter**: Converts nodes to YAML text

#### Open/Closed Principle
```go
// Extensible through interfaces, not modification
type Scanner interface {
    Next() (rune, error)
    Peek(n int) ([]rune, error)
    Position() Position
}

type TokenFilter interface {
    Filter(Token) (Token, bool)
}

type NodeTransformer interface {
    Transform(Node) (Node, error)
}
```

#### Liskov Substitution Principle
```go
// All node types are substitutable
type Node interface {
    Type() NodeType
    Value() interface{}
    Style() Style
    Comments() Comments
    Position() Position
}

type ScalarNode struct { /* implements Node */ }
type SequenceNode struct { /* implements Node */ }
type MappingNode struct { /* implements Node */ }
```

#### Interface Segregation Principle
```go
// Focused interfaces for specific capabilities
type Reader interface {
    Read() (Event, error)
}

type Writer interface {
    Write(Event) error
}

type Validator interface {
    Validate(Node) []ValidationError
}

type Queryable interface {
    Query(path string) []Node
}
```

#### Dependency Inversion Principle
```go
// High-level modules depend on abstractions
type Parser struct {
    scanner Scanner // interface, not concrete
    emitter EventEmitter // interface
    config  Config // interface
}

// Dependency injection through constructors
func NewParser(scanner Scanner, opts ...ParserOption) *Parser {
    p := &Parser{
        scanner: scanner,
        emitter: NewEventEmitter(),
    }
    for _, opt := range opts {
        opt(p)
    }
    return p
}
```

## Layered Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    API Layer                            │
│  Marshal/Unmarshal | Load/Dump | Transform | Query      │
├─────────────────────────────────────────────────────────┤
│                 Presentation Layer                      │
│    Emitter | Formatter | StyleEngine | Comments         │
├─────────────────────────────────────────────────────────┤
│                  Business Layer                         │
│   NodeTree | Schema | Validation | Anchors | Tags       │
├─────────────────────────────────────────────────────────┤
│                   Core Layer                            │
│    Parser | EventStream | NodeBuilder | TypeResolver    │
├─────────────────────────────────────────────────────────┤
│                Infrastructure Layer                     │
│   Scanner | Tokenizer | Buffer | Position | Errors      │
└─────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Scanner Component

```go
package scanner

// Configuration using functional options
type Option func(*Scanner)

func WithTabWidth(width int) Option {
    return func(s *Scanner) {
        s.tabWidth = width
    }
}

// Scanner with dependency injection
type Scanner struct {
    reader   io.Reader
    buffer   *bufio.Reader
    position Position
    tabWidth int
}

func New(r io.Reader, opts ...Option) *Scanner {
    s := &Scanner{
        reader:   r,
        buffer:   bufio.NewReader(r),
        tabWidth: 8,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### 2. Token System

```go
package token

// Token types as iota constants (no strings)
type Type uint16

const (
    Invalid Type = iota
    EOF

    // Scalars
    PlainScalar
    SingleQuoted
    DoubleQuoted
    Literal
    Folded

    // Structure
    SequenceStart
    SequenceEnd
    MappingStart
    MappingEnd

    // Markers
    DocumentStart
    DocumentEnd
    Anchor
    Alias
    Tag
)

// Immutable token
type Token struct {
    typ      Type
    value    string
    position Position
    style    Style
}

// Builder pattern for complex tokens
type TokenBuilder struct {
    token Token
}

func (b *TokenBuilder) Type(t Type) *TokenBuilder {
    b.token.typ = t
    return b
}

func (b *TokenBuilder) Build() Token {
    return b.token
}
```

### 3. Event-Driven Parser

```go
package parser

// Event types for streaming
type EventType uint8

const (
    StreamStart EventType = iota
    StreamEnd
    DocumentStart
    DocumentEnd
    MappingStart
    MappingEnd
    SequenceStart
    SequenceEnd
    Scalar
    Alias
)

// Immutable event
type Event struct {
    Type     EventType
    Value    interface{}
    Style    Style
    Anchor   string
    Tag      string
    Position Position
}

// Pull parser interface
type PullParser interface {
    Next() (Event, error)
    Peek() (Event, error)
}

// Push parser with observer pattern
type PushParser interface {
    Parse(io.Reader) error
    OnEvent(EventHandler)
}

type EventHandler func(Event) error
```

### 4. Node Tree with Visitor Pattern

```go
package node

// Visitor pattern for traversal
type Visitor interface {
    VisitScalar(*ScalarNode) error
    VisitSequence(*SequenceNode) error
    VisitMapping(*MappingNode) error
    VisitDocument(*DocumentNode) error
}

// Base node with composition
type baseNode struct {
    position  Position
    comments  Comments
    style     Style
    anchor    string
    tag       string
}

// Concrete nodes embed base
type ScalarNode struct {
    baseNode
    value string
}

func (n *ScalarNode) Accept(v Visitor) error {
    return v.VisitScalar(n)
}

// Builder with method chaining
type NodeBuilder struct {
    node Node
}

func NewScalar(value string) *NodeBuilder {
    return &NodeBuilder{
        node: &ScalarNode{value: value},
    }
}

func (b *NodeBuilder) WithStyle(s Style) *NodeBuilder {
    // Set style
    return b
}

func (b *NodeBuilder) Build() Node {
    return b.node
}
```

### 5. Schema System with Strategy Pattern

```go
package schema

// Strategy pattern for type resolution
type Resolver interface {
    Resolve(string) (interface{}, error)
}

type Schema interface {
    Name() string
    Resolver(tag string) Resolver
}

// Concrete schemas
type CoreSchema struct{}
type JSONSchema struct{}
type FailsafeSchema struct{}

// Registry pattern
type Registry struct {
    schemas map[string]Schema
    current Schema
}

func (r *Registry) Register(name string, schema Schema) {
    r.schemas[name] = schema
}

func (r *Registry) Use(name string) error {
    schema, ok := r.schemas[name]
    if !ok {
        return ErrSchemaNotFound
    }
    r.current = schema
    return nil
}
```

### 6. Transformation Pipeline

```go
package transform

// Pipeline pattern for transformations
type Transformer func(Node) (Node, error)

type Pipeline struct {
    transformers []Transformer
}

func (p *Pipeline) Add(t Transformer) *Pipeline {
    p.transformers = append(p.transformers, t)
    return p
}

func (p *Pipeline) Transform(n Node) (Node, error) {
    var err error
    for _, t := range p.transformers {
        n, err = t(n)
        if err != nil {
            return nil, err
        }
    }
    return n, nil
}

// Predefined transformers
func RemoveComments() Transformer {
    return func(n Node) (Node, error) {
        // Implementation
        return n, nil
    }
}

func SortKeys() Transformer {
    return func(n Node) (Node, error) {
        // Implementation
        return n, nil
    }
}
```

### 7. Query Engine

```go
package query

// Query interface with builder pattern
type Query interface {
    Select(path string) Query
    Where(predicate Predicate) Query
    Execute() []Node
}

type Predicate func(Node) bool

type queryBuilder struct {
    root       Node
    path       string
    predicates []Predicate
}

func From(n Node) Query {
    return &queryBuilder{root: n}
}

func (q *queryBuilder) Select(path string) Query {
    q.path = path
    return q
}

func (q *queryBuilder) Where(p Predicate) Query {
    q.predicates = append(q.predicates, p)
    return q
}
```

## Error Handling with Must Pattern

```go
package yaml

// Must pattern for critical operations
func MustParse(input string) *NodeTree {
    tree, err := Parse(input)
    if err != nil {
        panic(err)
    }
    return tree
}

// Error types with context
type Error struct {
    Type     ErrorType
    Message  string
    Position Position
    Context  string
}

func (e Error) Error() string {
    return fmt.Sprintf("%s at %s: %s",
        e.Type, e.Position, e.Message)
}

// Error collection for validation
type Errors []Error

func (e Errors) Error() string {
    // Aggregate errors
}
```

## Dependency Injection Container

```go
package yaml

// Container for dependency injection
type Container struct {
    scanner  Scanner
    parser   Parser
    emitter  Emitter
    schema   Schema
}

// Builder pattern for configuration
type Builder struct {
    container Container
}

func NewBuilder() *Builder {
    return &Builder{
        container: Container{
            // Defaults
        },
    }
}

func (b *Builder) WithScanner(s Scanner) *Builder {
    b.container.scanner = s
    return b
}

func (b *Builder) Build() *YAML {
    return &YAML{
        container: b.container,
    }
}
```

## Testing Strategy

### Unit Testing with Mocks

```go
// Mock scanner for testing
type MockScanner struct {
    mock.Mock
}

func (m *MockScanner) Next() (rune, error) {
    args := m.Called()
    return args.Get(0).(rune), args.Error(1)
}

// Table-driven tests
func TestParser(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []Event
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Testing

```go
// YAML test suite integration
func TestYAMLTestSuite(t *testing.T) {
    files, _ := filepath.Glob("yaml-test-suite/*.yaml")
    for _, file := range files {
        t.Run(file, func(t *testing.T) {
            // Parse and validate
        })
    }
}
```

## Performance Considerations

### Memory Management
- Object pooling for frequently allocated objects
- Buffer reuse for string building
- Lazy evaluation for aliases
- Streaming for large documents

### Optimization Strategies
```go
// Object pool for nodes
var nodePool = sync.Pool{
    New: func() interface{} {
        return &ScalarNode{}
    },
}

// String builder pool
var builderPool = sync.Pool{
    New: func() interface{} {
        return new(strings.Builder)
    },
}
```

## Migration Path

### Compatibility Layer

```go
package compat

import (
    "gopkg.in/yaml.v3"
    newyaml "github.com/elioetibr/golang-yaml-advanced/v2"
)

// Wrapper for backward compatibility
type Node struct {
    *yaml.Node
    internal *newyaml.Node
}

// Adapter pattern
func Unmarshal(data []byte, v interface{}) error {
    // Convert to new API
    tree, err := newyaml.Parse(data)
    if err != nil {
        return err
    }
    return tree.Decode(v)
}
```

## Package Structure

```
github.com/elioetibr/golang-yaml-advanced/v2/
├── yaml.go           # Main API
├── scanner/          # Lexical scanning
│   ├── scanner.go
│   ├── buffer.go
│   └── position.go
├── token/            # Token definitions
│   ├── token.go
│   └── types.go
├── parser/           # Event parser
│   ├── parser.go
│   ├── events.go
│   └── state.go
├── node/             # AST nodes
│   ├── node.go
│   ├── builder.go
│   └── visitor.go
├── schema/           # Type system
│   ├── schema.go
│   ├── core.go
│   ├── json.go
│   └── failsafe.go
├── emitter/          # YAML output
│   ├── emitter.go
│   ├── formatter.go
│   └── style.go
├── transform/        # Transformations
│   ├── pipeline.go
│   └── transformers.go
├── query/            # Query engine
│   ├── query.go
│   └── xpath.go
├── validate/         # Validation
│   ├── validator.go
│   └── jsonschema.go
├── compat/           # Compatibility
│   └── yaml_v3.go
└── internal/         # Internal utilities
    ├── pool/
    ├── utf8/
    └── errors/
```

---

*This architecture ensures:*
- **Zero external dependencies**
- **Clean, testable design**
- **High performance**
- **Easy migration path**
- **Full YAML 1.2.2 compliance**
- **Extensibility through interfaces**