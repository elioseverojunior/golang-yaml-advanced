package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elioetibr/golang-yaml-advanced"
)

func main() {
	fmt.Println("=== YAML Advanced Features Demo ===")

	// 1. Schema Validation
	fmt.Println("1. SCHEMA VALIDATION")
	fmt.Println(strings.Repeat("-", 40))
	demoSchemaValidation()

	// 2. Streaming Parser
	fmt.Println("\n2. STREAMING PARSER")
	fmt.Println(strings.Repeat("-", 40))
	demoStreamingParser()

	// 3. Transformation DSL
	fmt.Println("\n3. TRANSFORMATION DSL")
	fmt.Println(strings.Repeat("-", 40))
	demoTransformationDSL()

	// 4. Query System
	fmt.Println("\n4. QUERY SYSTEM")
	fmt.Println(strings.Repeat("-", 40))
	demoQuerySystem()
}

func demoSchemaValidation() {
	// Sample YAML to validate
	yamlContent := `
name: "John Doe"
age: 30
email: "john@example.com"
address:
  street: "123 Main St"
  city: "New York"
  zipcode: "10001"
hobbies:
  - reading
  - coding
  - gaming
`

	// Define a schema
	schema := &golang_yaml_advanced.Schema{
		Type: "object",
		Properties: map[string]*golang_yaml_advanced.Schema{
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
			"email": {
				Type:   "string",
				Format: "email",
			},
			"address": {
				Type: "object",
				Properties: map[string]*golang_yaml_advanced.Schema{
					"street": {Type: "string"},
					"city":   {Type: "string"},
					"zipcode": {
						Type:    "string",
						Pattern: `^\d{5}$`,
					},
				},
				Required: []string{"street", "city"},
			},
			"hobbies": {
				Type:     "array",
				Items:    &golang_yaml_advanced.Schema{Type: "string"},
				MinItems: intPtr(1),
				MaxItems: intPtr(10),
			},
		},
		Required: []string{"name", "email"},
	}

	// Parse YAML
	tree, err := golang_yaml_advanced.UnmarshalYAML([]byte(yamlContent))
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Validate against schema
	if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
		// Get the content node (skip document wrapper)
		contentNode := tree.Documents[0].Root
		if contentNode.Kind == golang_yaml_advanced.DocumentNode && len(contentNode.Children) > 0 {
			contentNode = contentNode.Children[0]
		}

		errors := schema.Validate(contentNode, "$")
		if len(errors) == 0 {
			fmt.Println("✅ YAML is valid according to schema!")
		} else {
			fmt.Println("❌ Validation errors found:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err.Error())
			}
		}
	}

	// Test with invalid YAML
	invalidYAML := `
name: ""
age: 200
email: "not-an-email"
`
	tree2, _ := golang_yaml_advanced.UnmarshalYAML([]byte(invalidYAML))
	if len(tree2.Documents) > 0 && tree2.Documents[0].Root != nil {
		contentNode := tree2.Documents[0].Root
		if contentNode.Kind == golang_yaml_advanced.DocumentNode && len(contentNode.Children) > 0 {
			contentNode = contentNode.Children[0]
		}

		errors := schema.Validate(contentNode, "$")
		fmt.Println("\nValidating invalid YAML:")
		for _, err := range errors {
			fmt.Printf("  ❌ %s\n", err.Error())
		}
	}
}

func demoStreamingParser() {
	// Create a large multi-document YAML
	largeYAML := `---
# Document 1
app:
  name: "App1"
  version: "1.0.0"
---
# Document 2
database:
  host: "localhost"
  port: 5432
---
# Document 3
cache:
  type: "redis"
  ttl: 3600
...`

	// Create streaming parser
	reader := strings.NewReader(largeYAML)
	parser := golang_yaml_advanced.NewStreamParser(reader)

	docCount := 0
	parser.SetDocumentCallback(func(tree *golang_yaml_advanced.NodeTree) error {
		docCount++
		fmt.Printf("Processed document %d\n", docCount)

		// Process each document as it's parsed
		if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
			// Extract some info
			root := tree.Documents[0].Root
			if root.Kind == golang_yaml_advanced.DocumentNode && len(root.Children) > 0 {
				fmt.Printf("  Root has %d children\n", len(root.Children[0].Children)/2)
			}
		}
		return nil
	})

	if err := parser.Parse(); err != nil {
		log.Printf("Streaming parse error: %v", err)
	} else {
		fmt.Printf("\n✅ Successfully streamed %d documents\n", docCount)
	}
}

