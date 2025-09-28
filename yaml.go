package golang_yaml_advanced

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type NodeKind int

const (
	DocumentNode NodeKind = iota
	MappingNode
	SequenceNode
	ScalarNode
	AliasNode
	NullNode
)

func (k NodeKind) String() string {
	switch k {
	case DocumentNode:
		return "DocumentNode"
	case MappingNode:
		return "MappingNode"
	case SequenceNode:
		return "SequenceNode"
	case ScalarNode:
		return "ScalarNode"
	case AliasNode:
		return "AliasNode"
	case NullNode:
		return "NullNode"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

type NodeStyle int

const (
	DefaultStyle NodeStyle = iota
	LiteralStyle
	FoldedStyle
	QuotedStyle
	DoubleQuotedStyle
	FlowStyle
	TaggedStyle
	SingleQuotedStyle
)

func (s NodeStyle) String() string {
	switch s {
	case DefaultStyle:
		return "DefaultStyle"
	case LiteralStyle:
		return "LiteralStyle"
	case FoldedStyle:
		return "FoldedStyle"
	case QuotedStyle:
		return "QuotedStyle"
	case DoubleQuotedStyle:
		return "DoubleQuotedStyle"
	case FlowStyle:
		return "FlowStyle"
	case TaggedStyle:
		return "TaggedStyle"
	case SingleQuotedStyle:
		return "SingleQuotedStyle"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type Node struct {
	Kind        NodeKind
	Style       NodeStyle
	Tag         string
	Value       interface{}
	Anchor      string
	Alias       *Node
	Parent      *Node
	Children    []*Node
	Key         *Node
	Line        int
	Column      int
	HeadComment []string // We can have multiple lines
	LineComment string
	FootComment []string // We can have multiple lines
	EmptyLines  []int    // Track empty lines: position and count (e.g., [0,2] means 2 empty lines before node)
	Metadata    map[string]interface{}
}

type Document struct {
	Root       *Node
	Directives []Directive
	Version    string
	Anchors    map[string]*Node
}

type Directive struct {
	Name  string
	Value string
}

type NodeTree struct {
	Documents   []*Document
	Current     *Document
	CurrentNode *Node
}

func NewNodeTree() *NodeTree {
	return &NodeTree{
		Documents: make([]*Document, 0),
	}
}

func (nt *NodeTree) AddDocument() *Document {
	doc := &Document{
		Anchors: make(map[string]*Node),
	}
	nt.Documents = append(nt.Documents, doc)
	nt.Current = doc
	return doc
}

func (d *Document) SetRoot(node *Node) {
	d.Root = node
	if node != nil {
		node.Parent = nil
	}
}

func NewNode(kind NodeKind) *Node {
	return &Node{
		Kind:     kind,
		Style:    DefaultStyle,
		Children: make([]*Node, 0),
		Metadata: make(map[string]interface{}),
	}
}

func NewScalarNode(value interface{}) *Node {
	node := NewNode(ScalarNode)
	node.Value = value
	return node
}

func NewMappingNode() *Node {
	return NewNode(MappingNode)
}

func NewSequenceNode() *Node {
	return NewNode(SequenceNode)
}

func (n *Node) AddChild(child *Node) {
	if child != nil {
		child.Parent = n
		n.Children = append(n.Children, child)
	}
}

func (n *Node) AddKeyValue(key, value *Node) error {
	if n.Kind != MappingNode {
		return fmt.Errorf("can only add key-value pairs to mapping nodes")
	}
	if key != nil {
		key.Parent = n
		value.Parent = n
		value.Key = key
		n.Children = append(n.Children, key, value)
	}
	return nil
}

func (n *Node) AddSequenceItem(item *Node) error {
	if n.Kind != SequenceNode {
		return fmt.Errorf("can only add items to sequence nodes")
	}
	n.AddChild(item)
	return nil
}

func (n *Node) GetMapValue(key string) *Node {
	if n.Kind != MappingNode {
		return nil
	}
	for i := 0; i < len(n.Children)-1; i += 2 {
		keyNode := n.Children[i]
		if keyNode.Kind == ScalarNode && fmt.Sprintf("%v", keyNode.Value) == key {
			return n.Children[i+1]
		}
	}
	return nil
}

func (n *Node) GetSequenceItems() []*Node {
	if n.Kind != SequenceNode {
		return nil
	}
	return n.Children
}

func (n *Node) Walk(visitor func(*Node) bool) {
	n.walk(visitor)
}

func (n *Node) walk(visitor func(*Node) bool) bool {
	if !visitor(n) {
		return false
	}
	for _, child := range n.Children {
		if !child.walk(visitor) {
			return false
		}
	}
	return true
}

func (n *Node) Find(predicate func(*Node) bool) *Node {
	var result *Node
	n.Walk(func(node *Node) bool {
		if predicate(node) {
			result = node
			return false
		}
		return true
	})
	return result
}

func (n *Node) FindAll(predicate func(*Node) bool) []*Node {
	var results []*Node
	n.Walk(func(node *Node) bool {
		if predicate(node) {
			results = append(results, node)
		}
		return true
	})
	return results
}

func (n *Node) Path() string {
	if n.Parent == nil {
		return "$"
	}
	path := n.Parent.Path()

	switch n.Parent.Kind {
	case MappingNode:
		if n.Key != nil && n.Key.Kind == ScalarNode {
			return fmt.Sprintf("%s.%v", path, n.Key.Value)
		}
	case SequenceNode:
		for i, child := range n.Parent.Children {
			if child == n {
				return fmt.Sprintf("%s[%d]", path, i)
			}
		}
	}
	return path
}

func (n *Node) Clone() *Node {
	return n.cloneWithSeen(make(map[*Node]*Node))
}

func (n *Node) cloneWithSeen(seen map[*Node]*Node) *Node {
	if n == nil {
		return nil
	}

	// Check if we've already cloned this node (circular reference)
	if clone, ok := seen[n]; ok {
		return clone
	}

	clone := &Node{
		Kind:        n.Kind,
		Style:       n.Style,
		Tag:         n.Tag,
		Value:       n.Value,
		Anchor:      n.Anchor,
		Line:        n.Line,
		Column:      n.Column,
		HeadComment: append([]string(nil), n.HeadComment...),
		LineComment: n.LineComment,
		FootComment: append([]string(nil), n.FootComment...),
		EmptyLines:  append([]int(nil), n.EmptyLines...),
		Children:    make([]*Node, 0, len(n.Children)),
		Metadata:    make(map[string]interface{}),
	}

	// Mark this node as seen
	seen[n] = clone

	for k, v := range n.Metadata {
		clone.Metadata[k] = v
	}

	for _, child := range n.Children {
		childClone := child.cloneWithSeen(seen)
		if childClone != nil {
			childClone.Parent = clone
			clone.Children = append(clone.Children, childClone)
		}
	}

	if n.Key != nil {
		clone.Key = n.Key.cloneWithSeen(seen)
	}

	return clone
}

func (n *Node) String() string {
	return n.stringify(0)
}

// convertToYAMLNode converts our Node to yaml.Node for internal testing
func (n *Node) convertToYAMLNode() *yaml.Node {
	if n == nil {
		return nil
	}

	yamlNode := &yaml.Node{}

	// Convert Kind
	switch n.Kind {
	case DocumentNode:
		yamlNode.Kind = yaml.DocumentNode
	case MappingNode:
		yamlNode.Kind = yaml.MappingNode
	case SequenceNode:
		yamlNode.Kind = yaml.SequenceNode
	case ScalarNode:
		yamlNode.Kind = yaml.ScalarNode
	case AliasNode:
		yamlNode.Kind = yaml.AliasNode
	default:
		yamlNode.Kind = yaml.ScalarNode
		yamlNode.Tag = "!!null"
	}

	// Set value and other properties
	yamlNode.Tag = n.Tag
	yamlNode.Value = fmt.Sprintf("%v", n.Value)
	yamlNode.Anchor = n.Anchor

	// Convert children
	for _, child := range n.Children {
		yamlNode.Content = append(yamlNode.Content, child.convertToYAMLNode())
	}

	// Set position
	yamlNode.Line = n.Line
	yamlNode.Column = n.Column

	// Convert comments
	if len(n.HeadComment) > 0 {
		yamlNode.HeadComment = strings.Join(n.HeadComment, "\n")
	}
	yamlNode.LineComment = n.LineComment
	if len(n.FootComment) > 0 {
		yamlNode.FootComment = strings.Join(n.FootComment, "\n")
	}

	// Handle alias
	if n.Alias != nil {
		yamlNode.Alias = n.Alias.convertToYAMLNode()
	}

	return yamlNode
}

func (n *Node) stringify(indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch n.Kind {
	case ScalarNode:
		if n.Value == nil {
			return ""
		}
		return fmt.Sprintf("%v", n.Value)
	case MappingNode:
		var sb strings.Builder
		for i := 0; i < len(n.Children)-1; i += 2 {
			key := n.Children[i]
			value := n.Children[i+1]
			sb.WriteString(fmt.Sprintf("\n%s%s: %s", indentStr, key.stringify(0), value.stringify(indent+1)))
		}
		return sb.String()
	case SequenceNode:
		var sb strings.Builder
		for _, item := range n.Children {
			sb.WriteString(fmt.Sprintf("\n%s- %s", indentStr, item.stringify(indent+1)))
		}
		return sb.String()
	case DocumentNode:
		if len(n.Children) > 0 {
			return n.Children[0].stringify(indent)
		}
	case NullNode:
		return ""
	case AliasNode:
		if n.Alias != nil {
			return fmt.Sprintf("*%s", n.Value)
		}
	}
	return ""
}

func (n *Node) IsNull() bool {
	return n == nil || n.Kind == NullNode || (n.Kind == ScalarNode && n.Value == nil)
}

func (n *Node) Remove() error {
	if n.Parent == nil {
		return fmt.Errorf("cannot remove root node")
	}

	for i, child := range n.Parent.Children {
		if child == n {
			n.Parent.Children = append(n.Parent.Children[:i], n.Parent.Children[i+1:]...)
			n.Parent = nil
			return nil
		}
	}
	return fmt.Errorf("node not found in parent's children")
}

func (n *Node) ReplaceWith(replacement *Node) error {
	if n.Parent == nil {
		return fmt.Errorf("cannot replace root node")
	}

	for i, child := range n.Parent.Children {
		if child == n {
			n.Parent.Children[i] = replacement
			replacement.Parent = n.Parent
			replacement.Key = n.Key
			n.Parent = nil
			return nil
		}
	}
	return fmt.Errorf("node not found in parent's children")
}

func (nt *NodeTree) Merge(other *NodeTree) {
	for _, doc := range other.Documents {
		nt.Documents = append(nt.Documents, doc)
	}
}

// MergeNodes merges two nodes, preserving comments from both
func MergeNodes(base, overlay *Node) *Node {
	if base == nil {
		if overlay == nil {
			return nil
		}
		return overlay.Clone()
	}
	if overlay == nil {
		return base.Clone()
	}

	// Clone the base to avoid modifying the original
	result := base.Clone()

	// If both are mappings, merge their keys
	if base.Kind == MappingNode && overlay.Kind == MappingNode {
		// Create a map of base keys for quick lookup
		baseKeys := make(map[string]int)
		for i := 0; i < len(base.Children)-1; i += 2 {
			keyNode := base.Children[i]
			if keyNode.Kind == ScalarNode {
				keyStr := fmt.Sprintf("%v", keyNode.Value)
				baseKeys[keyStr] = i
			}
		}

		// Process overlay keys
		for i := 0; i < len(overlay.Children)-1; i += 2 {
			overlayKey := overlay.Children[i]
			overlayValue := overlay.Children[i+1]

			if overlayKey.Kind == ScalarNode {
				keyStr := fmt.Sprintf("%v", overlayKey.Value)

				if baseIdx, exists := baseKeys[keyStr]; exists {
					// Key exists in base - merge or replace the value
					baseValue := result.Children[baseIdx+1]

					// If both values are mappings, merge them recursively
					if baseValue.Kind == MappingNode && overlayValue.Kind == MappingNode {
						merged := MergeNodes(baseValue, overlayValue)
						result.Children[baseIdx+1] = merged
					} else {
						// Replace with overlay value, but preserve overlay's comments
						clonedValue := overlayValue.Clone()
						clonedValue.Key = result.Children[baseIdx].Clone()

						// Preserve the overlay key's comments on the existing key
						if len(overlayKey.HeadComment) > 0 {
							result.Children[baseIdx].HeadComment = overlayKey.HeadComment
						}
						if overlayKey.LineComment != "" {
							result.Children[baseIdx].LineComment = overlayKey.LineComment
						}
						if len(overlayKey.FootComment) > 0 {
							result.Children[baseIdx].FootComment = overlayKey.FootComment
						}

						result.Children[baseIdx+1] = clonedValue
					}
				} else {
					// Key doesn't exist in base - add it
					clonedKey := overlayKey.Clone()
					clonedValue := overlayValue.Clone()
					clonedKey.Parent = result
					clonedValue.Parent = result
					clonedValue.Key = clonedKey
					result.Children = append(result.Children, clonedKey, clonedValue)
				}
			}
		}
	} else if base.Kind == SequenceNode && overlay.Kind == SequenceNode {
		// For sequences, append overlay items to base
		for _, item := range overlay.Children {
			cloned := item.Clone()
			result.AddChild(cloned)
		}
	} else {
		// For other types, overlay replaces base but preserve base comments if overlay has none
		result = overlay.Clone()
		if len(result.HeadComment) == 0 && len(base.HeadComment) > 0 {
			result.HeadComment = base.HeadComment
		}
		if result.LineComment == "" && base.LineComment != "" {
			result.LineComment = base.LineComment
		}
		if len(result.FootComment) == 0 && len(base.FootComment) > 0 {
			result.FootComment = base.FootComment
		}
		return result
	}

	return result
}

// MergeDocuments merges two documents preserving comments
func MergeDocuments(base, overlay *Document) *Document {
	if base == nil && overlay == nil {
		return nil
	}
	if base == nil {
		return &Document{
			Root:       overlay.Root.Clone(),
			Directives: append([]Directive{}, overlay.Directives...),
			Version:    overlay.Version,
			Anchors:    make(map[string]*Node),
		}
	}
	if overlay == nil {
		return &Document{
			Root:       base.Root.Clone(),
			Directives: append([]Directive{}, base.Directives...),
			Version:    base.Version,
			Anchors:    make(map[string]*Node),
		}
	}

	merged := &Document{
		Directives: append([]Directive{}, base.Directives...),
		Version:    base.Version,
		Anchors:    make(map[string]*Node),
	}

	// Merge directives from overlay
	for _, dir := range overlay.Directives {
		found := false
		for _, baseDir := range merged.Directives {
			if baseDir.Name == dir.Name {
				found = true
				break
			}
		}
		if !found {
			merged.Directives = append(merged.Directives, dir)
		}
	}

	// Handle the actual content nodes
	// If base root is a Document node, we need to handle its children
	var baseContent, overlayContent *Node

	if base.Root != nil && base.Root.Kind == DocumentNode && len(base.Root.Children) > 0 {
		baseContent = base.Root.Children[0]
	} else if base.Root != nil && base.Root.Kind != DocumentNode {
		baseContent = base.Root
	}

	if overlay.Root != nil && overlay.Root.Kind == DocumentNode && len(overlay.Root.Children) > 0 {
		overlayContent = overlay.Root.Children[0]
	} else if overlay.Root != nil && overlay.Root.Kind != DocumentNode {
		overlayContent = overlay.Root
	}

	// Merge the actual content nodes
	var mergedContent *Node
	if baseContent != nil || overlayContent != nil {
		mergedContent = MergeNodes(baseContent, overlayContent)
	}

	// Create the document node
	merged.Root = NewNode(DocumentNode)

	// Preserve head comments from base document
	if base.Root != nil && len(base.Root.HeadComment) > 0 {
		merged.Root.HeadComment = base.Root.HeadComment
	}

	// If overlay has head comments and base doesn't, use overlay's
	if len(merged.Root.HeadComment) == 0 && overlay.Root != nil && len(overlay.Root.HeadComment) > 0 {
		merged.Root.HeadComment = overlay.Root.HeadComment
	}

	if mergedContent != nil {
		merged.Root.AddChild(mergedContent)
	}

	// Merge anchors
	for k, v := range base.Anchors {
		merged.Anchors[k] = v
	}
	for k, v := range overlay.Anchors {
		merged.Anchors[k] = v
	}

	return merged
}

// MergeTrees merges two NodeTrees preserving comments from both
func MergeTrees(base, overlay *NodeTree) *NodeTree {
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}

	result := NewNodeTree()

	// If both have documents, merge the first documents
	if len(base.Documents) > 0 && len(overlay.Documents) > 0 {
		merged := MergeDocuments(base.Documents[0], overlay.Documents[0])
		result.Documents = append(result.Documents, merged)
		result.Current = merged

		// Append any additional documents from base
		for i := 1; i < len(base.Documents); i++ {
			result.Documents = append(result.Documents, base.Documents[i])
		}

		// Append any additional documents from overlay
		for i := 1; i < len(overlay.Documents); i++ {
			result.Documents = append(result.Documents, overlay.Documents[i])
		}
	} else if len(base.Documents) > 0 {
		result.Documents = append(result.Documents, base.Documents...)
		if len(result.Documents) > 0 {
			result.Current = result.Documents[0]
		}
	} else if len(overlay.Documents) > 0 {
		result.Documents = append(result.Documents, overlay.Documents...)
		if len(result.Documents) > 0 {
			result.Current = result.Documents[0]
		}
	}

	return result
}

func (d *Document) RegisterAnchor(name string, node *Node) {
	if d.Anchors == nil {
		d.Anchors = make(map[string]*Node)
	}
	d.Anchors[name] = node
	node.Anchor = name
}

func (d *Document) GetAnchor(name string) *Node {
	return d.Anchors[name]
}

func (n *Node) ToYAMLNode() *yaml.Node {
	if n == nil {
		return nil
	}

	yamlNode := &yaml.Node{
		Value:  fmt.Sprintf("%v", n.Value),
		Tag:    n.Tag,
		Anchor: n.Anchor,
		Line:   n.Line,
		Column: n.Column,
	}

	if len(n.HeadComment) > 0 {
		yamlNode.HeadComment = strings.Join(n.HeadComment, "\n")
	}
	yamlNode.LineComment = n.LineComment
	if len(n.FootComment) > 0 {
		yamlNode.FootComment = strings.Join(n.FootComment, "\n")
	}

	switch n.Kind {
	case DocumentNode:
		yamlNode.Kind = yaml.DocumentNode
		yamlNode.Content = make([]*yaml.Node, 0, len(n.Children))
		for _, child := range n.Children {
			yamlNode.Content = append(yamlNode.Content, child.ToYAMLNode())
		}
	case MappingNode:
		yamlNode.Kind = yaml.MappingNode
		yamlNode.Content = make([]*yaml.Node, 0, len(n.Children))
		for _, child := range n.Children {
			yamlNode.Content = append(yamlNode.Content, child.ToYAMLNode())
		}
	case SequenceNode:
		yamlNode.Kind = yaml.SequenceNode
		yamlNode.Content = make([]*yaml.Node, 0, len(n.Children))
		for _, child := range n.Children {
			yamlNode.Content = append(yamlNode.Content, child.ToYAMLNode())
		}
	case ScalarNode:
		yamlNode.Kind = yaml.ScalarNode
		if n.Value == nil {
			yamlNode.Value = ""
		}
	case AliasNode:
		yamlNode.Kind = yaml.AliasNode
	case NullNode:
		yamlNode.Kind = yaml.ScalarNode
		yamlNode.Tag = "!!null"
		yamlNode.Value = ""
	}

	switch n.Style {
	case LiteralStyle:
		yamlNode.Style = yaml.LiteralStyle
	case FoldedStyle:
		yamlNode.Style = yaml.FoldedStyle
	case QuotedStyle:
		yamlNode.Style = yaml.SingleQuotedStyle
	case DoubleQuotedStyle:
		yamlNode.Style = yaml.DoubleQuotedStyle
	case FlowStyle:
		yamlNode.Style = yaml.FlowStyle
	default:
		yamlNode.Style = 0
	}

	return yamlNode
}

func (d *Document) ToYAML() ([]byte, error) {
	if d.Root == nil {
		return []byte{}, nil
	}

	// Special case: if we have a document node with only comments
	if d.Root.Kind == DocumentNode && len(d.Root.Children) == 0 && len(d.Root.HeadComment) > 0 {
		// Return just the comments as plain text
		result := strings.Join(d.Root.HeadComment, "\n")
		if !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		return []byte(result), nil
	}

	// Special case: empty document node - return empty YAML
	if d.Root.Kind == DocumentNode && len(d.Root.Children) == 0 {
		return []byte(""), nil
	}

	yamlNode := d.Root.ToYAMLNode()

	// Use encoder with 2-space indentation
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // Set 2-space indentation

	if err := encoder.Encode(yamlNode); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (nt *NodeTree) ToYAML() ([]byte, error) {
	if len(nt.Documents) == 0 {
		return []byte{}, nil
	}

	result := []byte{}
	for i, doc := range nt.Documents {
		docBytes, err := doc.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document %d: %w", i, err)
		}

		// Post-process to add empty lines before @schema comments
		processedBytes := addEmptyLinesBeforeSchemaComments(docBytes)

		if i > 0 {
			result = append(result, []byte("---\n")...)
		}
		result = append(result, processedBytes...)
	}

	return result, nil
}

// addEmptyLinesBeforeSchemaComments adds empty lines before @schema comment blocks
func addEmptyLinesBeforeSchemaComments(input []byte) []byte {
	lines := strings.Split(string(input), "\n")
	var output []string

	for i := 0; i < len(lines); i++ {
		currentLine := lines[i]

		// Check if we need to add an empty line before this line
		if i > 0 {
			prevLine := lines[i-1]
			trimmedPrev := strings.TrimSpace(prevLine)

			// Add empty line before @schema if previous line is not empty and not a comment
			if strings.Contains(currentLine, "# @schema") &&
				prevLine != "" &&
				!strings.HasPrefix(trimmedPrev, "#") {
				output = append(output, "")
			}
		}

		output = append(output, currentLine)
	}

	return []byte(strings.Join(output, "\n"))
}

// UnmarshalYAML is a custom unmarshal function that preserves comments even when there's no content
func UnmarshalYAML(data []byte) (*NodeTree, error) {
	tree := NewNodeTree()

	// Handle completely empty input
	if len(data) == 0 {
		doc := tree.AddDocument()
		doc.SetRoot(nil)
		return tree, nil
	}

	// Split by document separator to handle multi-document YAML
	content := string(data)
	documents := splitDocuments(content)

	for _, docContent := range documents {
		// Check if the content has any non-comment YAML
		lines := strings.Split(docContent, "\n")
		hasYAMLContent := false
		var commentLines []string

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && trimmed != "---" {
				hasYAMLContent = true
				break
			}
			// Preserve all lines including empty ones to maintain formatting
			if trimmed != "---" && (trimmed != "" || len(commentLines) > 0) {
				// Only add empty lines if we already have some comments (to avoid leading empty lines)
				commentLines = append(commentLines, line)
			}
		}

		if !hasYAMLContent && len(commentLines) > 0 {
			// Remove trailing empty lines
			for len(commentLines) > 0 && strings.TrimSpace(commentLines[len(commentLines)-1]) == "" {
				commentLines = commentLines[:len(commentLines)-1]
			}

			// Create a document with only comments
			doc := tree.AddDocument()
			rootNode := NewNode(DocumentNode)

			// Store the comments in the HeadComment field
			rootNode.HeadComment = commentLines
			doc.SetRoot(rootNode)
			continue
		}

		// Normal YAML parsing
		var node yaml.Node
		err := yaml.Unmarshal([]byte(docContent), &node)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		doc := tree.AddDocument()
		rootNode := ConvertFromYAMLNode(&node)

		// Handle anchor resolution
		if rootNode != nil {
			resolveAnchors(rootNode, doc)
		}

		doc.SetRoot(rootNode)
	}

	return tree, nil
}

// splitDocuments splits a YAML string into separate documents by --- separator
func splitDocuments(content string) []string {
	lines := strings.Split(content, "\n")
	var documents []string
	var currentDoc strings.Builder
	inDocument := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			// Document separator found
			if currentDoc.Len() > 0 {
				documents = append(documents, currentDoc.String())
				currentDoc.Reset()
			}
			inDocument = true
		} else if trimmed == "..." {
			// Document end marker
			if currentDoc.Len() > 0 {
				documents = append(documents, currentDoc.String())
				currentDoc.Reset()
			}
			inDocument = false
		} else {
			// Regular content line
			if !inDocument && len(documents) == 0 {
				// First document without explicit --- marker
				inDocument = true
			}
			if inDocument {
				if currentDoc.Len() > 0 {
					currentDoc.WriteString("\n")
				}
				currentDoc.WriteString(line)
			}
		}
	}

	// Add the last document if any
	if currentDoc.Len() > 0 {
		documents = append(documents, currentDoc.String())
	}

	// If no documents were found, treat the entire content as one document
	if len(documents) == 0 && len(content) > 0 {
		documents = append(documents, content)
	}

	return documents
}

