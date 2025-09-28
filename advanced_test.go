package golang_yaml_advanced

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test Schema Validation
func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name      string
		schema    *Schema
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid object",
			schema: &Schema{
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
						Maximum: float64Ptr(150),
					},
				},
				Required: []string{"name"},
			},
			input: `
name: "John Doe"
age: 30`,
			wantError: false,
		},
		{
			name: "missing required field",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {Type: "string"},
				},
				Required: []string{"name"},
			},
			input:     `age: 30`,
			wantError: true,
			errorMsg:  "required",
		},
		{
			name: "string validation",
			schema: &Schema{
				Type:      "string",
				MinLength: intPtr(5),
				MaxLength: intPtr(10),
				Pattern:   "^[a-z]+$",
			},
			input:     `"hello"`,
			wantError: false,
		},
		{
			name: "string too short",
			schema: &Schema{
				Type:      "string",
				MinLength: intPtr(5),
			},
			input:     `"hi"`,
			wantError: true,
			errorMsg:  "length",
		},
		{
			name: "string pattern mismatch",
			schema: &Schema{
				Type:    "string",
				Pattern: "^[0-9]+$",
			},
			input:     `"abc"`,
			wantError: true,
			errorMsg:  "pattern",
		},
		{
			name: "email format",
			schema: &Schema{
				Type:   "string",
				Format: "email",
			},
			input:     `"test@example.com"`,
			wantError: false,
		},
		{
			name: "invalid email",
			schema: &Schema{
				Type:   "string",
				Format: "email",
			},
			input:     `"not-an-email"`,
			wantError: true,
			errorMsg:  "email",
		},
		{
			name: "integer validation",
			schema: &Schema{
				Type:    "integer",
				Minimum: float64Ptr(10),
				Maximum: float64Ptr(100),
			},
			input:     `50`,
			wantError: false,
		},
		{
			name: "integer too small",
			schema: &Schema{
				Type:    "integer",
				Minimum: float64Ptr(10),
			},
			input:     `5`,
			wantError: true,
			errorMsg:  "minimum",
		},
		{
			name: "number validation",
			schema: &Schema{
				Type:    "number",
				Minimum: float64Ptr(0),
				Maximum: float64Ptr(1),
			},
			input:     `0.5`,
			wantError: false,
		},
		{
			name: "boolean validation",
			schema: &Schema{
				Type: "boolean",
			},
			input:     `true`,
			wantError: false,
		},
		{
			name: "array validation",
			schema: &Schema{
				Type:     "array",
				Items:    &Schema{Type: "string"},
				MinItems: intPtr(1),
				MaxItems: intPtr(3),
			},
			input:     `["a", "b"]`,
			wantError: false,
		},
		{
			name: "array too long",
			schema: &Schema{
				Type:     "array",
				Items:    &Schema{Type: "string"},
				MaxItems: intPtr(2),
			},
			input:     `["a", "b", "c"]`,
			wantError: true,
			errorMsg:  "items",
		},
		{
			name: "array with invalid items",
			schema: &Schema{
				Type:  "array",
				Items: &Schema{Type: "number"},
			},
			input:     `["not", "numbers"]`,
			wantError: true,
			errorMsg:  "type",
		},
		{
			name: "nested object validation",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"user": {
						Type: "object",
						Properties: map[string]*Schema{
							"name":  {Type: "string"},
							"email": {Type: "string", Format: "email"},
						},
						Required: []string{"name", "email"},
					},
				},
			},
			input: `
user:
  name: "John"
  email: "john@example.com"`,
			wantError: false,
		},
		{
			name: "null validation",
			schema: &Schema{
				Type: "null",
			},
			input:     `null`,
			wantError: false,
		},
		{
			name: "enum validation",
			schema: &Schema{
				Type: "string",
				Enum: []interface{}{"red", "green", "blue"},
			},
			input:     `"green"`,
			wantError: false,
		},
		{
			name: "enum validation fail",
			schema: &Schema{
				Type: "string",
				Enum: []interface{}{"red", "green", "blue"},
			},
			input:     `"yellow"`,
			wantError: true,
			errorMsg:  "enum",
		},
		{
			name: "additional properties",
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {Type: "string"},
				},
				AdditionalProperties: false,
			},
			input: `
name: "test"
extra: "field"`,
			wantError: true,
			errorMsg:  "additional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input YAML
			tree, err := UnmarshalYAML([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			// Get the root node
			var node *Node
			if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
				node = tree.Documents[0].Root
				if node.Kind == DocumentNode && len(node.Children) > 0 {
					node = node.Children[0]
				}
			}

			// Validate
			errors := tt.schema.Validate(node, "$")

			if tt.wantError {
				if len(errors) == 0 {
					t.Error("Expected validation errors but got none")
				} else if tt.errorMsg != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(strings.ToLower(err.Error()), tt.errorMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no errors but got: %v", errors)
				}
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name         string
		node         *Node
		schemaType   string
		wantValidate bool
	}{
		{
			name:         "valid string",
			node:         &Node{Kind: ScalarNode, Value: "test"},
			schemaType:   "string",
			wantValidate: true,
		},
		{
			name:         "valid integer",
			node:         &Node{Kind: ScalarNode, Value: int64(42)},
			schemaType:   "integer",
			wantValidate: true,
		},
		{
			name:         "valid number",
			node:         &Node{Kind: ScalarNode, Value: float64(42.5)},
			schemaType:   "number",
			wantValidate: true,
		},
		{
			name:         "valid boolean",
			node:         &Node{Kind: ScalarNode, Value: true},
			schemaType:   "boolean",
			wantValidate: true,
		},
		{
			name:         "valid null",
			node:         &Node{Kind: ScalarNode, Value: nil},
			schemaType:   "null",
			wantValidate: true,
		},
		{
			name:         "valid array",
			node:         &Node{Kind: SequenceNode},
			schemaType:   "array",
			wantValidate: true,
		},
		{
			name:         "valid object",
			node:         &Node{Kind: MappingNode},
			schemaType:   "object",
			wantValidate: true,
		},
		{
			name:         "type mismatch",
			node:         &Node{Kind: ScalarNode, Value: "test"},
			schemaType:   "number",
			wantValidate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &Schema{Type: tt.schemaType}
			errors := schema.Validate(tt.node, "$")
			hasError := len(errors) > 0

			if tt.wantValidate && hasError {
				t.Errorf("Expected validation to pass for %s type, but got errors: %v", tt.schemaType, errors)
			} else if !tt.wantValidate && !hasError {
				t.Errorf("Expected validation to fail for %s type, but got no errors", tt.schemaType)
			}
		})
	}
}

