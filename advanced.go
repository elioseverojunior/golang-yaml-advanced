package golang_yaml_advanced

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Schema represents a validation schema for YAML nodes
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

// ValidationError represents a schema validation error
type ValidationError struct {
	Path       string
	Message    string
	SchemaPath string
	Value      interface{}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Validation error at %s: %s (value: %v)", e.Path, e.Message, e.Value)
}

// Validate checks if a node conforms to the schema
func (s *Schema) Validate(node *Node, path string) []ValidationError {
	var errors []ValidationError

	if node == nil {
		if s.Type != "" && s.Type != "null" {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("expected type %s but got null", s.Type),
				Value:   nil,
			})
		}
		return errors
	}

	// Type validation
	if s.Type != "" {
		nodeType := getNodeType(node)
		if !matchesType(nodeType, s.Type) {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("expected type %s but got %s", s.Type, nodeType),
				Value:   node.Value,
			})
			return errors // Type mismatch, no point in further validation
		}
	}

	// Enum validation
	if len(s.Enum) > 0 {
		if !isInEnum(node, s.Enum) {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("value must be one of enum values %v", s.Enum),
				Value:   node.Value,
			})
		}
	}

	// String validations
	if node.Kind == ScalarNode && s.Type == "string" {
		str := fmt.Sprintf("%v", node.Value)

		if s.MinLength != nil && len(str) < *s.MinLength {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("string length %d is less than minimum %d", len(str), *s.MinLength),
				Value:   node.Value,
			})
		}

		if s.MaxLength != nil && len(str) > *s.MaxLength {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("string length %d exceeds maximum %d", len(str), *s.MaxLength),
				Value:   node.Value,
			})
		}

		if s.Pattern != "" {
			if matched, _ := regexp.MatchString(s.Pattern, str); !matched {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("string does not match pattern %s", s.Pattern),
					Value:   node.Value,
				})
			}
		}

		if s.Format != "" {
			if !validateFormat(str, s.Format) {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("string does not match format %s", s.Format),
					Value:   node.Value,
				})
			}
		}
	}

	// Number validations
	if node.Kind == ScalarNode && (s.Type == "number" || s.Type == "integer") {
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", node.Value), 64); err == nil {
			if s.Minimum != nil && num < *s.Minimum {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("value %f is less than minimum %f", num, *s.Minimum),
					Value:   node.Value,
				})
			}

			if s.Maximum != nil && num > *s.Maximum {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("value %f exceeds maximum %f", num, *s.Maximum),
					Value:   node.Value,
				})
			}
		}
	}

	// Object validations
	if node.Kind == MappingNode {
		// Check required properties
		existingKeys := make(map[string]bool)
		for i := 0; i < len(node.Children)-1; i += 2 {
			if node.Children[i].Kind == ScalarNode {
				existingKeys[fmt.Sprintf("%v", node.Children[i].Value)] = true
			}
		}

		for _, required := range s.Required {
			if !existingKeys[required] {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("required property '%s' is missing", required),
					Value:   nil,
				})
			}
		}

		// Validate properties
		for i := 0; i < len(node.Children)-1; i += 2 {
			keyNode := node.Children[i]
			valueNode := node.Children[i+1]

			if keyNode.Kind == ScalarNode {
				key := fmt.Sprintf("%v", keyNode.Value)
				childPath := fmt.Sprintf("%s.%s", path, key)

				if propSchema, ok := s.Properties[key]; ok {
					// Validate against specific property schema
					propErrors := propSchema.Validate(valueNode, childPath)
					errors = append(errors, propErrors...)
				} else if s.AdditionalProperties != nil {
					// Handle additional properties
					switch ap := s.AdditionalProperties.(type) {
					case bool:
						if !ap {
							errors = append(errors, ValidationError{
								Path:    childPath,
								Message: "additional properties are not allowed",
								Value:   valueNode.Value,
							})
						}
					case *Schema:
						propErrors := ap.Validate(valueNode, childPath)
						errors = append(errors, propErrors...)
					}
				}
			}
		}
	}

	// Array validations
	if node.Kind == SequenceNode {
		arrayLen := len(node.Children)

		if s.MinItems != nil && arrayLen < *s.MinItems {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("array length %d is less than minimum %d", arrayLen, *s.MinItems),
				Value:   arrayLen,
			})
		}

		if s.MaxItems != nil && arrayLen > *s.MaxItems {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("array has too many items: %d (maximum %d)", arrayLen, *s.MaxItems),
				Value:   arrayLen,
			})
		}

		if s.UniqueItems {
			seen := make(map[string]bool)
			for i, child := range node.Children {
				key := nodeToString(child)
				if seen[key] {
					errors = append(errors, ValidationError{
						Path:    fmt.Sprintf("%s[%d]", path, i),
						Message: "duplicate items not allowed",
						Value:   child.Value,
					})
				}
				seen[key] = true
			}
		}

		// Validate items
		if s.Items != nil {
			for i, child := range node.Children {
				childPath := fmt.Sprintf("%s[%d]", path, i)
				itemErrors := s.Items.Validate(child, childPath)
				errors = append(errors, itemErrors...)
			}
		}
	}

	// OneOf, AnyOf, AllOf validations
	if len(s.OneOf) > 0 {
		validCount := 0
		for _, schema := range s.OneOf {
			if len(schema.Validate(node, path)) == 0 {
				validCount++
			}
		}
		if validCount != 1 {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("value must match exactly one schema (matched %d)", validCount),
				Value:   node.Value,
			})
		}
	}

	if len(s.AnyOf) > 0 {
		validCount := 0
		for _, schema := range s.AnyOf {
			if len(schema.Validate(node, path)) == 0 {
				validCount++
			}
		}
		if validCount == 0 {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "value must match at least one schema",
				Value:   node.Value,
			})
		}
	}

	if len(s.AllOf) > 0 {
		for _, schema := range s.AllOf {
			subErrors := schema.Validate(node, path)
			errors = append(errors, subErrors...)
		}
	}

	if s.Not != nil {
		if len(s.Not.Validate(node, path)) == 0 {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "value must not match the schema",
				Value:   node.Value,
			})
		}
	}

	return errors
}

