# Deprecations

This document lists deprecated features, planned removals, and migration guides for the golang-yaml-advanced library.

## Deprecation Policy

- **Minor versions (1.x.0)**: May deprecate features but will not remove them
- **Major versions (2.0.0)**: May remove previously deprecated features
- **Deprecation period**: Features will be deprecated for at least 2 minor versions before removal

## Current Deprecations (v1.2.0)

### Node.EmptyLines Field
**Status**: Deprecated since v1.2.0
**Planned Removal**: v2.0.0
**Replacement**: Use `Node.EmptyLinesBefore` and `Node.EmptyLinesAfter`

```go
// Deprecated
node.EmptyLines = []int{1, 2}

// Recommended
node.EmptyLinesBefore = 1
node.EmptyLinesAfter = 0
```

**Reason**: The new fields provide clearer semantics and better support for merge operations.

### UnmarshalYAMLLegacy Function
**Status**: Deprecated since v1.2.0
**Planned Removal**: v2.0.0
**Replacement**: Use `UnmarshalYAML` or `UnmarshalYAMLWithEmptyLines`

```go
// Deprecated
tree, err := UnmarshalYAMLLegacy(data)

// Recommended
tree, err := UnmarshalYAML(data)
```

**Reason**: The legacy function lacks empty line tracking and improved comment handling.

## Planned Deprecations (v1.3.0)

### addEmptyLinesBeforeSchemaComments Function
**Status**: Will be deprecated in v1.3.0
**Planned Removal**: v2.0.0
**Replacement**: Use `addEmptyLinesBeforeCommentBlocks`

```go
// Will be deprecated
output := addEmptyLinesBeforeSchemaComments(input)

// Recommended
output := addEmptyLinesBeforeCommentBlocks(input)
```

**Reason**: The new function name better reflects its general purpose beyond schema comments.

## Future Considerations (v2.0.0)

### Breaking Changes Planned for v2.0.0

1. **Remove all deprecated fields and functions**
   - `Node.EmptyLines` field will be removed
   - `UnmarshalYAMLLegacy` function will be removed
   - `addEmptyLinesBeforeSchemaComments` function will be removed

2. **API Improvements**
   - Consider renaming `NodeTree` to `Tree` for simplicity
   - Consider making `EmptyLineConfig` a required parameter rather than optional

3. **Performance Optimizations**
   - Potential changes to internal data structures for better performance
   - May affect custom implementations that rely on internal details

## Migration Guides

### Migrating from EmptyLines to EmptyLinesBefore/After

Before (deprecated):
```go
node := &Node{
    EmptyLines: []int{2}, // Ambiguous meaning
}
```

After (recommended):
```go
node := &Node{
    EmptyLinesBefore: 2, // Clear: 2 empty lines before this node
    EmptyLinesAfter: 0,  // Clear: no empty lines after this node
}
```

### Migrating Empty Line Configuration

Before (implicit):
```go
tree, _ := UnmarshalYAML(data)
output, _ := tree.ToYAML() // Uses default behavior
```

After (explicit):
```go
tree, _ := UnmarshalYAML(data)
tree.EmptyLineConfig = DefaultEmptyLineConfig() // Explicit configuration
output, _ := tree.ToYAML()
```

## How to Prepare

1. **Update your code**: Replace deprecated features with recommended alternatives
2. **Test thoroughly**: Ensure your code works with the new APIs
3. **Pin versions**: If you can't migrate immediately, pin to a specific minor version
4. **Monitor releases**: Watch for deprecation notices in release notes

## Deprecation Notices

Deprecated features will generate warnings in future versions (planned for v1.3.0):

```go
// Example warning (future implementation)
// WARNING: Node.EmptyLines is deprecated and will be removed in v2.0.0
// Use Node.EmptyLinesBefore and Node.EmptyLinesAfter instead
```

## Questions or Concerns?

If you have questions about deprecations or need help migrating:

1. Check the [CHANGELOG](CHANGELOG.md) for detailed version notes
2. Review the [API Documentation](docs/API.md) for current best practices
3. Open an [issue](https://github.com/elioetibr/golang-yaml-advanced/issues) for migration help

---

*Last updated: 2025-09-28 for v1.2.0*