// resolveAnchors processes a node tree and registers anchors with the document
func resolveAnchors(node *Node, doc *Document) {
	if node == nil {
		return
	}

	// Register anchor if present
	if node.Anchor != "" {
		doc.RegisterAnchor(node.Anchor, node)
	}

	// Handle alias nodes
	if node.Kind == AliasNode && node.Value != nil {
		// The Value field contains the alias name
		aliasName := fmt.Sprintf("%v", node.Value)
		if aliasName != "" && aliasName[0] == '*' {
			aliasName = aliasName[1:] // Remove the * prefix
		}
		node.Alias = doc.GetAnchor(aliasName)
	}

	// Process children recursively
	for _, child := range node.Children {
		resolveAnchors(child, doc)
	}
}

// DiffResult represents the difference between two nodes
type DiffResult struct {
	Type        DiffType
	Path        string
	OldValue    interface{}
	NewValue    interface{}
	OldNode     *Node
	NewNode     *Node
	OldComment  []string
	NewComment  []string
	Description string
}

type DiffType int

const (
	DiffNone DiffType = iota
	DiffAdded
	DiffRemoved
	DiffModified
	DiffCommentChanged
	DiffStyleChanged
	DiffReordered
)

func (dt DiffType) String() string {
	switch dt {
	case DiffNone:
		return "None"
	case DiffAdded:
		return "Added"
	case DiffRemoved:
		return "Removed"
	case DiffModified:
		return "Modified"
	case DiffCommentChanged:
		return "CommentChanged"
	case DiffStyleChanged:
		return "StyleChanged"
	case DiffReordered:
		return "Reordered"
	default:
		return "Unknown"
	}
}