// Test Streaming Parser
func TestStreamParser(t *testing.T) {
	multiDoc := `---
doc: 1
value: "first"
---
doc: 2
value: "second"
---
doc: 3
value: "third"
...`

	t.Run("parse all documents", func(t *testing.T) {
		reader := strings.NewReader(multiDoc)
		parser := NewStreamParser(reader)

		docCount := 0
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			docCount++
			if len(tree.Documents) != 1 {
				t.Errorf("Expected 1 document per callback, got %d", len(tree.Documents))
			}
			return nil
		})

		err := parser.Parse()
		if err != nil {
			t.Errorf("Parse failed: %v", err)
		}

		if docCount != 3 {
			t.Errorf("Expected 3 documents, got %d", docCount)
		}
	})

	t.Run("stop parsing early", func(t *testing.T) {
		reader := strings.NewReader(multiDoc)
		parser := NewStreamParser(reader)

		docCount := 0
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			docCount++
			if docCount == 2 {
				return errors.New("stop")
			}
			return nil
		})

		err := parser.Parse()
		if err == nil || !strings.Contains(err.Error(), "stop") {
			t.Error("Expected error to stop parsing")
		}

		if docCount != 2 {
			t.Errorf("Expected parsing to stop at 2 documents, got %d", docCount)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		reader := strings.NewReader("")
		parser := NewStreamParser(reader)

		docCount := 0
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			docCount++
			return nil
		})

		err := parser.Parse()
		if err != nil {
			t.Errorf("Parse failed: %v", err)
		}

		// Empty input might still produce one empty document
		if docCount > 1 {
			t.Errorf("Expected at most 1 document for empty input, got %d", docCount)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidYAML := `---
valid: document
---
invalid: [
  unclosed
---
another: document`

		reader := strings.NewReader(invalidYAML)
		parser := NewStreamParser(reader)

		docCount := 0
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			docCount++
			return nil
		})

		err := parser.Parse()

		if err == nil {
			t.Error("Expected parsing error for invalid YAML")
		}
	})

	t.Run("large document streaming", func(t *testing.T) {
		// Create a large multi-document YAML
		var sb strings.Builder
		for i := 0; i < 100; i++ {
			sb.WriteString(fmt.Sprintf("---\ndoc: %d\ndata: value%d\n", i, i))
		}

		reader := strings.NewReader(sb.String())
		parser := NewStreamParser(reader)

		docCount := 0
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			docCount++
			return nil
		})

		err := parser.Parse()
		if err != nil {
			t.Errorf("Failed to parse large document: %v", err)
		}

		if docCount != 100 {
			t.Errorf("Expected 100 documents, got %d", docCount)
		}
	})
}

