package golang_yaml_advanced

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestNodeKindString tests the String method of NodeKind
func TestNodeKindStringComplete(t *testing.T) {
	tests := []struct {
		name string
		kind NodeKind
		want string
	}{
		{"DocumentNode", DocumentNode, "DocumentNode"},
		{"MappingNode", MappingNode, "MappingNode"},
		{"SequenceNode", SequenceNode, "SequenceNode"},
		{"ScalarNode", ScalarNode, "ScalarNode"},
		{"AliasNode", AliasNode, "AliasNode"},
		{"NullNode", NullNode, "NullNode"},
		{"Unknown", NodeKind(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("NodeKind.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNodeStyleString tests the String method of NodeStyle
func TestNodeStyleStringComplete(t *testing.T) {
	tests := []struct {
		name  string
		style NodeStyle
		want  string
	}{
		{"DefaultStyle", DefaultStyle, "DefaultStyle"},
		{"LiteralStyle", LiteralStyle, "LiteralStyle"},
		{"FoldedStyle", FoldedStyle, "FoldedStyle"},
		{"QuotedStyle", QuotedStyle, "QuotedStyle"},
		{"DoubleQuotedStyle", DoubleQuotedStyle, "DoubleQuotedStyle"},
		{"FlowStyle", FlowStyle, "FlowStyle"},
		{"TaggedStyle", TaggedStyle, "TaggedStyle"},
		{"SingleQuotedStyle", SingleQuotedStyle, "SingleQuotedStyle"},
		{"Unknown", NodeStyle(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.String(); got != tt.want {
				t.Errorf("NodeStyle.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewNode tests the NewNode constructor
func TestNewNode(t *testing.T) {
	tests := []struct {
		name string
		kind NodeKind
	}{
		{"DocumentNode", DocumentNode},
		{"MappingNode", MappingNode},
		{"SequenceNode", SequenceNode},
		{"ScalarNode", ScalarNode},
		{"AliasNode", AliasNode},
		{"NullNode", NullNode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewNode(tt.kind)
			if node.Kind != tt.kind {
				t.Errorf("NewNode() Kind = %v, want %v", node.Kind, tt.kind)
			}
			if node.Style != DefaultStyle {
				t.Errorf("NewNode() Style = %v, want DefaultStyle", node.Style)
			}
			if len(node.Children) != 0 {
				t.Errorf("NewNode() Children length = %v, want 0", len(node.Children))
			}
			if len(node.Metadata) != 0 {
				t.Errorf("NewNode() Metadata length = %v, want 0", len(node.Metadata))
			}
		})
	}
}

// TestNewScalarNode tests the NewScalarNode constructor
func TestNewScalarNode(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"String", "test"},
		{"Integer", 42},
		{"Float", 3.14},
		{"Boolean", true},
		{"Nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewScalarNode(tt.value)
			if node.Kind != ScalarNode {
				t.Errorf("NewScalarNode() Kind = %v, want ScalarNode", node.Kind)
			}
			if !reflect.DeepEqual(node.Value, tt.value) {
				t.Errorf("NewScalarNode() Value = %v, want %v", node.Value, tt.value)
			}
		})
	}
}

// TestNewMappingNodeComplete tests the NewMappingNode constructor
func TestNewMappingNodeComplete(t *testing.T) {
	node := NewMappingNode()
	if node.Kind != MappingNode {
		t.Errorf("NewMappingNode() Kind = %v, want MappingNode", node.Kind)
	}
}

// TestNewSequenceNodeComplete tests the NewSequenceNode constructor
func TestNewSequenceNodeComplete(t *testing.T) {
	node := NewSequenceNode()
	if node.Kind != SequenceNode {
		t.Errorf("NewSequenceNode() Kind = %v, want SequenceNode", node.Kind)
	}
}

// TestNodeAddChild tests the AddChild method
func TestNodeAddChild(t *testing.T) {
	parent := NewNode(DocumentNode)
	child := NewNode(ScalarNode)
	child.Value = "test"

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("AddChild() Children length = %v, want 1", len(parent.Children))
	}
	if parent.Children[0] != child {
		t.Errorf("AddChild() child not added correctly")
	}
	if child.Parent != parent {
		t.Errorf("AddChild() parent not set correctly")
	}

	// Test nil child
	parent.AddChild(nil)
	if len(parent.Children) != 1 {
		t.Errorf("AddChild() with nil should not add child, length = %v, want 1", len(parent.Children))
	}
}

// TestNodeAddKeyValue tests the AddKeyValue method
func TestNodeAddKeyValue(t *testing.T) {
	t.Run("ValidMapping", func(t *testing.T) {
		mapping := NewMappingNode()
		key := NewScalarNode("key")
		value := NewScalarNode("value")

		err := mapping.AddKeyValue(key, value)
		if err != nil {
			t.Errorf("AddKeyValue() error = %v", err)
		}

		if len(mapping.Children) != 2 {
			t.Errorf("AddKeyValue() Children length = %v, want 2", len(mapping.Children))
		}
		if mapping.Children[0] != key {
			t.Errorf("AddKeyValue() key not at correct position")
		}
		if mapping.Children[1] != value {
			t.Errorf("AddKeyValue() value not at correct position")
		}
		if value.Key != key {
			t.Errorf("AddKeyValue() key reference not set on value")
		}
	})

	t.Run("NonMappingNode", func(t *testing.T) {
		sequence := NewSequenceNode()
		key := NewScalarNode("key")
		value := NewScalarNode("value")

		err := sequence.AddKeyValue(key, value)
		if err == nil {
			t.Errorf("AddKeyValue() on non-mapping should return error")
		}
	})

	t.Run("NilKey", func(t *testing.T) {
		mapping := NewMappingNode()
		value := NewScalarNode("value")

		err := mapping.AddKeyValue(nil, value)
		if err != nil {
			t.Errorf("AddKeyValue() with nil key error = %v", err)
		}
		if len(mapping.Children) != 0 {
			t.Errorf("AddKeyValue() with nil key should not add children")
		}
	})
}

// TestNodeAddSequenceItem tests the AddSequenceItem method
func TestNodeAddSequenceItem(t *testing.T) {
	t.Run("ValidSequence", func(t *testing.T) {
		sequence := NewSequenceNode()
		item := NewScalarNode("item")

		err := sequence.AddSequenceItem(item)
		if err != nil {
			t.Errorf("AddSequenceItem() error = %v", err)
		}

		if len(sequence.Children) != 1 {
			t.Errorf("AddSequenceItem() Children length = %v, want 1", len(sequence.Children))
		}
		if sequence.Children[0] != item {
			t.Errorf("AddSequenceItem() item not added correctly")
		}
	})

	t.Run("NonSequenceNode", func(t *testing.T) {
		mapping := NewMappingNode()
		item := NewScalarNode("item")

		err := mapping.AddSequenceItem(item)
		if err == nil {
			t.Errorf("AddSequenceItem() on non-sequence should return error")
		}
	})
}

// TestNodeGetMapValue tests the GetMapValue method
func TestNodeGetMapValue(t *testing.T) {
	t.Run("ExistingKey", func(t *testing.T) {
		mapping := NewMappingNode()
		key1 := NewScalarNode("key1")
		value1 := NewScalarNode("value1")
		key2 := NewScalarNode("key2")
		value2 := NewScalarNode("value2")

		mapping.AddKeyValue(key1, value1)
		mapping.AddKeyValue(key2, value2)

		result := mapping.GetMapValue("key1")
		if result != value1 {
			t.Errorf("GetMapValue() = %v, want %v", result, value1)
		}

		result = mapping.GetMapValue("key2")
		if result != value2 {
			t.Errorf("GetMapValue() = %v, want %v", result, value2)
		}
	})

	t.Run("NonExistentKey", func(t *testing.T) {
		mapping := NewMappingNode()
		key := NewScalarNode("key")
		value := NewScalarNode("value")
		mapping.AddKeyValue(key, value)

		result := mapping.GetMapValue("nonexistent")
		if result != nil {
			t.Errorf("GetMapValue() for non-existent key = %v, want nil", result)
		}
	})

	t.Run("NonMappingNode", func(t *testing.T) {
		sequence := NewSequenceNode()
		result := sequence.GetMapValue("key")
		if result != nil {
			t.Errorf("GetMapValue() on non-mapping = %v, want nil", result)
		}
	})
}

// TestNodeGetSequenceItems tests the GetSequenceItems method
func TestNodeGetSequenceItems(t *testing.T) {
	t.Run("ValidSequence", func(t *testing.T) {
		sequence := NewSequenceNode()
		item1 := NewScalarNode("item1")
		item2 := NewScalarNode("item2")

		sequence.AddSequenceItem(item1)
		sequence.AddSequenceItem(item2)

		items := sequence.GetSequenceItems()
		if len(items) != 2 {
			t.Errorf("GetSequenceItems() length = %v, want 2", len(items))
		}
		if items[0] != item1 || items[1] != item2 {
			t.Errorf("GetSequenceItems() items not in correct order")
		}
	})

	t.Run("NonSequenceNode", func(t *testing.T) {
		mapping := NewMappingNode()
		items := mapping.GetSequenceItems()
		if items != nil {
			t.Errorf("GetSequenceItems() on non-sequence = %v, want nil", items)
		}
	})
}

// TestNodeWalk tests the Walk method
func TestNodeWalk(t *testing.T) {
	root := NewMappingNode()
	key1 := NewScalarNode("key1")
	value1 := NewScalarNode("value1")
	key2 := NewScalarNode("key2")
	value2 := NewSequenceNode()
	item1 := NewScalarNode("item1")
	item2 := NewScalarNode("item2")

	root.AddKeyValue(key1, value1)
	root.AddKeyValue(key2, value2)
	value2.AddSequenceItem(item1)
	value2.AddSequenceItem(item2)

	var visited []*Node
	root.Walk(func(n *Node) bool {
		visited = append(visited, n)
		return true
	})

	if len(visited) != 7 {
		t.Errorf("Walk() visited %v nodes, want 7", len(visited))
	}

	// Test early termination
	visited = visited[:0]
	root.Walk(func(n *Node) bool {
		visited = append(visited, n)
		return len(visited) < 3
	})

	if len(visited) != 3 {
		t.Errorf("Walk() with early termination visited %v nodes, want 3", len(visited))
	}
}

// TestNodeFind tests the Find method
func TestNodeFind(t *testing.T) {
	root := NewMappingNode()
	key := NewScalarNode("target")
	value := NewScalarNode("found")
	root.AddKeyValue(key, value)

	result := root.Find(func(n *Node) bool {
		return n.Kind == ScalarNode && n.Value == "found"
	})

	if result != value {
		t.Errorf("Find() = %v, want %v", result, value)
	}

	result = root.Find(func(n *Node) bool {
		return n.Kind == ScalarNode && n.Value == "notfound"
	})

	if result != nil {
		t.Errorf("Find() for non-existent = %v, want nil", result)
	}
}

// TestNodeFindAll tests the FindAll method
func TestNodeFindAll(t *testing.T) {
	root := NewMappingNode()
	key1 := NewScalarNode("key1")
	value1 := NewScalarNode("value")
	key2 := NewScalarNode("key2")
	value2 := NewScalarNode("value")

	root.AddKeyValue(key1, value1)
	root.AddKeyValue(key2, value2)

	results := root.FindAll(func(n *Node) bool {
		return n.Kind == ScalarNode && n.Value == "value"
	})

	if len(results) != 2 {
		t.Errorf("FindAll() found %v nodes, want 2", len(results))
	}
}

// TestNodePathComplete tests the Path method
func TestNodePathComplete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Node
		expected string
	}{
		{
			name: "RootNode",
			setup: func() *Node {
				return NewNode(DocumentNode)
			},
			expected: "$",
		},
		{
			name: "MappingChild",
			setup: func() *Node {
				root := NewMappingNode()
				key := NewScalarNode("child")
				value := NewScalarNode("value")
				root.AddKeyValue(key, value)
				return value
			},
			expected: "$.child",
		},
		{
			name: "SequenceChild",
			setup: func() *Node {
				root := NewSequenceNode()
				item := NewScalarNode("item")
				root.AddSequenceItem(item)
				return item
			},
			expected: "$[0]",
		},
		{
			name: "NestedPath",
			setup: func() *Node {
				root := NewMappingNode()
				key := NewScalarNode("parent")
				seq := NewSequenceNode()
				root.AddKeyValue(key, seq)
				item := NewScalarNode("item")
				seq.AddSequenceItem(item)
				return item
			},
			expected: "$.parent[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.setup()
			path := node.Path()
			if path != tt.expected {
				t.Errorf("Path() = %v, want %v", path, tt.expected)
			}
		})
	}
}

// TestNodeClone tests the Clone method
func TestNodeClone(t *testing.T) {
	t.Run("SimpleNode", func(t *testing.T) {
		original := NewScalarNode("test")
		original.Tag = "!!str"
		original.Anchor = "anchor1"
		original.HeadComment = []string{"# comment"}
		original.LineComment = "inline"
		original.FootComment = []string{"# foot"}
		original.Metadata["key"] = "value"

		clone := original.Clone()

		if clone == original {
			t.Error("Clone() returned same instance")
		}
		if clone.Value != original.Value {
			t.Errorf("Clone() Value = %v, want %v", clone.Value, original.Value)
		}
		if clone.Tag != original.Tag {
			t.Errorf("Clone() Tag = %v, want %v", clone.Tag, original.Tag)
		}
		if clone.Anchor != original.Anchor {
			t.Errorf("Clone() Anchor = %v, want %v", clone.Anchor, original.Anchor)
		}
		if !reflect.DeepEqual(clone.HeadComment, original.HeadComment) {
			t.Error("Clone() HeadComment not equal")
		}
		if clone.LineComment != original.LineComment {
			t.Error("Clone() LineComment not equal")
		}
		if !reflect.DeepEqual(clone.FootComment, original.FootComment) {
			t.Error("Clone() FootComment not equal")
		}
		if !reflect.DeepEqual(clone.Metadata, original.Metadata) {
			t.Error("Clone() Metadata not equal")
		}
	})

	t.Run("ComplexTree", func(t *testing.T) {
		original := NewMappingNode()
		key := NewScalarNode("key")
		seq := NewSequenceNode()
		original.AddKeyValue(key, seq)
		item1 := NewScalarNode("item1")
		item2 := NewScalarNode("item2")
		seq.AddSequenceItem(item1)
		seq.AddSequenceItem(item2)

		clone := original.Clone()

		if clone == original {
			t.Error("Clone() returned same instance")
		}
		if len(clone.Children) != len(original.Children) {
			t.Errorf("Clone() Children length = %v, want %v", len(clone.Children), len(original.Children))
		}

		// Verify deep clone
		clonedSeq := clone.Children[1]
		if clonedSeq == seq {
			t.Error("Clone() did not deep clone children")
		}
		if len(clonedSeq.Children) != 2 {
			t.Error("Clone() sequence items not cloned correctly")
		}
	})

	t.Run("NilNode", func(t *testing.T) {
		var original *Node
		clone := original.Clone()
		if clone != nil {
			t.Errorf("Clone() of nil = %v, want nil", clone)
		}
	})
}

// TestNodeStringComplete tests the String method
func TestNodeStringComplete(t *testing.T) {
	tests := []struct {
		name     string
		node     func() *Node
		contains string
	}{
		{
			name: "ScalarNode",
			node: func() *Node {
				return NewScalarNode("test")
			},
			contains: "test",
		},
		{
			name: "MappingNode",
			node: func() *Node {
				mapping := NewMappingNode()
				mapping.AddKeyValue(NewScalarNode("key"), NewScalarNode("value"))
				return mapping
			},
			contains: "key: value",
		},
		{
			name: "SequenceNode",
			node: func() *Node {
				seq := NewSequenceNode()
				seq.AddSequenceItem(NewScalarNode("item1"))
				seq.AddSequenceItem(NewScalarNode("item2"))
				return seq
			},
			contains: "- item",
		},
		{
			name: "NullNode",
			node: func() *Node {
				return NewNode(NullNode)
			},
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node().String()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("String() = %v, want to contain %v", result, tt.contains)
			}
		})
	}
}

// TestNodeIsNull tests the IsNull method
func TestNodeIsNull(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want bool
	}{
		{"NilNode", nil, true},
		{"NullKindNode", NewNode(NullNode), true},
		{"ScalarNilValue", &Node{Kind: ScalarNode, Value: nil}, true},
		{"ScalarWithValue", NewScalarNode("test"), false},
		{"MappingNode", NewMappingNode(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.IsNull(); got != tt.want {
				t.Errorf("IsNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNodeRemoveComplete tests the Remove method
func TestNodeRemoveComplete(t *testing.T) {
	t.Run("RemoveFromMapping", func(t *testing.T) {
		parent := NewMappingNode()
		key := NewScalarNode("key")
		value := NewScalarNode("value")
		parent.AddKeyValue(key, value)

		err := value.Remove()
		if err != nil {
			t.Errorf("Remove() error = %v", err)
		}

		if len(parent.Children) != 1 {
			t.Errorf("Remove() parent children = %v, want 1", len(parent.Children))
		}
		if value.Parent != nil {
			t.Error("Remove() did not clear parent reference")
		}
	})

	t.Run("RemoveFromSequence", func(t *testing.T) {
		parent := NewSequenceNode()
		item1 := NewScalarNode("item1")
		item2 := NewScalarNode("item2")
		parent.AddSequenceItem(item1)
		parent.AddSequenceItem(item2)

		err := item1.Remove()
		if err != nil {
			t.Errorf("Remove() error = %v", err)
		}

		if len(parent.Children) != 1 {
			t.Errorf("Remove() parent children = %v, want 1", len(parent.Children))
		}
		if parent.Children[0] != item2 {
			t.Error("Remove() did not shift remaining children correctly")
		}
	})

	t.Run("RemoveRootNode", func(t *testing.T) {
		root := NewNode(DocumentNode)
		err := root.Remove()
		if err == nil {
			t.Error("Remove() on root should return error")
		}
	})
}

// TestNodeReplaceWithComplete tests the ReplaceWith method
func TestNodeReplaceWithComplete(t *testing.T) {
	t.Run("ReplaceInMapping", func(t *testing.T) {
		parent := NewMappingNode()
		key := NewScalarNode("key")
		oldValue := NewScalarNode("old")
		parent.AddKeyValue(key, oldValue)

		newValue := NewScalarNode("new")
		err := oldValue.ReplaceWith(newValue)
		if err != nil {
			t.Errorf("ReplaceWith() error = %v", err)
		}

		if parent.Children[1] != newValue {
			t.Error("ReplaceWith() did not replace node correctly")
		}
		if newValue.Parent != parent {
			t.Error("ReplaceWith() did not set parent correctly")
		}
		if newValue.Key != key {
			t.Error("ReplaceWith() did not preserve key reference")
		}
		if oldValue.Parent != nil {
			t.Error("ReplaceWith() did not clear old parent reference")
		}
	})

	t.Run("ReplaceRootNode", func(t *testing.T) {
		root := NewNode(DocumentNode)
		replacement := NewNode(DocumentNode)
		err := root.ReplaceWith(replacement)
		if err == nil {
			t.Error("ReplaceWith() on root should return error")
		}
	})
}

// TestNodeConvertToYAMLNode tests the convertToYAMLNode method
func TestNodeConvertToYAMLNode(t *testing.T) {
	t.Run("ScalarNode", func(t *testing.T) {
		node := NewScalarNode("test")
		node.Tag = "!!str"
		node.Anchor = "anchor1"
		node.HeadComment = []string{"# head"}
		node.LineComment = "inline"
		node.FootComment = []string{"# foot"}

		yamlNode := node.convertToYAMLNode()

		if yamlNode.Kind != yaml.ScalarNode {
			t.Errorf("convertToYAMLNode() Kind = %v, want ScalarNode", yamlNode.Kind)
		}
		if yamlNode.Value != "test" {
			t.Errorf("convertToYAMLNode() Value = %v, want test", yamlNode.Value)
		}
		if yamlNode.Tag != "!!str" {
			t.Errorf("convertToYAMLNode() Tag = %v, want !!str", yamlNode.Tag)
		}
		if yamlNode.Anchor != "anchor1" {
			t.Errorf("convertToYAMLNode() Anchor = %v, want anchor1", yamlNode.Anchor)
		}
	})

	t.Run("MappingNode", func(t *testing.T) {
		node := NewMappingNode()
		key := NewScalarNode("key")
		value := NewScalarNode("value")
		node.AddKeyValue(key, value)

		yamlNode := node.convertToYAMLNode()

		if yamlNode.Kind != yaml.MappingNode {
			t.Errorf("convertToYAMLNode() Kind = %v, want MappingNode", yamlNode.Kind)
		}
		if len(yamlNode.Content) != 2 {
			t.Errorf("convertToYAMLNode() Content length = %v, want 2", len(yamlNode.Content))
		}
	})

	t.Run("NilNode", func(t *testing.T) {
		var node *Node
		yamlNode := node.convertToYAMLNode()
		if yamlNode != nil {
			t.Errorf("convertToYAMLNode() of nil = %v, want nil", yamlNode)
		}
	})
}

// TestNodeToYAMLNode tests the ToYAMLNode method
func TestNodeToYAMLNode(t *testing.T) {
	t.Run("CompleteConversion", func(t *testing.T) {
		node := NewMappingNode()
		node.Style = FlowStyle
		key := NewScalarNode("key")
		value := NewScalarNode("value")
		value.Style = DoubleQuotedStyle
		node.AddKeyValue(key, value)

		yamlNode := node.ToYAMLNode()

		if yamlNode.Kind != yaml.MappingNode {
			t.Errorf("ToYAMLNode() Kind = %v, want MappingNode", yamlNode.Kind)
		}
		if yamlNode.Style != yaml.FlowStyle {
			t.Errorf("ToYAMLNode() Style = %v, want FlowStyle", yamlNode.Style)
		}
		if len(yamlNode.Content) != 2 {
			t.Errorf("ToYAMLNode() Content length = %v, want 2", len(yamlNode.Content))
		}
		if yamlNode.Content[1].Style != yaml.DoubleQuotedStyle {
			t.Error("ToYAMLNode() did not preserve child style")
		}
	})
}

// TestNewNodeTree tests the NewNodeTree constructor
func TestNewNodeTree(t *testing.T) {
	tree := NewNodeTree()
	if tree == nil {
		t.Fatal("NewNodeTree() returned nil")
	}
	if len(tree.Documents) != 0 {
		t.Errorf("NewNodeTree() Documents length = %v, want 0", len(tree.Documents))
	}
}

// TestNodeTreeAddDocument tests the AddDocument method
func TestNodeTreeAddDocument(t *testing.T) {
	tree := NewNodeTree()
	doc := tree.AddDocument()

	if doc == nil {
		t.Fatal("AddDocument() returned nil")
	}
	if len(tree.Documents) != 1 {
		t.Errorf("AddDocument() Documents length = %v, want 1", len(tree.Documents))
	}
	if tree.Current != doc {
		t.Error("AddDocument() did not set Current document")
	}
	if doc.Anchors == nil {
		t.Error("AddDocument() did not initialize Anchors map")
	}
}

// TestDocumentSetRoot tests the SetRoot method
func TestDocumentSetRoot(t *testing.T) {
	doc := &Document{Anchors: make(map[string]*Node)}
	root := NewNode(DocumentNode)
	root.Parent = NewNode(MappingNode) // Set a parent to test it gets cleared

	doc.SetRoot(root)

	if doc.Root != root {
		t.Error("SetRoot() did not set root")
	}
	if root.Parent != nil {
		t.Error("SetRoot() did not clear parent reference")
	}

	// Test nil root
	doc.SetRoot(nil)
	if doc.Root != nil {
		t.Error("SetRoot(nil) did not clear root")
	}
}

// TestDocumentRegisterAnchor tests the RegisterAnchor method
func TestDocumentRegisterAnchor(t *testing.T) {
	doc := &Document{}
	node := NewScalarNode("test")

	doc.RegisterAnchor("anchor1", node)

	if doc.Anchors == nil {
		t.Fatal("RegisterAnchor() did not initialize Anchors map")
	}
	if doc.Anchors["anchor1"] != node {
		t.Error("RegisterAnchor() did not register anchor correctly")
	}
	if node.Anchor != "anchor1" {
		t.Errorf("RegisterAnchor() did not set node anchor = %v, want anchor1", node.Anchor)
	}
}

// TestDocumentGetAnchor tests the GetAnchor method
func TestDocumentGetAnchor(t *testing.T) {
	doc := &Document{Anchors: make(map[string]*Node)}
	node := NewScalarNode("test")
	doc.RegisterAnchor("anchor1", node)

	result := doc.GetAnchor("anchor1")
	if result != node {
		t.Errorf("GetAnchor() = %v, want %v", result, node)
	}

	result = doc.GetAnchor("nonexistent")
	if result != nil {
		t.Errorf("GetAnchor() for non-existent = %v, want nil", result)
	}
}

// TestNodeTreeMergeComplete tests the Merge method
func TestNodeTreeMergeComplete(t *testing.T) {
	tree1 := NewNodeTree()
	doc1 := tree1.AddDocument()
	doc1.SetRoot(NewScalarNode("doc1"))

	tree2 := NewNodeTree()
	doc2 := tree2.AddDocument()
	doc2.SetRoot(NewScalarNode("doc2"))

	tree1.Merge(tree2)

	if len(tree1.Documents) != 2 {
		t.Errorf("Merge() Documents length = %v, want 2", len(tree1.Documents))
	}
	if tree1.Documents[1] != doc2 {
		t.Error("Merge() did not append documents correctly")
	}
}

// TestMergeNodesComplete tests the MergeNodes function
func TestMergeNodesComplete(t *testing.T) {
	t.Run("BothNil", func(t *testing.T) {
		result := MergeNodes(nil, nil)
		if result != nil {
			t.Errorf("MergeNodes(nil, nil) = %v, want nil", result)
		}
	})

	t.Run("BaseNil", func(t *testing.T) {
		overlay := NewScalarNode("test")
		result := MergeNodes(nil, overlay)
		if result == nil {
			t.Fatal("MergeNodes(nil, overlay) returned nil")
		}
		if result == overlay {
			t.Error("MergeNodes() should return clone, not original")
		}
		if result.Value != overlay.Value {
			t.Error("MergeNodes() did not preserve overlay value")
		}
	})

	t.Run("OverlayNil", func(t *testing.T) {
		base := NewScalarNode("test")
		result := MergeNodes(base, nil)
		if result == nil {
			t.Fatal("MergeNodes(base, nil) returned nil")
		}
		if result == base {
			t.Error("MergeNodes() should return clone, not original")
		}
		if result.Value != base.Value {
			t.Error("MergeNodes() did not preserve base value")
		}
	})

	t.Run("MergeMappings", func(t *testing.T) {
		base := NewMappingNode()
		base.AddKeyValue(NewScalarNode("key1"), NewScalarNode("base1"))
		base.AddKeyValue(NewScalarNode("key2"), NewScalarNode("base2"))

		overlay := NewMappingNode()
		overlay.AddKeyValue(NewScalarNode("key2"), NewScalarNode("overlay2"))
		overlay.AddKeyValue(NewScalarNode("key3"), NewScalarNode("overlay3"))

		result := MergeNodes(base, overlay)

		if result.Kind != MappingNode {
			t.Error("MergeNodes() did not preserve mapping kind")
		}

		val1 := result.GetMapValue("key1")
		if val1 == nil || val1.Value != "base1" {
			t.Error("MergeNodes() did not preserve base-only key")
		}

		val2 := result.GetMapValue("key2")
		if val2 == nil || val2.Value != "overlay2" {
			t.Error("MergeNodes() did not override with overlay value")
		}

		val3 := result.GetMapValue("key3")
		if val3 == nil || val3.Value != "overlay3" {
			t.Error("MergeNodes() did not add overlay-only key")
		}
	})

	t.Run("MergeSequences", func(t *testing.T) {
		base := NewSequenceNode()
		base.AddSequenceItem(NewScalarNode("item1"))

		overlay := NewSequenceNode()
		overlay.AddSequenceItem(NewScalarNode("item2"))

		result := MergeNodes(base, overlay)

		if result.Kind != SequenceNode {
			t.Error("MergeNodes() did not preserve sequence kind")
		}
		if len(result.Children) != 2 {
			t.Errorf("MergeNodes() sequence length = %v, want 2", len(result.Children))
		}
	})

	t.Run("DifferentTypes", func(t *testing.T) {
		base := NewScalarNode("scalar")
		overlay := NewMappingNode()

		result := MergeNodes(base, overlay)

		if result.Kind != MappingNode {
			t.Error("MergeNodes() should use overlay type for different types")
		}
	})

	t.Run("PreserveComments", func(t *testing.T) {
		base := NewScalarNode("base")
		base.HeadComment = []string{"# base comment"}

		overlay := NewScalarNode("overlay")
		overlay.LineComment = "overlay inline"

		result := MergeNodes(base, overlay)

		if result.Value != "overlay" {
			t.Error("MergeNodes() did not use overlay value")
		}
		if len(result.HeadComment) == 0 || result.HeadComment[0] != "# base comment" {
			t.Error("MergeNodes() did not preserve base comment when overlay has none")
		}
		if result.LineComment != "overlay inline" {
			t.Error("MergeNodes() did not preserve overlay comment")
		}
	})
}

// TestMergeDocuments tests the MergeDocuments function
func TestMergeDocuments(t *testing.T) {
	t.Run("BothNil", func(t *testing.T) {
		result := MergeDocuments(nil, nil)
		if result != nil {
			t.Errorf("MergeDocuments(nil, nil) = %v, want nil", result)
		}
	})

	t.Run("MergeRoots", func(t *testing.T) {
		base := &Document{
			Root:    NewScalarNode("base"),
			Version: "1.1",
			Anchors: make(map[string]*Node),
		}

		overlay := &Document{
			Root:    NewScalarNode("overlay"),
			Version: "1.2",
			Anchors: make(map[string]*Node),
		}

		result := MergeDocuments(base, overlay)

		if result == nil {
			t.Fatal("MergeDocuments() returned nil")
		}
		if result.Version != "1.1" {
			t.Error("MergeDocuments() did not preserve base version")
		}
		// Check that the root content was merged
		if result.Root == nil || result.Root.Kind != DocumentNode {
			t.Error("MergeDocuments() did not create document node")
		}
	})

	t.Run("MergeDirectives", func(t *testing.T) {
		base := &Document{
			Directives: []Directive{
				{Name: "TAG", Value: "base"},
				{Name: "YAML", Value: "1.1"},
			},
			Anchors: make(map[string]*Node),
		}

		overlay := &Document{
			Directives: []Directive{
				{Name: "TAG", Value: "overlay"},
				{Name: "NEW", Value: "value"},
			},
			Anchors: make(map[string]*Node),
		}

		result := MergeDocuments(base, overlay)

		if len(result.Directives) != 3 {
			t.Errorf("MergeDocuments() Directives length = %v, want 3", len(result.Directives))
		}
	})
}

// TestMergeTreesComplete tests the MergeTrees function
func TestMergeTreesComplete(t *testing.T) {
	t.Run("BaseNil", func(t *testing.T) {
		overlay := NewNodeTree()
		result := MergeTrees(nil, overlay)
		if result != overlay {
			t.Error("MergeTrees(nil, overlay) should return overlay")
		}
	})

	t.Run("OverlayNil", func(t *testing.T) {
		base := NewNodeTree()
		result := MergeTrees(base, nil)
		if result != base {
			t.Error("MergeTrees(base, nil) should return base")
		}
	})

	t.Run("MergeFirstDocuments", func(t *testing.T) {
		base := NewNodeTree()
		baseDoc := base.AddDocument()
		baseDoc.Root = NewScalarNode("base")

		overlay := NewNodeTree()
		overlayDoc := overlay.AddDocument()
		overlayDoc.Root = NewScalarNode("overlay")

		result := MergeTrees(base, overlay)

		if len(result.Documents) != 1 {
			t.Errorf("MergeTrees() Documents length = %v, want 1", len(result.Documents))
		}
		if result.Current == nil {
			t.Error("MergeTrees() did not set Current document")
		}
	})

	t.Run("MultipleDocuments", func(t *testing.T) {
		base := NewNodeTree()
		base.AddDocument()
		base.AddDocument()

		overlay := NewNodeTree()
		overlay.AddDocument()
		overlay.AddDocument()
		overlay.AddDocument()

		result := MergeTrees(base, overlay)

		if len(result.Documents) != 4 {
			t.Errorf("MergeTrees() Documents length = %v, want 4", len(result.Documents))
		}
	})
}

// TestDocumentToYAML tests the ToYAML method
func TestDocumentToYAML(t *testing.T) {
	t.Run("EmptyDocument", func(t *testing.T) {
		doc := &Document{}
		result, err := doc.ToYAML()
		if err != nil {
			t.Errorf("ToYAML() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("ToYAML() for empty document = %v bytes, want 0", len(result))
		}
	})

	t.Run("DocumentWithContent", func(t *testing.T) {
		doc := &Document{}
		root := NewNode(DocumentNode)
		mapping := NewMappingNode()
		mapping.AddKeyValue(NewScalarNode("key"), NewScalarNode("value"))
		root.AddChild(mapping)
		doc.Root = root

		result, err := doc.ToYAML()
		if err != nil {
			t.Errorf("ToYAML() error = %v", err)
		}
		if !strings.Contains(string(result), "key: value") {
			t.Errorf("ToYAML() = %v, want to contain 'key: value'", string(result))
		}
	})

	t.Run("DocumentWithOnlyComments", func(t *testing.T) {
		doc := &Document{}
		root := NewNode(DocumentNode)
		root.HeadComment = []string{"# Comment only"}
		doc.Root = root

		result, err := doc.ToYAML()
		if err != nil {
			t.Errorf("ToYAML() error = %v", err)
		}
		if !strings.Contains(string(result), "# Comment only") {
			t.Errorf("ToYAML() = %v, want to contain comment", string(result))
		}
	})
}

// TestNodeTreeToYAML tests the ToYAML method
func TestNodeTreeToYAML(t *testing.T) {
	t.Run("EmptyTree", func(t *testing.T) {
		tree := NewNodeTree()
		result, err := tree.ToYAML()
		if err != nil {
			t.Errorf("ToYAML() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("ToYAML() for empty tree = %v bytes, want 0", len(result))
		}
	})

	t.Run("MultipleDocuments", func(t *testing.T) {
		tree := NewNodeTree()

		doc1 := tree.AddDocument()
		root1 := NewNode(DocumentNode)
		root1.AddChild(NewScalarNode("doc1"))
		doc1.Root = root1

		doc2 := tree.AddDocument()
		root2 := NewNode(DocumentNode)
		root2.AddChild(NewScalarNode("doc2"))
		doc2.Root = root2

		result, err := tree.ToYAML()
		if err != nil {
			t.Errorf("ToYAML() error = %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, "---") {
			t.Error("ToYAML() should contain document separator")
		}
		if !strings.Contains(resultStr, "doc1") {
			t.Error("ToYAML() should contain first document")
		}
		if !strings.Contains(resultStr, "doc2") {
			t.Error("ToYAML() should contain second document")
		}
	})
}

// TestUnmarshalYAMLComplete tests the UnmarshalYAML function
func TestUnmarshalYAMLComplete(t *testing.T) {
	t.Run("EmptyInput", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte{})
		if err != nil {
			t.Errorf("UnmarshalYAML() error = %v", err)
		}
		if tree == nil {
			t.Fatal("UnmarshalYAML() returned nil tree")
		}
		if len(tree.Documents) != 1 {
			t.Errorf("UnmarshalYAML() Documents length = %v, want 1", len(tree.Documents))
		}
	})

	t.Run("SimpleYAML", func(t *testing.T) {
		input := []byte("key: value")
		tree, err := UnmarshalYAML(input)
		if err != nil {
			t.Errorf("UnmarshalYAML() error = %v", err)
		}
		if tree == nil || len(tree.Documents) == 0 {
			t.Fatal("UnmarshalYAML() returned empty tree")
		}

		doc := tree.Documents[0]
		if doc.Root == nil {
			t.Fatal("UnmarshalYAML() document has nil root")
		}
	})

	t.Run("OnlyComments", func(t *testing.T) {
		input := []byte("# Just a comment\n# Another comment")
		tree, err := UnmarshalYAML(input)
		if err != nil {
			t.Errorf("UnmarshalYAML() error = %v", err)
		}
		if tree == nil || len(tree.Documents) == 0 {
			t.Fatal("UnmarshalYAML() returned empty tree for comments-only")
		}

		doc := tree.Documents[0]
		if doc.Root == nil {
			t.Fatal("UnmarshalYAML() should create root for comments")
		}
		if len(doc.Root.HeadComment) == 0 {
			t.Error("UnmarshalYAML() did not preserve comments")
		}
	})

	t.Run("MultiDocument", func(t *testing.T) {
		input := []byte("---\nkey1: value1\n---\nkey2: value2")
		tree, err := UnmarshalYAML(input)
		if err != nil {
			t.Errorf("UnmarshalYAML() error = %v", err)
		}
		if len(tree.Documents) != 2 {
			t.Errorf("UnmarshalYAML() Documents length = %v, want 2", len(tree.Documents))
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		input := []byte("invalid: [unclosed")
		_, err := UnmarshalYAML(input)
		if err == nil {
			t.Error("UnmarshalYAML() should return error for invalid YAML")
		}
	})
}

// TestSplitDocuments tests the splitDocuments function
func TestSplitDocuments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Empty", "", 0},
		{"SingleImplicit", "key: value", 1},
		{"SingleExplicit", "---\nkey: value", 1},
		{"Multiple", "---\ndoc1\n---\ndoc2", 2},
		{"WithEnd", "---\ndoc1\n...", 1},
		{"MixedMarkers", "doc1\n---\ndoc2\n...\n---\ndoc3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitDocuments(tt.input)
			if len(result) != tt.expected {
				t.Errorf("splitDocuments() returned %v documents, want %v", len(result), tt.expected)
			}
		})
	}
}

// TestResolveAnchorsComplete tests the resolveAnchors function
func TestResolveAnchorsComplete(t *testing.T) {
	doc := &Document{Anchors: make(map[string]*Node)}

	// Create nodes with anchors
	node1 := NewScalarNode("value1")
	node1.Anchor = "anchor1"

	// Create alias node
	aliasNode := NewNode(AliasNode)
	aliasNode.Value = "*anchor1"

	// Create root with both nodes
	root := NewMappingNode()
	root.AddKeyValue(NewScalarNode("original"), node1)
	root.AddKeyValue(NewScalarNode("alias"), aliasNode)

	resolveAnchors(root, doc)

	// Check anchor was registered
	if doc.Anchors["anchor1"] != node1 {
		t.Error("resolveAnchors() did not register anchor")
	}

	// Check alias was resolved
	if aliasNode.Alias != node1 {
		t.Error("resolveAnchors() did not resolve alias")
	}
}

// TestDiffType tests the String method of DiffType
func TestDiffTypeStringComplete(t *testing.T) {
	tests := []struct {
		diffType DiffType
		expected string
	}{
		{DiffNone, "None"},
		{DiffAdded, "Added"},
		{DiffRemoved, "Removed"},
		{DiffModified, "Modified"},
		{DiffCommentChanged, "CommentChanged"},
		{DiffStyleChanged, "StyleChanged"},
		{DiffReordered, "Reordered"},
		{DiffType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.diffType.String(); got != tt.expected {
				t.Errorf("DiffType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDiffNodes tests the DiffNodes function
func TestDiffNodesComplete(t *testing.T) {
	t.Run("BothNil", func(t *testing.T) {
		diffs := DiffNodes(nil, nil, "$")
		if len(diffs) != 0 {
			t.Errorf("DiffNodes(nil, nil) returned %v diffs, want 0", len(diffs))
		}
	})

	t.Run("NodeAdded", func(t *testing.T) {
		newNode := NewScalarNode("new")
		diffs := DiffNodes(nil, newNode, "$")
		if len(diffs) != 1 {
			t.Errorf("DiffNodes() returned %v diffs, want 1", len(diffs))
		}
		if diffs[0].Type != DiffAdded {
			t.Errorf("DiffNodes() Type = %v, want DiffAdded", diffs[0].Type)
		}
	})

	t.Run("NodeRemoved", func(t *testing.T) {
		oldNode := NewScalarNode("old")
		diffs := DiffNodes(oldNode, nil, "$")
		if len(diffs) != 1 {
			t.Errorf("DiffNodes() returned %v diffs, want 1", len(diffs))
		}
		if diffs[0].Type != DiffRemoved {
			t.Errorf("DiffNodes() Type = %v, want DiffRemoved", diffs[0].Type)
		}
	})

	t.Run("ValueChanged", func(t *testing.T) {
		oldNode := NewScalarNode("old")
		newNode := NewScalarNode("new")
		diffs := DiffNodes(oldNode, newNode, "$")
		if len(diffs) != 1 {
			t.Errorf("DiffNodes() returned %v diffs, want 1", len(diffs))
		}
		if diffs[0].Type != DiffModified {
			t.Errorf("DiffNodes() Type = %v, want DiffModified", diffs[0].Type)
		}
		if diffs[0].OldValue != "old" || diffs[0].NewValue != "new" {
			t.Error("DiffNodes() did not capture values correctly")
		}
	})

	t.Run("TypeChanged", func(t *testing.T) {
		oldNode := NewScalarNode("scalar")
		newNode := NewMappingNode()
		diffs := DiffNodes(oldNode, newNode, "$")

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffModified &&
			   diff.OldValue == ScalarNode &&
			   diff.NewValue == MappingNode {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffNodes() did not detect type change")
		}
	})

	t.Run("StyleChanged", func(t *testing.T) {
		oldNode := NewScalarNode("value")
		oldNode.Style = DefaultStyle
		newNode := NewScalarNode("value")
		newNode.Style = DoubleQuotedStyle

		diffs := DiffNodes(oldNode, newNode, "$")

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffStyleChanged {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffNodes() did not detect style change")
		}
	})

	t.Run("CommentChanged", func(t *testing.T) {
		oldNode := NewScalarNode("value")
		oldNode.HeadComment = []string{"# old"}
		newNode := NewScalarNode("value")
		newNode.HeadComment = []string{"# new"}

		diffs := DiffNodes(oldNode, newNode, "$")

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffCommentChanged {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffNodes() did not detect comment change")
		}
	})

	t.Run("MappingKeyAdded", func(t *testing.T) {
		oldMap := NewMappingNode()
		oldMap.AddKeyValue(NewScalarNode("key1"), NewScalarNode("value1"))

		newMap := NewMappingNode()
		newMap.AddKeyValue(NewScalarNode("key1"), NewScalarNode("value1"))
		newMap.AddKeyValue(NewScalarNode("key2"), NewScalarNode("value2"))

		diffs := DiffNodes(oldMap, newMap, "$")

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffAdded && strings.Contains(diff.Description, "key2") {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffNodes() did not detect added mapping key")
		}
	})

	t.Run("SequenceItemAdded", func(t *testing.T) {
		oldSeq := NewSequenceNode()
		oldSeq.AddSequenceItem(NewScalarNode("item1"))

		newSeq := NewSequenceNode()
		newSeq.AddSequenceItem(NewScalarNode("item1"))
		newSeq.AddSequenceItem(NewScalarNode("item2"))

		diffs := DiffNodes(oldSeq, newSeq, "$")

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffAdded && strings.Contains(diff.Path, "[1]") {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffNodes() did not detect added sequence item")
		}
	})
}

// TestDiffTreesComplete tests the DiffTrees function
func TestDiffTreesComplete(t *testing.T) {
	t.Run("BothNil", func(t *testing.T) {
		diffs := DiffTrees(nil, nil)
		if len(diffs) != 0 {
			t.Errorf("DiffTrees(nil, nil) returned %v diffs, want 0", len(diffs))
		}
	})

	t.Run("TreeAdded", func(t *testing.T) {
		newTree := NewNodeTree()
		doc := newTree.AddDocument()
		doc.Root = NewScalarNode("new")

		diffs := DiffTrees(nil, newTree)
		if len(diffs) != 1 {
			t.Errorf("DiffTrees() returned %v diffs, want 1", len(diffs))
		}
		if diffs[0].Type != DiffAdded {
			t.Errorf("DiffTrees() Type = %v, want DiffAdded", diffs[0].Type)
		}
	})

	t.Run("DocumentAdded", func(t *testing.T) {
		oldTree := NewNodeTree()
		oldTree.AddDocument()

		newTree := NewNodeTree()
		newTree.AddDocument()
		newTree.AddDocument()

		diffs := DiffTrees(oldTree, newTree)

		found := false
		for _, diff := range diffs {
			if diff.Type == DiffAdded && strings.Contains(diff.Path, "document:1") {
				found = true
				break
			}
		}
		if !found {
			t.Error("DiffTrees() did not detect added document")
		}
	})
}

// TestEqualStringSlicesComplete tests the equalStringSlices function
func TestEqualStringSlicesComplete(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{"BothNil", nil, nil, true},
		{"BothEmpty", []string{}, []string{}, true},
		{"Equal", []string{"a", "b"}, []string{"a", "b"}, true},
		{"DifferentLength", []string{"a"}, []string{"a", "b"}, false},
		{"DifferentContent", []string{"a", "b"}, []string{"a", "c"}, false},
		{"OneNil", []string{"a"}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalStringSlices(tt.a, tt.b); got != tt.want {
				t.Errorf("equalStringSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConvertFromYAMLNode tests the ConvertFromYAMLNode function
func TestConvertFromYAMLNodeComplete(t *testing.T) {
	t.Run("NilNode", func(t *testing.T) {
		result := ConvertFromYAMLNode(nil)
		if result != nil {
			t.Errorf("ConvertFromYAMLNode(nil) = %v, want nil", result)
		}
	})

	t.Run("ScalarNode", func(t *testing.T) {
		yamlNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "test",
			Tag:   "!!str",
		}

		result := ConvertFromYAMLNode(yamlNode)
		if result.Kind != ScalarNode {
			t.Errorf("ConvertFromYAMLNode() Kind = %v, want ScalarNode", result.Kind)
		}
		if result.Value != "test" {
			t.Errorf("ConvertFromYAMLNode() Value = %v, want test", result.Value)
		}
		if result.Tag != "!!str" {
			t.Errorf("ConvertFromYAMLNode() Tag = %v, want !!str", result.Tag)
		}
	})

	t.Run("BooleanValues", func(t *testing.T) {
		tests := []struct {
			value    string
			expected interface{}
		}{
			{"true", true},
			{"false", false},
			{"null", nil},
			{"~", nil},
		}

		for _, tt := range tests {
			yamlNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: tt.value,
			}
			result := ConvertFromYAMLNode(yamlNode)
			if !reflect.DeepEqual(result.Value, tt.expected) {
				t.Errorf("ConvertFromYAMLNode() Value = %v, want %v", result.Value, tt.expected)
			}
		}
	})

	t.Run("NumberValues", func(t *testing.T) {
		// Test integer
		yamlNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "42",
		}
		result := ConvertFromYAMLNode(yamlNode)
		if result.Value != int64(42) {
			t.Errorf("ConvertFromYAMLNode() integer Value = %v (%T), want 42", result.Value, result.Value)
		}

		// Test float
		yamlNode = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "3.14",
		}
		result = ConvertFromYAMLNode(yamlNode)
		if result.Value != 3.14 {
			t.Errorf("ConvertFromYAMLNode() float Value = %v, want 3.14", result.Value)
		}
	})

	t.Run("MappingNode", func(t *testing.T) {
		yamlNode := &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "key"},
				{Kind: yaml.ScalarNode, Value: "value"},
			},
		}

		result := ConvertFromYAMLNode(yamlNode)
		if result.Kind != MappingNode {
			t.Errorf("ConvertFromYAMLNode() Kind = %v, want MappingNode", result.Kind)
		}
		if len(result.Children) != 2 {
			t.Errorf("ConvertFromYAMLNode() Children length = %v, want 2", len(result.Children))
		}
	})

	t.Run("SequenceNode", func(t *testing.T) {
		yamlNode := &yaml.Node{
			Kind: yaml.SequenceNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "item1"},
				{Kind: yaml.ScalarNode, Value: "item2"},
			},
		}

		result := ConvertFromYAMLNode(yamlNode)
		if result.Kind != SequenceNode {
			t.Errorf("ConvertFromYAMLNode() Kind = %v, want SequenceNode", result.Kind)
		}
		if len(result.Children) != 2 {
			t.Errorf("ConvertFromYAMLNode() Children length = %v, want 2", len(result.Children))
		}
	})

	t.Run("StyleConversion", func(t *testing.T) {
		styles := []struct {
			yaml yaml.Style
			node NodeStyle
		}{
			{yaml.LiteralStyle, LiteralStyle},
			{yaml.FoldedStyle, FoldedStyle},
			{yaml.SingleQuotedStyle, QuotedStyle},
			{yaml.DoubleQuotedStyle, DoubleQuotedStyle},
			{yaml.FlowStyle, FlowStyle},
		}

		for _, s := range styles {
			yamlNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "test",
				Style: s.yaml,
			}
			result := ConvertFromYAMLNode(yamlNode)
			if result.Style != s.node {
				t.Errorf("ConvertFromYAMLNode() Style = %v, want %v", result.Style, s.node)
			}
		}
	})
}

// TestUnmarshal tests the Unmarshal function
func TestUnmarshal(t *testing.T) {
	type TestStruct struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	input := []byte("name: test\nvalue: 42")
	var result TestStruct

	err := Unmarshal(input, &result)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Unmarshal() Name = %v, want test", result.Name)
	}
	if result.Value != 42 {
		t.Errorf("Unmarshal() Value = %v, want 42", result.Value)
	}
}

// TestMarshal tests the Marshal function
func TestMarshal(t *testing.T) {
	type TestStruct struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	input := TestStruct{
		Name:  "test",
		Value: 42,
	}

	result, err := Marshal(input)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "name: test") {
		t.Errorf("Marshal() = %v, want to contain 'name: test'", resultStr)
	}
	if !strings.Contains(resultStr, "value: 42") {
		t.Errorf("Marshal() = %v, want to contain 'value: 42'", resultStr)
	}
}

// TestUnmarshalStrict tests the UnmarshalStrict function
func TestUnmarshalStrict(t *testing.T) {
	type TestStruct struct {
		Name string `yaml:"name"`
	}

	t.Run("ValidFields", func(t *testing.T) {
		input := []byte("name: test")
		var result TestStruct

		err := UnmarshalStrict(input, &result)
		if err != nil {
			t.Errorf("UnmarshalStrict() error = %v", err)
		}
		if result.Name != "test" {
			t.Errorf("UnmarshalStrict() Name = %v, want test", result.Name)
		}
	})

	t.Run("UnknownFields", func(t *testing.T) {
		input := []byte("name: test\nunknown: field")
		var result TestStruct

		err := UnmarshalStrict(input, &result)
		if err == nil {
			t.Error("UnmarshalStrict() should error on unknown fields")
		}
	})
}

// TestMarshalIndent tests the MarshalIndent function
func TestMarshalIndent(t *testing.T) {
	type TestStruct struct {
		Name   string `yaml:"name"`
		Nested struct {
			Value int `yaml:"value"`
		} `yaml:"nested"`
	}

	input := TestStruct{
		Name: "test",
	}
	input.Nested.Value = 42

	result, err := MarshalIndent(input, 4)
	if err != nil {
		t.Errorf("MarshalIndent() error = %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "    value:") {
		t.Error("MarshalIndent() should use 4-space indentation")
	}
}

// TestAddEmptyLinesBeforeSchemaComments tests the addEmptyLinesBeforeSchemaComments function
func TestAddEmptyLinesBeforeSchemaComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NoSchemaComment",
			input:    "key: value",
			expected: "key: value",
		},
		{
			name:     "SchemaAfterContent",
			input:    "key: value\n# @schema",
			expected: "key: value\n\n# @schema",
		},
		{
			name:     "SchemaAfterComment",
			input:    "# comment\n# @schema",
			expected: "# comment\n# @schema",
		},
		{
			name:     "SchemaAfterEmpty",
			input:    "key: value\n\n# @schema",
			expected: "key: value\n\n# @schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addEmptyLinesBeforeSchemaComments([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("addEmptyLinesBeforeSchemaComments() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}