// Helper functions for schema validation
func getNodeType(node *Node) string {
	switch node.Kind {
	case MappingNode:
		return "object"
	case SequenceNode:
		return "array"
	case ScalarNode:
		if node.Value == nil || node.IsNull() {
			return "null"
		}
		// Try to determine scalar type
		str := fmt.Sprintf("%v", node.Value)
		if str == "null" {
			return "null"
		}
		if str == "true" || str == "false" {
			return "boolean"
		}
		if _, err := strconv.ParseInt(str, 10, 64); err == nil {
			return "integer"
		}
		if _, err := strconv.ParseFloat(str, 64); err == nil {
			return "number"
		}
		return "string"
	default:
		return "unknown"
	}
}

func matchesType(nodeType, schemaType string) bool {
	if nodeType == schemaType {
		return true
	}
	// Integer is a subset of number
	if nodeType == "integer" && schemaType == "number" {
		return true
	}
	return false
}

func isInEnum(node *Node, enum []interface{}) bool {
	if node == nil {
		for _, v := range enum {
			if v == nil {
				return true
			}
		}
		return false
	}

	if node.Kind == ScalarNode {
		nodeStr := fmt.Sprintf("%v", node.Value)
		for _, v := range enum {
			if fmt.Sprintf("%v", v) == nodeStr {
				return true
			}
		}
	}
	return false
}

func nodeToString(node *Node) string {
	if node == nil {
		return ""
	}
	if node.Kind != ScalarNode {
		return ""
	}
	if node.Value == nil {
		return ""
	}
	return fmt.Sprintf("%v", node.Value)
}

func nodeToInterface(node *Node) interface{} {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ScalarNode:
		return node.Value
	case MappingNode:
		result := make(map[string]interface{})
		for i := 0; i < len(node.Children)-1; i += 2 {
			key := fmt.Sprintf("%v", node.Children[i].Value)
			result[key] = nodeToInterface(node.Children[i+1])
		}
		return result
	case SequenceNode:
		result := make([]interface{}, 0, len(node.Children))
		for _, child := range node.Children {
			result = append(result, nodeToInterface(child))
		}
		return result
	default:
		return node.Value
	}
}