// Test Transform DSL
func TestTransformDSL(t *testing.T) {
	baseYAML := `
config:
  database:
    host: "localhost"
    port: 5432
    username: "admin"
    password: "secret"
  cache:
    type: "redis"
    ttl: 3600
  features:
    - authentication
    - logging
    - monitoring
metadata:
  version: "1.0.0"
  author: "test"
`

	tree, _ := UnmarshalYAML([]byte(baseYAML))

	t.Run("RemoveKey", func(t *testing.T) {
		dsl := NewTransformDSL().RemoveKey("password")
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		// Verify password is removed
		output, _ := result.ToYAML()
		if strings.Contains(string(output), "password") {
			t.Error("Password should be removed")
		}
		if !strings.Contains(string(output), "username") {
			t.Error("Username should still exist")
		}
	})

	t.Run("RenameKey", func(t *testing.T) {
		dsl := NewTransformDSL().RenameKey("username", "user")
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		if strings.Contains(string(output), "username:") {
			t.Error("Old key 'username' should be renamed")
		}
		if !strings.Contains(string(output), "user:") {
			t.Error("New key 'user' should exist")
		}
	})

	t.Run("SortKeys", func(t *testing.T) {
		dsl := NewTransformDSL().SortKeys()
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		// Check that keys appear in alphabetical order
		configIdx := strings.Index(string(output), "config:")
		metadataIdx := strings.Index(string(output), "metadata:")
		if configIdx < 0 || metadataIdx < 0 {
			t.Error("Expected both config and metadata keys")
		}
		if configIdx > metadataIdx {
			t.Error("Keys should be sorted alphabetically")
		}
	})

	t.Run("AddComment", func(t *testing.T) {
		comment := "Generated by system"
		dsl := NewTransformDSL().AddComment(comment)
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		if !strings.Contains(string(output), comment) {
			t.Error("Comment should be added")
		}
	})

	t.Run("Flatten", func(t *testing.T) {
		dsl := NewTransformDSL().Flatten()
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		// Flattened structure should have keys like "config.database.host"
		if !strings.Contains(string(output), "config.database.host") &&
			!strings.Contains(string(output), "config_database_host") {
			t.Error("Structure should be flattened")
		}
	})

	t.Run("Select", func(t *testing.T) {
		// Keep only config section nodes
		dsl := NewTransformDSL().Select(func(node *Node) bool {
			// Keep config key and its children
			if node.Key != nil {
				if keyStr, ok := node.Key.Value.(string); ok && keyStr == "config" {
					return true
				}
			}
			// Keep nodes that are children of config
			if node.Parent != nil && node.Parent.Key != nil {
				if keyStr, ok := node.Parent.Key.Value.(string); ok && keyStr == "config" {
					return true
				}
			}
			return false
		})
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		if !strings.Contains(string(output), "config:") {
			t.Errorf("Config should be kept. Output: %s", string(output))
		}
		if strings.Contains(string(output), "metadata:") {
			t.Error("Metadata should be filtered out")
		}
	})

	t.Run("Map", func(t *testing.T) {
		// Custom transform to uppercase all string values
		dsl := NewTransformDSL().Map(func(node *Node) *Node {
			if node.Kind == ScalarNode && node.Value != nil {
				if str, ok := node.Value.(string); ok {
					node.Value = strings.ToUpper(str)
				}
			}
			return node
		})
		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()
		if !strings.Contains(string(output), "LOCALHOST") {
			t.Error("String values should be uppercase")
		}
	})

	t.Run("Chained transforms", func(t *testing.T) {
		dsl := NewTransformDSL().
			RemoveKey("password").
			RenameKey("username", "user").
			SortKeys().
			AddComment("Modified")

		result, err := dsl.Apply(tree)
		if err != nil {
			t.Fatalf("Transform failed: %v", err)
		}

		output, _ := result.ToYAML()

		// Verify all transforms were applied
		if strings.Contains(string(output), "password") {
			t.Error("Password should be removed")
		}
		if strings.Contains(string(output), "username:") {
			t.Error("Username should be renamed")
		}
		if !strings.Contains(string(output), "user:") {
			t.Error("User key should exist")
		}
		if !strings.Contains(string(output), "Modified") {
			t.Error("Comment should be added")
		}
	})

	t.Run("Transform with nil tree", func(t *testing.T) {
		dsl := NewTransformDSL().RemoveKey("test")
		result, err := dsl.Apply(nil)
		if err == nil {
			t.Error("Should error on nil tree")
		}
		if result != nil {
			t.Error("Result should be nil on error")
		}
	})

	t.Run("Transform empty tree", func(t *testing.T) {
		emptyTree := &NodeTree{}
		dsl := NewTransformDSL().AddComment("test")
		result, err := dsl.Apply(emptyTree)
		if err != nil {
			t.Errorf("Should handle empty tree: %v", err)
		}
		if result == nil {
			t.Error("Should return non-nil result")
		}
	})
}

