package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/elioetibr/golang-yaml-advanced"
)

type YamlData struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

var (
	yamlCommonTestData = map[string]*YamlData{
		"headersOnly": {
			Title: "Headers Only YAML",
			Content: `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.
`,
		},
		"simple": {
			Title: "Simple YAML with Single Field",
			Content: `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# @schema
# additionalProperties: false
# @schema
# -- Kubernetes deployment strategy for managing pod updates and ensuring zero-downtime deployments
fullnameOverride: "test"
`,
		},
		"complex": {
			Title: "Complex YAML with Nested Structure",
			Content: `# yaml-language-server: $schema=values.schema.json
# Default values for base-chart.
# This is a YAML-formatted file.

# Declare variables to be passed into your templates.

# @schema
# additionalProperties: false
# @schema
# -- Kubernetes deployment strategy for managing pod updates and ensuring zero-downtime deployments
strategy:

  # -- Strategy type: RollingUpdate for gradual replacement, Recreate for immediate replacement
  type: RollingUpdate

  # -- Rolling Update Configuration
  rollingUpdate:
    # -- Maximum extra pods allowed during update (can be number or percentage like "25%")
    maxSurge: 1
    # -- Maximum pods that can be unavailable during update to maintain service availability
    maxUnavailable: 0

# @schema
# additionalProperties: true
# @schema
# -- List of secret names containing registry credentials for pulling images from private repositories like AWS ECR
imagePullSecrets: []
# Example: [{name: "aws-ecr-secret"}]
`,
		},
	}
	yamlMergeTestData = map[string]*YamlData{
		"toMerge": {
			Title: "YAML Content to Merge",
			Content: `# @schema
# additionalProperties: true
# @schema
# -- List of secret names containing registry credentials for pulling images from private repositories like AWS ECR
imagePullSecrets:
# Example: [{name: "aws-ecr-secret"}]
  - name: "aws-ecr-secret"
`,
		},
	}

	yamlAdvancedTestData = map[string]*YamlData{
		"anchorsAndAliases": {
			Title: "YAML with Anchors and Aliases",
			Content: `# Database configuration with shared settings
defaults: &defaults
  adapter: postgresql
  encoding: unicode
  pool: 5

development:
  # Inheriting from defaults
  <<: *defaults
  database: myapp_development
  host: localhost

test:
  # Also inheriting defaults
  <<: *defaults
  database: myapp_test
  host: localhost

production:
  <<: *defaults
  database: myapp_production
  host: prod.example.com
  pool: 10  # Override pool size for production
`,
		},
		"multiDocument": {
			Title: "Multi-document YAML",
			Content: `# First document - Application config
---
app:
  name: MyApp
  version: 1.0.0
---
# Second document - Database config
database:
  host: localhost
  port: 5432
---
# Third document - Cache config
cache:
  type: redis
  ttl: 3600
`,
		},
		"complexTags": {
			Title: "YAML with Complex Tags",
			Content: `# Custom tags and explicit types
string_value: !!str 123
integer_value: !!int "456"
float_value: !!float 789
boolean_value: !!bool "yes"
null_value: !!null ""
timestamp: !!timestamp 2024-09-24T12:00:00Z
binary: !!binary |
  R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5
set: !!set
  ? item1
  ? item2
  ? item3
omap: !!omap
  - Mark: 65
  - Sammy: 63
  - Key: 58
`,
		},
		"folded": {
			Title: "YAML with Folded and Literal Scalars",
			Content: `# Different scalar styles
literal_block: |
  This is a literal block scalar.
  Line breaks are preserved.
  	Indentation too.

  Even blank lines.

folded_block: >
  This is a folded block scalar.
  Line breaks are folded into spaces.

  But blank lines create paragraphs.

plain_scalar: This is a plain scalar that can span multiple lines if needed

single_quoted: 'This is a single-quoted scalar with ''escaped'' quotes'

double_quoted: "This supports\nescapes like \\n and \\t"

# Comments between scalars
flow_sequence: [item1, item2, # inline comment
                item3]

flow_mapping: {key1: value1, key2: value2}
`,
		},
		"deepNesting": {
			Title: "Deeply Nested YAML",
			Content: `# Deeply nested structure with comments at every level
level1:
  # Level 2 comment
  level2:
    # Level 3 comment
    level3:
      # Level 4 comment
      level4:
        # Level 5 comment
        level5:
          # Level 6 comment
          level6:
            # Deep value with metadata
            value: "deep"
            # Sibling at depth
            sibling: "also deep"
          # Back at level 6
          parallel6:
            data: "parallel"
        # Array at level 5
        array5:
          - item1  # First item
          - item2  # Second item
          - # Third item is a map
            nested: true
            complex: yes
      # Back to level 4
      level4b:
        key: value
`,
		},
		"mergeKeys": {
			Title: "YAML with Merge Keys",
			Content: `# Complex merge key scenarios
base: &base
  name: Base Name
  settings:
    timeout: 30
    retries: 3

extended: &extended
  <<: *base
  settings:
    timeout: 60  # Override timeout
    retries: 3   # Keep retries
    new_setting: true  # Add new setting

multiple_merge:
  # Merging multiple anchors
  <<: [*base, *extended]
  name: "Multiple Merge"  # Override name
  extra: "additional value"
`,
		},
	}
)

