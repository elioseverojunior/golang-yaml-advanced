package golang_yaml_advanced

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test data fixtures
var (
	simpleYAML = `# Header comment
name: test
value: 123`

	complexYAML = `# Document header
app:
  name: MyApp  # inline comment
  version: 1.0.0
  # Settings section
  settings:
    debug: true
    port: 8080
`

	multiDocYAML = `---
# First document
doc: 1
---
# Second document
doc: 2
...`

	anchorsYAML = `defaults: &defaults
  timeout: 30
  retries: 3

development:
  <<: *defaults
  host: localhost`

	emptyYAML = `# Just comments
# No actual content`

	invalidYAML = `
invalid: [
  unclosed array
`
)

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, tree *NodeTree)
	}{
		{
			name:    "simple YAML",
			input:   simpleYAML,
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				if len(tree.Documents) != 1 {
					t.Errorf("expected 1 document, got %d", len(tree.Documents))
				}
				if tree.Documents[0].Root == nil {
					t.Error("expected non-nil root")
				}
			},
		},
		{
			name:    "complex nested YAML",
			input:   complexYAML,
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				root := tree.Documents[0].Root
				if root.Kind != DocumentNode {
					t.Errorf("expected DocumentNode, got %v", root.Kind)
				}
			},
		},
		{
			name:    "multi-document YAML",
			input:   multiDocYAML,
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				if len(tree.Documents) != 2 {
					t.Errorf("expected 2 documents, got %d", len(tree.Documents))
				}
			},
		},
		{
			name:    "YAML with anchors",
			input:   anchorsYAML,
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				doc := tree.Documents[0]
				if len(doc.Anchors) == 0 {
					t.Error("expected anchors to be processed")
				}
			},
		},
		{
			name:    "empty YAML with comments",
			input:   emptyYAML,
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				if len(tree.Documents) != 1 {
					t.Error("expected 1 document even for empty YAML")
				}
			},
		},
		{
			name:    "invalid YAML",
			input:   invalidYAML,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, tree *NodeTree) {
				if len(tree.Documents) != 1 {
					t.Error("expected 1 empty document")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := UnmarshalYAML([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, tree)
			}
		})
	}
}

func TestToYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", simpleYAML},
		{"complex", complexYAML},
		{"multi-doc", multiDocYAML},
		{"anchors", anchorsYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse
			tree, err := UnmarshalYAML([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Serialize
			output, err := tree.ToYAML()
			if err != nil {
				t.Fatalf("Failed to serialize: %v", err)
			}

			// Parse again to verify
			tree2, err := UnmarshalYAML(output)
			if err != nil {
				t.Fatalf("Failed to parse serialized output: %v", err)
			}

			// Check document count matches
			if len(tree.Documents) != len(tree2.Documents) {
				t.Errorf("Document count mismatch: %d vs %d", len(tree.Documents), len(tree2.Documents))
			}
		})
	}
}

func TestNode_Walk(t *testing.T) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	root := tree.Documents[0].Root

	count := 0
	root.Walk(func(n *Node) bool {
		count++
		return true // continue
	})

	if count == 0 {
		t.Error("Walk should visit at least one node")
	}

	// Test early termination
	limitedCount := 0
	root.Walk(func(n *Node) bool {
		limitedCount++
		return limitedCount < 3 // stop after 3 nodes
	})

	if limitedCount != 3 {
		t.Errorf("Expected walk to stop at 3 nodes, got %d", limitedCount)
	}
}

func TestNode_Find(t *testing.T) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	root := tree.Documents[0].Root

	// Find a node with specific value
	found := root.Find(func(n *Node) bool {
		return n.Kind == ScalarNode && n.Value == "MyApp"
	})

	if found == nil {
		t.Error("Should find node with value 'MyApp'")
	}

	// Find non-existent node
	notFound := root.Find(func(n *Node) bool {
		return n.Kind == ScalarNode && n.Value == "NonExistent"
	})

	if notFound != nil {
		t.Error("Should not find non-existent node")
	}
}