// Test Query System
func TestQuery(t *testing.T) {
	yamlContent := `
users:
  - name: "Alice"
    age: 30
    roles:
      - admin
      - developer
  - name: "Bob"
    age: 25
    roles:
      - user
  - name: "Charlie"
    age: 35
    roles:
      - admin
      - user
settings:
  theme: "dark"
  language: "en"
  notifications:
    email: true
    push: false
    sms: true
database:
  connections:
    - host: "db1.example.com"
      port: 5432
    - host: "db2.example.com"
      port: 5433
`

	tree, _ := UnmarshalYAML([]byte(yamlContent))
	root := tree.Documents[0].Root
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	tests := []struct {
		name     string
		query    string
		expected int
		check    func(t *testing.T, results []*Node)
	}{
		{
			name:     "simple path",
			query:    "settings",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				if results[0].Kind != MappingNode {
					t.Error("Should return mapping node")
				}
			},
		},
		{
			name:     "nested path",
			query:    "settings/theme",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				if results[0].Value != "dark" {
					t.Errorf("Expected 'dark', got %v", results[0].Value)
				}
			},
		},
		{
			name:     "deep nested path",
			query:    "settings/notifications/email",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				boolVal, ok := results[0].Value.(bool)
				if !ok || boolVal != true {
					t.Errorf("Expected true (bool), got %v (%T)", results[0].Value, results[0].Value)
				}
			},
		},
		{
			name:     "array index",
			query:    "users/[0]/name",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				if results[0].Value != "Alice" {
					t.Errorf("Expected 'Alice', got %v", results[0].Value)
				}
			},
		},
		{
			name:     "array last index",
			query:    "users/[2]/name",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				if results[0].Value != "Charlie" {
					t.Errorf("Expected 'Charlie', got %v", results[0].Value)
				}
			},
		},
		{
			name:     "wildcard",
			query:    "users/*/name",
			expected: 3,
			check: func(t *testing.T, results []*Node) {
				names := make(map[string]bool)
				for _, r := range results {
					if str, ok := r.Value.(string); ok {
						names[str] = true
					}
				}
				if !names["Alice"] || !names["Bob"] || !names["Charlie"] {
					t.Error("Should return all names")
				}
			},
		},
		{
			name:     "nested wildcard",
			query:    "users/*/roles/*",
			expected: -1, // Variable number
			check: func(t *testing.T, results []*Node) {
				if len(results) < 5 {
					t.Error("Should return multiple roles")
				}
			},
		},
		{
			name:     "wildcard in middle",
			query:    "database/*/[0]/host",
			expected: 1,
			check: func(t *testing.T, results []*Node) {
				if results[0].Value != "db1.example.com" {
					t.Errorf("Expected 'db1.example.com', got %v", results[0].Value)
				}
			},
		},
		{
			name:     "non-existent path",
			query:    "nonexistent/path",
			expected: 0,
		},
		{
			name:     "invalid array index",
			query:    "users/[99]/name",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Query(root, tt.query)

			if tt.expected >= 0 && len(results) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(results))
			}

			if tt.check != nil && len(results) > 0 {
				tt.check(t, results)
			}
		})
	}
}