func validateFormat(value, format string) bool {
	switch format {
	case "email":
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+$`
		matched, _ := regexp.MatchString(emailRegex, value)
		return matched
	case "uri", "url":
		urlRegex := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
		matched, _ := regexp.MatchString(urlRegex, value)
		return matched
	case "ipv4":
		ipv4Regex := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
		matched, _ := regexp.MatchString(ipv4Regex, value)
		return matched
	case "ipv6":
		ipv6Regex := `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$`
		matched, _ := regexp.MatchString(ipv6Regex, value)
		return matched
	case "uuid":
		uuidRegex := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
		matched, _ := regexp.MatchString(uuidRegex, value)
		return matched
	case "date":
		dateRegex := `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`
		matched, _ := regexp.MatchString(dateRegex, value)
		if !matched {
			return false
		}
		// Additional date validation
		parts := strings.Split(value, "-")
		year, _ := strconv.Atoi(parts[0])
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[2])
		// Check days in month
		daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			daysInMonth[1] = 29 // Leap year
		}
		if month < 1 || month > 12 || day < 1 || day > daysInMonth[month-1] {
			return false
		}
		return true
	case "date-time", "datetime":
		dateTimeRegex := `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.\d+)?(Z|[+-]\d{2}:\d{2})$`
		matched, _ := regexp.MatchString(dateTimeRegex, value)
		return matched
	case "time":
		timeRegex := `^([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`
		matched, _ := regexp.MatchString(timeRegex, value)
		return matched
	default:
		return true // Unknown format, assume valid
	}
}

// StreamParser provides streaming YAML parsing for large files
type StreamParser struct {
	reader           *bufio.Reader
	currentLine      int
	buffer           []string
	inDocument       bool
	documentCallback func(*NodeTree) error
}

// NewStreamParser creates a new streaming YAML parser
func NewStreamParser(reader io.Reader) *StreamParser {
	return &StreamParser{
		reader:      bufio.NewReader(reader),
		currentLine: 0,
		buffer:      make([]string, 0),
	}
}

// SetDocumentCallback sets the callback function for each parsed document
func (sp *StreamParser) SetDocumentCallback(callback func(*NodeTree) error) {
	sp.documentCallback = callback
}

// Parse starts the streaming parse process
func (sp *StreamParser) Parse() error {
	for {
		line, err := sp.reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading line %d: %w", sp.currentLine, err)
		}

		if err != io.EOF {
			sp.currentLine++
			trimmed := strings.TrimSpace(line)

			// Check for document separator
			if trimmed == "---" {
				// Process current buffer if not empty
				if len(sp.buffer) > 0 {
					if err := sp.processBuffer(); err != nil {
						return err
					}
				}
				sp.inDocument = true
			} else if trimmed == "..." {
				// Document end marker
				if len(sp.buffer) > 0 {
					if err := sp.processBuffer(); err != nil {
						return err
					}
				}
				sp.inDocument = false
			} else {
				// Regular content
				if !sp.inDocument && len(sp.buffer) == 0 && trimmed != "" {
					// Start of first document without explicit ---
					sp.inDocument = true
				}
				if sp.inDocument || trimmed != "" {
					sp.buffer = append(sp.buffer, strings.TrimRight(line, "\n"))
				}
			}
		} else {
			// Handle partial line without newline at EOF
			if len(line) > 0 {
				sp.buffer = append(sp.buffer, line)
			}
		}

		if err == io.EOF {
			// Process any remaining buffer
			if len(sp.buffer) > 0 {
				if err := sp.processBuffer(); err != nil {
					return err
				}
			}
			break
		}
	}

	return nil
}

func (sp *StreamParser) processBuffer() error {
	if len(sp.buffer) == 0 {
		return nil
	}

	content := strings.Join(sp.buffer, "\n")
	tree, err := UnmarshalYAML([]byte(content))
	if err != nil {
		return fmt.Errorf("error parsing document at line %d: %w", sp.currentLine-len(sp.buffer), err)
	}

	if sp.documentCallback != nil {
		if err := sp.documentCallback(tree); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	sp.buffer = sp.buffer[:0] // Clear buffer
	return nil
}

// Transform represents a transformation operation on nodes
type Transform struct {
	name        string
	description string
	operation   func(*Node) (*Node, error)
}

// TransformDSL provides a fluent interface for YAML transformations
type TransformDSL struct {
	transforms []Transform
	errors     []error
}

// NewTransformDSL creates a new transformation DSL
func NewTransformDSL() *TransformDSL {
	return &TransformDSL{
		transforms: make([]Transform, 0),
		errors:     make([]error, 0),
	}
}

// Select creates a transform that filters nodes matching a predicate
func (dsl *TransformDSL) Select(predicate func(*Node) bool) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "select",
		description: "Filter nodes by predicate",
		operation: func(node *Node) (*Node, error) {
			return selectNode(node, predicate)
		},
	})
	return dsl
}

func selectNode(node *Node, predicate func(*Node) bool) (*Node, error) {
	if node == nil {
		return nil, nil
	}

	// For document nodes, apply select to the child
	if node.Kind == DocumentNode {
		if len(node.Children) > 0 {
			filteredChild, err := selectNode(node.Children[0], predicate)
			if err != nil {
				return nil, err
			}
			if filteredChild != nil {
				result := node.Clone()
				result.Children = []*Node{filteredChild}
				return result, nil
			}
		}
		return nil, nil
	}

	// For mapping nodes, filter key-value pairs
	if node.Kind == MappingNode {
		newChildren := make([]*Node, 0)
		for i := 0; i < len(node.Children)-1; i += 2 {
			keyNode := node.Children[i]
			valueNode := node.Children[i+1]

			// Set the Key field on the value node for the predicate check
			valueNode.Key = keyNode

			// Check if this key-value pair should be kept
			if predicate(valueNode) {
				// Keep the entire value including nested content
				newChildren = append(newChildren, keyNode, valueNode)
			} else if valueNode.Kind == MappingNode {
				// Recursively check nested mappings
				filteredValue, err := selectNode(valueNode, predicate)
				if err != nil {
					return nil, err
				}
				if filteredValue != nil && len(filteredValue.Children) > 0 {
					newChildren = append(newChildren, keyNode, filteredValue)
				}
			}
		}

		if len(newChildren) > 0 {
			result := node.Clone()
			result.Children = newChildren
			return result, nil
		}
		return nil, nil
	}

	// For non-mapping nodes, check the predicate directly
	if predicate(node) {
		return node, nil
	}
	return nil, nil
}

// Map applies a transformation function to each node
func (dsl *TransformDSL) Map(fn func(*Node) *Node) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "map",
		description: "Transform each node",
		operation: func(node *Node) (*Node, error) {
			return fn(node), nil
		},
	})
	return dsl
}

// SetValue sets the value of scalar nodes
func (dsl *TransformDSL) SetValue(value interface{}) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "setValue",
		description: fmt.Sprintf("Set value to %v", value),
		operation: func(node *Node) (*Node, error) {
			if node.Kind == ScalarNode {
				node.Value = value
			}
			return node, nil
		},
	})
	return dsl
}

// AddComment adds a comment to nodes
func (dsl *TransformDSL) AddComment(comment string) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "addComment",
		description: "Add comment to nodes",
		operation: func(node *Node) (*Node, error) {
			node.HeadComment = append(node.HeadComment, comment)
			return node, nil
		},
	})
	return dsl
}

// RemoveKey removes a key from mapping nodes
func (dsl *TransformDSL) RemoveKey(key string) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "removeKey",
		description: fmt.Sprintf("Remove key '%s'", key),
		operation: func(node *Node) (*Node, error) {
			if node.Kind == MappingNode {
				newChildren := make([]*Node, 0)
				for i := 0; i < len(node.Children)-1; i += 2 {
					keyNode := node.Children[i]
					valueNode := node.Children[i+1]
					if keyNode.Kind != ScalarNode || fmt.Sprintf("%v", keyNode.Value) != key {
						newChildren = append(newChildren, keyNode, valueNode)
					}
				}
				node.Children = newChildren
			}
			return node, nil
		},
	})
	return dsl
}

// RenameKey renames a key in mapping nodes
func (dsl *TransformDSL) RenameKey(oldKey, newKey string) *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "renameKey",
		description: fmt.Sprintf("Rename key '%s' to '%s'", oldKey, newKey),
		operation: func(node *Node) (*Node, error) {
			if node.Kind == MappingNode {
				for i := 0; i < len(node.Children)-1; i += 2 {
					keyNode := node.Children[i]
					if keyNode.Kind == ScalarNode && fmt.Sprintf("%v", keyNode.Value) == oldKey {
						keyNode.Value = newKey
						break
					}
				}
			}
			return node, nil
		},
	})
	return dsl
}

// SortKeys sorts mapping keys alphabetically
func (dsl *TransformDSL) SortKeys() *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "sortKeys",
		description: "Sort mapping keys alphabetically",
		operation: func(node *Node) (*Node, error) {
			if node.Kind == MappingNode {
				// Extract key-value pairs
				type kvPair struct {
					key   *Node
					value *Node
				}
				pairs := make([]kvPair, 0)
				for i := 0; i < len(node.Children)-1; i += 2 {
					pairs = append(pairs, kvPair{
						key:   node.Children[i],
						value: node.Children[i+1],
					})
				}

				// Sort by key
				for i := 0; i < len(pairs); i++ {
					for j := i + 1; j < len(pairs); j++ {
						key1 := fmt.Sprintf("%v", pairs[i].key.Value)
						key2 := fmt.Sprintf("%v", pairs[j].key.Value)
						if key1 > key2 {
							pairs[i], pairs[j] = pairs[j], pairs[i]
						}
					}
				}

				// Rebuild children
				newChildren := make([]*Node, 0)
				for _, pair := range pairs {
					newChildren = append(newChildren, pair.key, pair.value)
				}
				node.Children = newChildren
			}
			return node, nil
		},
	})
	return dsl
}

// Flatten flattens nested mappings using dot notation
func (dsl *TransformDSL) Flatten() *TransformDSL {
	dsl.transforms = append(dsl.transforms, Transform{
		name:        "flatten",
		description: "Flatten nested mappings",
		operation: func(node *Node) (*Node, error) {
			if node.Kind != MappingNode {
				return node, nil
			}

			flatMap := make(map[string]*Node)
			flattenRecursive(node, "", flatMap)

			// Create new flat mapping
			newNode := NewNode(MappingNode)
			for key, value := range flatMap {
				keyNode := NewScalarNode(key)
				newNode.AddKeyValue(keyNode, value)
			}

			return newNode, nil
		},
	})
	return dsl
}

func flattenRecursive(node *Node, prefix string, result map[string]*Node) {
	if node.Kind != MappingNode {
		if prefix != "" {
			result[prefix] = node
		}
		return
	}

	for i := 0; i < len(node.Children)-1; i += 2 {
		keyNode := node.Children[i]
		valueNode := node.Children[i+1]

		if keyNode.Kind == ScalarNode {
			key := fmt.Sprintf("%v", keyNode.Value)
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}

			if valueNode.Kind == MappingNode {
				flattenRecursive(valueNode, newPrefix, result)
			} else {
				result[newPrefix] = valueNode
			}
		}
	}
}

// Apply executes all transformations on a node tree
func (dsl *TransformDSL) Apply(tree *NodeTree) (*NodeTree, error) {
	if tree == nil {
		return nil, fmt.Errorf("tree is nil")
	}

	if len(dsl.errors) > 0 {
		return nil, fmt.Errorf("DSL has %d errors", len(dsl.errors))
	}

	if len(tree.Documents) == 0 {
		return tree, nil
	}

	resultTree := NewNodeTree()

	for _, doc := range tree.Documents {
		if doc == nil {
			return nil, fmt.Errorf("tree contains nil document")
		}
		newDoc := &Document{
			Directives: doc.Directives,
			Version:    doc.Version,
			Anchors:    make(map[string]*Node),
		}

		if doc.Root != nil {
			transformedRoot, err := dsl.applyToNode(doc.Root)
			if err != nil {
				return nil, err
			}
			newDoc.Root = transformedRoot
		}

		resultTree.Documents = append(resultTree.Documents, newDoc)
		resultTree.Current = newDoc
	}

	return resultTree, nil
}

func (dsl *TransformDSL) applyToNode(node *Node) (*Node, error) {
	if node == nil {
		return nil, nil
	}

	// Special handling for Select transforms - they need to be applied differently
	for _, transform := range dsl.transforms {
		if transform.name == "select" {
			return transform.operation(node)
		}
	}

	result := node.Clone()

	for _, transform := range dsl.transforms {
		var err error
		result, err = transform.operation(result)
		if err != nil {
			return nil, fmt.Errorf("transform '%s' failed: %w", transform.name, err)
		}
		if result == nil {
			return nil, nil // Node filtered out
		}
	}

	// Apply transformations to children recursively
	if result.Kind == MappingNode || result.Kind == SequenceNode || result.Kind == DocumentNode {
		newChildren := make([]*Node, 0)
		for _, child := range result.Children {
			transformedChild, err := dsl.applyToNode(child)
			if err != nil {
				return nil, err
			}
			if transformedChild != nil {
				newChildren = append(newChildren, transformedChild)
			}
		}
		result.Children = newChildren
	}

	return result, nil
}

// Query provides XPath-like querying for YAML
func Query(node *Node, query string) []*Node {
	// Simple query parser
	parts := strings.Split(query, "/")
	results := []*Node{node}

	for _, part := range parts {
		if part == "" {
			continue
		}

		newResults := []*Node{}
		for _, n := range results {
			if n == nil {
				continue
			}

			if part == "*" {
				// Wildcard - get all children
				newResults = append(newResults, n.Children...)
			} else if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
				// Array index
				indexStr := part[1 : len(part)-1]
				if index, err := strconv.Atoi(indexStr); err == nil {
					if n.Kind == SequenceNode && index >= 0 && index < len(n.Children) {
						newResults = append(newResults, n.Children[index])
					}
				}
			} else {
				// Key name
				if n.Kind == MappingNode {
					for i := 0; i < len(n.Children)-1; i += 2 {
						keyNode := n.Children[i]
						valueNode := n.Children[i+1]
						if keyNode.Kind == ScalarNode && fmt.Sprintf("%v", keyNode.Value) == part {
							newResults = append(newResults, valueNode)
							break
						}
					}
				}
			}
		}
		results = newResults
	}

	return results
}