func TestNode_FindAll(t *testing.T) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	root := tree.Documents[0].Root

	// Find all scalar nodes
	scalars := root.FindAll(func(n *Node) bool {
		return n.Kind == ScalarNode
	})

	if len(scalars) == 0 {
		t.Error("Should find scalar nodes")
	}

	// Find all mapping nodes
	mappings := root.FindAll(func(n *Node) bool {
		return n.Kind == MappingNode
	})

	if len(mappings) == 0 {
		t.Error("Should find mapping nodes")
	}
}

func TestNode_GetMapValue(t *testing.T) {
	yamlContent := `
root:
  child1: value1
  child2:
    nested: value2
`
	tree, _ := UnmarshalYAML([]byte(yamlContent))
	root := tree.Documents[0].Root

	// Get root content (skip document node)
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	// Test getting existing value
	rootMap := root.GetMapValue("root")
	if rootMap == nil {
		t.Error("Should find 'root' key")
	}

	child1 := rootMap.GetMapValue("child1")
	if child1 == nil || child1.Value != "value1" {
		t.Error("Should find 'child1' with value 'value1'")
	}

	// Test nested access
	child2 := rootMap.GetMapValue("child2")
	if child2 == nil {
		t.Error("Should find 'child2'")
	}

	nested := child2.GetMapValue("nested")
	if nested == nil || nested.Value != "value2" {
		t.Error("Should find 'nested' with value 'value2'")
	}

	// Test non-existent key
	notFound := rootMap.GetMapValue("nonexistent")
	if notFound != nil {
		t.Error("Should return nil for non-existent key")
	}

	// Test on non-mapping node
	nonMap := child1.GetMapValue("anything")
	if nonMap != nil {
		t.Error("Should return nil when called on non-mapping node")
	}
}

func TestNode_GetSequenceItems(t *testing.T) {
	yamlContent := `
list:
  - item1
  - item2
  - item3
scalar: value
`
	tree, _ := UnmarshalYAML([]byte(yamlContent))
	root := tree.Documents[0].Root

	// Get root content
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	// Test getting sequence items
	list := root.GetMapValue("list")
	if list == nil {
		t.Fatal("Should find 'list'")
	}

	items := list.GetSequenceItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Verify item values
	expectedValues := []string{"item1", "item2", "item3"}
	for i, item := range items {
		if item.Value != expectedValues[i] {
			t.Errorf("Item %d: expected %s, got %v", i, expectedValues[i], item.Value)
		}
	}

	// Test on non-sequence node
	scalar := root.GetMapValue("scalar")
	if scalar == nil {
		t.Fatal("Should find 'scalar'")
	}

	nonSeqItems := scalar.GetSequenceItems()
	if nonSeqItems != nil {
		t.Error("Should return nil for non-sequence node")
	}
}

func TestMergeTrees(t *testing.T) {
	base := `
app:
  name: BaseApp
  version: 1.0.0
  settings:
    debug: false
    port: 3000
`

	overlay := `
app:
  version: 2.0.0
  settings:
    debug: true
    timeout: 30
`

	baseTree, _ := UnmarshalYAML([]byte(base))
	overlayTree, _ := UnmarshalYAML([]byte(overlay))

	merged := MergeTrees(baseTree, overlayTree)
	output, err := merged.ToYAML()
	if err != nil {
		t.Fatalf("Failed to serialize merged tree: %v", err)
	}

	// Parse merged output to verify
	resultTree, _ := UnmarshalYAML(output)
	root := resultTree.Documents[0].Root
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	app := root.GetMapValue("app")
	if app == nil {
		t.Fatal("Should have 'app' in merged result")
	}

	// Check merged values
	version := app.GetMapValue("version")
	if version == nil || version.Value != "2.0.0" {
		t.Error("Version should be overridden to 2.0.0")
	}

	name := app.GetMapValue("name")
	if name == nil || name.Value != "BaseApp" {
		t.Error("Name should be preserved from base")
	}

	settings := app.GetMapValue("settings")
	if settings == nil {
		t.Fatal("Should have settings")
	}

	debug := settings.GetMapValue("debug")
	if debug == nil || debug.Value != true {
		t.Error("Debug should be overridden to true")
	}

	timeout := settings.GetMapValue("timeout")
	if timeout == nil || timeout.Value != int64(30) {
		t.Errorf("Timeout should be added from overlay, got %v (%T)", timeout.Value, timeout.Value)
	}

	port := settings.GetMapValue("port")
	if port == nil || port.Value != int64(3000) {
		t.Errorf("Port should be preserved from base, got %v (%T)", port.Value, port.Value)
	}
}