func TestQueryEdgeCases(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		results := Query(nil, "any/path")
		if len(results) != 0 {
			t.Error("Nil node should return empty results")
		}
	})

	t.Run("empty query", func(t *testing.T) {
		node := &Node{Kind: MappingNode}
		results := Query(node, "")
		if len(results) != 1 || results[0] != node {
			t.Error("Empty query should return the node itself")
		}
	})

	t.Run("query with spaces", func(t *testing.T) {
		yamlContent := `"key with spaces": value`
		tree, _ := UnmarshalYAML([]byte(yamlContent))
		root := tree.Documents[0].Root
		if root.Kind == DocumentNode && len(root.Children) > 0 {
			root = root.Children[0]
		}

		results := Query(root, "key with spaces")
		if len(results) != 1 || results[0].Value != "value" {
			t.Error("Should handle keys with spaces")
		}
	})

	t.Run("negative array index", func(t *testing.T) {
		yamlContent := `list: [a, b, c]`
		tree, _ := UnmarshalYAML([]byte(yamlContent))
		root := tree.Documents[0].Root
		if root.Kind == DocumentNode && len(root.Children) > 0 {
			root = root.Children[0]
		}

		results := Query(root, "list/[-1]")
		// Negative indices might be supported to get from end
		// or might return empty - implementation dependent
		_ = results
	})
}

// Test ValidationError
func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Path:    "$.user.email",
		Message: "invalid email format",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "$.user.email") {
		t.Error("Error should contain path")
	}
	if !strings.Contains(errStr, "invalid email format") {
		t.Error("Error should contain message")
	}
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// Test error callbacks and edge cases
func TestStreamParserErrorHandling(t *testing.T) {
	t.Run("reader error", func(t *testing.T) {
		// Create a reader that returns an error
		reader := &errorReader{err: errors.New("read error")}
		parser := NewStreamParser(reader)

		err := parser.Parse()
		if err == nil {
			t.Error("Expected error from reader")
		}
	})

	t.Run("partial document", func(t *testing.T) {
		partial := `---
key: value
incomplete: [`

		reader := strings.NewReader(partial)
		parser := NewStreamParser(reader)

		err := parser.Parse()

		if err == nil {
			t.Error("Should detect incomplete document")
		}
	})
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// Test transform error cases
func TestTransformDSLErrors(t *testing.T) {
	t.Run("invalid tree structure", func(t *testing.T) {
		// Create a tree with nil document
		tree := &NodeTree{
			Documents: []*Document{nil},
		}

		dsl := NewTransformDSL().RemoveKey("test")
		_, err := dsl.Apply(tree)
		if err == nil {
			t.Error("Should error on nil document")
		}
	})

	t.Run("circular reference", func(t *testing.T) {
		// Create circular reference
		node1 := &Node{Kind: MappingNode}
		node2 := &Node{Kind: MappingNode}
		node1.Children = []*Node{node2}
		node2.Children = []*Node{node1} // circular

		tree := &NodeTree{
			Documents: []*Document{
				{Root: node1},
			},
		}

		dsl := NewTransformDSL().SortKeys()
		// Should handle circular references without infinite loop
		// Implementation should detect and break cycles
		result, err := dsl.Apply(tree)
		_ = result
		_ = err
		// Test passes if it doesn't hang
	})
}

// Benchmark tests for advanced features
func BenchmarkSchemaValidation(b *testing.B) {
	schema := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name":  {Type: "string"},
			"age":   {Type: "integer"},
			"email": {Type: "string", Format: "email"},
		},
		Required: []string{"name", "email"},
	}

	yamlContent := `
name: "John Doe"
age: 30
email: "john@example.com"
`

	tree, _ := UnmarshalYAML([]byte(yamlContent))
	node := tree.Documents[0].Root
	if node.Kind == DocumentNode && len(node.Children) > 0 {
		node = node.Children[0]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate(node, "$")
	}
}