func main() {
	// Process all YAML test data
	trees := make(map[string]*golang_yaml_advanced.NodeTree)

	for key, data := range yamlCommonTestData {
		fmt.Printf("\n=== Parsing %s: %s ===\n", key, data.Title)
		tree := parseAndDisplayYAML(data.Title, data.Content)
		trees[key] = tree
	}

	fmt.Println("\n=== Writing YAML back to files ===")

	for key, tree := range trees {
		filename := fmt.Sprintf("output_%s.yaml", key)
		writeYAMLToFile(tree, filename)
		fmt.Printf("Written %s to %s\n", yamlCommonTestData[key].Title, filename)
	}

	fmt.Println("\n=== Verifying written files ===")
	for key := range yamlCommonTestData {
		filename := fmt.Sprintf("output_%s.yaml", key)
		verifyWrittenFile(filename)
	}

	// Test merging
	fmt.Println("\n=== Testing YAML Merge ===")
	testYAMLMerge()

	// Test advanced YAML features
	fmt.Println("\n=== Testing Advanced YAML Features ===")
	testAdvancedYAML()
}

func parseAndDisplayYAML(name string, content string) *golang_yaml_advanced.NodeTree {
	// Use our custom unmarshal function
	tree, err := golang_yaml_advanced.UnmarshalYAML([]byte(content))
	if err != nil {
		log.Fatalf("Failed to parse %s YAML: %v", name, err)
	}

	fmt.Printf("Document Type: %s\n", name)
	fmt.Printf("Number of documents: %d\n", len(tree.Documents))

	if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
		rootNode := tree.Documents[0].Root
		fmt.Println("Root node structure:")
		displayNodeTree(rootNode, 0)

		fmt.Println("\nExtracting comments and metadata:")
		extractComments(rootNode)
	}

	return tree
}

func displayNodeTree(node *golang_yaml_advanced.Node, indent int) {
	if node == nil {
		return
	}

	indentStr := strings.Repeat("  ", indent)

	switch node.Kind {
	case golang_yaml_advanced.DocumentNode:
		fmt.Printf("%sDocument:\n", indentStr)
		for _, child := range node.Children {
			displayNodeTree(child, indent+1)
		}
	case golang_yaml_advanced.MappingNode:
		fmt.Printf("%sMapping:\n", indentStr)
		for i := 0; i < len(node.Children)-1; i += 2 {
			key := node.Children[i]
			value := node.Children[i+1]
			fmt.Printf("%s  %v:\n", indentStr, key.Value)
			displayNodeTree(value, indent+2)
		}
	case golang_yaml_advanced.SequenceNode:
		fmt.Printf("%sSequence:\n", indentStr)
		for _, item := range node.Children {
			fmt.Printf("%s  - ", indentStr)
			displayNodeTree(item, indent+2)
		}
	case golang_yaml_advanced.ScalarNode:
		if node.Value != nil {
			fmt.Printf("%sScalar: %v\n", indentStr, node.Value)
		}
	}
}

func extractComments(node *golang_yaml_advanced.Node) {
	node.Walk(func(n *golang_yaml_advanced.Node) bool {
		hasComments := false

		if len(n.HeadComment) > 0 {
			fmt.Printf("Node at line %d has head comments:\n", n.Line)
			for _, comment := range n.HeadComment {
				fmt.Printf("  %s\n", comment)
			}
			hasComments = true
		}

		if n.LineComment != "" {
			fmt.Printf("Node at line %d has line comment: %s\n", n.Line, n.LineComment)
			hasComments = true
		}

		if len(n.FootComment) > 0 {
			fmt.Printf("Node at line %d has foot comments:\n", n.Line)
			for _, comment := range n.FootComment {
				fmt.Printf("  %s\n", comment)
			}
			hasComments = true
		}

		if hasComments && n.Kind == golang_yaml_advanced.ScalarNode && n.Value != nil {
			fmt.Printf("  -> Value: %v\n", n.Value)
		}

		return true
	})
}