func TestMergeNodes(t *testing.T) {
	// Test nil cases
	node1 := &Node{Kind: ScalarNode, Value: "test"}

	result := MergeNodes(nil, node1)
	if result == nil || result.Value != node1.Value {
		t.Error("Merging nil with node should return node")
	}

	result = MergeNodes(node1, nil)
	if result == nil || result.Value != node1.Value {
		t.Error("Merging node with nil should return node")
	}

	result = MergeNodes(nil, nil)
	if result != nil {
		t.Error("Merging nil with nil should return nil")
	}

	// Test different kind merges
	scalarNode := &Node{Kind: ScalarNode, Value: "scalar"}
	mappingNode := &Node{Kind: MappingNode, Children: []*Node{}}

	result = MergeNodes(scalarNode, mappingNode)
	if result.Kind != mappingNode.Kind {
		t.Error("Different kinds should return overlay")
	}

	// Test comment preservation
	nodeWithComment := &Node{
		Kind:        ScalarNode,
		Value:       "value1",
		HeadComment: []string{"# Original comment"},
	}
	nodeWithoutComment := &Node{
		Kind:  ScalarNode,
		Value: "value2",
	}

	result = MergeNodes(nodeWithComment, nodeWithoutComment)
	if len(result.HeadComment) == 0 {
		t.Error("Should preserve comments from base when overlay has none")
	}
}

func TestDiffTrees(t *testing.T) {
	yaml1 := `
name: test
value: 123
list:
  - item1
  - item2
`

	yaml2 := `
name: test
value: 456  # changed
list:
  - item1
  - item2
  - item3  # added
new_key: new_value
`

	tree1, _ := UnmarshalYAML([]byte(yaml1))
	tree2, _ := UnmarshalYAML([]byte(yaml2))

	diffs := DiffTrees(tree1, tree2)
	if len(diffs) == 0 {
		t.Error("Should detect differences")
	}

	// Check for specific diff types
	var hasValueChange, hasAddition bool
	for _, diff := range diffs {
		if diff.Type == DiffModified {
			hasValueChange = true
		}
		if diff.Type == DiffAdded {
			hasAddition = true
		}
	}

	if !hasValueChange {
		t.Error("Should detect value change")
	}
	if !hasAddition {
		t.Error("Should detect additions")
	}
}

func TestDiffNodes(t *testing.T) {
	// Test identical nodes
	node1 := &Node{Kind: ScalarNode, Value: "same"}
	node2 := &Node{Kind: ScalarNode, Value: "same"}

	diffs := DiffNodes(node1, node2, "/root")
	if len(diffs) != 0 {
		t.Error("Identical nodes should have no differences")
	}

	// Test value change
	node3 := &Node{Kind: ScalarNode, Value: "different"}
	diffs = DiffNodes(node1, node3, "/root")
	if len(diffs) != 1 || diffs[0].Type != DiffModified {
		t.Error("Should detect value change")
	}

	// Test nil cases
	diffs = DiffNodes(nil, node1, "/root")
	if len(diffs) != 1 || diffs[0].Type != DiffAdded {
		t.Error("Should detect addition when old is nil")
	}

	diffs = DiffNodes(node1, nil, "/root")
	if len(diffs) != 1 || diffs[0].Type != DiffRemoved {
		t.Error("Should detect deletion when new is nil")
	}

	// Test comment changes
	nodeWithComment := &Node{
		Kind:        ScalarNode,
		Value:       "value",
		HeadComment: []string{"# Comment"},
	}
	nodeWithoutComment := &Node{
		Kind:  ScalarNode,
		Value: "value",
	}

	diffs = DiffNodes(nodeWithoutComment, nodeWithComment, "/root")
	hasCommentChange := false
	for _, diff := range diffs {
		if diff.Type == DiffCommentChanged {
			hasCommentChange = true
			break
		}
	}
	if !hasCommentChange {
		t.Error("Should detect comment change")
	}
}

