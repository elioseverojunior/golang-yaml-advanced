# Examples

This directory contains practical examples demonstrating various features of the Golang YAML Advanced library.

## Available Examples

### 1. Basic Usage (`/basic`)
Simple examples showing fundamental operations:
- Parsing YAML files
- Accessing node values
- Preserving comments
- Serializing back to YAML

### 2. Final Implementation (`/final`)
Complete working example showing:
- Loading multiple YAML files
- Merging configurations
- Override handling
- Diff comparison
- Writing merged results

**Files:**
- `main.go` - Main implementation
- `yaml-values.yaml` - Base configuration
- `yaml-values-override.yaml` - Override values
- `yaml-values-merged.yaml` - Expected merged output

**Usage:**
```bash
cd examples/final
go run main.go
```

**Output:**
```
Step 1: Loading yaml-values.yaml...
  ✓ Loaded base values

Step 2: Loading yaml-values-override.yaml...
  ✓ Loaded override values

Step 3: Merging YAML files...
  ✓ Merged YAML trees

Step 4: Writing merged result to yaml-values-merged.yaml...
  ✓ Successfully wrote merged YAML

Step 5: Verifying output...
  ✓ Output matches expected result - no diff!
```

## Running Examples

### Prerequisites
```bash
# Install the library
go get github.com/elioetibr/golang-yaml-advanced
```

### Running Individual Examples
```bash
# Run basic example
cd examples/basic
go run main.go

# Run final example
cd examples/final
go run main.go
```

## Example Patterns

### 1. Comment Preservation
```go
// Comments are automatically preserved
tree, _ := golang_yaml_advanced.UnmarshalYAML(yamlData)
output, _ := tree.ToYAML()
// Output maintains all original comments
```

### 2. Large Integer Preservation
```yaml
# Input
account_id: 123456789012

# Preserved as integer, not scientific notation
# Other libraries might output: 1.23457e+11
```

### 3. Merging with Override
```go
base, _ := golang_yaml_advanced.UnmarshalYAML(baseYAML)
override, _ := golang_yaml_advanced.UnmarshalYAML(overrideYAML)
merged := golang_yaml_advanced.MergeTrees(base, override)
```

### 4. Finding Nodes
```go
node := root.Find(func(n *Node) bool {
    return n.Kind == ScalarNode && n.Value == "target"
})
```

### 5. Walking the Tree
```go
root.Walk(func(n *Node) bool {
    fmt.Printf("Node: %v\n", n.Value)
    return true // Continue traversal
})
```

### 6. Diff Comparison
```go
diffs := golang_yaml_advanced.DiffTrees(tree1, tree2)
for _, diff := range diffs {
    fmt.Printf("Change: %s at %s\n", diff.Type, diff.Path)
}
```

## Common Use Cases

### Configuration Management
Merge multiple configuration files with environment-specific overrides:
```go
baseConfig := loadConfig("config.yaml")
envConfig := loadConfig(fmt.Sprintf("config.%s.yaml", env))
finalConfig := golang_yaml_advanced.MergeTrees(baseConfig, envConfig)
```

### Helm Values Processing
Process Helm chart values with overrides:
```go
values := loadValues("values.yaml")
overrides := loadValues("values.override.yaml")
merged := golang_yaml_advanced.MergeTrees(values, overrides)
```

### Schema Validation
Validate YAML against expected structure:
```go
tree, _ := golang_yaml_advanced.UnmarshalYAML(data)
errors := validateSchema(tree.Documents[0].Root, schema)
```

### YAML Transformation
Transform YAML structure programmatically:
```go
tree.Documents[0].Root.Walk(func(n *Node) bool {
    if n.Kind == ScalarNode && n.Tag == "!!secret" {
        n.Value = "***REDACTED***"
    }
    return true
})
```

### Multi-Document Processing
Handle files with multiple YAML documents:
```go
tree, _ := golang_yaml_advanced.UnmarshalYAML(multiDocYAML)
for i, doc := range tree.Documents {
    fmt.Printf("Document %d has root kind: %v\n", i, doc.Root.Kind)
}
```

## Best Practices

1. **Always Handle Errors**: Check all error returns
2. **Clone Before Modifying**: Use `Clone()` to preserve originals
3. **Use Type Assertions Carefully**: Check node kinds before casting
4. **Preserve Formatting**: Maintain original style when possible
5. **Test with Real Files**: Validate with actual YAML from your use case

## Troubleshooting

### Issue: Comments Not Preserved
**Solution**: Ensure you're using `UnmarshalYAML()` instead of standard `Unmarshal()`

### Issue: Large Integers Become Scientific Notation
**Solution**: The library handles this automatically, ensure you're using the latest version

### Issue: Merge Not Working as Expected
**Solution**: Check that both trees have the same structure at the merge point

### Issue: Performance with Large Files
**Solution**: Consider using streaming parser for files > 10MB

## Contributing Examples

To add a new example:
1. Create a new directory under `/examples`
2. Include a `main.go` with your example
3. Add sample YAML files if needed
4. Update this README with description
5. Ensure the example is self-contained and runnable

## License

All examples are provided under the same MIT license as the main library.