// DiffNodes performs a deep comparison of two nodes and returns differences
func DiffNodes(oldNode, newNode *Node, path string) []DiffResult {
	var diffs []DiffResult

	// Handle nil cases
	if oldNode == nil && newNode == nil {
		return diffs
	}
	if oldNode == nil {
		diffs = append(diffs, DiffResult{
			Type:        DiffAdded,
			Path:        path,
			NewValue:    newNode.Value,
			NewNode:     newNode,
			Description: fmt.Sprintf("Added node at %s", path),
		})
		return diffs
	}
	if newNode == nil {
		diffs = append(diffs, DiffResult{
			Type:        DiffRemoved,
			Path:        path,
			OldValue:    oldNode.Value,
			OldNode:     oldNode,
			Description: fmt.Sprintf("Removed node at %s", path),
		})
		return diffs
	}

	// Check for type changes
	if oldNode.Kind != newNode.Kind {
		diffs = append(diffs, DiffResult{
			Type:        DiffModified,
			Path:        path,
			OldValue:    oldNode.Kind,
			NewValue:    newNode.Kind,
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Node type changed at %s", path),
		})
		return diffs
	}

	// Check for value changes (for scalar nodes)
	if oldNode.Kind == ScalarNode {
		if fmt.Sprintf("%v", oldNode.Value) != fmt.Sprintf("%v", newNode.Value) {
			diffs = append(diffs, DiffResult{
				Type:        DiffModified,
				Path:        path,
				OldValue:    oldNode.Value,
				NewValue:    newNode.Value,
				OldNode:     oldNode,
				NewNode:     newNode,
				Description: fmt.Sprintf("Value changed at %s from '%v' to '%v'", path, oldNode.Value, newNode.Value),
			})
		}
	}

	// Check for style changes
	if oldNode.Style != newNode.Style {
		diffs = append(diffs, DiffResult{
			Type:        DiffStyleChanged,
			Path:        path,
			OldValue:    oldNode.Style,
			NewValue:    newNode.Style,
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Style changed at %s", path),
		})
	}

	// Check for comment changes
	if !equalStringSlices(oldNode.HeadComment, newNode.HeadComment) {
		diffs = append(diffs, DiffResult{
			Type:        DiffCommentChanged,
			Path:        path,
			OldComment:  oldNode.HeadComment,
			NewComment:  newNode.HeadComment,
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Head comment changed at %s", path),
		})
	}

	if oldNode.LineComment != newNode.LineComment {
		diffs = append(diffs, DiffResult{
			Type:        DiffCommentChanged,
			Path:        path,
			OldComment:  []string{oldNode.LineComment},
			NewComment:  []string{newNode.LineComment},
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Line comment changed at %s", path),
		})
	}

	if !equalStringSlices(oldNode.FootComment, newNode.FootComment) {
		diffs = append(diffs, DiffResult{
			Type:        DiffCommentChanged,
			Path:        path,
			OldComment:  oldNode.FootComment,
			NewComment:  newNode.FootComment,
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Foot comment changed at %s", path),
		})
	}

	// Check for tag changes
	if oldNode.Tag != newNode.Tag {
		diffs = append(diffs, DiffResult{
			Type:        DiffModified,
			Path:        path,
			OldValue:    oldNode.Tag,
			NewValue:    newNode.Tag,
			OldNode:     oldNode,
			NewNode:     newNode,
			Description: fmt.Sprintf("Tag changed at %s from '%s' to '%s'", path, oldNode.Tag, newNode.Tag),
		})
	}

	// Compare children for mapping nodes
	if oldNode.Kind == MappingNode {
		oldKeys := make(map[string]*Node)
		newKeys := make(map[string]*Node)

		// Build key maps
		for i := 0; i < len(oldNode.Children)-1; i += 2 {
			key := oldNode.Children[i]
			value := oldNode.Children[i+1]
			if key.Kind == ScalarNode {
				keyStr := fmt.Sprintf("%v", key.Value)
				oldKeys[keyStr] = value
			}
		}

		for i := 0; i < len(newNode.Children)-1; i += 2 {
			key := newNode.Children[i]
			value := newNode.Children[i+1]
			if key.Kind == ScalarNode {
				keyStr := fmt.Sprintf("%v", key.Value)
				newKeys[keyStr] = value
			}
		}

		// Check for removed keys
		for key, oldValue := range oldKeys {
			if _, exists := newKeys[key]; !exists {
				childPath := fmt.Sprintf("%s.%s", path, key)
				diffs = append(diffs, DiffResult{
					Type:        DiffRemoved,
					Path:        childPath,
					OldValue:    oldValue.Value,
					OldNode:     oldValue,
					Description: fmt.Sprintf("Key '%s' removed", key),
				})
			}
		}

		// Check for added or modified keys
		for key, newValue := range newKeys {
			childPath := fmt.Sprintf("%s.%s", path, key)
			if oldValue, exists := oldKeys[key]; exists {
				// Key exists in both - check for differences
				childDiffs := DiffNodes(oldValue, newValue, childPath)
				diffs = append(diffs, childDiffs...)
			} else {
				// New key added
				diffs = append(diffs, DiffResult{
					Type:        DiffAdded,
					Path:        childPath,
					NewValue:    newValue.Value,
					NewNode:     newValue,
					Description: fmt.Sprintf("Key '%s' added", key),
				})
			}
		}
	}

	// Compare children for sequence nodes
	if oldNode.Kind == SequenceNode {
		oldLen := len(oldNode.Children)
		newLen := len(newNode.Children)

		minLen := oldLen
		if newLen < minLen {
			minLen = newLen
		}

		// Compare common elements
		for i := 0; i < minLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			childDiffs := DiffNodes(oldNode.Children[i], newNode.Children[i], childPath)
			diffs = append(diffs, childDiffs...)
		}

		// Check for removed elements
		for i := minLen; i < oldLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			diffs = append(diffs, DiffResult{
				Type:        DiffRemoved,
				Path:        childPath,
				OldValue:    oldNode.Children[i].Value,
				OldNode:     oldNode.Children[i],
				Description: fmt.Sprintf("Array element at index %d removed", i),
			})
		}

		// Check for added elements
		for i := minLen; i < newLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			diffs = append(diffs, DiffResult{
				Type:        DiffAdded,
				Path:        childPath,
				NewValue:    newNode.Children[i].Value,
				NewNode:     newNode.Children[i],
				Description: fmt.Sprintf("Array element at index %d added", i),
			})
		}
	}

	// Compare children for document nodes
	if oldNode.Kind == DocumentNode && len(oldNode.Children) > 0 && len(newNode.Children) > 0 {
		childDiffs := DiffNodes(oldNode.Children[0], newNode.Children[0], path)
		diffs = append(diffs, childDiffs...)
	}

	return diffs
}

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// DiffTrees compares two NodeTrees and returns all differences
func DiffTrees(oldTree, newTree *NodeTree) []DiffResult {
	var allDiffs []DiffResult

	if oldTree == nil && newTree == nil {
		return allDiffs
	}

	if oldTree == nil {
		// All documents in new tree are added
		for i, doc := range newTree.Documents {
			allDiffs = append(allDiffs, DiffResult{
				Type:        DiffAdded,
				Path:        fmt.Sprintf("$[document:%d]", i),
				NewNode:     doc.Root,
				Description: fmt.Sprintf("Document %d added", i),
			})
		}
		return allDiffs
	}

	if newTree == nil {
		// All documents in old tree are removed
		for i, doc := range oldTree.Documents {
			allDiffs = append(allDiffs, DiffResult{
				Type:        DiffRemoved,
				Path:        fmt.Sprintf("$[document:%d]", i),
				OldNode:     doc.Root,
				Description: fmt.Sprintf("Document %d removed", i),
			})
		}
		return allDiffs
	}

	maxDocs := len(oldTree.Documents)
	if len(newTree.Documents) > maxDocs {
		maxDocs = len(newTree.Documents)
	}

	for i := 0; i < maxDocs; i++ {
		docPath := fmt.Sprintf("$[document:%d]", i)

		if i >= len(oldTree.Documents) {
			// New document added
			allDiffs = append(allDiffs, DiffResult{
				Type:        DiffAdded,
				Path:        docPath,
				NewNode:     newTree.Documents[i].Root,
				Description: fmt.Sprintf("Document %d added", i),
			})
			continue
		}

		if i >= len(newTree.Documents) {
			// Document removed
			allDiffs = append(allDiffs, DiffResult{
				Type:        DiffRemoved,
				Path:        docPath,
				OldNode:     oldTree.Documents[i].Root,
				Description: fmt.Sprintf("Document %d removed", i),
			})
			continue
		}

		// Compare documents
		docDiffs := DiffNodes(oldTree.Documents[i].Root, newTree.Documents[i].Root, docPath)
		allDiffs = append(allDiffs, docDiffs...)
	}

	return allDiffs
}