func TestDocument_ToYAML(t *testing.T) {
	// Test empty document
	doc := &Document{}
	output, err := doc.ToYAML()
	if err != nil {
		t.Errorf("Empty document should serialize without error: %v", err)
	}
	if len(output) != 0 {
		t.Error("Empty document should produce empty output")
	}

	// Test document with content
	doc = &Document{
		Root: &Node{
			Kind: DocumentNode,
			Children: []*Node{
				{
					Kind: MappingNode,
					Children: []*Node{
						{Kind: ScalarNode, Value: "key"},
						{Kind: ScalarNode, Value: "value"},
					},
				},
			},
		},
	}

	output, err = doc.ToYAML()
	if err != nil {
		t.Errorf("Should serialize document: %v", err)
	}
	if len(output) == 0 {
		t.Error("Should produce non-empty output")
	}
}

func TestNodeTree_ToYAML(t *testing.T) {
	// Test empty tree
	tree := &NodeTree{}
	output, err := tree.ToYAML()
	if err != nil {
		t.Errorf("Empty tree should serialize without error: %v", err)
	}

	// Test tree with multiple documents
	tree = &NodeTree{
		Documents: []*Document{
			{Root: &Node{Kind: DocumentNode}},
			{Root: &Node{Kind: DocumentNode}},
		},
	}

	output, err = tree.ToYAML()
	if err != nil {
		t.Errorf("Should serialize multi-doc tree: %v", err)
	}

	// Should contain document separator
	if !bytes.Contains(output, []byte("---")) {
		t.Error("Multi-document output should contain separator")
	}
}

func TestConvertFromYAMLNode(t *testing.T) {
	// Test nil input
	result := ConvertFromYAMLNode(nil)
	if result != nil {
		t.Error("Should return nil for nil input")
	}

	// Test scalar node
	yamlNode := &yaml.Node{
		Kind:        yaml.ScalarNode,
		Value:       "test",
		Tag:         "!!str",
		Style:       yaml.DoubleQuotedStyle,
		Line:        1,
		Column:      1,
		HeadComment: "# head",
		LineComment: "# line",
		FootComment: "# foot",
	}

	node := ConvertFromYAMLNode(yamlNode)
	if node == nil {
		t.Fatal("Should convert scalar node")
	}

	if node.Kind != ScalarNode {
		t.Errorf("Expected ScalarNode, got %v", node.Kind)
	}
	if node.Value != "test" {
		t.Errorf("Expected value 'test', got %v", node.Value)
	}
	if node.Tag != "!!str" {
		t.Errorf("Expected tag '!!str', got %s", node.Tag)
	}
	if len(node.HeadComment) == 0 {
		t.Error("Should preserve head comment")
	}

	// Test mapping node
	yamlNode = &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "key"},
			{Kind: yaml.ScalarNode, Value: "value"},
		},
	}

	node = convertFromYAMLNode(yamlNode, nil, nil)
	if node.Kind != MappingNode {
		t.Error("Should convert mapping node")
	}
	if len(node.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(node.Children))
	}

	// Test sequence node
	yamlNode = &yaml.Node{
		Kind: yaml.SequenceNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "item1"},
			{Kind: yaml.ScalarNode, Value: "item2"},
		},
	}

	node = convertFromYAMLNode(yamlNode, nil, nil)
	if node.Kind != SequenceNode {
		t.Error("Should convert sequence node")
	}
	if len(node.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(node.Children))
	}

	// Test anchor node
	anchors := make(map[string]*Node)
	yamlNode = &yaml.Node{
		Kind:   yaml.ScalarNode,
		Value:  "anchor_value",
		Anchor: "test_anchor",
	}

	node = convertFromYAMLNode(yamlNode, nil, anchors)
	if node.Anchor != "test_anchor" {
		t.Error("Should preserve anchor")
	}
	if anchors["test_anchor"] != node {
		t.Error("Should register anchor in map")
	}

	// Test alias node
	yamlNode = &yaml.Node{
		Kind:  yaml.AliasNode,
		Alias: anchors["test_anchor"].convertToYAMLNode(),
	}

	node = convertFromYAMLNode(yamlNode, nil, anchors)
	if node.Alias == nil {
		t.Error("Should set alias reference")
	}
}