func BenchmarkStreamParser(b *testing.B) {
	// Create multi-document YAML
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("---\ndoc: %d\n", i))
	}
	data := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(data)
		parser := NewStreamParser(reader)
		parser.SetDocumentCallback(func(tree *NodeTree) error {
			return nil
		})
		_ = parser.Parse()
	}
}

func BenchmarkTransformDSL(b *testing.B) {
	yamlContent := `
config:
  database:
    host: "localhost"
    port: 5432
    username: "admin"
    password: "secret"
`

	tree, _ := UnmarshalYAML([]byte(yamlContent))
	dsl := NewTransformDSL().
		RemoveKey("password").
		RenameKey("username", "user").
		SortKeys()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dsl.Apply(tree)
	}
}

func BenchmarkQuery(b *testing.B) {
	yamlContent := `
deeply:
  nested:
    structure:
      with:
        many:
          levels:
            value: "found"
`

	tree, _ := UnmarshalYAML([]byte(yamlContent))
	root := tree.Documents[0].Root
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Query(root, "deeply/nested/structure/with/many/levels/value")
	}
}

// Test concurrent access (if applicable)
func TestConcurrentAccess(t *testing.T) {
	yamlContent := `
test: value
number: 42
`

	tree, _ := UnmarshalYAML([]byte(yamlContent))

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Perform various read operations
			_, _ = tree.ToYAML()
			root := tree.Documents[0].Root
			root.Walk(func(n *Node) bool {
				return true
			})
			_ = root.Find(func(n *Node) bool {
				return n.Value == "value"
			})
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passed if no race conditions or panics
}

// Test memory-intensive operations
func TestMemoryIntensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory-intensive test in short mode")
	}

	t.Run("large nested structure", func(t *testing.T) {
		// Create a very deep nested structure
		var sb strings.Builder
		depth := 100
		for i := 0; i < depth; i++ {
			sb.WriteString(strings.Repeat("  ", i))
			sb.WriteString(fmt.Sprintf("level%d:\n", i))
		}
		sb.WriteString(strings.Repeat("  ", depth))
		sb.WriteString("value: deep\n")

		tree, err := UnmarshalYAML([]byte(sb.String()))
		if err != nil {
			t.Fatalf("Failed to parse deep structure: %v", err)
		}

		// Verify we can traverse it
		count := 0
		tree.Documents[0].Root.Walk(func(n *Node) bool {
			count++
			return true
		})

		if count < depth {
			t.Error("Should traverse all levels")
		}
	})

	t.Run("large flat structure", func(t *testing.T) {
		// Create a structure with many keys
		var sb strings.Builder
		for i := 0; i < 10000; i++ {
			sb.WriteString(fmt.Sprintf("key%d: value%d\n", i, i))
		}

		start := time.Now()
		tree, err := UnmarshalYAML([]byte(sb.String()))
		if err != nil {
			t.Fatalf("Failed to parse large structure: %v", err)
		}

		duration := time.Since(start)
		if duration > 5*time.Second {
			t.Errorf("Parsing took too long: %v", duration)
		}

		// Verify round-trip
		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}

		if len(output) == 0 {
			t.Error("Output should not be empty")
		}
	})
}