// convertFromYAMLNode is internal function for testing that converts yaml.Node with anchor tracking
func convertFromYAMLNode(yamlNode *yaml.Node, parent *Node, anchors map[string]*Node) *Node {
	if yamlNode == nil {
		return nil
	}

	if anchors == nil {
		anchors = make(map[string]*Node)
	}

	var nodeKind NodeKind
	switch yamlNode.Kind {
	case yaml.DocumentNode:
		nodeKind = DocumentNode
	case yaml.MappingNode:
		nodeKind = MappingNode
	case yaml.SequenceNode:
		nodeKind = SequenceNode
	case yaml.ScalarNode:
		nodeKind = ScalarNode
	case yaml.AliasNode:
		nodeKind = AliasNode
	default:
		nodeKind = ScalarNode
	}

	node := NewNode(nodeKind)
	node.Parent = parent
	node.Value = yamlNode.Value
	node.Tag = yamlNode.Tag
	node.Anchor = yamlNode.Anchor
	node.Line = yamlNode.Line
	node.Column = yamlNode.Column

	// Store anchor for future alias resolution
	if yamlNode.Anchor != "" {
		anchors[yamlNode.Anchor] = node
	}

	// Handle alias
	if yamlNode.Kind == yaml.AliasNode && yamlNode.Alias != nil {
		if aliasTarget, ok := anchors[yamlNode.Alias.Anchor]; ok {
			node.Alias = aliasTarget
		}
	}

	if yamlNode.HeadComment != "" {
		node.HeadComment = strings.Split(yamlNode.HeadComment, "\n")
	}
	node.LineComment = yamlNode.LineComment
	if yamlNode.FootComment != "" {
		node.FootComment = strings.Split(yamlNode.FootComment, "\n")
	}

	// Convert children
	for _, child := range yamlNode.Content {
		childNode := convertFromYAMLNode(child, node, anchors)
		node.Children = append(node.Children, childNode)
	}

	return node
}