func TestConvertToYAMLNode(t *testing.T) {
	// Test nil input
	node := (*Node)(nil)
	result := node.convertToYAMLNode()
	if result != nil {
		t.Error("Should return nil for nil input")
	}

	// Test scalar node
	node = &Node{
		Kind:        ScalarNode,
		Value:       "test",
		Tag:         "!!str",
		Style:       LiteralStyle,
		HeadComment: []string{"# head1", "# head2"},
		LineComment: "# line",
		FootComment: []string{"# foot"},
	}

	yamlNode := node.convertToYAMLNode()
	if yamlNode == nil {
		t.Fatal("Should convert scalar node")
	}

	if yamlNode.Kind != yaml.ScalarNode {
		t.Error("Should convert to yaml.ScalarNode")
	}
	if yamlNode.Value != "test" {
		t.Errorf("Expected value 'test', got %s", yamlNode.Value)
	}
	if yamlNode.Tag != "!!str" {
		t.Errorf("Expected tag '!!str', got %s", yamlNode.Tag)
	}
	if !strings.Contains(yamlNode.HeadComment, "head1") {
		t.Error("Should combine head comments")
	}

	// Test mapping node with children
	node = &Node{
		Kind: MappingNode,
		Children: []*Node{
			{Kind: ScalarNode, Value: "key1"},
			{Kind: ScalarNode, Value: "value1"},
			{Kind: ScalarNode, Value: "key2"},
			{Kind: ScalarNode, Value: "value2"},
		},
	}

	yamlNode = node.convertToYAMLNode()
	if yamlNode.Kind != yaml.MappingNode {
		t.Error("Should convert to yaml.MappingNode")
	}
	if len(yamlNode.Content) != 4 {
		t.Errorf("Expected 4 content nodes, got %d", len(yamlNode.Content))
	}

	// Test alias node
	targetNode := &Node{Kind: ScalarNode, Value: "target"}
	node = &Node{
		Kind:  AliasNode,
		Alias: targetNode,
	}

	yamlNode = node.convertToYAMLNode()
	if yamlNode.Kind != yaml.AliasNode {
		t.Error("Should convert to yaml.AliasNode")
	}
}