// Test for potential security issues
func TestSecurityConcerns(t *testing.T) {
	t.Run("yaml bomb prevention", func(t *testing.T) {
		// Test against billion laughs attack
		bomb := `
a: &a ["lol", "lol", "lol", "lol", "lol", "lol", "lol", "lol", "lol"]
b: &b [*a, *a, *a, *a, *a, *a, *a, *a, *a]
c: &c [*b, *b, *b, *b, *b, *b, *b, *b, *b]
d: &d [*c, *c, *c, *c, *c, *c, *c, *c, *c]
`
		// Parser should handle this without consuming excessive memory
		// The underlying yaml.v3 should have protections
		tree, err := UnmarshalYAML([]byte(bomb))
		_ = tree
		_ = err
		// Test passes if it doesn't consume excessive memory or crash
	})

	t.Run("path traversal in queries", func(t *testing.T) {
		yamlContent := `
safe: value
`
		tree, _ := UnmarshalYAML([]byte(yamlContent))
		root := tree.Documents[0].Root

		// Attempt path traversal
		dangerous := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"/etc/passwd",
			"C:\\Windows\\System32",
		}

		for _, path := range dangerous {
			results := Query(root, path)
			if len(results) > 0 {
				t.Errorf("Should not traverse to system paths: %s", path)
			}
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("processComments", func(t *testing.T) {
		// Simple comment processing function for testing
		processComments := func(input string) []string {
			if input == "" {
				return nil
			}
			return strings.Split(input, "\n")
		}

		tests := []struct {
			input    string
			expected []string
		}{
			{"", nil},
			{"# single", []string{"# single"}},
			{"# line1\n# line2", []string{"# line1", "# line2"}},
			{"# line1\n\n# line2", []string{"# line1", "", "# line2"}},
		}

		for _, tt := range tests {
			result := processComments(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("processComments(%q) = %v, want %v",
					tt.input, result, tt.expected)
			}
		}
	})
}

// Test JSON conversion utilities
func TestJSONConversion(t *testing.T) {
	t.Run("YAML to JSON structure", func(t *testing.T) {
		yamlContent := `
string: "value"
number: 42
float: 3.14
boolean: true
null: null
array: [1, 2, 3]
object:
  nested: "value"
`
		tree, _ := UnmarshalYAML([]byte(yamlContent))

		// Convert to JSON-compatible structure
		output, _ := tree.ToYAML()

		// Parse back and verify types
		tree2, _ := UnmarshalYAML(output)
		if tree2 == nil {
			t.Fatal("Failed to parse converted YAML")
		}

		root := tree2.Documents[0].Root
		if root.Kind == DocumentNode && len(root.Children) > 0 {
			root = root.Children[0]
		}

		// Verify types are preserved
		strNode := root.GetMapValue("string")
		if strNode == nil || strNode.Value != "value" {
			t.Error("String not preserved")
		}

		numNode := root.GetMapValue("number")
		if numNode == nil {
			t.Error("Number not preserved")
		}

		boolNode := root.GetMapValue("boolean")
		if boolNode == nil || boolNode.Value != true {
			t.Error("Boolean not preserved")
		}

		nullNode := root.GetMapValue("null")
		if nullNode == nil {
			t.Error("Null node is missing")
		} else if nullNode.Value != nil {
			t.Errorf("Null value not preserved, got %v (%T)", nullNode.Value, nullNode.Value)
		}
	})
}

// Test advanced validation scenarios
func TestAdvancedValidation(t *testing.T) {
	t.Run("oneOf validation", func(t *testing.T) {
		schema := &Schema{
			OneOf: []*Schema{
				{Type: "string"},
				{Type: "number"},
			},
		}

		// Test valid string
		node := &Node{Kind: ScalarNode, Value: "test"}
		errors := schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("String should be valid for oneOf")
		}

		// Test valid number
		node = &Node{Kind: ScalarNode, Value: float64(42)}
		errors = schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("Number should be valid for oneOf")
		}

		// Test invalid type
		node = &Node{Kind: MappingNode}
		errors = schema.Validate(node, "$")
		if len(errors) == 0 {
			t.Error("Object should be invalid for oneOf string/number")
		}
	})

	t.Run("anyOf validation", func(t *testing.T) {
		schema := &Schema{
			AnyOf: []*Schema{
				{Type: "string", MinLength: intPtr(5)},
				{Type: "number", Minimum: float64Ptr(10)},
			},
		}

		// Test valid long string
		node := &Node{Kind: ScalarNode, Value: "hello"}
		errors := schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("Long string should be valid")
		}

		// Test valid large number
		node = &Node{Kind: ScalarNode, Value: float64(15)}
		errors = schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("Large number should be valid")
		}

		// Test invalid - short string
		node = &Node{Kind: ScalarNode, Value: "hi"}
		errors = schema.Validate(node, "$")
		if len(errors) == 0 {
			t.Error("Short string should be invalid")
		}
	})

	t.Run("allOf validation", func(t *testing.T) {
		schema := &Schema{
			AllOf: []*Schema{
				{Type: "object"},
				{Properties: map[string]*Schema{
					"name": {Type: "string"},
				}},
				{Required: []string{"name"}},
			},
		}

		// Test valid object with name
		yamlContent := `name: "test"`
		tree, _ := UnmarshalYAML([]byte(yamlContent))
		node := tree.Documents[0].Root
		if node.Kind == DocumentNode && len(node.Children) > 0 {
			node = node.Children[0]
		}

		errors := schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("Object with name should be valid")
		}

		// Test invalid - missing name
		yamlContent = `other: "value"`
		tree, _ = UnmarshalYAML([]byte(yamlContent))
		node = tree.Documents[0].Root
		if node.Kind == DocumentNode && len(node.Children) > 0 {
			node = node.Children[0]
		}

		errors = schema.Validate(node, "$")
		if len(errors) == 0 {
			t.Error("Object without name should be invalid")
		}
	})

	t.Run("not validation", func(t *testing.T) {
		schema := &Schema{
			Not: &Schema{Type: "string"},
		}

		// Test valid (not string)
		node := &Node{Kind: ScalarNode, Value: float64(42)}
		errors := schema.Validate(node, "$")
		if len(errors) > 0 {
			t.Error("Number should be valid (not string)")
		}

		// Test invalid (is string)
		node = &Node{Kind: ScalarNode, Value: "test"}
		errors = schema.Validate(node, "$")
		if len(errors) == 0 {
			t.Error("String should be invalid")
		}
	})
}