func demoTransformationDSL() {
	yamlContent := `
config:
  database:
    host: "localhost"
    port: 5432
    username: "admin"
    password: "secret123"
  cache:
    type: "redis"
    ttl: 3600
  features:
    - authentication
    - logging
    - monitoring
`

	tree, err := golang_yaml_advanced.UnmarshalYAML([]byte(yamlContent))
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Example 1: Remove sensitive data
	fmt.Println("Example 1: Remove sensitive data")
	dsl1 := golang_yaml_advanced.NewTransformDSL().
		RemoveKey("password").
		RemoveKey("username")

	result1, err := dsl1.Apply(tree)
	if err != nil {
		log.Printf("Transform error: %v", err)
	} else {
		output, _ := result1.ToYAML()
		fmt.Println("Result (sensitive data removed):")
		fmt.Println(string(output))
	}

	// Example 2: Add comments and rename keys
	fmt.Println("\nExample 2: Add comments and rename keys")
	dsl2 := golang_yaml_advanced.NewTransformDSL().
		AddComment("Configuration file").
		RenameKey("database", "db").
		SortKeys()

	result2, err := dsl2.Apply(tree)
	if err != nil {
		log.Printf("Transform error: %v", err)
	} else {
		output, _ := result2.ToYAML()
		fmt.Println("Result (with comments and renamed keys):")
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i > 10 {
				fmt.Println("  ... (truncated)")
				break
			}
			fmt.Println("  " + line)
		}
	}

	// Example 3: Flatten nested structure
	fmt.Println("\nExample 3: Flatten nested structure")
	dsl3 := golang_yaml_advanced.NewTransformDSL().Flatten()

	result3, err := dsl3.Apply(tree)
	if err != nil {
		log.Printf("Transform error: %v", err)
	} else {
		output, _ := result3.ToYAML()
		fmt.Println("Result (flattened):")
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
				fmt.Println("  " + line)
			}
		}
	}
}

func demoQuerySystem() {
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
`

	tree, err := golang_yaml_advanced.UnmarshalYAML([]byte(yamlContent))
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
		root := tree.Documents[0].Root
		if root.Kind == golang_yaml_advanced.DocumentNode && len(root.Children) > 0 {
			root = root.Children[0]
		}

		// Query examples
		queries := []string{
			"users",
			"users/[0]/name",
			"users/[1]/roles",
			"settings/notifications/email",
			"settings/theme",
		}

		fmt.Println("Query Examples:")
		for _, query := range queries {
			results := golang_yaml_advanced.Query(root, query)
			fmt.Printf("\nQuery: %s\n", query)
			if len(results) > 0 {
				for _, node := range results {
					if node.Kind == golang_yaml_advanced.ScalarNode {
						fmt.Printf("  Result: %v\n", node.Value)
					} else if node.Kind == golang_yaml_advanced.SequenceNode {
						fmt.Printf("  Result: Array with %d items\n", len(node.Children))
					} else if node.Kind == golang_yaml_advanced.MappingNode {
						fmt.Printf("  Result: Object with %d keys\n", len(node.Children)/2)
					}
				}
			} else {
				fmt.Println("  No results found")
			}
		}
	}

	// Advanced: Find all users who are admins
	fmt.Println("\n\nAdvanced: Find all admin users")
	if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
		root := tree.Documents[0].Root
		if root.Kind == golang_yaml_advanced.DocumentNode && len(root.Children) > 0 {
			root = root.Children[0]
		}

		usersNode := root.GetMapValue("users")
		if usersNode != nil && usersNode.Kind == golang_yaml_advanced.SequenceNode {
			for i, userNode := range usersNode.Children {
				if userNode.Kind == golang_yaml_advanced.MappingNode {
					nameNode := userNode.GetMapValue("name")
					rolesNode := userNode.GetMapValue("roles")

					if rolesNode != nil && rolesNode.Kind == golang_yaml_advanced.SequenceNode {
						for _, roleNode := range rolesNode.Children {
							if roleNode.Kind == golang_yaml_advanced.ScalarNode && fmt.Sprintf("%v", roleNode.Value) == "admin" {
								if nameNode != nil && nameNode.Kind == golang_yaml_advanced.ScalarNode {
									fmt.Printf("  User %d: %v is an admin\n", i, nameNode.Value)
								}
								break
							}
						}
					}
				}
			}
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func jsonPretty(data interface{}) string {
	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}