func writeYAMLToFile(tree *golang_yaml_advanced.NodeTree, filename string) {
	yamlBytes, err := tree.ToYAML()
	if err != nil {
		log.Fatalf("Failed to serialize YAML: %v", err)
	}

	err = os.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to write file %s: %v", filename, err)
	}
}

func verifyWrittenFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", filename, err)
	}

	fmt.Printf("\n--- Content of %s ---\n", filename)
	fmt.Print(string(content))
	fmt.Printf("--- End of %s ---\n", filename)
}

func testYAMLMerge() {
	// Parse the base YAML (simple)
	baseData := yamlCommonTestData["simple"]
	baseTree, err := golang_yaml_advanced.UnmarshalYAML([]byte(baseData.Content))
	if err != nil {
		log.Fatalf("Failed to parse base YAML: %v", err)
	}

	// Parse the overlay YAML (toMerge)
	overlayData := yamlMergeTestData["toMerge"]
	overlayTree, err := golang_yaml_advanced.UnmarshalYAML([]byte(overlayData.Content))
	if err != nil {
		log.Fatalf("Failed to parse overlay YAML: %v", err)
	}

	fmt.Println("Base YAML:")
	fmt.Println(baseData.Content)

	fmt.Println("Overlay YAML to merge:")
	fmt.Println(overlayData.Content)

	// Merge the trees
	mergedTree := golang_yaml_advanced.MergeTrees(baseTree, overlayTree)

	// Write the merged result
	mergedBytes, err := mergedTree.ToYAML()
	if err != nil {
		log.Fatalf("Failed to serialize merged YAML: %v", err)
	}

	fmt.Println("==================================================")
	fmt.Println("MERGED YAML CONTENT:")
	fmt.Println("==================================================")
	fmt.Print(string(mergedBytes))

	// Save to file for inspection
	err = os.WriteFile("output_merged.yaml", mergedBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to write merged file: %v", err)
	}
	fmt.Println("\nMerged YAML written to output_merged.yaml")
}

func testAdvancedYAML() {
	for key, data := range yamlAdvancedTestData {
		fmt.Printf("\n=== Testing %s: %s ===\n", key, data.Title)

		// Parse the YAML
		tree, err := golang_yaml_advanced.UnmarshalYAML([]byte(data.Content))
		if err != nil {
			log.Printf("Failed to parse %s: %v", key, err)
			continue
		}

		// Display statistics
		if len(tree.Documents) > 0 {
			fmt.Printf("Documents: %d\n", len(tree.Documents))

			for i, doc := range tree.Documents {
				if doc.Root != nil {
					stats := getNodeStats(doc.Root)
					fmt.Printf("Document %d - Nodes: %d, Max Depth: %d, Anchors: %d\n",
						i+1, stats.nodeCount, stats.maxDepth, len(doc.Anchors))
				}
			}
		}

		// Write it back
		output, err := tree.ToYAML()
		if err != nil {
			log.Printf("Failed to serialize %s: %v", key, err)
			continue
		}

		// Save to file
		filename := fmt.Sprintf("output_advanced_%s.yaml", key)
		err = os.WriteFile(filename, output, 0644)
		if err != nil {
			log.Printf("Failed to write %s: %v", filename, err)
			continue
		}

		fmt.Printf("Written to %s\n", filename)

		// Show first few lines of output
		lines := strings.Split(string(output), "\n")
		preview := 5
		if len(lines) < preview {
			preview = len(lines)
		}
		fmt.Println("Preview:")
		for i := 0; i < preview; i++ {
			fmt.Printf("  %s\n", lines[i])
		}
		if len(lines) > preview {
			fmt.Printf("  ... (%d more lines)\n", len(lines)-preview)
		}
	}
}

type NodeStats struct {
	nodeCount int
	maxDepth  int
}

func getNodeStats(node *golang_yaml_advanced.Node) NodeStats {
	stats := NodeStats{}
	getNodeStatsRecursive(node, 0, &stats)
	return stats
}

func getNodeStatsRecursive(node *golang_yaml_advanced.Node, depth int, stats *NodeStats) {
	if node == nil {
		return
	}

	stats.nodeCount++
	if depth > stats.maxDepth {
		stats.maxDepth = depth
	}

	for _, child := range node.Children {
		getNodeStatsRecursive(child, depth+1, stats)
	}
}