// ConvertFromYAMLNode converts a yaml.Node to our Node structure
func ConvertFromYAMLNode(yamlNode *yaml.Node) *Node {
	if yamlNode == nil {
		return nil
	}

	var nodeKind NodeKind
	switch yamlNode.Kind {
	case yaml.DocumentNode:
		nodeKind = DocumentNode
	case yaml.MappingNode:
		nodeKind = MappingNode
	case yaml.SequenceNode:
		nodeKind = SequenceNode
	case yaml.ScalarNode:
		nodeKind = ScalarNode
	case yaml.AliasNode:
		nodeKind = AliasNode
	default:
		nodeKind = ScalarNode
	}

	node := NewNode(nodeKind)

	// For scalar nodes, decode the value properly
	if nodeKind == ScalarNode {
		var value interface{}
		if yamlNode.Tag == "!!str" || yamlNode.Tag == "" {
			// Check if it's a boolean, number, or null
			switch yamlNode.Value {
			case "true":
				value = true
			case "false":
				value = false
			case "null", "~":
				value = nil
			default:
				// Try to parse as integer first to preserve large numbers
				if intVal, err := strconv.ParseInt(yamlNode.Value, 10, 64); err == nil {
					value = intVal
				} else if floatVal, err := strconv.ParseFloat(yamlNode.Value, 64); err == nil {
					value = floatVal
				} else {
					value = yamlNode.Value
				}
			}
		} else if yamlNode.Tag == "!!bool" {
			value = yamlNode.Value == "true"
		} else if yamlNode.Tag == "!!int" {
			if intVal, err := strconv.ParseInt(yamlNode.Value, 10, 64); err == nil {
				value = intVal
			} else {
				value = yamlNode.Value
			}
		} else if yamlNode.Tag == "!!float" {
			value, _ = strconv.ParseFloat(yamlNode.Value, 64)
		} else if yamlNode.Tag == "!!null" {
			value = nil
		} else {
			value = yamlNode.Value
		}
		node.Value = value
	} else {
		node.Value = yamlNode.Value
	}

	node.Tag = yamlNode.Tag
	node.Anchor = yamlNode.Anchor
	node.Line = yamlNode.Line
	node.Column = yamlNode.Column

	if yamlNode.HeadComment != "" {
		node.HeadComment = strings.Split(yamlNode.HeadComment, "\n")
	}
	node.LineComment = yamlNode.LineComment
	if yamlNode.FootComment != "" {
		node.FootComment = strings.Split(yamlNode.FootComment, "\n")
	}

	// Convert style
	switch yamlNode.Style {
	case yaml.LiteralStyle:
		node.Style = LiteralStyle
	case yaml.FoldedStyle:
		node.Style = FoldedStyle
	case yaml.SingleQuotedStyle:
		node.Style = QuotedStyle
	case yaml.DoubleQuotedStyle:
		node.Style = DoubleQuotedStyle
	case yaml.FlowStyle:
		node.Style = FlowStyle
	default:
		node.Style = DefaultStyle
	}

	// Process children
	if nodeKind == MappingNode {
		for i := 0; i < len(yamlNode.Content)-1; i += 2 {
			keyYamlNode := yamlNode.Content[i]
			key := ConvertFromYAMLNode(keyYamlNode)
			// For mapping keys, preserve the literal string value
			if key.Kind == ScalarNode && keyYamlNode.Value == "null" {
				key.Value = "null"
			}
			value := ConvertFromYAMLNode(yamlNode.Content[i+1])
			if err := node.AddKeyValue(key, value); err != nil {
				// Log but continue processing
				fmt.Printf("Warning: failed to add key-value: %v\n", err)
			}
		}
	} else if nodeKind == SequenceNode || nodeKind == DocumentNode {
		for _, child := range yamlNode.Content {
			childNode := ConvertFromYAMLNode(child)
			node.AddChild(childNode)
		}
	}

	return node
}