// Test custom formats
func TestCustomFormats(t *testing.T) {
	formats := []struct {
		format  string
		valid   []string
		invalid []string
	}{
		{
			format:  "email",
			valid:   []string{"test@example.com", "user+tag@domain.co.uk"},
			invalid: []string{"not-email", "@example.com", "test@"},
		},
		{
			format:  "uri",
			valid:   []string{"http://example.com", "https://example.com/path", "ftp://files.com"},
			invalid: []string{"not-uri", "http://", "://example.com"},
		},
		{
			format:  "date",
			valid:   []string{"2024-01-15", "2000-12-31"},
			invalid: []string{"2024-13-01", "2024-01-32", "not-date"},
		},
		{
			format:  "time",
			valid:   []string{"14:30:00", "23:59:59", "00:00:00"},
			invalid: []string{"25:00:00", "14:60:00", "not-time"},
		},
		{
			format:  "date-time",
			valid:   []string{"2024-01-15T14:30:00Z", "2024-01-15T14:30:00+01:00"},
			invalid: []string{"2024-01-15", "14:30:00", "not-datetime"},
		},
	}

	for _, ft := range formats {
		schema := &Schema{
			Type:   "string",
			Format: ft.format,
		}

		for _, valid := range ft.valid {
			node := &Node{Kind: ScalarNode, Value: valid}
			errors := schema.Validate(node, "$")
			if len(errors) > 0 {
				t.Errorf("Format %s: '%s' should be valid", ft.format, valid)
			}
		}

		for _, invalid := range ft.invalid {
			node := &Node{Kind: ScalarNode, Value: invalid}
			errors := schema.Validate(node, "$")
			if len(errors) == 0 {
				t.Errorf("Format %s: '%s' should be invalid", ft.format, invalid)
			}
		}
	}
}
