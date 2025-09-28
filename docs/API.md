# API Documentation

## Overview

The golang-yaml-advanced library provides a comprehensive Go API for YAML processing with full comment preservation, advanced merging capabilities, schema validation, streaming support, and powerful transformation DSL. This document covers the complete API surface and usage guidelines.

## Table of Contents

1. [Core API](#core-api)
2. [Advanced Features](#advanced-features)
3. [Schema Validation](#schema-validation)
4. [Streaming Parser](#streaming-parser)
5. [Transform DSL](#transform-dsl)
6. [Query System](#query-system)
7. [Community Usage Guidelines](#community-usage-guidelines)
8. [Performance Guidelines](#performance-guidelines)
9. [Error Handling](#error-handling)
10. [Examples](#examples)

## Core API

### Primary Types

#### NodeTree
The main container for YAML documents.

```go
type NodeTree struct {
    Documents   []*Document
    Current     *Document
    CurrentNode *Node
}

// Primary methods
func UnmarshalYAML(data []byte) (*NodeTree, error)
func (tree *NodeTree) ToYAML() ([]byte, error)
func NewNodeTree() *NodeTree
func (nt *NodeTree) AddDocument() *Document
```

#### Document
Represents a single YAML document within a stream.

```go
type Document struct {
    Root       *Node
    Anchors    map[string]*Node
    Directives []string
    Version    string
}

// Methods
func (doc *Document) ToYAML() ([]byte, error)
func (doc *Document) SetRoot(node *Node)
func (doc *Document) RegisterAnchor(name string, node *Node)
```

#### Node
The fundamental YAML node with complete metadata preservation.

```go
type Node struct {
    Kind        NodeKind
    Value       interface{}
    Tag         string
    Anchor      string
    Style       NodeStyle
    Children    []*Node
    Parent      *Node
    Key         *Node
    HeadComment []string
    LineComment string
    FootComment []string
    Line        int
    Column      int
}

// Constructor methods
func NewNode(kind NodeKind) *Node
func NewScalarNode(value interface{}) *Node
func NewMappingNode() *Node
func NewSequenceNode() *Node

// Core methods
func (n *Node) AddChild(child *Node)
func (n *Node) AddKeyValue(key, value *Node) error
func (n *Node) AddSequenceItem(item *Node) error
func (n *Node) GetMapValue(key string) *Node
func (n *Node) GetSequenceItems() []*Node
func (n *Node) Clone() *Node
func (n *Node) String() string
func (n *Node) IsNull() bool
func (n *Node) Path() string
func (n *Node) Remove() error
func (n *Node) ReplaceWith(replacement *Node) error

// Traversal methods
func (n *Node) Walk(visitor func(*Node) bool)
func (n *Node) Find(predicate func(*Node) bool) *Node
func (n *Node) FindAll(predicate func(*Node) bool) []*Node
```

#### NodeKind
Enumeration of YAML node types.

```go
type NodeKind int

const (
    DocumentNode NodeKind = iota
    SequenceNode
    MappingNode
    ScalarNode
    AliasNode
)

func (k NodeKind) String() string
```

#### NodeStyle
Enumeration of YAML node styles.

```go
type NodeStyle int

const (
    DefaultStyle NodeStyle = iota
    TaggedStyle
    DoubleQuotedStyle
    SingleQuotedStyle
    QuotedStyle
    LiteralStyle
    FoldedStyle
    FlowStyle
)

func (s NodeStyle) String() string
```

### Serialization and Formatting

The library provides intelligent serialization with format preservation:

#### Default Settings
- **Indentation**: 2 spaces (YAML standard)
- **Number Preservation**: Large integers remain as integers (no scientific notation)
- **Empty Lines**: Intelligently preserved before `@schema` comment blocks
- **Comment Preservation**: All comments maintained in their original positions

#### Example
```go
// Large numbers are preserved without scientific notation
yamlContent := `
account_id: 123456789012  # Stays as int64, not 1.23457e+11
aws_account: "012345678901"  # String with leading zeros
port: 8080                 # Small numbers remain as integers
price: 19.99              # Floats preserved as floats
`

tree, _ := yaml.UnmarshalYAML([]byte(yamlContent))
output, _ := tree.ToYAML()  // Uses 2-space indentation, preserves formats
```

### Merging Operations

```go
// Merge two YAML trees
func MergeTrees(base, overlay *NodeTree) *NodeTree

// Merge two nodes
func MergeNodes(base, overlay *Node) *Node

// Merge two documents
func MergeDocuments(base, overlay *Document) *Document
```

### Diff Operations

```go
// Diff result types
type DiffType int

const (
    DiffNone DiffType = iota
    DiffAdded
    DiffRemoved
    DiffModified
    DiffCommentChanged
)

type DiffResult struct {
    Type        DiffType
    Path        string
    OldValue    interface{}
    NewValue    interface{}
    OldComment  []string
    NewComment  []string
    OldNode     *Node
    NewNode     *Node
    Description string
}

// Diff functions
func DiffTrees(oldTree, newTree *NodeTree) []DiffResult
func DiffNodes(oldNode, newNode *Node, path string) []DiffResult
```

## Advanced Features

### Schema Validation

The library provides JSON Schema-compatible validation for YAML documents.

```go
type Schema struct {
    Type                 string             `json:"type,omitempty"`
    Properties           map[string]*Schema `json:"properties,omitempty"`
    Items                *Schema            `json:"items,omitempty"`
    Required             []string           `json:"required,omitempty"`
    Enum                 []interface{}      `json:"enum,omitempty"`
    Pattern              string             `json:"pattern,omitempty"`
    MinLength            *int               `json:"minLength,omitempty"`
    MaxLength            *int               `json:"maxLength,omitempty"`
    Minimum              *float64           `json:"minimum,omitempty"`
    Maximum              *float64           `json:"maximum,omitempty"`
    MinItems             *int               `json:"minItems,omitempty"`
    MaxItems             *int               `json:"maxItems,omitempty"`
    UniqueItems          bool               `json:"uniqueItems,omitempty"`
    Format               string             `json:"format,omitempty"`
    Description          string             `json:"description,omitempty"`
    Default              interface{}        `json:"default,omitempty"`
    AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`
    OneOf                []*Schema          `json:"oneOf,omitempty"`
    AnyOf                []*Schema          `json:"anyOf,omitempty"`
    AllOf                []*Schema          `json:"allOf,omitempty"`
    Not                  *Schema            `json:"not,omitempty"`
}

// Validation
func (s *Schema) Validate(node *Node, path string) []ValidationError

type ValidationError struct {
    Path       string
    Message    string
    SchemaPath string
    Value      interface{}
}
```

#### Supported Formats

- `email`: RFC-compliant email addresses
- `uri`, `url`: Valid URLs with protocol
- `date`: ISO 8601 dates (YYYY-MM-DD)
- `date-time`, `datetime`: ISO 8601 date-times
- `time`: HH:MM:SS format
- `ipv4`: IPv4 addresses
- `ipv6`: IPv6 addresses
- `uuid`: RFC 4122 UUIDs

### Streaming Parser

For processing large YAML files with minimal memory usage.

```go
type StreamParser struct {
    // Internal fields
}

// Constructor
func NewStreamParser(reader io.Reader) *StreamParser

// Methods
func (sp *StreamParser) SetDocumentCallback(callback func(*NodeTree) error)
func (sp *StreamParser) Parse() error
```

Example usage:
```go
file, _ := os.Open("large.yaml")
parser := NewStreamParser(file)

parser.SetDocumentCallback(func(tree *NodeTree) error {
    // Process each document as it's parsed
    fmt.Printf("Processing document with %d nodes\n", countNodes(tree))
    return nil
})

err := parser.Parse()
```

### Transform DSL

A fluent interface for complex YAML transformations.

```go
type TransformDSL struct {
    // Internal fields
}

// Constructor
func NewTransformDSL() *TransformDSL

// Transform methods (all return *TransformDSL for chaining)
func (dsl *TransformDSL) Select(predicate func(*Node) bool) *TransformDSL
func (dsl *TransformDSL) Map(fn func(*Node) *Node) *TransformDSL
func (dsl *TransformDSL) RemoveKey(key string) *TransformDSL
func (dsl *TransformDSL) RenameKey(oldKey, newKey string) *TransformDSL
func (dsl *TransformDSL) SortKeys() *TransformDSL
func (dsl *TransformDSL) AddComment(comment string) *TransformDSL
func (dsl *TransformDSL) Flatten() *TransformDSL
func (dsl *TransformDSL) SetValue(value interface{}) *TransformDSL

// Apply transformations
func (dsl *TransformDSL) Apply(tree *NodeTree) (*NodeTree, error)
```

Example usage:
```go
// Chain multiple transformations
result, err := NewTransformDSL().
    RemoveKey("password").
    RenameKey("username", "user").
    SortKeys().
    Apply(tree)

// Filter nodes with Select
result, err := NewTransformDSL().
    Select(func(node *Node) bool {
        // Keep only production configuration
        return node.Key != nil && node.Key.Value == "production"
    }).
    Apply(tree)
```

### Query System

XPath-like querying for YAML documents.

```go
func Query(node *Node, query string) []*Node
```

Query syntax:
- `/key`: Direct child access
- `/parent/child`: Nested access
- `/array/[0]`: Array index access
- `/*`: Wildcard (all children)
- `/*/nested`: Wildcard in path

Examples:
```go
// Get database host
hosts := Query(root, "/config/database/host")

// Get all items in an array
items := Query(root, "/items/*")

// Get specific array element
firstItem := Query(root, "/items/[0]")
```

## Schema Validation

### Basic Validation

```go
schema := &Schema{
    Type: "object",
    Properties: map[string]*Schema{
        "name": {
            Type:      "string",
            MinLength: intPtr(1),
            MaxLength: intPtr(100),
        },
        "age": {
            Type:    "integer",
            Minimum: float64Ptr(0),
            Maximum: float64Ptr(120),
        },
        "email": {
            Type:   "string",
            Format: "email",
        },
    },
    Required: []string{"name", "email"},
}

errors := schema.Validate(node, "$")
for _, err := range errors {
    fmt.Printf("Validation error at %s: %s\n", err.Path, err.Message)
}
```

### Advanced Validation

```go
// OneOf validation
schema := &Schema{
    OneOf: []*Schema{
        {Type: "string"},
        {Type: "number"},
    },
}

// Pattern validation
schema := &Schema{
    Type:    "string",
    Pattern: "^[A-Z][0-9]+$",
}

// Nested object validation
schema := &Schema{
    Type: "object",
    Properties: map[string]*Schema{
        "address": {
            Type: "object",
            Properties: map[string]*Schema{
                "street": {Type: "string"},
                "city":   {Type: "string"},
                "zip":    {Type: "string", Pattern: "^[0-9]{5}$"},
            },
        },
    },
}
```

## Streaming Parser

### Basic Usage

```go
func processLargeFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    parser := NewStreamParser(file)

    docCount := 0
    parser.SetDocumentCallback(func(tree *NodeTree) error {
        docCount++

        // Process each document
        if tree.Documents[0].Root != nil {
            fmt.Printf("Document %d: %s\n", docCount, tree.Documents[0].Root.String())
        }

        // Return error to stop parsing
        if docCount >= 100 {
            return fmt.Errorf("limit reached")
        }

        return nil
    })

    return parser.Parse()
}
```

### Memory-Efficient Processing

```go
// Process and aggregate without loading entire file
func countKeysInLargeFile(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    totalKeys := 0
    parser := NewStreamParser(file)

    parser.SetDocumentCallback(func(tree *NodeTree) error {
        tree.Documents[0].Root.Walk(func(node *Node) bool {
            if node.Kind == MappingNode {
                totalKeys += len(node.Children) / 2
            }
            return true
        })
        return nil
    })

    err = parser.Parse()
    return totalKeys, err
}
```

## Transform DSL

### Complex Transformations

```go
// Remove sensitive data and reorganize
result, err := NewTransformDSL().
    RemoveKey("password").
    RemoveKey("apiKey").
    RemoveKey("secret").
    RenameKey("username", "user").
    RenameKey("hostname", "host").
    SortKeys().
    Apply(tree)

// Flatten nested configuration
result, err := NewTransformDSL().
    Flatten().  // Converts nested.key to "nested.key"
    Apply(tree)

// Add documentation
result, err := NewTransformDSL().
    AddComment("Generated by build system").
    AddComment("Do not edit manually").
    Apply(tree)
```

### Custom Transformations

```go
// Transform all string values to uppercase
result, err := NewTransformDSL().
    Map(func(node *Node) *Node {
        if node.Kind == ScalarNode {
            if str, ok := node.Value.(string); ok {
                node.Value = strings.ToUpper(str)
            }
        }
        return node
    }).
    Apply(tree)

// Filter configuration by environment
result, err := NewTransformDSL().
    Select(func(node *Node) bool {
        // Keep only production and common settings
        if node.Key != nil {
            key := fmt.Sprintf("%v", node.Key.Value)
            return key == "production" || key == "common"
        }
        return false
    }).
    Apply(tree)
```

## Community Usage Guidelines

### Best Practices

1. **Comment Preservation**: The library automatically preserves all comments. Use this for configuration documentation.

2. **Memory Management**: For large files (>100MB), use the StreamParser instead of loading entire files.

3. **Error Handling**: Always check errors from parsing and transformation operations.

4. **Thread Safety**: Node structures are not thread-safe. Clone nodes before concurrent access.

5. **Performance**: Use batch operations when possible. The Query system is optimized for repeated searches.

### Common Patterns

#### Configuration Management
```go
// Load configuration with environment overlay
base, _ := UnmarshalYAML(defaultConfig)
env, _ := UnmarshalYAML(envConfig)
merged := MergeTrees(base, env)
```

#### Schema-Driven Validation
```go
// Validate before deployment
schema, _ := loadSchemaFromFile("config-schema.json")
config, _ := UnmarshalYAML(configData)
if errors := schema.Validate(config.Documents[0].Root, "$"); len(errors) > 0 {
    return fmt.Errorf("invalid configuration: %v", errors)
}
```

#### Safe Transformations
```go
// Remove sensitive data before logging
sanitized, _ := NewTransformDSL().
    RemoveKey("password").
    RemoveKey("token").
    RemoveKey("secret").
    Apply(tree)
```

## Performance Guidelines

### Memory Optimization

- Use StreamParser for files larger than 100MB
- Clone nodes only when necessary
- Clear unused references to allow garbage collection

### CPU Optimization

- Use Query for repeated searches instead of Walk
- Batch transformations with TransformDSL
- Avoid repeated parsing of the same content

### Benchmarks

```go
// Parsing performance
BenchmarkUnmarshalYAML-8         50000      30567 ns/op
BenchmarkToYAML-8               100000      15234 ns/op

// Query performance
BenchmarkQuery-8               1000000       1045 ns/op
BenchmarkWalk-8                 500000       3421 ns/op

// Transform performance
BenchmarkTransformDSL-8          30000      45123 ns/op
BenchmarkMergeTrees-8            50000      28456 ns/op
```

## Error Handling

### Parse Errors

```go
tree, err := UnmarshalYAML(data)
if err != nil {
    // Handle parsing errors
    if syntaxErr, ok := err.(*yaml.SyntaxError); ok {
        fmt.Printf("Syntax error at line %d: %s\n", syntaxErr.Line, syntaxErr.Message)
    }
    return err
}
```

### Validation Errors

```go
errors := schema.Validate(node, "$")
if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Validation failed at %s: %s (value: %v)\n",
            err.Path, err.Message, err.Value)
    }
    return fmt.Errorf("%d validation errors", len(errors))
}
```

### Transform Errors

```go
result, err := dsl.Apply(tree)
if err != nil {
    // Transform errors include context
    fmt.Printf("Transform failed: %v\n", err)
    return err
}
```

## Examples

### Complete Example: Config Processing Pipeline

```go
package main

import (
    "fmt"
    "os"
    yaml "github.com/elioetibr/golang-yaml-advanced"
)

func processConfig(filename string) error {
    // 1. Load configuration
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }

    tree, err := yaml.UnmarshalYAML(data)
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    // 2. Validate against schema
    schema := &yaml.Schema{
        Type: "object",
        Properties: map[string]*yaml.Schema{
            "version": {Type: "string"},
            "services": {
                Type: "object",
                AdditionalProperties: &yaml.Schema{
                    Type: "object",
                    Properties: map[string]*yaml.Schema{
                        "port": {Type: "integer", Minimum: float64Ptr(1), Maximum: float64Ptr(65535)},
                        "host": {Type: "string", Format: "uri"},
                    },
                },
            },
        },
        Required: []string{"version", "services"},
    }

    if errors := schema.Validate(tree.Documents[0].Root, "$"); len(errors) > 0 {
        return fmt.Errorf("validation failed: %v", errors)
    }

    // 3. Transform configuration
    transformed, err := yaml.NewTransformDSL().
        RemoveKey("debug").           // Remove debug settings
        RenameKey("srv", "services"). // Normalize naming
        SortKeys().                   // Consistent ordering
        Apply(tree)

    if err != nil {
        return fmt.Errorf("transform error: %w", err)
    }

    // 4. Query specific values
    ports := yaml.Query(transformed.Documents[0].Root, "/services/*/port")
    for _, port := range ports {
        fmt.Printf("Service port: %v\n", port.Value)
    }

    // 5. Save processed configuration
    output, err := transformed.ToYAML()
    if err != nil {
        return fmt.Errorf("serialization error: %w", err)
    }

    return os.WriteFile("processed-"+filename, output, 0644)
}

func float64Ptr(v float64) *float64 { return &v }
```

### Stream Processing Example

```go
func analyzeYAMLStream(reader io.Reader) (*Statistics, error) {
    stats := &Statistics{
        Documents: 0,
        Keys:      make(map[string]int),
    }

    parser := yaml.NewStreamParser(reader)
    parser.SetDocumentCallback(func(tree *yaml.NodeTree) error {
        stats.Documents++

        // Analyze document structure
        tree.Documents[0].Root.Walk(func(node *yaml.Node) bool {
            if node.Kind == yaml.MappingNode {
                for i := 0; i < len(node.Children)-1; i += 2 {
                    if key := node.Children[i]; key.Kind == yaml.ScalarNode {
                        keyStr := fmt.Sprintf("%v", key.Value)
                        stats.Keys[keyStr]++
                    }
                }
            }
            return true
        })

        return nil
    })

    err := parser.Parse()
    return stats, err
}
```

## Migration Guide

### From Standard Library

```go
// Before (using gopkg.in/yaml.v3)
var data map[string]interface{}
err := yaml.Unmarshal(yamlBytes, &data)

// After (using this library)
tree, err := yaml.UnmarshalYAML(yamlBytes)
// Access data through tree.Documents[0].Root
```

### From Other YAML Libraries

The library provides compatibility through the standard `yaml.Node` conversion:

```go
// Convert from yaml.Node
yamlNode := &yaml.Node{} // from gopkg.in/yaml.v3
node := yaml.ConvertFromYAMLNode(yamlNode)

// Convert to yaml.Node
node := tree.Documents[0].Root
yamlNode := node.ToYAMLNode()
```

## Support

For issues, feature requests, and contributions, please visit:
https://github.com/elioetibr/golang-yaml-advanced