func TestNodeKindString(t *testing.T) {
	tests := []struct {
		kind     NodeKind
		expected string
	}{
		{DocumentNode, "DocumentNode"},
		{SequenceNode, "SequenceNode"},
		{MappingNode, "MappingNode"},
		{ScalarNode, "ScalarNode"},
		{AliasNode, "AliasNode"},
		{NodeKind(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("NodeKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNodeStyleString(t *testing.T) {
	tests := []struct {
		style    NodeStyle
		expected string
	}{
		{TaggedStyle, "TaggedStyle"},
		{DoubleQuotedStyle, "DoubleQuotedStyle"},
		{SingleQuotedStyle, "SingleQuotedStyle"},
		{LiteralStyle, "LiteralStyle"},
		{FoldedStyle, "FoldedStyle"},
		{FlowStyle, "FlowStyle"},
		{NodeStyle(0), "DefaultStyle"}, // 0 is DefaultStyle
		{NodeStyle(99), "Unknown(99)"}, // Unknown style returns formatted string
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.style.String(); got != tt.expected {
				t.Errorf("NodeStyle.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCommentProcessing(t *testing.T) {
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
		{
			input:    "# Single line",
			expected: []string{"# Single line"},
		},
		{
			input:    "# Line 1\n# Line 2",
			expected: []string{"# Line 1", "# Line 2"},
		},
		{
			input:    "",
			expected: nil,
		},
		{
			input:    "# Line with\n\n# empty line between",
			expected: []string{"# Line with", "", "# empty line between"},
		},
	}

	for _, tt := range tests {
		result := processComments(tt.input)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("processComments(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("Circular reference detection", func(t *testing.T) {
		// Create a circular reference
		node1 := &Node{Kind: MappingNode}
		node2 := &Node{Kind: MappingNode, Parent: node1}
		node1.Children = []*Node{node2}
		node2.Children = []*Node{node1} // circular

		// This shouldn't cause infinite loop
		count := 0
		visited := make(map[*Node]bool)
		var walk func(*Node)
		walk = func(n *Node) {
			if visited[n] {
				return
			}
			visited[n] = true
			count++
			for _, child := range n.Children {
				walk(child)
			}
		}
		walk(node1)

		if count != 2 {
			t.Errorf("Should handle circular reference, visited %d nodes", count)
		}
	})

	t.Run("Deep nesting", func(t *testing.T) {
		// Create deeply nested structure
		root := &Node{Kind: MappingNode}
		current := root
		depth := 100

		for i := 0; i < depth; i++ {
			child := &Node{Kind: MappingNode, Parent: current}
			current.Children = []*Node{
				{Kind: ScalarNode, Value: fmt.Sprintf("key%d", i)},
				child,
			}
			current = child
		}

		// Should handle deep nesting without stack overflow
		count := 0
		root.Walk(func(n *Node) bool {
			count++
			return true
		})

		if count == 0 {
			t.Error("Should walk deep nested structure")
		}
	})

	t.Run("Large document", func(t *testing.T) {
		// Create a large YAML document
		var sb strings.Builder
		for i := 0; i < 1000; i++ {
			sb.WriteString(fmt.Sprintf("key%d: value%d\n", i, i))
		}

		tree, err := UnmarshalYAML([]byte(sb.String()))
		if err != nil {
			t.Fatalf("Should handle large document: %v", err)
		}

		if len(tree.Documents) == 0 {
			t.Error("Should parse large document")
		}
	})

	t.Run("Special characters in values", func(t *testing.T) {
		specialYAML := `
special: "quotes\"and\\slashes"
unicode: "emoji 😀 and symbols ★"
multiline: |
  Line 1
  Line 2
  Line 3
control: "\n\t\r"
`
		tree, err := UnmarshalYAML([]byte(specialYAML))
		if err != nil {
			t.Fatalf("Should handle special characters: %v", err)
		}

		// Verify round-trip preserves special characters
		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Should serialize special characters: %v", err)
		}

		tree2, err := UnmarshalYAML(output)
		if err != nil {
			t.Fatalf("Should parse serialized special characters: %v", err)
		}

		if len(tree.Documents) != len(tree2.Documents) {
			t.Error("Round-trip should preserve structure")
		}
	})
}

func TestParentChildRelationships(t *testing.T) {
	yamlContent := `
parent:
  child1:
    grandchild1: value1
  child2:
    grandchild2: value2
`
	tree, _ := UnmarshalYAML([]byte(yamlContent))
	root := tree.Documents[0].Root

	// Navigate to nested nodes
	if root.Kind == DocumentNode && len(root.Children) > 0 {
		root = root.Children[0]
	}

	parent := root.GetMapValue("parent")
	if parent == nil {
		t.Fatal("Should find parent")
	}

	child1 := parent.GetMapValue("child1")
	if child1 == nil {
		t.Fatal("Should find child1")
	}

	// Verify parent relationship
	if child1.Parent == nil {
		t.Error("child1 should have parent reference")
	}

	grandchild1 := child1.GetMapValue("grandchild1")
	if grandchild1 == nil {
		t.Fatal("Should find grandchild1")
	}

	// Walk up the tree
	current := grandchild1
	levels := 0
	for current != nil && levels < 10 {
		current = current.Parent
		levels++
	}

	if levels < 3 {
		t.Error("Should be able to walk up the tree")
	}
}

// Benchmark tests
func BenchmarkUnmarshalYAML(b *testing.B) {
	data := []byte(complexYAML)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalYAML(data)
	}
}

func BenchmarkToYAML(b *testing.B) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.ToYAML()
	}
}

func BenchmarkMergeTrees(b *testing.B) {
	base, _ := UnmarshalYAML([]byte(complexYAML))
	overlay, _ := UnmarshalYAML([]byte(simpleYAML))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeTrees(base, overlay)
	}
}

func BenchmarkWalk(b *testing.B) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	root := tree.Documents[0].Root
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Walk(func(n *Node) bool {
			return true
		})
	}
}

func BenchmarkFind(b *testing.B) {
	tree, _ := UnmarshalYAML([]byte(complexYAML))
	root := tree.Documents[0].Root
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.Find(func(n *Node) bool {
			return n.Kind == ScalarNode && n.Value == "MyApp"
		})
	}
}
